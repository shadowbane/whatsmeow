// cmd/api/main.bak.go

package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"gomeow/cmd/api/router"
	"gomeow/pkg/application"
	"gomeow/pkg/exithandler"
	"gomeow/pkg/server"
	"gomeow/pkg/wmeow"
	"runtime"
	"time"
)

func main() {
	var cpuCount = runtime.NumCPU()
	if cpuCount > 1 {
		runtime.GOMAXPROCS(cpuCount)
	}

	// load .env
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Error loading .env file: %v\n", err)
	}

	// start application
	app, err := application.Start()
	if err != nil {
		zap.S().Fatal(err.Error())
	}

	// ToDo: Move somewhere else, after the device is connected.
	// The queue should be pushed according to user's ID
	// Load queue
	//go app.LoadQueue(app.Meow.DeviceStore.ID.String())

	srv := server.
		Get().
		WithAddr(app.Cfg.GetAPIPort()).
		WithRouter(router.Get(app)).
		WithErrLogger(zap.S())

	// start the server
	go func() {
		zap.S().Info("starting server at ", app.Cfg.GetAPIPort())

		if err := srv.Start(); err != nil {
			zap.S().Warn(err.Error())
		}
	}()

	wmeow.ConnectOnStartup(app)

	// ToDo: Move after device is connected
	// each device should have their own queue
	// queue runner
	// will run every second
	//go func() {
	//	zap.S().Info("starting queue runner")
	//	app.RunQueue()
	//}()

	exithandler.Init(func() {
		zap.S().Info("Closing Application")
		zap.S().Info("Waiting for all the processes to finish")

		// here, we need to wait for everything to be closed gracefully
		// we will use context timeout, and wait for all the goroutines to finish
		// if it's not closed after 5 seconds, we will force close it
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		wmeow.Shutdown()

		if err := srv.Close(); err != nil {
			zap.S().Error(err.Error())
		}

		select {
		case <-ctx.Done():
			zap.S().Debugf("Gracefully closed")
		}
		zap.S().Info("Application Closed")
	})
}
