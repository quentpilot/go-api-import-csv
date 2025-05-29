package server

import (
	"go-csv-import/internal/config"
	"go-csv-import/internal/handlers"
	"go-csv-import/internal/middleware"
	"go-csv-import/internal/service/phonebook"

	"github.com/gin-gonic/gin"
)

type HttpRouter interface {
	Load(s *gin.Engine)
}

type UploadRouter struct {
	HttpConfig *config.HttpConfig
	AmqpConfig *config.ApmqConfig
}

func (r UploadRouter) Load(s *gin.Engine) {
	publisher := phonebook.NewPhonebookPublisher(r.AmqpConfig, r.HttpConfig)

	s.POST("/upload", middleware.LimitRequestSize(r.HttpConfig.MaxContentLength), handlers.Upload(publisher))
}
