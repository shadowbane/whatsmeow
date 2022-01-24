package config

import (
	"flag"
	"go.uber.org/zap"
	"os"

	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"

	_ "github.com/mattn/go-sqlite3"
)

type Config struct {
	appEnv string
	dbUser string
	dbPswd string
	dbHost string
	dbPort string
	dbName string

	apiPort string
	migrate string

	redisHost string
	redisPort string
}

func Get() *Config {
	conf := &Config{}

	/** App Environment **/
	flag.StringVar(&conf.appEnv, "appenv", getenv("APP_ENV", "production"), "Application Environment")

	/** Database Configurations **/
	flag.StringVar(&conf.dbName, "dbname", getenv("DB_DATABASE", "meow.db"), "DB name")

	/** API Port Config **/
	flag.StringVar(&conf.apiPort, "apiPort", getenv("API_PORT", "8080"), "API Port")

	flag.Parse()

	return conf
}

func getenv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func (c *Config) GetAppEnv() string {
	return c.appEnv
}

func (c *Config) GetDBConnStr() string {
	return "file:" + c.dbName + "?_foreign_keys=on"
}

func (c *Config) GetAPIPort() string {
	return ":" + c.apiPort
}

func (c *Config) ConnectToDatabase() *sqlstore.Container {

	logLevel := "ERROR"
	if c.appEnv != "production" {
		logLevel = "DEBUG"
	}

	dbLog := waLog.Stdout("Database", logLevel, true)
	db, err := sqlstore.New("sqlite3", c.GetDBConnStr(), dbLog)
	if err != nil {
		zap.S().Panicf("Failed to connect to database: %s", err)
		panic(err)
	}

	return db
}
