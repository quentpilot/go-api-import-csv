package server

import (
	"go-csv-import/internal/config"
	"go-csv-import/internal/container"
	"go-csv-import/internal/handlers"
	"go-csv-import/internal/logger"
	"go-csv-import/internal/middleware"
	"time"

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
	s.POST("/upload", middleware.Timeout(5*time.Second), middleware.LimitRequestSize(r.HttpConfig.MaxContentLength), handlers.Upload(r.Services.PhonebookUploader))
	s.GET("/upload/status/:uuid", handlers.UploadStatus(r.Services.PhonebookUploader))

	s.LoadHTMLGlob("templates/*")
	s.GET("/upload-form", handlers.HtmlUpload())
	logger.Debug("Upload route registered")
}
