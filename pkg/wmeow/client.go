package wmeow

import (
	"context"
	"errors"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"go.uber.org/zap"
	"gomeow/cmd/models"
	"gorm.io/gorm"

	waLog "go.mau.fi/whatsmeow/util/log"
)

type MeowClient struct {
	ClientLog      waLog.Logger
	DB             *gorm.DB
	DeviceStore    *store.Device
	eventHandlerID uint32
	subscriptions  []string
	User           *models.User
	WAClient       *whatsmeow.Client
	connection     *connectionContext
}

type connectionContext struct {
	ctx   context.Context
	close context.CancelFunc
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

func (mycli *MeowClient) ConnectAndLogin(err error) (error, bool) {
	for {
		select {
		case <-mycli.connection.ctx.Done():

			return nil, false
		default:
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

			mycli.connection.close()

			return nil, false
		}
	}
}
