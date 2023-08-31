module gomeow

go 1.20

require (
	github.com/go-playground/locales v0.14.1
	github.com/go-playground/universal-translator v0.18.1
	github.com/go-playground/validator/v10 v10.15.3
	github.com/joho/godotenv v1.5.1
	github.com/julienschmidt/httprouter v1.3.0
	github.com/mattn/go-sqlite3 v1.14.17
	github.com/mdp/qrterminal/v3 v3.1.1
	github.com/oklog/ulid/v2 v2.1.0
	github.com/shadowbane/go-logger v0.1.0-alpha
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
	go.mau.fi/whatsmeow v0.0.0-20230824151650-6da2abde6b7c
	go.uber.org/zap v1.25.0
	google.golang.org/protobuf v1.31.0
	gorm.io/driver/mysql v1.5.1
	gorm.io/gorm v1.25.4
)

require (
	filippo.io/edwards25519 v1.0.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/go-sql-driver/mysql v1.7.1 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	go.mau.fi/libsignal v0.1.0 // indirect
	go.mau.fi/util v0.0.0-20230805171708-199bf3eec776 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.12.0 // indirect
	golang.org/x/net v0.14.0 // indirect
	golang.org/x/sys v0.11.0 // indirect
	golang.org/x/text v0.12.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	rsc.io/qr v0.2.0 // indirect
)

// Use local path for github.com/shadowbane/go-logger
replace github.com/shadowbane/go-logger => /Users/adliifkar/projects/go/go-logger
