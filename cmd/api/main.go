package main

import (
	"go-csv-import/internal/app"
	"go-csv-import/internal/bootstrap"
	"go-csv-import/internal/handlers"
	"go-csv-import/internal/logger"
	"go-csv-import/internal/middleware"
	"go-csv-import/internal/queue"

	"github.com/gin-gonic/gin"
)

func main() {
	bootstrap.Init(app.Config{
		LoggerName: "api",
	})

	r := gin.Default()

	publisher := &queue.RabbitPublisher{}
	r.POST("/upload", middleware.LimitRequestSize(10<<20), handlers.Upload(publisher))

	logger.Current.Info("API Server runs on localhost:8080")
	r.Run(":8080")
}
