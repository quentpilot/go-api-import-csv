package main

import (
	"go-csv-import/internal/app"
	"go-csv-import/internal/bootstrap"
	"go-csv-import/internal/server"
	"go-csv-import/internal/server/routes"
)

func main() {
	bootstrap.Init(app.AppConfig{
		LoggerName: "api",
	})

	s := server.Init()

	server.LoadRoutes(s, routes.UploadRouter{})

	server.Run(s, app.HttpConfig().Port)
}
