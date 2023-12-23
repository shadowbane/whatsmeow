package wmeow

import (
	"context"
	"database/sql"
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
// Key: DeviceID
var ClientPointer = make(map[int64]*MeowClient)

var KillChannel = make(map[int64]chan bool)

// StartClient starts a new whatsmeow client.
// This will open a new connection, essentially creating a new 'web' session.
func StartClient(device *models.Device, app *application.Application, jid string, subscriptions []string) error {
	if isConnected(device) == true {
		return nil
	}

	var deviceStore *store.Device
	var err error

	// if JID is empty, create a new device
	if jid == "" {
		deviceStore = app.DB.NewDevice()
	} else {
		jid, _ := ParseJID(jid)

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
	zap.S().Debugf("WA Client Log Level: %s", app.Cfg.GetWALogLevel())
	clientLog := InitZapLogger("Client", app.Cfg.GetWALogLevel())

	var client *whatsmeow.Client
	client = whatsmeow.NewClient(deviceStore, clientLog)

	mycli := MeowClient{
		ClientLog:      clientLog,
		DB:             app.Models,
		DeviceStore:    deviceStore,
		eventHandlerID: 1,
		subscriptions:  subscriptions,
		Device:         device,
		WAClient:       client,
	}
	mycli.eventHandlerID = mycli.WAClient.AddEventHandler(mycli.myEventHandler)
	ClientPointer[device.ID] = &mycli

	KillChannel[device.ID] = make(chan bool, 1)

	zap.S().Debugf("KillChannel: %+v", KillChannel)

	// make mycli.ConnectAndLogin run in separate goroutine
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
	zap.S().Debugf("WMEOW\tKeeping DeviceID %d connected", device.ID)

	exithandler.WaitGroup.Add(1)

	// Keep connected client live until disconnected/killed
	go func() {
		<-KillChannel[device.ID]

		mycli.connection.close()

		zap.S().Debugf("WMEOW\tKilling channel with DeviceID %d", device.ID)

		// Clear QR code after pairing
		device.QRCode = sql.NullString{
			String: "",
		}

		result := app.Models.Save(&device)
		if result.Error != nil {
			zap.S().Errorf("WMEOW\tError updating device: %+v", result)
		}

		client.Disconnect()
		delete(ClientPointer, device.ID)

		defer func() {
			if exithandler.WaitGroup.GetCount() > 0 {
				exithandler.WaitGroup.Done()
			}
		}()
	}()

	return nil
}

// isConnected checks if a client is connected.
func isConnected(device *models.Device) bool {
	if ClientPointer[device.ID] != nil {
		isConnected := ClientPointer[device.ID].WAClient.IsConnected()
		if isConnected == true {
			return true
		}
	}

	return false
}

func StripJID(arg types.JID) string {
	jidVal := arg.String()

	if jidVal[0] == '+' {
		jidVal = jidVal[1:]
	}

	jidVal = strings.Split(jidVal, "@")[0]
	jidVal = strings.Split(jidVal, ".")[0]
	jidVal = strings.Split(jidVal, ":")[0]

	return jidVal
}

// ParseJID parses a JID from a string.
func ParseJID(arg string) (types.JID, bool) {
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

	// We need to remove everything after ":"
	phonenumber = strings.Split(phonenumber, ":")[0]

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
			zap.S().Error("WMEOW\tInvalid jid. No server specified")

			return recipient, false
		}
		return recipient, true
	}
}

// ConnectOnStartup This is the main entry point of the application
// if the device is logged in, it will be marked as connected
// and we will auto connect it to whatsapp
func ConnectOnStartup(app *application.Application) {
	var devices []models.Device
	result := app.Models.
		Where("is_connected = 1").
		Find(&devices)

	if result.Error != nil {
		zap.S().Errorf("WMEOW\tError getting devices: %+v", result)

		panic(result.Error)
	}

	if len(devices) == 0 {
		zap.S().Debug("WMEOW\tNo devices to connect on startup")

		return
	}

	var subscribedEvents []string

	for _, device := range devices {
		zap.S().Debugf("WMEOW\tAuto Connecting device %s", device.Name)

		subscribedEvents = strings.Split(device.Events, ",")

		go func(device *models.Device, app *application.Application, jid string, subscribedEvents []string) {
			err := StartClient(device, app, jid, subscribedEvents)
			if err != nil {
				zap.S().Errorf("WMEOW\tError starting client: %+v", err)
			}
		}(&device, app, device.JID.String, subscribedEvents)
	}

	zap.S().Debug("WMEOW\tConnected all devices on startup")
}

// Shutdown shuts down all clients.
func Shutdown() {
	zap.S().Debug("WMEOW\tShutting down WhatsMeow...")
	for k, _ := range KillChannel {
		KillChannel[k] <- true
	}
}
