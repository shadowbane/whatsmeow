package wmeow

import (
	"context"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
	"go.uber.org/zap"
	"gomeow/cmd/models"
	"gomeow/pkg/application"
	"gomeow/pkg/exithandler"
	"strings"
)

// ClientPointer is a map of whatsmeow clients.
// This is the heart of the application.
//
// Key: UserID
var ClientPointer = make(map[int64]*MeowClient)

var KillChannel = make(map[int64](chan bool))

func StartClient(user *models.User, app *application.Application, jid string, subscriptions []string) error {
	if isConnected(user) == true {
		return nil
	}

	var deviceStore *store.Device
	var err error

	// if JID is empty, create a new device
	if jid == "" {
		deviceStore = app.DB.NewDevice()
	} else {
		jid, _ := parseJID(jid)

		deviceStore, err = app.DB.GetDevice(jid)
		if err != nil {
			zap.S().Errorf("WMEOW\tError: %v", err)

			return err
		}
	}

	osName := "Mac OS 13"
	store.DeviceProps.PlatformType = waProto.DeviceProps_UNKNOWN.Enum()
	store.DeviceProps.Os = &osName

	// init client log
	logLevel := "ERROR"
	if app.Cfg.GetAppEnv() != "production" {
		logLevel = app.Cfg.GetLogLevel()
	}

	clientLog := InitZapLogger("Client", logLevel)
	var client *whatsmeow.Client
	client = whatsmeow.NewClient(deviceStore, clientLog)

	mycli := MeowClient{
		ClientLog:      clientLog,
		DB:             app.Models,
		DeviceStore:    deviceStore,
		eventHandlerID: 1,
		subscriptions:  subscriptions,
		User:           user,
		WAClient:       client,
	}
	mycli.eventHandlerID = mycli.WAClient.AddEventHandler(mycli.myEventHandler)
	ClientPointer[user.ID] = &mycli

	KillChannel[user.ID] = make(chan bool, 1)

	zap.S().Debugf("KillChannel: %+v", KillChannel)
	//err, _ = mycli.ConnectAndLogin(err)
	//if err != nil {
	//	zap.S().Errorf("WMEOW\tError: %v", err)
	//}

	// make mycli.ConnectAndLogin run in separate goroutine
	// ToDo: This is still buggy as hell
	// If we disconnect a client that has not been logged in,
	// sometimes the qrChan is not closed and still receiving events
	ctx, cancel := context.WithCancel(context.Background())
	mycli.connection = &connectionContext{
		ctx:   ctx,
		close: cancel,
	}

	go func() {
		err, _ := mycli.ConnectAndLogin(err)
		if err != nil {
			zap.S().Errorf("WMEOW\tError: %v", err)
		}
	}()

	// Keep connected client live until disconnected/killed
	zap.S().Debugf("WMEOW\tKeeping UserID %d connected", user.ID)

	exithandler.WaitGroup.Add(1)

	// Keep connected client live until disconnected/killed
	go func() {
		<-KillChannel[user.ID]

		mycli.connection.close()

		zap.S().Debugf("WMEOW\tKilling channel with UserID %d", user.ID)

		// Clear QR code after pairing
		user.QRCode = ""

		result := app.Models.Save(&user)
		if result.Error != nil {
			zap.S().Errorf("WMEOW\tError updating user: %+v", result)
		}

		client.Disconnect()
		delete(ClientPointer, user.ID)

		defer func() {
			if exithandler.WaitGroup.GetCount() > 0 {
				exithandler.WaitGroup.Done()
			}
		}()
	}()

	return nil
}

// isConnected checks if a client is connected.
func isConnected(user *models.User) bool {
	if ClientPointer[user.ID] != nil {
		isConnected := ClientPointer[user.ID].WAClient.IsConnected()
		if isConnected == true {
			return true
		}
	}

	return false
}

// parseJID parses a JID from a string.
func parseJID(arg string) (types.JID, bool) {
	if arg == "" {
		return types.NewJID("", types.DefaultUserServer), false
	}
	if arg[0] == '+' {
		arg = arg[1:]
	}

	// Basic only digit check for recipient phone number, we want to remove @server and .session
	phonenumber := ""
	phonenumber = strings.Split(arg, "@")[0]
	phonenumber = strings.Split(phonenumber, ".")[0]
	b := true
	for _, c := range phonenumber {
		if c < '0' || c > '9' {
			b = false
			break
		}
	}
	if b == false {
		zap.S().Warn("WMEOW\tInvalid JID format")
		recipient, _ := types.ParseJID("")

		return recipient, false
	}

	if !strings.ContainsRune(arg, '@') {
		return types.NewJID(arg, types.DefaultUserServer), true
	} else {
		recipient, err := types.ParseJID(arg)
		if err != nil {
			zap.S().Errorf("WMEOW\tError: %s", err)
			zap.S().Errorf("WMEOW\tInvalid jid: %s", arg)

			return recipient, false
		} else if recipient.User == "" {
			zap.S().Errorf("WMEOW\tError: %s", err)
			zap.S().Error("WMEOW\tInvalid jid. No server specifieds")

			return recipient, false
		}
		return recipient, true
	}
}

// ConnectOnStartup This is the main entry point of the application
// if the device is logged in, it will be marked as connected
// and we will auto connect it to whatsapp
func ConnectOnStartup(app *application.Application) {
	var users []models.User
	result := app.Models.
		Where("is_connected = 1").
		Find(&users)

	if result.Error != nil {
		zap.S().Errorf("WMEOW\tError getting users: %+v", result)

		panic(result.Error)
	}

	if len(users) == 0 {
		zap.S().Debug("WMEOW\tNo users to connect on startup")

		return
	}

	var subscribedEvents []string

	for _, user := range users {
		zap.S().Debugf("WMEOW\tAuto Connecting user %s", user.Name)

		subscribedEvents = strings.Split(user.Events, ",")

		go func(user *models.User, app *application.Application, jid string, subscribedEvents []string) {
			err := StartClient(user, app, jid, subscribedEvents)
			if err != nil {
				zap.S().Errorf("WMEOW\tError starting client: %+v", err)
			}
		}(&user, app, user.JID, subscribedEvents)
	}

	zap.S().Debug("WMEOW\tConnected all users on startup")
}

// Shutdown shuts down all clients.
func Shutdown() {
	zap.S().Debug("WMEOW\tShutting down WhatsMeow...")
	for k, _ := range KillChannel {
		KillChannel[k] <- true
	}
}
