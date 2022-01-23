// cmd/api/main.bak.go

package main

import (
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"gomeow/cmd/api/router"
	"gomeow/pkg/application"
	"gomeow/pkg/exithandler"
	"gomeow/pkg/logger"
	"gomeow/pkg/server"
)

func main() {
	logger.Init()
	zap.S().Info("Starting Application")

	if err := godotenv.Load(); err != nil {
		zap.S().Warnf("Failed to load env vars!")
	}

	app, err := application.Start()
	if err != nil {
		zap.S().Fatal(err.Error())
	}

	srv := server.
		Get().
		WithAddr(app.Cfg.GetAPIPort()).
		WithRouter(router.Get(app)).
		WithErrLogger(zap.S())

	go func() {
		zap.S().Info("starting server at ", app.Cfg.GetAPIPort())

		if err := srv.Start(); err != nil {
			zap.S().Fatal(err.Error())
		}
	}()

	exithandler.Init(func() {
		if err := srv.Close(); err != nil {
			zap.S().Error(err.Error())
		}
		zap.S().Info("Exiting Application")
		app.Meow.Exit()
	})
}
