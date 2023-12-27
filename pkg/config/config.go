package config

import (
	"flag"
	"fmt"
	"github.com/shadowbane/go-logger"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.uber.org/zap"
	"gomeow/pkg/config/dblogger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"

	_ "github.com/mattn/go-sqlite3"
	waLog "go.mau.fi/whatsmeow/util/log"
	gormlogger "gorm.io/gorm/logger"
)

type Config struct {
	appEnv      string
	whatsmeowDB string

	dbUser                 string
	dbPass                 string
	dbPort                 string
	dbHost                 string
	dbName                 string
	dbLogLevel             string
	dbSlowQueryThreshold   int
	dbParameterizedQueries bool

	apiPort      string
	logLevel     string
	logDirectory string

	waLogLevel string
}

func Get() *Config {
	conf := &Config{}

	/** App Environment **/
	flag.StringVar(&conf.appEnv, "appenv", getenv("APP_ENV", "production"), "Application Environment")

	/** Database Configurations **/
	flag.StringVar(&conf.whatsmeowDB, "dbname", getenv("WHATSMEOW_DATABASE", "meow.db"), "DB name")

	/** API Port Config **/
	flag.StringVar(&conf.apiPort, "apiPort", getenv("API_PORT", "8080"), "API Port")

	/**
	 * Models Database - MySQL
	 */
	flag.StringVar(&conf.dbUser, "dbUser", getenv("DB_USERNAME", "root"), "Database user name")
	flag.StringVar(&conf.dbPass, "dbPass", getenv("DB_PASSWORD", "password"), "Database pass")
	flag.StringVar(&conf.dbPort, "dbPort", getenv("DB_PORT", "3306"), "Database port")
	flag.StringVar(&conf.dbHost, "dbHost", getenv("DB_HOST", "localhost"), "Database host")
	flag.StringVar(&conf.dbName, "dbName", getenv("DB_DATABASE", "forge"), "Database name")
	// Additional configurations - MySQL
	dbSlowQueryThreshold, _ := strconv.Atoi(getenv("DB_SLOW_QUERY_THRESHOLD", "1"))
	dbParameterizedQueries, _ := strconv.ParseBool(getenv("DB_PARAMETERIZED_QUERY", "true"))
	flag.StringVar(&conf.dbLogLevel, "dbLogLevel", getenv("DB_LOG_LEVEL", "warn"), "Database log level")
	flag.IntVar(&conf.dbSlowQueryThreshold, "dbSlowQueryThreshold", dbSlowQueryThreshold, "Database slow query threshold in seconds")
	flag.BoolVar(&conf.dbParameterizedQueries, "dbParameterizedQueries", dbParameterizedQueries, "Database parameterized queries")

	// If the log level is not set, set it to the value of conf.appEnv
	if getenv("LOG_LEVEL", "") == "" {
		os.Setenv("LOG_LEVEL", getDefaultLogLevel(conf.appEnv))
	}

	// If the log directory is not set, set it to storage/logs
	if getenv("LOG_DIRECTORY", "") == "" {
		os.Setenv("LOG_DIRECTORY", "storage/logs")
	}

	flag.StringVar(&conf.logLevel, "log level", getenv("LOG_LEVEL", "info"), "Log level")
	flag.StringVar(&conf.logDirectory, "log directory", getenv("LOG_DIRECTORY", "storage/logs"), "Log Directory")
	flag.StringVar(&conf.waLogLevel, "whatsapp log level", getenv("WA_LOG_LEVEL", "error"), "WhatsApp Log level")

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

func (c *Config) Get(name string) interface{} {
	fieldValue := reflect.ValueOf(c).Elem().FieldByName(name)

	return getUnexportedField(fieldValue)
}

func getUnexportedField(field reflect.Value) interface{} {
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Interface()
}

func (c *Config) GetAppEnv() string {
	return c.appEnv
}

func (c *Config) GetDBConnStr() string {
	return "file:" + c.whatsmeowDB + "?_foreign_keys=on"
}

func (c *Config) GetMySQLConnectionString() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?multiStatements=true&charset=utf8mb4&parseTime=True&loc=Local",
		c.dbUser,
		c.dbPass,
		c.dbHost,
		c.dbPort,
		c.dbName,
	)
}

func (c *Config) GetAPIPort() string {
	return ":" + c.apiPort
}

func (c *Config) GetLogLevel() string {
	return strings.ToUpper(c.logLevel)
}

func (c *Config) GetWALogLevel() string {
	return strings.ToUpper(c.waLogLevel)
}

func (c *Config) ConnectToWhatsmeowDB() *sqlstore.Container {
	logLevel := "ERROR"
	if c.appEnv != "production" {
		logLevel = c.GetLogLevel()
	}

	dbLog := waLog.Stdout("Database", logLevel, true)
	zap.S().Debugf("Connecting to sqlite database at @%s", c.GetDBConnStr())
	db, err := sqlstore.New("sqlite3", c.GetDBConnStr(), dbLog)
	if err != nil {
		zap.S().Panicf("Failed to connect to database: %s", err)
		panic(err)
	}

	return db
}

func (c *Config) ConnectToDB() *gorm.DB {
	zap.S().Debugf("Connecting to database @%s:%s", c.dbHost, c.dbPort)

	db, err := gorm.Open(mysql.Open(c.GetMySQLConnectionString()), &gorm.Config{
		Logger: &dblogger.ZapLogger{
			Config: gormlogger.Config{
				SlowThreshold:             time.Second * time.Duration(c.dbSlowQueryThreshold),
				LogLevel:                  dblogger.LogLevelToGormLevel(c.dbLogLevel),
				IgnoreRecordNotFoundError: true,
				ParameterizedQueries:      c.dbParameterizedQueries,
			},
		},
	})
	if err != nil {
		zap.S().Panicf("Failed to connect to database: %s", err)
		panic(err)
	}

	dbInstance, err := db.DB()
	if err != nil {
		zap.S().Panicf("Failed to connect to database: %s", err)
		panic(err)
	}

	dbInstance.SetMaxIdleConns(10)
	dbInstance.SetMaxOpenConns(100)
	dbInstance.SetConnMaxLifetime(time.Hour)

	return db
}

// getDefaultLogLevel returns the default log level for the given environment.
// This is used when the log level is not set
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
