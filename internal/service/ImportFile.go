package service

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"go-csv-import/internal/config"
	"go-csv-import/internal/job"
	"go-csv-import/internal/model"
	"go-csv-import/internal/repository"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type ImportFileService interface {
	Import(file job.TempFileJob) error
}

type ImportFile struct {
	HttpConfig *config.HttpConfig
	DbConfig   *config.DbConfig
	Repository repository.ContactRepository
}

func NewImportFile(h *config.HttpConfig, d *config.DbConfig) *ImportFile {
	return &ImportFile{
		HttpConfig: h,
		DbConfig:   d,
		Repository: repository.NewContactRepository(),
	}
}

func (i *ImportFile) Import(file *job.ImportJob) error {
	chunk, err := i.mustChunkFile(file)
	if err != nil {
		slog.Error("Error checking chunk file:", "error", err)
		return fmt.Errorf("error checking chunk file: %w", err)
	}

	i.Repository.Truncate()

	files := &job.JobStat{FilePath: file.FilePath, TotalRows: 0, ProcessTime: 0}
	if chunk {
		files, err := i.chunkFile(file)
		if err != nil {
			slog.Error("Error chunking file:", "file", err)
			return fmt.Errorf("error chunking file: %w", err)
		}

		return i.processSeveralFiles(files)
	}

	slog.Info("Processing single file:", "file", file.FilePath)
	return i.processSingleFile(files)
}

// Parse CSV file
func (i *ImportFile) processSingleFile(file *job.JobStat) error {
	slog.Info("Processing current file:", "file", file.FilePath)

	start := time.Now()

	f, err := os.Open(file.FilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.Comma = ';'

	// Skip header
	headers, err := reader.Read()
	if err != nil {
		return err
	}
	slog.Debug("CSV Headers:", "headers", headers)

	file.TotalRows = 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			slog.Error("failed to read row", "row", record, "error", err)
			return fmt.Errorf("failed to read row: %w", err)
		}

		// Print each line
		slog.Debug("Read current line", "row", record)
		contact, err := i.insert(headers, record)
		if err != nil {
			slog.Error("Failed to insert current contact", "error", err.Error())
		} else {
			slog.Debug("New contact inserted", "contact", fmt.Sprintf("%#v", contact))
		}

		file.TotalRows++
	}

	file.ProcessTime = time.Since(start)

	slog.Info(file.PrintStat())

	return nil
}

func (i *ImportFile) processSeveralFiles(files []job.JobStat) error {
	slog.Info("Processing several files", "files", files)

	// Define max CPU usage to avoir using all CPU cores
	maxCPU := runtime.NumCPU()
	if maxCPU > 4 {
		maxCPU -= 2
	}
	runtime.GOMAXPROCS(maxCPU)

	sem := make(chan struct{}, maxCPU)

	var wg sync.WaitGroup

	for _, file := range files {
		wg.Add(1)
		go func(f job.JobStat) {
			defer wg.Done()

			sem <- struct{}{}

			defer func() { sem <- struct{}{} }()

			if err := i.processSingleFile(&f); err != nil {
				slog.Error(fmt.Sprintf("Error processing file %s: %v", f.FilePath, err))
			}

		}(file)
	}
	wg.Wait()
	slog.Info("All files processed")

	for _, file := range files {
		if err := file.Remove(); err != nil {
			slog.Error("Error removing file", "file", file.FilePath, "error", err)
		} else {
			slog.Info("File has been removed successfully", "file", file.FilePath)
		}
	}

	return nil
}

/*
Check if the file has more than maximum rows configured.

This determines if the file should be chunked or not.
*/
func (i *ImportFile) mustChunkFile(file *job.ImportJob) (bool, error) {
	f, err := os.Open(file.FilePath)
	if err != nil {
		return false, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	count := 0
	reached := false
	for scanner.Scan() {
		count++
		if count > file.MaxRows {
			reached = true
			break
		}
	}

	return reached, nil
}

/*
Splits the file into smaller chunks.
Each chunk will have a maximum number of rows defined by file.MaxRows.
The chunk files will be created in the /tmp directory.
The chunk files will be named as <original_file_name>-part-<chunk_index>.csv.
The chunk files will contain the same header as the original file.
The chunk files will be returned as a slice of strings.
The original file will not be modified.
*/
func (i *ImportFile) chunkFile(file *job.ImportJob) ([]job.JobStat, error) {
	f, err := os.Open(file.FilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	// Lire le header une fois
	if !scanner.Scan() {
		return nil, fmt.Errorf("failed to read header for file %s: %w", file.FilePath, scanner.Err())
	}
	header := scanner.Text()

	var chunkFiles []job.JobStat
	var out *os.File
	var writer *bufio.Writer
	var currentLine int
	chunkIndex := 1

	createNewChunk := func() error {
		if writer != nil {
			writer.Flush()
			out.Close()
		}

		// Create a new chunk file
		filename := fmt.Sprintf("%v-part-%d.csv", filepath.Base(file.FilePath), chunkIndex)
		filename = filepath.Join("/tmp", filename)
		chunkIndex++
		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		out = file
		writer = bufio.NewWriter(out)

		// Write the header to the new chunk file
		if _, err := writer.WriteString(header + "\n"); err != nil {
			return err
		}

		jobStat := job.JobStat{FilePath: filename, TotalRows: 0, ProcessTime: 0}
		chunkFiles = append(chunkFiles, jobStat)
		currentLine = 0
		return nil
	}

	if err := createNewChunk(); err != nil {
		return nil, err
	}

	for scanner.Scan() {
		line := scanner.Text()

		if currentLine >= file.MaxRows {
			if err := createNewChunk(); err != nil {
				return nil, err
			}
		}

		if _, err := writer.WriteString(line + "\n"); err != nil {
			return nil, err
		}
		currentLine++
	}

	if writer != nil {
		writer.Flush()
		out.Close()
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return chunkFiles, nil
}

// Combines header as key with row as value
func (i *ImportFile) combine(header []string, row []string) (map[string]string, error) {
	if len(header) != len(row) {
		return nil, errors.New("header and row slices mismatch")
	}

	r := make(map[string]string, len(header))
	for i, k := range header {
		r[k] = row[i]
	}

	return r, nil
}

func (i *ImportFile) createContactFromRow(header []string, row []string) (*model.Contact, error) {
	r, err := i.combine(header, row)
	if err != nil {
		return &model.Contact{}, err
	}
	slog.Debug("Combine row result", "combine", fmt.Sprintf("%#v", r))

	required := []string{"Phone", "Firstname", "Lastname"}
	for i := 0; i < len(required); i++ {
		key := required[i]
		if _, exists := r[key]; exists {
			continue
		} else {
			return &model.Contact{}, fmt.Errorf("columns %s is missing", key)
		}
	}

	return &model.Contact{
		Phone:     r["Phone"],
		Firstname: r["Firstname"],
		Lastname:  r["Lastname"],
	}, nil
}

func (i *ImportFile) insert(header []string, row []string) (*model.Contact, error) {
	c, err := i.createContactFromRow(header, row)
	if err != nil {
		return &model.Contact{}, err
	}
	slog.Debug("Model created", "contact", fmt.Sprintf("%#v", c))

	err = i.Repository.Insert(c)

	return c, err
}
