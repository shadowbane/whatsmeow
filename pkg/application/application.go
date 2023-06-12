package application

import (
	"container/list"
	"github.com/jinzhu/gorm"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	"go.uber.org/zap"
	"gomeow/pkg/config"
	"gomeow/pkg/queues"
	"time"

	"gomeow/cmd/models"
)

type Application struct {
	Cfg          *config.Config
	Meow         *Meow
	DB           *sqlstore.Container
	MessageStore *gorm.DB
	Queue        *queues.Queue
}

func Start() (*Application, error) {
	cfg := config.Get()
	zap.S().Info("Starting application")
	db := cfg.ConnectToDatabase()
	meow := Init(cfg, db)
	queue := queues.InitQueue()
	msgStore := cfg.ConnectToMessageStore()

	// run automigration
	zap.S().Info("Running auto migration")
	msgStore.AutoMigrate(&models.Message{})

	meow.Connect()

	return &Application{
		Cfg:          cfg,
		DB:           db,
		Meow:         meow,
		Queue:        queue,
		MessageStore: msgStore,
	}, nil
}

func (app *Application) AddReadEventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Receipt:
		if v.Type == events.ReceiptTypeRead {
			zap.S().Debugf("Received a read receipt [%s]", v.MessageIDs)

			go func() {
				for _, messageId := range v.MessageIDs {
					app.MessageStore.Model(&models.Message{}).
						Where("message_id = ?", messageId).
						Update("read", true)
				}
			}()
		}
	}
}

func (app *Application) LoadQueue(jid string) {
	zap.S().Info("Loading queue")

	var messages []models.Message
	app.MessageStore.Where("sent = ? AND jid = ?", "0", jid).Find(&messages)

	zap.S().Info("Found ", len(messages), " messages to send")

	for _, message := range messages {

		pendingMessage := PendingMessage{
			To:        message.Destination,
			MessageId: message.MessageId,
			Message:   message.Body,
		}

		app.Queue.Add(pendingMessage)
	}
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
	//zap.S().Debugf("Queue length: %d", messageLength)

	if messageLength > 0 {
		e := app.Queue.Messages.Front()
		value := e.Value

		// remove from queue and database
		app.RemoveFromQueue(e)

		err := app.Meow.SendMessage(value.(PendingMessage))

		// Requeue if error happens.
		if err != nil {
			zap.S().Warnf("Error Sending Message: %s. Pushing message back to queue", err.Error())
			app.Queue.Add(value.(PendingMessage))
			return
		} else {
			// mark as sent
			app.MarkAsSent(e)
		}
	}
}

func (app *Application) RemoveFromQueue(e *list.Element) {
	app.Queue.Messages.Remove(e)
}

func (app *Application) MarkAsSent(e *list.Element) {

	zap.S().Debugf("Marking message as sent")

	// find record in database
	storedMessage := app.findMessageById(e.Value.(PendingMessage).MessageId)

	// mark as sent
	storedMessage.Sent = true
	app.MessageStore.Save(&storedMessage)
}

func (app *Application) findMessageById(messageId string) *models.Message {
	message := models.Message{}
	app.MessageStore.Where("message_id = ?", messageId).First(&message)

	return &message
}
