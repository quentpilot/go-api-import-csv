package service

import (
	"encoding/json"
	"go-csv-import/internal/config"
	"go-csv-import/internal/job"
	"go-csv-import/internal/queue"
	"log/slog"
	"time"
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

	for msg := range q.Queue.Consume(q.AmqpConfig.Queue, true) {
		var job job.ImportJob
		if err := json.Unmarshal(msg.Body, &job); err != nil {
			slog.Error("Invalid Job format:", "body", msg.Body, "error", err)
			continue
		}

		start := time.Now()
		slog.Info("Try to treat file:", "file", job.FilePath)

		if err := q.Imporer.Import(&job); err != nil {
			slog.Error("Error Treatment:", "error", err)
		} else {
			slog.Info("File has been successful treated", "file", job.FilePath, "time", time.Since(start))

			err = job.Remove()
			if err != nil {
				slog.Error("Cannot properly remove file '", "file", job.FilePath, "error", err)
			} else {
				slog.Info("File has been successful deleted:", "file", job.FilePath)
			}
		}

		//msg.Ack(false)
		slog.Info("Message acknowledged")
	}
}
