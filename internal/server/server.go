package server

import (
	"go-csv-import/internal/config"
	"go-csv-import/internal/logger"

	"github.com/gin-gonic/gin"
)

type Server struct {
	Engine *gin.Engine
	Config *config.HttpConfig
}

func New(c *config.HttpConfig) *Server {
	return &Server{
		Engine: gin.Default(),
		Config: c,
	}
}

func (s *Server) LoadRoutes(r HttpRouter) {
	r.Load(s.Engine)
}

func (s *Server) Run() {
	url := s.Config.Host + s.Config.Port
	logger.Info("API Server runs on " + url)

	s.Engine.Run(s.Config.Port)
}
