package wmeow

import (
	"context"
	"database/sql"
	"errors"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
	"go.uber.org/zap"
	"gomeow/cmd/models"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
	"math/rand"
	"time"

	waProto "go.mau.fi/whatsmeow/binary/proto"
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

// Logout logs out of WAClient
// Deletes the JID from the DB, and logout connected devices.
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
	zap.S().Debugf("Sending message with ID: %s to: %s", messageId, to)

	go func() {
		// sleep between 1 and 10 seconds
		// ToDo: Implement Queue
		time.Sleep(time.Duration(rand.Intn(10-1)+1) * time.Second)

		_, err := mycli.WAClient.SendMessage(
			mycli.connection.ctx,
			types.NewJID(to, "s.whatsapp.net"),
			&waProto.Message{
				ExtendedTextMessage: &waProto.ExtendedTextMessage{
					Text: proto.String(message),
				},
			},
			whatsmeow.SendRequestExtra{
				ID:   messageId,
				Peer: false,
			})

		if err != nil {
			zap.S().Errorf("Error when sending text message: %s", err.Error())
			mycli.DB.Model(&models.Message{}).
				Where("message_id = ?", messageId).
				Update("failed", true).
				Update("failed_at", time.Now())
			return
		}

		mycli.DB.Model(&models.Message{}).
			Where("message_id = ?", messageId).
			Update("sent", true).
			Update("sent_at", time.Now())
	}()

	return nil
}

// SendPollMessage sends poll message to a given JID
func (mycli *MeowClient) SendPollMessage(messageId string, to string, dto models.PollDTO) error {
	zap.S().Debugf("Sending PollMessage with ID: %s to: %s", messageId, to)

	go func() {
		// sleep between 1 and 10 seconds
		// ToDo: Implement Queue
		time.Sleep(time.Duration(rand.Intn(10-1)+1) * time.Second)

		// pluck only the options from pollDTO.Details
		options := make([]string, len(dto.Details))
		for i, detail := range dto.Details {
			options[i] = detail.Option
		}

		// Send Poll
		pollMsg := mycli.WAClient.BuildPollCreation(
			dto.Question,
			options,
			1,
		)

		_, err := mycli.WAClient.SendMessage(
			mycli.connection.ctx,
			types.NewJID(to, "s.whatsapp.net"),
			pollMsg,
			whatsmeow.SendRequestExtra{
				ID:   messageId,
				Peer: false,
			})

		if err != nil {
			zap.S().Errorf("Error when sending poll message: %s", err.Error())
			mycli.DB.Model(&models.PollMessage{}).
				Where("message_id = ?", messageId).
				Update("failed", true).
				Update("failed_at", time.Now())
			return
		}

		mycli.DB.Model(&models.PollMessage{}).
			Where("message_id = ?", messageId).
			Update("sent", true).
			Update("sent_at", time.Now())
	}()

	return nil
}
