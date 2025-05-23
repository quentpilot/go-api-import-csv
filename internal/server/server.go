package server

import (
	"go-csv-import/internal/app"

	"github.com/gin-gonic/gin"
)

func Init() *gin.Engine {
	return gin.Default()
}

func Run(s *gin.Engine, addr ...string) {
	app.Logger().Info("API Server runs on http://localhost:8080")

	s.Run(addr...)
}
