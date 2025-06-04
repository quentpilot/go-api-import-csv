package phonebook

import (
	"context"
	"encoding/csv"
	"fmt"
	"go-csv-import/internal/config"
	"go-csv-import/internal/db"
	"go-csv-import/internal/handlers/worker"
	"go-csv-import/internal/logger"
	"go-csv-import/internal/repository"
	"go-csv-import/internal/utils"
	"io"
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
	HttpConfig    *config.HttpConfig
	DbConfig      *config.DbConfig
	Repository    *repository.ContactRepository
	ProgressStore *worker.MessageProgressStore
}

func NewContactUploader(h *config.HttpConfig, d *config.DbConfig, p *worker.MessageProgressStore) *ContactUploader {
	return &ContactUploader{
		HttpConfig:    h,
		DbConfig:      d,
		Repository:    repository.NewContactRepository(),
		ProgressStore: p,
	}
}

func (c *ContactUploader) reset() {
	c.Repository.Truncate()
}

func (c *ContactUploader) Upload(ctx context.Context, file *FileMessage) error {
	c.reset()
	totalRows, err := utils.FileCountRowsCsv(file.FilePath)
	if err != nil {
		return NewFileError(file.FilePath, fmt.Errorf("error counting file rows: %w", err))
	}
	c.ProgressStore.Init(file.Uuid, int64(totalRows))

	chunk, err := c.mustChunkFile(file)
	if err != nil {
		return NewFileError(file.FilePath, fmt.Errorf("error checking chunk file: %w", err))
	}

	files := []FilePart{{FilePath: file.FilePath, Uuid: file.Uuid, TotalRows: 0, ProcessTime: 0}}
	if chunk {
		files, err = c.chunkFile(file)
		if err != nil {
			return NewFileError(file.FilePath, fmt.Errorf("error chunking file: %w", err))
		}
	}

	return c.handleFiles(ctx, files)
}

func (c *ContactUploader) handleFiles(ctx context.Context, files []FilePart) error {
	logger.Debug("Processing chunked files")
	logger.Trace("Files to process", "files", fmt.Sprintf("%#v", files))

	// Define max CPU usage to avoir using all CPU cores
	maxCPU := runtime.NumCPU()
	if maxCPU > 4 {
		maxCPU -= 2
	}
	logger.Debug("Setting max CPU usage", "maxCPU", maxCPU)
	runtime.GOMAXPROCS(maxCPU)

	jobs := make(chan FilePart)
	errs := make(chan error, len(files))
	var aErrs *multierror.Error

	var wg sync.WaitGroup

	logger.Trace("Launching workers", "workers", maxCPU)
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
				logger.Debug(file.PrintStat())
			}

		}()
	}

	logger.Trace("Dispatching files to workers channels")
	// Send files to jobs channel
	go func() {
		for _, file := range files {
			jobs <- file
		}
		close(jobs)
	}()

	logger.Trace("Waiting for workers to finish")
	// Wait for workers and close errors channel
	go func() {
		wg.Wait()
		close(errs)
	}()

	logger.Trace("Collecting errors from workers")
	// Collect and join errors
	for err := range errs {
		aErrs = multierror.Append(aErrs, err)
	}

	return aErrs.ErrorOrNil()
}

func (c *ContactUploader) uploadFile(ctx context.Context, file *FilePart) error {
	logger.Debug("Start processing routine file", "file", file.FilePath, "uuid", file.Uuid)

	ctxT, cancel := context.WithTimeout(ctx, c.HttpConfig.FileTimeout)
	defer cancel()
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
	logger.Trace("CSV Headers", "headers", headers)

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
		logger.Trace("RAW line", "line", fmt.Sprintf("%#v", record))

		//return &FileError{FilePath: file.FilePath, Err: fmt.Errorf("simulate error file")}

		// Batch insert contacts
		br := c.handleBatchInsert(ctxT, file, batch, headers, record, false)
		if br != nil {
			return db.NewDbError(br)
		}

		file.TotalRows++
		file.ProcessTime = time.Since(start)
	}

	// Batch insert contacts
	br := c.handleBatchInsert(ctxT, file, batch, headers, []string{}, true)
	if br != nil {
		return db.NewDbError(fmt.Errorf("error while forcing insert batch contacts: %w", br))
	}

	logger.Debug("End processing routine file", "file", file.FilePath, "uuid", file.Uuid)
	return nil
}
