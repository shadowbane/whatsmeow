package controllers

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"go.mau.fi/whatsmeow"
	"go.uber.org/zap"
	"gomeow/pkg/application"
	"io"
	"log"
	"net/http"
)

type arrayOfMessage struct {
	Data []textMessageData `json:"data"`
}

type textMessageData struct {
	Phone   string `json:"phone"`
	Message string `json:"message"`
	Secret  bool   `json:"secret"`
	Retry   bool   `json:"retry"`
	IsGroup bool   `json:"isGroup"`
}

func MessageSend(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Fatal(err)
			}
		}(r.Body)

		var requestData arrayOfMessage
		err := json.NewDecoder(r.Body).Decode(&requestData)

		// debug requestData
		zap.S().Debugf("Request Data: %+v", requestData)
		if err != nil {
			writeErrorResponse(w, http.StatusUnprocessableEntity, "Unprocessable Entity")
			return
		}

		messageArr := requestData.Data[0]

		// remove first character if it is a '+' sign
		firstCharacter := messageArr.Phone[0:1]
		if firstCharacter == "+" {
			messageArr.Phone = messageArr.Phone[1:]
		}

		// dump request data
		zap.S().Debugf("Request Data: %+v", messageArr)

		if len(messageArr.Phone) == 0 || len(messageArr.Message) == 0 {
			writeErrorResponse(w, 422, "Invalid Parameter Supplied")
			return
		}

		// generate new messageID
		newMessageId := whatsmeow.GenerateMessageID()

		zap.S().Debugf("Queueing message with ID: %s and content: %s to %s", newMessageId, messageArr.Message, messageArr.Phone)

		pendingMessage := application.PendingMessage{
			To:        messageArr.Phone,
			Message:   messageArr.Message,
			MessageId: newMessageId,
		}

		// store to message store
		storeToMessageStore(
			app,
			app.Meow.DeviceStore.ID.String(),
			messageArr.Phone,
			messageArr.Message,
			newMessageId,
		)

		// add to queue
		app.Queue.Add(pendingMessage)

		formattedValues := returnData{
			Status:  true,
			Message: "Message queued",
			Data:    pendingMessage,
		}

		response, _ := json.Marshal(formattedValues)
		_, err = w.Write(response)
		if err != nil {
			zap.S().Errorf(err.Error())
			writeErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
		}
	}
}
