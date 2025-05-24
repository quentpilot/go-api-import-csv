package queue

import (
	"encoding/json"
	"go-csv-import/internal/app"
	"go-csv-import/internal/importer"
	"go-csv-import/internal/job"
	"time"

	"github.com/streadway/amqp"
)

type RabbitPublisher struct{}

func (r *RabbitPublisher) PublishImportJob(path string, maxRows int) error {
	return PublishImportJob(path, maxRows)
}

func getChannel() (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(app.AmqpConfig().Dsn)
	if err != nil {
		return nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, nil, err
	}

	_, err = ch.QueueDeclare("import_queue", true, false, false, false, nil)
	return conn, ch, err
}

func PublishImportJob(filepath string, maxRows int) error {
	conn, ch, err := getChannel()
	if err != nil {
		return err
	}
	defer conn.Close()
	defer ch.Close()

	job := job.ImportJob{FilePath: filepath, MaxRows: maxRows}
	body, _ := json.Marshal(job)

	return ch.Publish("", app.AmqpConfig().Queue, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}

func ConsumeImportJobs() {
	_, ch, err := getChannel()
	if err != nil {
		app.Log().Error("Connect to RabbitMQ", "error", err)
		panic(err)
	}

	msgs, err := ch.Consume(app.AmqpConfig().Queue, "", false, false, false, false, nil)
	if err != nil {
		app.Log().Error("Error while consuming message", "error", err)
		panic(err)
	}

	for msg := range msgs {
		var job job.ImportJob
		if err := json.Unmarshal(msg.Body, &job); err != nil {
			app.Log().Error("Invalid Job format:", "body", msg.Body, "error", err)
			continue
		}

		start := time.Now()
		app.Log().Info("Try to treat file:", "file", job.FilePath)
		if err := importer.ProcessFile(job); err != nil {
			app.Log().Error("Error Treatment:", "error", err)
		} else {
			app.Log().Info("File has been successful treated", "file", job.FilePath, "time", time.Since(start))

			err = job.Remove()
			if err != nil {
				app.Log().Error("Cannot properly remove file '", "file", job.FilePath, "error", err)
			} else {
				app.Log().Info("File has been successful deleted:", "file", job.FilePath)
			}
		}

		msg.Ack(false)
		app.Log().Info("Message acknowledged")
	}
}
