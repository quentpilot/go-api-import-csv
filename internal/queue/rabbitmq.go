package queue

import (
	"go-csv-import/internal/app"
	"go-csv-import/internal/config"

	"github.com/streadway/amqp"
)

type AmqpQueueHandler interface {
	getChannel() (*amqp.Connection, *amqp.Channel, error)
	Publish() error
	Consume() <-chan amqp.Delivery
}

type AmqpQueue struct {
	Config config.ApmqConfig
}

func NewAmqpQueue(c config.ApmqConfig) *AmqpQueue {
	return &AmqpQueue{Config: c}
}

func (q *AmqpQueue) getChannel(queue string) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(q.Config.Dsn)
	if err != nil {
		return nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, nil, err
	}

	_, err = ch.QueueDeclare(queue, true, false, false, false, nil)
	return conn, ch, err
}

func (q *AmqpQueue) Publish(queue string, body []byte) error {
	conn, ch, err := q.getChannel(queue)
	if err != nil {
		return err
	}
	defer conn.Close()
	defer ch.Close()

	return ch.Publish("", queue, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}

func (q *AmqpQueue) Consume(queue string, autoAck bool) <-chan amqp.Delivery {
	_, ch, err := q.getChannel(queue)
	if err != nil {
		app.Get().Log().Error("Connect to RabbitMQ", "error", err)
		panic(err)
	}

	msgs, err := ch.Consume(queue, "", autoAck, false, false, false, nil)
	if err != nil {
		app.Get().Log().Error("Error while consuming message", "error", err)
		panic(err)
	}

	return msgs
}

/* func ConsumeImportJobs() {
	msgs := Consume(app.AmqpConfig().Queue, false)

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
} */
