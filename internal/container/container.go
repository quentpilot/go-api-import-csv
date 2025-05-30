package container

import (
	"go-csv-import/internal/config"
	"go-csv-import/internal/logger"
	"go-csv-import/internal/service/phonebook"
)

type Services struct {
	PhonebookUploader *phonebook.PhonebookHandler
}

// LoadServices initializes and returns the services for the application.
func LoadConsumerServices(a *config.AppConfig) *Services {
	s := &Services{
		PhonebookUploader: phonebook.NewPhonebookConsumer(&a.Amqp, &a.Http, &a.Db),
	}

	logger.Trace("Consumer Services Loaded")

	return s
}

func LoadApiServices(a *config.AppConfig) *Services {
	s := &Services{
		PhonebookUploader: phonebook.NewPhonebookPublisher(&a.Amqp, &a.Http),
	}

	logger.Trace("API Services Loaded")

	return s
}
