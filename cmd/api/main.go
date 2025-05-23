package main

import (
	"go-csv-import/internal/app"
	"go-csv-import/internal/bootstrap"
	"go-csv-import/internal/server"
	"go-csv-import/internal/server/routes"
)

func main() {
	bootstrap.Init(app.Config{
		LoggerName: "api",
	})

	s := server.Init()

	r := routes.Route{}
	r.Load(s)

	server.Run(s, ":8080")
}
