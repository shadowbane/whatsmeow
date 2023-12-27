package application

import (
	"container/list"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.uber.org/zap"
	"gomeow/cmd/models"
	"gomeow/pkg/config"
	"gomeow/pkg/queues"
	"gomeow/pkg/validator"
	"gomeow/pkg/whatsmeow"
	"gorm.io/gorm"
	"time"
)

type Application struct {
	Cfg       *config.Config
	Meow      *whatsmeow.Meow
	DB        *sqlstore.Container
	Models    *gorm.DB
	Queue     *queues.Queue
	Validator *validator.Validator
}

func Start() (*Application, error) {
	cfg := config.Get()
	zap.S().Info("Starting application")
	meowdb := cfg.ConnectToWhatsmeowDB()
	queue := queues.InitQueue()
	database := cfg.ConnectToDB()

	// run automigration
	zap.S().Debug("Running auto migration")
	err := database.AutoMigrate([]interface{}{
		&models.Message{},
		&models.Device{},
		&models.Poll{},
		&models.PollDetail{},
		&models.PollHistory{},
	}...)
	if err != nil {
		zap.S().Fatalf("Error running auto migration: %v", err)
		panic(err)
	}

	app := &Application{
		DB:     meowdb,
		Cfg:    cfg,
		Queue:  queue,
		Models: database,
	}

	app.InitValidator()

	return app, nil
}

func (app *Application) InitValidator() {
	app.Validator = validator.InitValidator(app.Models)
}

// Deprecated: Unused
func (app *Application) LoadQueue(jid string) {
	zap.S().Info("Loading queue")

	var messages []models.Message
	app.Models.Where("sent = ? AND jid = ?", "0", jid).Find(&messages)

	zap.S().Info("Found ", len(messages), " messages to send")

	for _, message := range messages {

		pendingMessage := whatsmeow.PendingMessage{
			To:        message.Destination,
			MessageId: message.MessageId,
			Message:   message.Body,
		}

		app.Queue.Add(pendingMessage)
	}
}

// Deprecated: Unused
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

// Deprecated: Unused
func (app *Application) SendMeow() {
	messageLength := app.Queue.Messages.Len()

	if messageLength > 0 {
		zap.S().Debugf("Queue length: %d", messageLength)
		e := app.Queue.Messages.Front()
		value := e.Value

		// remove from queue and database
		app.RemoveFromQueue(e)

		err := app.Meow.SendMessage(value.(whatsmeow.PendingMessage))

		// Requeue if error happens.
		if err != nil {
			zap.S().Warnf("Error Sending Message: %s. Pushing message back to queue", err.Error())
			app.Queue.Add(value.(whatsmeow.PendingMessage))
			return
		} else {
			// mark as sent
			app.MarkAsSent(e)
		}
	}
}

// Deprecated: Unused
func (app *Application) RemoveFromQueue(e *list.Element) {
	app.Queue.Messages.Remove(e)
}

// Deprecated: Unused
func (app *Application) MarkAsSent(e *list.Element) {

	zap.S().Debugf("Marking message as sent")

	go func() {
		app.Models.
			Model(&models.Message{}).
			Where("message_id = ?", e.Value.(whatsmeow.PendingMessage).MessageId).
			Update("sent", true)
	}()
}
