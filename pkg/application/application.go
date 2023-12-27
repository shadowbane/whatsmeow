package application

import (
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.uber.org/zap"
	"gomeow/cmd/models"
	"gomeow/pkg/config"
	"gomeow/pkg/queues"
	"gomeow/pkg/validator"
	"gorm.io/gorm"
)

type Application struct {
	Cfg       *config.Config
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
