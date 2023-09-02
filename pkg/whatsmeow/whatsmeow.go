package whatsmeow

import (
	"context"
	"fmt"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"go.uber.org/zap"
	"gomeow/cmd/models"
	"gomeow/pkg/config"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
	"os"
	"strings"
)

// Deprecated: Not used in new version
type Meow struct {
	DeviceStore *store.Device
	ClientLog   waLog.Logger
	Client      *whatsmeow.Client
	DB          *gorm.DB
}

// Deprecated: Not used in new version
type PendingMessage struct {
	Message   string `json:"message"`
	To        string `json:"to"`
	MessageId string `json:"messageId"`
}

// Deprecated: Not used in new version
type CustomLogger waLog.Logger

// Deprecated: Not used in new version
type zapLogger struct {
	module  string
	minimum int
}

// Deprecated: Not used in new version
func (s *zapLogger) Errorf(msg string, args ...interface{}) {
	if levelToInt["ERROR"] < s.minimum {
		return
	}
	zap.S().Errorf("[WhatsApp]\t"+msg, args...)
}

// Deprecated: Not used in new version
func (s *zapLogger) Warnf(msg string, args ...interface{}) {
	if levelToInt["WARN"] < s.minimum {
		return
	}
	zap.S().Warnf("[WhatsApp]\t"+msg, args...)
}

// Deprecated: Not used in new version
func (s *zapLogger) Infof(msg string, args ...interface{}) {
	if levelToInt["INFO"] < s.minimum {
		return
	}
	zap.S().Infof("[WhatsApp]\t"+msg, args...)
}

// Deprecated: Not used in new version
func (s *zapLogger) Debugf(msg string, args ...interface{}) {
	if levelToInt["DEBUG"] < s.minimum {
		return
	}
	zap.S().Debugf("[WhatsApp]\t"+msg, args...)
}

// Deprecated: Not used in new version
func (s *zapLogger) Sub(module string) waLog.Logger {
	return &zapLogger{module: fmt.Sprintf("%s/%s", s.module, module), minimum: s.minimum}
}

// Deprecated: Not used in new version
var levelToInt = map[string]int{
	"":      -1,
	"DEBUG": 0,
	"INFO":  1,
	"WARN":  2,
	"ERROR": 3,
}

// Deprecated: Not used in new version
func InitZapLogger(module string, minLevel string) waLog.Logger {
	return &zapLogger{
		module:  module,
		minimum: levelToInt[strings.ToUpper(minLevel)],
	}
}

// Deprecated: Not used in new version
func Init(c *config.Config, container *sqlstore.Container, db *gorm.DB) *Meow {
	// init device store
	store.DeviceProps.PlatformType = waProto.DeviceProps_CHROME.Enum()
	//store.CompanionProps.Os = waProto.UserAgent_WINDOWS.String()
	//store.CompanionProps.Version = "1.0.0"
	zap.S().Info("Initializing DeviceStore")
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		zap.S().Panicf("Error: %s", err)
		panic(err)
	}

	zap.S().Debug("JID: ", deviceStore.ID)

	// init client log
	clientLog := InitZapLogger("Client", c.GetWALogLevel())

	// init client
	zap.S().Info("Initializing WhatsMeow Client")
	client := whatsmeow.NewClient(deviceStore, clientLog)

	return &Meow{
		DeviceStore: deviceStore,
		ClientLog:   clientLog,
		Client:      client,
		DB:          db,
	}
}

// Deprecated: Not used in new version
func (m *Meow) Connect() {
	if m.Client.Store.ID == nil {
		zap.S().Info("No credential found, creating new device")
		// No ID stored, new login
		qrChan, err := m.Client.GetQRChannel(context.Background())
		err = m.Client.Connect()
		if err != nil {
			zap.S().Info("Connecting to WhatsApp")
			zap.S().Panicf("Failed to connect to WhatsApp: %s", err)
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				// Render the QR code here
				// e.g. qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				// or just manually `echo 2@... | qrencode -t ansiutf8` in a terminal
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				fmt.Println("QR code:", evt.Code)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		zap.S().Info("Connecting to WhatsApp")
		err := m.Client.Connect()
		if err != nil {
			zap.S().Panicf("Failed to connect to WhatsApp: %s", err)
			panic(err)
		}

		m.Client.AddEventHandler(m.eventHandler)
	}
}

// Deprecated: Not used in new version
func (m *Meow) Exit() {
	m.Client.Disconnect()
}

// Deprecated: Not used in new version
func (m *Meow) SendMessage(message PendingMessage) error {
	zap.S().Debugf("Sending message with ID: %s and content: %s to: %s", message.MessageId, message.Message, message.To)

	newJid := types.NewJID(message.To, "s.whatsapp.net")
	newMessage := &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text: proto.String(message.Message),
		},
	}

	_, err := m.Client.SendMessage(context.Background(), newJid, newMessage)
	if err != nil {
		zap.S().Errorf(err.Error())
		return err
	}

	return nil
}

// Deprecated: Not used in new version
func (m *Meow) eventHandler(evt interface{}) {
	switch v := evt.(type) {

	case *events.Message:
		zap.S().Debugf("Received a message: %s", v.Message.GetConversation())

	case *events.Receipt:
		if v.Type == events.ReceiptTypeRead {
			zap.S().Debugf("Received a read receipt [%s]", v.MessageIDs)

			go func() {
				for _, messageId := range v.MessageIDs {
					m.DB.Model(&models.Message{}).
						Where("message_id = ?", messageId).
						Update("read", true)
				}
			}()
		}
	}
}
