package server

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

type Server struct {
	Engine *gin.Engine
}

func New() *Server {
	return &Server{
		Engine: gin.Default(),
	}
}

func (s *Server) LoadRoutes(r HttpRouter) {
	r.Load(s.Engine)
}

func (s *Server) Run(addr ...string) {
	url := "http://localhost" + addr[0]
	slog.Info("API Server runs on " + url)

	s.Engine.Run(addr...)
}
