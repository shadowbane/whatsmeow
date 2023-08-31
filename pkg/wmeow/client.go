package wmeow

import (
	"context"
	"database/sql"
	"errors"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"go.uber.org/zap"
	"gomeow/cmd/models"
	"google.golang.org/protobuf/proto"
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
	mycli.User.JID = sql.NullString{String: ""}
	mycli.User.Webhook = sql.NullString{String: ""}
	result := mycli.DB.Save(&mycli.User)
	if result.Error != nil {
		zap.S().Errorf("WMEOW\tError updating user: %+v", result)
	}
	KillChannel[mycli.User.ID] <- true
	zap.S().Infof("WMEOW\tLogged out of WAClient.")
}

// myEventHandler is the event handler for the WAClient.
// This will handle all the events from the WAClient.
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
		mycli.User.JID = sql.NullString{
			String: evt.ID.String(),
			Valid:  true,
		}
		mycli.User.IsConnected = true

		result := mycli.DB.Save(&mycli.User)
		if result.Error != nil {
			zap.S().Errorf("WMEOW\tError updating user: %+v", result)
		}
	case *events.StreamReplaced:
		zap.S().Warnf("Stream Replaced!")
	case *events.Message:
		if evt.Message.GetPollUpdateMessage() != nil {
			pollVote, err := mycli.WAClient.DecryptPollVote(evt)

			if err != nil {
				zap.S().Errorf("WMEOW\tError decrypting poll vote: %v", err)
				return
			}
			zap.S().Infof("WMEOW\tPoll vote received from %s In %s", evt.Info.Sender, evt.Message.GetMessageContextInfo())
			for _, hash := range pollVote.GetSelectedOptions() {
				zap.S().Infof("WMEOW\tPoll vote: - %X", hash)
			}
		} else {
			zap.S().Debugf("WMEOW\tReceived message form %s: %s", evt.Info.Sender, evt.Message.GetConversation())
		}

	case *events.Receipt:
		if evt.Type == events.ReceiptTypeRead {
			zap.S().Debugf("WMEOW\tReceived a read receipt [%s]", evt.MessageIDs)

			go func() {
				for _, messageId := range evt.MessageIDs {
					mycli.DB.Model(&models.Message{}).
						Where("message_id = ?", messageId).
						Update("read", true)
				}
			}()
		}
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

// ConnectAndLogin connects to the WAClient and logs in if necessary.
// This will also shows the QR code if the user is not logged in
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

		qrChan, err := mycli.WAClient.GetQRChannel(mycli.connection.ctx)
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

					mycli.User.QRCode = sql.NullString{
						String: evt.Code,
						Valid:  true,
					}

					result := mycli.DB.Save(&mycli.User)
					if result.Error != nil {
						zap.S().Errorf("WMEOW\tError updating mycli.User: %+v", result)

						return result.Error, true
					}
				} else if evt.Event == "timeout" {
					// Clear QR code from DB on timeout
					mycli.User.QRCode = sql.NullString{
						String: "",
					}

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
					mycli.User.QRCode = sql.NullString{String: ""}

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

// SendTextMessage sends a message to a given JID
func (mycli *MeowClient) SendTextMessage(messageId string, to string, message string) error {
	zap.S().Debugf("Sending message with ID: %s and content: %s to: %s", messageId, message, to)

	newJid := types.NewJID(to, "s.whatsapp.net")
	newMessage := &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text: proto.String(message),
		},
	}

	go func() {
		_, err := mycli.WAClient.SendMessage(mycli.connection.ctx, newJid, newMessage, whatsmeow.SendRequestExtra{
			ID:   messageId,
			Peer: false,
		})

		if err != nil {
			zap.S().Errorf(err.Error())
		}

		mycli.DB.Model(&models.Message{}).
			Where("message_id = ?", messageId).
			Update("sent", true)
	}()

	return nil
}
