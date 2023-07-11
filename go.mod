module gomeow

go 1.20

require (
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2
	github.com/joho/godotenv v1.5.1
	github.com/julienschmidt/httprouter v1.3.0
	github.com/mattn/go-sqlite3 v1.14.17
	github.com/mdp/qrterminal/v3 v3.1.1
	github.com/shadowbane/go-logger v0.1.0-alpha
	go.mau.fi/whatsmeow v0.0.0-20230710094417-93091c7024da
	go.uber.org/zap v1.24.0
	google.golang.org/protobuf v1.31.0
	gorm.io/driver/mysql v1.5.1
	gorm.io/gorm v1.25.2
)

require (
	filippo.io/edwards25519 v1.0.0 // indirect
	github.com/go-sql-driver/mysql v1.7.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	go.mau.fi/libsignal v0.1.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.11.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	rsc.io/qr v0.2.0 // indirect
)

// Use local path for github.com/shadowbane/go-logger
replace github.com/shadowbane/go-logger => /Users/adliifkar/projects/go/go-logger
