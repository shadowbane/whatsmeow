package config

import (
	"flag"
	"fmt"
	"github.com/jinzhu/gorm"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
	"go.uber.org/zap"
	"gomeow/pkg/logger"
	"os"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/mattn/go-sqlite3"
)

type Config struct {
	appEnv string
	dbName string

	msgstoreUser string
	msgstorePswd string
	msgstorePort string
	msgstoreHost string
	msgstoreName string

	apiPort string
}

func Get() *Config {
	conf := &Config{}

	/** App Environment **/
	flag.StringVar(&conf.appEnv, "appenv", getenv("APP_ENV", "production"), "Application Environment")

	/** Database Configurations **/
	flag.StringVar(&conf.dbName, "dbname", getenv("DB_DATABASE", "meow.db"), "DB name")

	/** API Port Config **/
	flag.StringVar(&conf.apiPort, "apiPort", getenv("API_PORT", "8080"), "API Port")

	/**
	 * Message Store Database
	 * Using MySQL database
	 */
	flag.StringVar(&conf.msgstoreUser, "msgstoreUser", getenv("DB_MSGSTORE_USERNAME", "root"), "Message store DB user name")
	flag.StringVar(&conf.msgstorePswd, "msgstorePswd", getenv("DB_MSGSTORE_PASSWORD", "password"), "Message store DB pass")
	flag.StringVar(&conf.msgstorePort, "msgstorePort", getenv("DB_MSGSTORE_PORT", "3306"), "Message store DB port")
	flag.StringVar(&conf.msgstoreHost, "msgstoreHost", getenv("DB_MSGSTORE_HOST", "localhost"), "Message store DB host")
	flag.StringVar(&conf.msgstoreName, "msgstoreName", getenv("DB_MSGSTORE_DATABASE", "microservice_sekolah"), "Message store DB name")

	flag.Parse()

	logger.Init(conf.appEnv)

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

func (c *Config) GetMsgStoreConnStr() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?multiStatements=true&charset=utf8mb4&parseTime=True&loc=Local",
		c.msgstoreUser,
		c.msgstorePswd,
		c.msgstoreHost,
		c.msgstorePort,
		c.msgstoreName,
	)
}

func (c *Config) GetAPIPort() string {
	return ":" + c.apiPort
}

func (c *Config) ConnectToDatabase() *sqlstore.Container {

	logLevel := "ERROR"
	if c.appEnv != "production" {
		logLevel = "INFO"
	}

	dbLog := waLog.Stdout("Database", logLevel, true)
	db, err := sqlstore.New("sqlite3", c.GetDBConnStr(), dbLog)
	if err != nil {
		zap.S().Panicf("Failed to connect to database: %s", err)
		panic(err)
	}

	return db
}

func (c *Config) ConnectToMessageStore() *gorm.DB {
	zap.S().Debugf("Connecting to message store database @ %s:%s\n", c.msgstoreHost, c.msgstorePort)

	db, err := gorm.Open("mysql", c.GetMsgStoreConnStr())
	if err != nil {
		zap.S().Panicf("Failed to connect to database: %s", err)
		panic(err)
	}

	return db
}
