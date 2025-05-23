package importer

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"go-csv-import/internal/app"
	"go-csv-import/internal/job"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

func ProcessFile(file job.ImportJob) error {
	chunk, err := mustChunkFile(file)
	if err != nil {
		app.Logger().Error("Error checking chunk file:", "error", err)
		return fmt.Errorf("error checking chunk file: %w", err)
	}

	files := job.JobStat{FilePath: file.FilePath, TotalRows: 0, ProcessTime: 0}
	if chunk {
		files, err := chunkFile(file)
		if err != nil {
			app.Logger().Error("Error chunking file:", "file", err)
			return fmt.Errorf("error chunking file: %w", err)
		}

		return processSeveralFiles(files)
	}

	app.Logger().Info("Processing single file:", "file", file.FilePath)
	return processSingleFile(files)
}

func processSingleFile(file job.JobStat) error {
	app.Logger().Info("Processing current file:", "file", file.FilePath)

	start := time.Now()

	f, err := os.Open(file.FilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := csv.NewReader(f)

	// Skip header
	headers, err := reader.Read()
	if err != nil {
		return err
	}
	app.Logger().Debug("CSV Headers:", "headers", headers)

	file.TotalRows = 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			app.Logger().Error("failed to read row", "row", record, "error", err)
			return fmt.Errorf("failed to read row: %w", err)
		}

		// Print each line
		app.Logger().Debug("Read current line", "row", record)
		file.TotalRows++
	}

	file.ProcessTime = time.Since(start)

	app.Logger().Info(file.PrintStat())

	return nil
}

func processSeveralFiles(files []job.JobStat) error {
	app.Logger().Info("Processing several files", "files", files)

	// Define max CPU usage to avoir using all CPU cores
	maxCPU := runtime.NumCPU()
	if maxCPU > 4 {
		maxCPU -= 2
	}
	runtime.GOMAXPROCS(maxCPU)

	var wg sync.WaitGroup

	for _, file := range files {
		wg.Add(1)
		go func(f job.JobStat) {
			defer wg.Done()

			if err := processSingleFile(f); err != nil {
				log.Printf("Error processing file %s: %v", f.FilePath, err)
			}

		}(file)
	}
	wg.Wait()
	app.Logger().Info("All files processed")

	for _, file := range files {
		if err := file.Remove(); err != nil {
			app.Logger().Error("Error removing file", "file", file.FilePath, "error", err)
		} else {
			app.Logger().Info("File has been removed successfully", "file", file.FilePath)
		}
	}

	return nil
}

// Reads first line of the CSV file and returns the headers
func getCsvHeaders(path string) (headers []string, error error) {
	f, err := os.Open(path)
	if err != nil {
		return headers, err
	}
	defer f.Close()

	reader := csv.NewReader(f)

	// Skip header
	headers, err = reader.Read()
	if err != nil {
		return headers, err
	}

	return headers, nil
}

/*
Check if the file has more than maximum rows configured.

This determines if the file should be chunked or not.
*/
func mustChunkFile(file job.ImportJob) (bool, error) {
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
func chunkFile(file job.ImportJob) ([]job.JobStat, error) {
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
