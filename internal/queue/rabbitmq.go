package queue

import (
	"encoding/json"
	"go-csv-import/internal/importer"
	"log"
	"os"

	"github.com/streadway/amqp"
)

type ImportJob struct {
	Filepath string `json:"filepath"`
}

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

func PublishImportJob(filepath string) error {
	conn, ch, err := getChannel()
	if err != nil {
		return err
	}
	defer conn.Close()
	defer ch.Close()

	job := ImportJob{Filepath: filepath}
	body, _ := json.Marshal(job)

	return ch.Publish("", "import_queue", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}

func ConsumeImportJobs() {
	_, ch, err := getChannel()
	if err != nil {
		log.Fatal("Connect to RabbitMQ:", err)
	}

	msgs, err := ch.Consume("import_queue", "", false, false, false, false, nil)
	if err != nil {
		log.Fatal("Error while consuming message:", err)
	}

	for msg := range msgs {
		var job ImportJob
		if err := json.Unmarshal(msg.Body, &job); err != nil {
			log.Println("Invalid Job format:", msg.Body, err)
			continue
		}

		log.Println("Try to treat file:", job.Filepath)
		if err := importer.ProcessFile(job.Filepath); err != nil {
			log.Println("Error Treatment:", err)
		} else {
			log.Println("File has been successful treated:", job.Filepath)

			err = os.Remove(job.Filepath)
			if err != nil {
				log.Println("Cannot properly remove file '", job.Filepath, "'. Error:", err)
			} else {
				log.Println("File has been successful deleted:", job.Filepath)
			}
		}
		msg.Ack(false)
		log.Println("Message acknowledged")
	}
}
