package server

import (
	"go-csv-import/internal/app"
	"go-csv-import/internal/server/routes"

	"github.com/gin-gonic/gin"
)

func Init() *gin.Engine {
	return gin.Default()
}

func LoadRoutes(s *gin.Engine, r routes.HttpRouter) {
	r.Load(s)
}

func Run(s *gin.Engine, addr ...string) {
	app.Logger().Info("API Server runs on http://localhost:8080")

	s.Run(addr...)
}
