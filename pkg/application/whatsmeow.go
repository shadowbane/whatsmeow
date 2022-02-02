package application

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
	"gomeow/pkg/config"
	"google.golang.org/protobuf/proto"
	"math/rand"
	"os"
	"time"
)

type Meow struct {
	DeviceStore *store.Device
	ClientLog   waLog.Logger
	Client      *whatsmeow.Client
}

type PendingMessage struct {
	Message   string `json:"message"`
	To        string `json:"to"`
	MessageId string `json:"messageId"`
}

func Init(c *config.Config, container *sqlstore.Container) *Meow {
	// init device store
	store.CompanionProps.PlatformType = waProto.CompanionProps_CHROME.Enum()
	//store.CompanionProps.Os = waProto.UserAgent_WINDOWS.String()
	//store.CompanionProps.Version = "1.0.0"
	zap.S().Info("Initializing DeviceStore")
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		zap.S().Panicf("Error: %s", err)
		panic(err)
	}

	zap.S().Debug("JID: ", deviceStore.ID)
	//zap.S().Fatal("Exited")

	//fmt.Printf("%+v\n", store.CompanionProps)
	//fmt.Printf("%+v\n", deviceStore.Platform)
	//panic("test")

	// init client log
	logLevel := "ERROR"
	if c.GetAppEnv() != "production" {
		logLevel = "INFO"
	}

	clientLog := waLog.Stdout("Client", logLevel, true)

	// init client
	zap.S().Info("Initializing WhatsMeow Client")
	client := whatsmeow.NewClient(deviceStore, clientLog)
	client.AddEventHandler(eventHandler)

	return &Meow{
		DeviceStore: deviceStore,
		ClientLog:   clientLog,
		Client:      client,
	}
}

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
	}
}

func (m *Meow) Exit() {
	m.Client.Disconnect()
}

func (m *Meow) SendMessage(message PendingMessage) error {
	// add random delay
	r := rand.Intn(10)
	time.Sleep(time.Duration(r) * time.Second)
	zap.S().Debugf("Sending message with ID: %s and content: %s to: %s", message.MessageId, message.Message, message.To)

	newJid := types.NewJID(message.To, "s.whatsapp.net")
	newMessage := &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text: proto.String(message.Message),
		},
	}

	_, err := m.Client.SendMessage(newJid, message.MessageId, newMessage)
	if err != nil {
		zap.S().Errorf(err.Error())
		return err
	}

	return nil
}

func eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		zap.S().Debugf("Received a message: %s", v.Message.GetConversation())
	}
}
