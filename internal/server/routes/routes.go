package routes

import (
	"go-csv-import/internal/handlers"
	"go-csv-import/internal/middleware"
	"go-csv-import/internal/queue"

	"github.com/gin-gonic/gin"
)

type HttpRouter interface {
	Load(s *gin.Engine)
}

type Route struct {
}

func (r Route) Load(s *gin.Engine) {
	publisher := &queue.RabbitPublisher{}

	s.POST("/upload", middleware.LimitRequestSize(10<<20), handlers.Upload(publisher))
}
