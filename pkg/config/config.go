package config

import (
	"flag"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/shadowbane/go-logger"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.uber.org/zap"
	"gomeow/pkg/config/dblogger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	waLog "go.mau.fi/whatsmeow/util/log"
	gormlogger "gorm.io/gorm/logger"
)

type Config struct {
	appEnv string
	dbName string

	msgstoreUser string
	msgstorePswd string
	msgstorePort string
	msgstoreHost string
	msgstoreName string

	apiPort  string
	logLevel string
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
	flag.StringVar(&conf.msgstoreName, "msgstoreName", getenv("DB_MSGSTORE_DATABASE", "forge"), "Message store DB name")

	// If the log level is not set, set it to the value of conf.appEnv
	if getenv("LOG_LEVEL", "") == "" {
		os.Setenv("LOG_LEVEL", getDefaultLogLevel(conf.appEnv))
	}

	flag.StringVar(&conf.logLevel, "log level", getenv("LOG_LEVEL", "info"), "Log level")

	// load logger config
	logger.Init(logger.LoadEnvForLogger())

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

func (c *Config) GetMySQLConnectionString() string {
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

func (c *Config) GetLogLevel() string {
	// return c.logLevel in uppercase
	return strings.ToUpper(c.logLevel)
}

func (c *Config) ConnectToWhatsmeowDB() *sqlstore.Container {

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

func (c *Config) ConnectToDB() *gorm.DB {
	zap.S().Debugf("Connecting to database @%s:%s", c.msgstoreHost, c.msgstorePort)

	db, err := gorm.Open(mysql.Open(c.GetMySQLConnectionString()), &gorm.Config{
		Logger: &dblogger.ZapLogger{
			Config: gormlogger.Config{
				SlowThreshold:             2 * time.Second,
				LogLevel:                  dblogger.LogLevelToGormLevel(c.GetLogLevel()),
				IgnoreRecordNotFoundError: true,
				ParameterizedQueries:      false,
			},
		},
	})
	if err != nil {
		zap.S().Panicf("Failed to connect to database: %s", err)
		panic(err)
	}

	govalidator.SetFieldsRequiredByDefault(false)

	return db
}

func getDefaultLogLevel(environment string) string {
	switch environment {
	case "local":
		return "debug"
	case "debug":
		return "debug"
	case "testing":
		return "debug"
	default:
		return "info"
	}
}
