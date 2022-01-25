package application

import (
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.uber.org/zap"
	"gomeow/pkg/config"
	"gomeow/pkg/queues"
	"time"
)

type Application struct {
	Cfg   *config.Config
	Meow  *Meow
	DB    *sqlstore.Container
	Queue *queues.Queue
}

func Start() (*Application, error) {
	cfg := config.Get()
	zap.S().Info("Starting application")
	db := cfg.ConnectToDatabase()
	meow := Init(cfg, db)
	queue := queues.InitQueue()

	meow.Connect()

	// queue runner
	// will run every second
	go func() {

	}()

	return &Application{
		Cfg:   cfg,
		DB:    db,
		Meow:  meow,
		Queue: queue,
	}, nil
}

func (app *Application) RunQueue() {
	/**
	 * For now, run the queue every second
	 * ToDo: make this configurable, and add every x seconds
	 */
	ticker := time.NewTicker(1 * time.Second)
	quit := make(chan struct{})

	for {
		select {
		case <-ticker.C:
			// ToDo: Implement other commands
			go app.SendMeow()
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

func (app *Application) SendMeow() {
	messageLength := app.Queue.Messages.Len()
	zap.S().Debugf("Queue length: %d", messageLength)

	if messageLength > 0 {
		e := app.Queue.Messages.Front()
		value := e.Value

		err := app.Meow.SendMessage(value.(PendingMessage))

		// push to back if error happens, then return
		if err != nil {
			app.Queue.Messages.MoveToBack(e)
			return
		}

		// remove from queue
		app.Queue.Messages.Remove(e)
	}
}
