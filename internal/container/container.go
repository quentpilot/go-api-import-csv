package container

import (
	"go-csv-import/internal/config"
	"go-csv-import/internal/service/phonebook"
)

type Services struct {
	PhonebookUploader *phonebook.PhonebookHandler
}

// LoadServices initializes and returns the services for the application.
func LoadServices(a *config.AppConfig) *Services {
	return &Services{
		PhonebookUploader: phonebook.NewPhonebookConsumer(&a.Amqp, &a.Http, &a.Db),
	}
}
