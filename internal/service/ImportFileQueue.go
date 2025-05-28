package service

import (
	"context"
	"encoding/json"
	"errors"
	"go-csv-import/internal/config"
	"go-csv-import/internal/db"
	"go-csv-import/internal/job"
	"go-csv-import/internal/queue"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/hashicorp/go-multierror"
)

type AmqpQueueService interface {
	Publish(job job.TempFileJob) error
	Consume()
}

type ImportFileQueue struct {
	AmqpConfig *config.ApmqConfig
	HttpConfig *config.HttpConfig
	Queue      *queue.AmqpQueue
	Imporer    *ImportFile
}

func NewImportFileQueuePublisher(a *config.ApmqConfig, h *config.HttpConfig) *ImportFileQueue {
	self := &ImportFileQueue{
		AmqpConfig: a,
		HttpConfig: h,
	}

	self.Queue = queue.NewAmqpQueue(a)

	return self
}

func NewImportFileQueueConsumer(a *config.ApmqConfig, h *config.HttpConfig, d *config.DbConfig) *ImportFileQueue {
	self := NewImportFileQueuePublisher(a, h)

	self.Imporer = NewImportFile(h, d)

	return self
}

func (q *ImportFileQueue) Publish(job *job.ImportJob) error {
	body, _ := json.Marshal(job)

	return q.Queue.Publish(q.AmqpConfig.Queue, body)
}

func (q *ImportFileQueue) Consume() {
	//msgs := q.Queue.Consume(q.AmqpConfig.Queue, true)
	ctxInt, stopInt := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stopInt()

	for msg := range q.Queue.Consume(q.AmqpConfig.Queue, true) {
		var job job.ImportJob
		if err := json.Unmarshal(msg.Body, &job); err != nil {
			slog.Error("Invalid Job format:", "body", msg.Body, "error", err)
			continue
		}

		start := time.Now()
		slog.Info("Try to treat file:", "file", job.FilePath)

		if err := q.Imporer.Import(ctxInt, &job); err != nil {
			q.printTypedErrors(err, job)
		} else {
			slog.Info("File has been successful treated", "file", job.FilePath, "time", time.Since(start))
		}
		job.Remove()
		//msg.Ack(false)
		slog.Info("Message acknowledged")
	}
}

func (q *ImportFileQueue) printTypedErrors(err error, job job.ImportJob) {
	if errs, ok := err.(*multierror.Error); ok {
		for _, e := range errs.Errors {
			var fe *FileError
			if errors.As(e, &fe) {
				slog.Error("Error processing file", "file", fe.FilePath, "error", fe)
				continue
			}

			var de *db.DbError
			if errors.As(e, &de) {
				slog.Error("Database error", "error", de)
				continue
			}

			if errors.Is(e, context.Canceled) {
				slog.Error("Interrupted by SIGINT", "error", e)
				continue
			}

			if errors.Is(e, context.DeadlineExceeded) {
				slog.Error("Timeout", "error", e)
				continue
			}

			slog.Error("Unexpected error", "file", job.FilePath, "error", e)
		}
	} else {

		if ie, ok := err.(*FileError); ok {
			slog.Error("Error processing single file", "file", job.FilePath, "error", ie.Err)
		} else if de, ok := err.(*db.DbError); ok {
			slog.Error("Database error for single file", "error", de.Err)
		} else if errors.Is(err, context.Canceled) {
			slog.Error("Interrupted by SIGINT", "error", err)
		} else if errors.Is(err, context.DeadlineExceeded) {
			slog.Error("Timeout", "error", err)
		} else {
			slog.Error("Unexpected error for single file", "file", job.FilePath, "error", err)
		}
	}
}
