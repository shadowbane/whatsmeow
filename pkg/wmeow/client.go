package wmeow

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/vincent-petithory/dataurl"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
	"go.uber.org/zap"
	"gomeow/cmd/dto"
	"gomeow/cmd/models"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
	"math/rand"
	"mime"
	"net/http"
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
	Device         *models.Device
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
	mycli.Device.IsConnected = false
	mycli.Device.JID = sql.NullString{String: ""}
	mycli.Device.Webhook = sql.NullString{String: ""}
	result := mycli.DB.Save(&mycli.Device)
	if result.Error != nil {
		zap.S().Errorf("WMEOW\tError updating device: %+v", result)
	}
	KillChannel[mycli.Device.ID] <- true
	zap.S().Infof("WMEOW\tLogged out of WAClient.")
}

// ConnectAndLogin connects to the WAClient and logs in if necessary.
// This will also shows the QR code if the device is not logged in
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

					mycli.Device.QRCode = sql.NullString{
						String: evt.Code,
						Valid:  true,
					}

					result := mycli.DB.Save(&mycli.Device)
					if result.Error != nil {
						zap.S().Errorf("WMEOW\tError updating mycli.Device: %+v", result)

						return result.Error, true
					}
				} else if evt.Event == "timeout" {
					// Clear QR code from DB on timeout
					mycli.Device.QRCode = sql.NullString{
						String: "",
					}

					result := mycli.DB.Save(&mycli.Device)
					if result.Error != nil {
						zap.S().Errorf("WMEOW\tError updating mycli.Device: %+v", result)

						return result.Error, true
					}

					zap.S().Errorf("WMEOW\tQR Code Timeout... Killing channel...")

					delete(ClientPointer, mycli.Device.ID)
					KillChannel[mycli.Device.ID] <- true
				} else if evt.Event == "success" {
					zap.S().Debugf("WMEOW\tQR pairing ok!")
					// Clear QR code after pairing
					mycli.Device.QRCode = sql.NullString{String: ""}

					result := mycli.DB.Save(&mycli.Device)
					if result.Error != nil {
						zap.S().Errorf("WMEOW\tError updating device: %+v", result)

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
		mycli.markTyping(to)

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
		mycli.markTyping(to)

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

// SendImageMessage send image message to a given JID
func (mycli *MeowClient) SendImageMessage(messageId string, dto *dto.ImageDTO) error {
	zap.S().Debugf("Sending ImageMessage with ID: %s to: %s", messageId, dto.Destination)

	mycli.markTyping(dto.Destination)

	var uploaded whatsmeow.UploadResponse
	var fileData []byte

	// upload image
	dataURL, err := dataurl.DecodeString(dto.Base64Image)
	if err != nil {
		zap.S().Debug("Could not decode base64 payload!")
		return errors.New("could not decode base64 payload")
	}

	fileData = dataURL.Data
	uploaded, err = mycli.WAClient.Upload(mycli.connection.ctx, fileData, whatsmeow.MediaImage)
	if err != nil {
		zap.S().Debug("Failed to upload image!")
		return errors.New("could not upload image to WhatsApp server")
	}

	msg := &waProto.Message{
		ImageMessage: &waProto.ImageMessage{
			Caption:       proto.String(dto.Message),
			Url:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String(http.DetectContentType(fileData)),
			FileEncSha256: uploaded.FileEncSHA256,
			FileSha256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(fileData))),
		},
	}

	if dto.ContextInfo.StanzaId != nil {
		msg.ExtendedTextMessage.ContextInfo = &waProto.ContextInfo{
			StanzaId:      proto.String(*dto.ContextInfo.StanzaId),
			Participant:   proto.String(*dto.ContextInfo.Participant),
			QuotedMessage: &waProto.Message{Conversation: proto.String("")},
		}
	}

	_, err = mycli.WAClient.SendMessage(
		mycli.connection.ctx,
		types.NewJID(dto.Destination, types.DefaultUserServer),
		msg,
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
		return fmt.Errorf("error when sending text message: %s", err.Error())
	}

	mycli.DB.Model(&models.Message{}).
		Where("message_id = ?", messageId).
		Update("sent", true).
		Update("sent_at", time.Now())

	return nil
}

// SendFileMessage send image message to a given JID
func (mycli *MeowClient) SendFileMessage(messageId string, dto *dto.FileDTO) error {
	zap.S().Debugf("Sending FileMessage with ID: %s to: %s", messageId, dto.Destination)

	mycli.markTyping(dto.Destination)

	var uploaded whatsmeow.UploadResponse
	var fileData []byte

	// upload image
	dataURL, err := dataurl.DecodeString(dto.File)
	if err != nil {
		zap.S().Debug("Could not decode base64 payload!")
		return errors.New("could not decode base64 payload")
	}

	fileData = dataURL.Data
	uploaded, err = mycli.WAClient.Upload(mycli.connection.ctx, fileData, whatsmeow.MediaDocument)
	if err != nil {
		zap.S().Debug("Failed to upload file!")
		return errors.New("could not upload file to WhatsApp server")
	}

	extension, err := mime.ExtensionsByType(http.DetectContentType(fileData))
	if err != nil {
		zap.S().Errorf("No extension detected: %s", err.Error())
		return errors.New("no extension detected")
	}
	fileName := dto.FileName + extension[len(extension)-1]

	msg := &waProto.Message{
		DocumentMessage: &waProto.DocumentMessage{
			FileName:      &fileName,
			Url:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String(http.DetectContentType(fileData)),
			FileEncSha256: uploaded.FileEncSHA256,
			FileSha256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(fileData))),
		},
	}

	if dto.ContextInfo.StanzaId != nil {
		msg.ExtendedTextMessage.ContextInfo = &waProto.ContextInfo{
			StanzaId:      proto.String(*dto.ContextInfo.StanzaId),
			Participant:   proto.String(*dto.ContextInfo.Participant),
			QuotedMessage: &waProto.Message{Conversation: proto.String("")},
		}
	}

	_, err = mycli.WAClient.SendMessage(
		mycli.connection.ctx,
		types.NewJID(dto.Destination, types.DefaultUserServer),
		msg,
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
		return fmt.Errorf("error when sending text message: %s", err.Error())
	}

	mycli.DB.Model(&models.Message{}).
		Where("message_id = ?", messageId).
		Update("sent", true).
		Update("sent_at", time.Now()).
		Update("file_name", &fileName)

	return nil
}

// markTyping mark myself as typing, and add random delay before sending message
func (mycli *MeowClient) markTyping(to string) {
	_ = mycli.WAClient.SendChatPresence(
		types.NewJID(to, "s.whatsapp.net"),
		types.ChatPresenceComposing,
		types.ChatPresenceMediaText,
	)

	time.Sleep(time.Duration(rand.Intn(10-1)+1) * time.Second)
}
