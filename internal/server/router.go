package server

import (
	"go-csv-import/internal/config"
	"go-csv-import/internal/container"
	"go-csv-import/internal/handlers"
	"go-csv-import/internal/logger"
	"go-csv-import/internal/middleware"

	"github.com/gin-gonic/gin"
)

type HttpRouter interface {
	Load(s *gin.Engine)
}

type UploadRouter struct {
	HttpConfig *config.HttpConfig
	AmqpConfig *config.ApmqConfig
	Services   *container.Services
}

func (r UploadRouter) Load(s *gin.Engine) {
	s.GET("/ping", handlers.HealthCheck)
	s.POST("/upload", middleware.LimitRequestSize(r.HttpConfig.MaxContentLength), handlers.Upload(r.Services.PhonebookUploader))
	s.GET("/upload/status/:uuid", handlers.UploadStatus(r.Services.PhonebookUploader))
	logger.Debug("Upload route registered")
}
