package application

import (
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.uber.org/zap"
	"gomeow/pkg/config"
)

type Application struct {
	Cfg  *config.Config
	Meow *Meow
	DB   *sqlstore.Container
}

func Start() (*Application, error) {
	zap.S().Info("Starting application")
	cfg := config.Get()
	db := cfg.ConnectToDatabase()
	meow := Init(cfg, db)

	meow.Connect()

	return &Application{
		Cfg:  cfg,
		DB:   db,
		Meow: meow,
	}, nil
}
