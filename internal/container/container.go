package container

import (
	"go-csv-import/internal/config"
	"go-csv-import/internal/service"
)

type Services struct {
	ImportFileQueue *service.ImportFileQueue
}

func LoadServices(a *config.AppConfig) *Services {
	return &Services{
		ImportFileQueue: service.NewImportFileQueueConsumer(&a.Amqp, &a.Http, &a.Db),
	}
}
