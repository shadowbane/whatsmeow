package wmeow

import (
	"database/sql"
	"fmt"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"go.uber.org/zap"
	"gomeow/cmd/models"

	waProto "go.mau.fi/whatsmeow/binary/proto"
)

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
			zap.S().Debug("Marked self as available")
		}

		mycli.Device.IsConnected = true
		result := mycli.DB.Save(&mycli.Device)
		if result.Error != nil {
			zap.S().Errorf("WMEOW\tError updating device: %+v", result)
		}
	/**
	 * Do not handle disconnected event.
	 */
	//case *events.Disconnected:
	//zap.S().Debugf("WMEOW\tReceiving disconnect request")
	//mycli.Device.IsConnected = false
	//result := mycli.DB.Save(&mycli.Device)
	//if result.Error != nil {
	//	zap.S().Errorf("WMEOW\tError updating device: %+v", result)
	//}
	case *events.PairSuccess:
		zap.S().Infof("WMEOW\tPairing success for %d. JID: %s", mycli.Device.ID, evt.ID.String())
		mycli.Device.JID = sql.NullString{
			String: evt.ID.String(),
			Valid:  true,
		}
		mycli.Device.IsConnected = true

		result := mycli.DB.Save(&mycli.Device)
		if result.Error != nil {
			zap.S().Errorf("WMEOW\tError updating device: %+v", result)
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

			mycli.PollVote(pollVote, evt)
		} else {
			zap.S().Debugf("WMEOW\tReceived message form %s: %s", evt.Info.Sender, evt.Message.GetConversation())
		}

	case *events.Receipt:
		if evt.Type == events.ReceiptTypeRead {
			zap.S().Debugf("WMEOW\tReceived a read receipt [%s]", evt.MessageIDs)

			// Mark message as read
			go func() {
				for _, messageId := range evt.MessageIDs {
					mycli.DB.Model(&models.Message{}).
						Where("message_id = ?", messageId).
						Update("read", true).
						Update("read_at", evt.Timestamp)
				}
			}()

			// Mark poll message as read
			go func() {
				for _, messageId := range evt.MessageIDs {
					mycli.DB.Model(&models.Message{}).
						Where("message_id = ?", messageId).
						Update("read", true).
						Update("read_at", evt.Timestamp)
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
		zap.S().Debugf("WMEOW\tGot CallOffer event - %+v", evt)
	case *events.CallAccept:
		zap.S().Debugf("WMEOW\tGot CallAccept event - %+v", evt)
	case *events.CallTerminate:
		zap.S().Debugf("WMEOW\tGot CallTerminate event - %+v", evt)
	case *events.CallOfferNotice:
		zap.S().Debugf("WMEOW\tGot CallOfferNotice event - %+v", evt)
	case *events.CallRelayLatency:
		zap.S().Debugf("WMEOW\tGot CallRelayLatency event - %+v", evt)
	default:
		zap.S().Debugf("WMEOW\tUnhandled event: %+v", evt)

	}
}

func (mycli *MeowClient) PollVote(pollVote *waProto.PollVoteMessage, evt *events.Message) {
	zap.S().Debugf("WMEOW\tPoll vote received from %s", evt.Info.Sender)

	var optionVal string
	for _, hash := range pollVote.GetSelectedOptions() {
		optionVal = fmt.Sprintf("%x", hash)
	}
	//sender := StripJID(evt.Info.Sender)

	pollData := evt.Message.GetPollUpdateMessage()
	// Get message ID from context info
	pollId := pollData.PollCreationMessageKey.GetId()

	// Get PollMessage from DB
	var pollMessage models.Message
	result := mycli.DB.
		Preload("Poll").
		Where("message_id = ?", pollId).
		First(&pollMessage)
	if result.Error != nil {
		zap.S().Errorf("WMEOW\tError getting poll message: %+v", result.Error)
		return
	}

	// Find PollDetail from DB
	var pollDetail models.PollDetail
	result = mycli.DB.
		Where("poll_id = ? AND option_sha256 = ?", pollMessage.Poll.ID, optionVal).
		First(&pollDetail)

	tx := mycli.DB.Begin()

	// Store Poll Vote
	pollMessage.AnsweredAt = sql.NullTime{
		Time:  evt.Info.Timestamp,
		Valid: true,
	}
	pollMessage.PollDetailId = pollDetail.ID
	result = tx.Save(&pollMessage)
	if result.Error != nil {
		tx.Rollback()
		zap.S().Errorf("WMEOW\tError saving poll message: %+v", result.Error)
		return
	}

	// Store Poll History
	result = tx.Save(&models.PollHistory{
		PollId:       pollMessage.PollId,
		DeviceId:     pollMessage.DeviceId,
		PollDetailId: pollMessage.PollDetailId,
		MessageId:    pollMessage.MessageId,
		Destination:  pollMessage.Destination,
		AnsweredAt:   pollMessage.AnsweredAt,
	})
	if result.Error != nil {
		tx.Rollback()
		zap.S().Errorf("WMEOW\tError saving poll history: %+v", result.Error)
		return
	}

	// commit transaction
	tx.Commit()
	return
}
