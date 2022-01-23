package exithandler

import (
	"log"
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
		log.Println("exit reason: ", sig)
		terminate <- true
	}()

	<-terminate
	cb()
	log.Print("exiting program")
}
