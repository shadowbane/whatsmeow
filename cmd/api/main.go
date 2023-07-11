// cmd/api/main.bak.go

package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"gomeow/cmd/api/router"
	"gomeow/pkg/application"
	"gomeow/pkg/exithandler"
	"gomeow/pkg/server"
	"runtime"
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

	// ToDo: Move after device is connected
	// each device should have their own queue
	// queue runner
	// will run every second
	//go func() {
	//	zap.S().Info("starting queue runner")
	//	app.RunQueue()
	//}()

	exithandler.Init(func() {
		if err := srv.Close(); err != nil {
			zap.S().Error(err.Error())
		}
		zap.S().Info("Closing Application")

		// ToDo: Exit all device connected
		//app.Meow.Exit()
	})
}
