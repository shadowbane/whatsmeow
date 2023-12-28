// cmd/api/main.bak.go

package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"gomeow/cmd/api/router"
	"gomeow/pkg/application"
	"gomeow/pkg/exithandler"
	"gomeow/pkg/grpc"
	"gomeow/pkg/server"
	"gomeow/pkg/wmeow"
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
		fmt.Println("Please ensure you load correct environment variables")
	}

	// start application
	app, err := application.Start()
	if err != nil {
		zap.S().Fatal(err.Error())
	}

	srv := server.
		Get().
		WithAddr(app.Cfg.GetAPIPort()).
		WithRouter(router.Api(app)).
		WithErrLogger(zap.S())

	// Init gRPC server
	grpcServer := grpc.
		Get().
		WithAddr(app.Cfg.GetGrpcPort())

	// start the api server
	go func() {
		zap.S().Info("starting api server at ", app.Cfg.GetAPIPort())

		if err := srv.Start(); err != nil {
			zap.S().Warn(err.Error())
		}
	}()

	// start the grpc server
	go func() {
		zap.S().Info("starting grpc server at ", app.Cfg.GetGrpcPort())

		if err := grpcServer.Start(app); err != nil {
			zap.S().Warn(err.Error())
		}
	}()

	// This is the main entry point of the application
	// if the device is logged in, it will be marked as connected
	// and we will auto connect it to whatsapp
	wmeow.ConnectOnStartup(app)

	exithandler.Init(func() {
		zap.S().Info("Closing Application")
		zap.S().Info("Waiting for all the processes to finish")

		// here, we need to wait for everything to be closed gracefully
		// we will use context timeout, and wait for all the goroutines to finish
		// if it's not closed after 5 seconds, we will force close it
		//ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		wmeow.Shutdown()

		if err := srv.Close(); err != nil {
			zap.S().Error(err.Error())
		}
		if err := grpcServer.Close(); err != nil {
			zap.S().Error(err.Error())
		}

		//select {
		//case <-ctx.Done():
		//	zap.S().Debugf("Gracefully closed")
		//}
		zap.S().Info("Application Closed")
	})
}
