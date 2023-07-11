package exithandler

import (
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

func Init(cb func()) {
	sigs := make(chan os.Signal, 1)
	terminate := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		zap.S().Info("exit reason: ", sig)
		terminate <- true
	}()

	<-terminate
	cb()
	zap.S().Info("Application Closed")
}
