package wmeow

import (
	"context"
	"errors"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"go.uber.org/zap"
	"gomeow/cmd/models"
	"gomeow/pkg/application"
	"gorm.io/gorm"
	"strings"
)

type MeowClient struct {
	ClientLog      waLog.Logger
	DB             *gorm.DB
	DeviceStore    *store.Device
	eventHandlerID uint32
	subscriptions  []string
	User           *models.User
	WAClient       *whatsmeow.Client
}

func (mycli *MeowClient) Logout() {
	mycli.User.IsConnected = false
	mycli.User.JID = ""
	mycli.User.Webhook = ""
	result := mycli.DB.Save(&mycli.User)
	if result.Error != nil {
		zap.S().Errorf("WMEOW\tError updating user: %+v", result)
	}
	KillChannel[mycli.User.ID] <- true
	zap.S().Infof("WMEOW\tLogged out of WAClient.")
}

func (mycli *MeowClient) myEventHandler(rawEvt interface{}) {
	switch evt := rawEvt.(type) {
	case *events.AppStateSyncComplete:
	case *events.Connected, *events.PushNameSetting:
		zap.S().Debugf("WMEOW\tConnected to WAClient.")
		if len(mycli.WAClient.Store.PushName) == 0 {
			return
		}

		err := mycli.WAClient.SendPresence(types.PresenceAvailable)
		if err != nil {
			zap.S().Errorf("WMEOW\tFailed to send available presence: %+v", err)
		} else {
			zap.S().Info("Marked self as available")
		}

		mycli.User.IsConnected = true
		result := mycli.DB.Save(&mycli.User)
		if result.Error != nil {
			zap.S().Errorf("WMEOW\tError updating user: %+v", result)
		}
	case *events.PairSuccess:
		zap.S().Infof("WMEOW\tPairing success for %d. JID: %s", mycli.User.ID, evt.ID.String())
		mycli.User.JID = evt.ID.String()
		mycli.User.IsConnected = true

		result := mycli.DB.Save(&mycli.User)
		if result.Error != nil {
			zap.S().Errorf("WMEOW\tError updating user: %+v", result)
		}
	case *events.StreamReplaced:
		zap.S().Warnf("Stream Replaced!")
	case *events.Message:
	case *events.Receipt:
	case *events.Presence:
	case *events.HistorySync:
	case *events.AppState:
	case *events.LoggedOut:
		mycli.Logout()
	case *events.ChatPresence:
	case *events.CallOffer:
		zap.S().Infof("Got CallOffer event - %+v", evt)
	case *events.CallAccept:
		zap.S().Infof("Got CallAccept event - %+v", evt)
	case *events.CallTerminate:
		zap.S().Infof("Got CallTerminate event - %+v", evt)
	case *events.CallOfferNotice:
		zap.S().Infof("Got CallOfferNotice event - %+v", evt)
	case *events.CallRelayLatency:
		zap.S().Infof("Got CallRelayLatency event - %+v", evt)
	default:
		zap.S().Debugf("WMEOW\tUnhandled event: %v", evt)

	}
}

// ClientPointer is a map of whatsmeow clients.
// This is the heart of the application.
//
// Key: UserID
var ClientPointer = make(map[int64]*MeowClient)

// var KillChannel = make(map[int64]chan bool)
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
	// make mycli.ConnectAndLogin run in separate goroutine
	go func() {
		err, _ := mycli.ConnectAndLogin(err)
		if err != nil {
			zap.S().Errorf("WMEOW\tError: %v", err)
		}
	}()

	// Keep connected client live until disconnected/killed
	zap.S().Debugf("WMEOW\tKeeping UserID %d connected", user.ID)

	// Keep connected client live until disconnected/killed
	go func() {
		<-KillChannel[user.ID]
		zap.S().Debugf("WMEOW\tKilling channel with UserID %d", user.ID)

		// Clear QR code after pairing
		user.QRCode = ""

		result := app.Models.Save(&user)
		if result.Error != nil {
			zap.S().Errorf("WMEOW\tError updating user: %+v", result)
		}

		client.Disconnect()
		delete(ClientPointer, user.ID)
	}()

	return nil
}

func (mycli *MeowClient) ConnectAndLogin(err error) (error, bool) {
	if mycli.WAClient.Store.ID != nil {
		zap.S().Debugf("WMEOW\tDevice %s already logged in", mycli.WAClient.Store.ID)
		err = mycli.WAClient.Connect()
		if err != nil {
			zap.S().Panicf("WMEOW\tError: %v", err)

			panic(err)
		}

		return nil, true
	} else {
		zap.S().Debug("WMEOW\tlogging in")

		qrChan, err := mycli.WAClient.GetQRChannel(context.Background())
		if err != nil {
			if !errors.Is(err, whatsmeow.ErrQRStoreContainsID) {
				zap.S().Errorf("WMEOW\tFailed to get QR channel: %v", err)
			}

			return err, true
		} else {
			zap.S().Debug("WMEOW\tGetting QR code")
			err = mycli.WAClient.Connect()
			if err != nil {
				zap.S().Panicf("WMEOW\tError: %v", err)
				panic(err)
			}

			for evt := range qrChan {
				if evt.Event == "code" {

					mycli.User.QRCode = evt.Code

					result := mycli.DB.Save(&mycli.User)
					if result.Error != nil {
						zap.S().Errorf("WMEOW\tError updating mycli.User: %+v", result)

						return result.Error, true
					}
				} else if evt.Event == "timeout" {
					// Clear QR code from DB on timeout
					mycli.User.QRCode = ""

					result := mycli.DB.Save(&mycli.User)
					if result.Error != nil {
						zap.S().Errorf("WMEOW\tError updating mycli.User: %+v", result)

						return result.Error, true
					}

					zap.S().Errorf("WMEOW\tQR Code Timeout... Killing channel...")

					delete(ClientPointer, mycli.User.ID)
					KillChannel[mycli.User.ID] <- true
				} else if evt.Event == "success" {
					zap.S().Debugf("WMEOW\tQR pairing ok!")
					// Clear QR code after pairing
					mycli.User.QRCode = ""

					result := mycli.DB.Save(&mycli.User)
					if result.Error != nil {
						zap.S().Errorf("WMEOW\tError updating user: %+v", result)

						return result.Error, true
					}
				} else {
					zap.S().Debugf("WMEOW\tLogin event: %v", evt.Event)
				}
			}
		}
	}
	return nil, false
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

func Shutdown() {
	zap.S().Debug("WMEOW\tShutting down WhatsMeow...")
	for k, _ := range KillChannel {
		KillChannel[k] <- true
	}
}
