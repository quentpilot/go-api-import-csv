package routes

import (
	"go-csv-import/internal/app"
	"go-csv-import/internal/handlers"
	"go-csv-import/internal/middleware"
	"go-csv-import/internal/queue"

	"github.com/gin-gonic/gin"
)

type HttpRouter interface {
	Load(s *gin.Engine)
}

type UploadRouter struct {
}

func (r UploadRouter) Load(s *gin.Engine) {
	publisher := &queue.RabbitPublisher{}

	s.POST("/upload", middleware.LimitRequestSize(app.HttpConfig().MaxContentLength), handlers.Upload(publisher))
}
