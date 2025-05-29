package phonebook

import (
	"context"
	"encoding/csv"
	"fmt"
	"go-csv-import/internal/config"
	"go-csv-import/internal/db"
	"go-csv-import/internal/repository"
	"io"
	"log/slog"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
)

type UploaderService interface {
	Upload(ctx context.Context, file FileMessage) error
}

type ContactUploader struct {
	HttpConfig *config.HttpConfig
	DbConfig   *config.DbConfig
	Repository repository.ContactRepository
}

func NewContactUploader(h *config.HttpConfig, d *config.DbConfig) *ContactUploader {
	return &ContactUploader{
		HttpConfig: h,
		DbConfig:   d,
		Repository: repository.NewContactRepository(),
	}
}

func (c *ContactUploader) reset() {
	c.Repository.Truncate()
}

func (c *ContactUploader) Upload(ctx context.Context, file *FileMessage) error {
	c.reset()
	chunk, err := c.mustChunkFile(file)
	if err != nil {
		return NewFileError(file.FilePath, fmt.Errorf("error checking chunk file: %w", err))
	}

	files := []FilePart{{FilePath: file.FilePath, TotalRows: 0, ProcessTime: 0}}
	if chunk {
		files, err = c.chunkFile(file)
		if err != nil {
			return &FileError{FilePath: file.FilePath, Err: fmt.Errorf("error chunking file: %w", err)}
		}
	}

	return c.handleFiles(ctx, files)
}

func (c *ContactUploader) handleFiles(ctx context.Context, files []FilePart) error {
	slog.Debug("Processing several files", "files", files)

	// Define max CPU usage to avoir using all CPU cores
	maxCPU := runtime.NumCPU()
	if maxCPU > 4 {
		maxCPU -= 2
	}
	runtime.GOMAXPROCS(maxCPU)

	jobs := make(chan FilePart)
	errs := make(chan error, len(files))
	var aErrs *multierror.Error

	var wg sync.WaitGroup

	// Consume files
	for w := 0; w < maxCPU; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range jobs {
				if err := c.uploadFile(ctx, &file); err != nil {
					errs <- fmt.Errorf("file %s: %w", file.FilePath, err)
					continue
				}
				slog.Info(file.PrintStat())
			}

		}()
	}

	// Send files to jobs channel
	go func() {
		for _, file := range files {
			jobs <- file
		}
		close(jobs)
	}()

	// Wait for workers and close errors channel
	go func() {
		wg.Wait()
		close(errs)
	}()

	// Collect and join errors
	for err := range errs {
		aErrs = multierror.Append(aErrs, err)
	}

	return aErrs.ErrorOrNil()
}

func (c *ContactUploader) uploadFile(ctx context.Context, file *FilePart) error {
	ctxT, cancel := context.WithTimeout(ctx, c.HttpConfig.FileTimeout)
	defer cancel()

	slog.Info("Processing current file:", "file", file.FilePath)
	defer file.Remove()

	start := time.Now()

	f, err := os.Open(file.FilePath)
	if err != nil {
		return NewFileError(file.FilePath, err)
	}
	defer f.Close()
	defer file.Remove()

	reader := csv.NewReader(f)
	reader.Comma = ';'

	// Skip header
	headers, err := reader.Read()
	if err != nil {
		return err
	}
	slog.Debug("CSV Headers:", "headers", headers)

	batch := NewBatch()

	file.TotalRows = 0
	for {
		select {
		case <-ctxT.Done():
			return fmt.Errorf("%w on FilePart", ctxT.Err())
		default:
		}

		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return NewFileError(file.FilePath, fmt.Errorf("failed to read row: %w", err))
		}

		//return &FileError{FilePath: file.FilePath, Err: fmt.Errorf("simulate error file")}

		// Batch insert contacts
		br := c.handleBatchInsert(batch, headers, record, false)
		if br != nil {
			return db.NewDbError(br)
		}

		file.TotalRows++
		file.ProcessTime = time.Since(start)
	}

	// Batch insert contacts
	br := c.handleBatchInsert(batch, headers, []string{}, true)
	if br != nil {
		return db.NewDbError(fmt.Errorf("error while forcing insert batch contacts: %w", br))
	}

	return nil
}
