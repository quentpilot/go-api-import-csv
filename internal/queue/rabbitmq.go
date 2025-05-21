package queue

import (
	"encoding/json"
	"go-csv-import/internal/importer"
	"go-csv-import/internal/job"
	"go-csv-import/internal/logger"
	"log"

	"github.com/streadway/amqp"
)

func getChannel() (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
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

	return ch.Publish("", "import_queue", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}

func ConsumeImportJobs() {
	if err := logger.InitCurrent("worker", false); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	_, ch, err := getChannel()
	if err != nil {
		log.Fatal("Connect to RabbitMQ:", err)
	}

	msgs, err := ch.Consume("import_queue", "", false, false, false, false, nil)
	if err != nil {
		log.Fatal("Error while consuming message:", err)
	}

	for msg := range msgs {
		var job job.ImportJob
		if err := json.Unmarshal(msg.Body, &job); err != nil {
			logger.Current.Info("Invalid Job format:", "body", msg.Body, "error", err)
			continue
		}

		logger.Current.Info("Try to treat file:", "file", job.FilePath)
		if err := importer.ProcessFile(job); err != nil {
			logger.Current.Info("Error Treatment:", "error", err)
		} else {
			logger.Current.Info("File has been successful treated:", "file", job.FilePath)

			err = job.Remove()
			if err != nil {
				logger.Current.Info("Cannot properly remove file '", "file", job.FilePath, "error", err)
			} else {
				logger.Current.Info("File has been successful deleted:", "file", job.FilePath)
			}
		}

		msg.Ack(false)
		logger.Current.Info("Message acknowledged")
	}
}
