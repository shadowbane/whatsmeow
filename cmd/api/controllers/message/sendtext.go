package message

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
	"gomeow/cmd/models"
	"gomeow/pkg/application"
	"gomeow/pkg/wmeow"
	"net/http"

	apiformattertrait "gomeow/cmd/api/controllers/traits"
)

func SendText(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		device := r.Context().Value("device").(models.Device)

		// check if device is connected
		if wmeow.ClientPointer[device.ID] == nil {
			apiformattertrait.WriteErrorResponse(w, http.StatusBadRequest, "Device not connected")
			return
		}

		// check if device is logged in
		client := wmeow.ClientPointer[device.ID].WAClient
		if !client.IsConnected() || !client.IsLoggedIn() {
			apiformattertrait.WriteErrorResponse(w, http.StatusBadRequest, "Device is not logged in")
			return
		}

		// create request instance
		var request TextMessage
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			apiformattertrait.WriteErrorResponse(w, http.StatusUnprocessableEntity, err.Error())

			return
		}

		// convert phone number to acceptable format
		request.Destination = stripPhoneNumber(request.Destination)

		// validating request
		validationError := app.Validator.Validate(request)
		if validationError != nil {
			zap.S().Debugf("Error validating request: %+v", validationError)
			apiformattertrait.WriteMultipleErrorResponse(w, http.StatusUnprocessableEntity, validationError)

			return
		}

		// Do everything else here

		// generate new messageID
		newMessageId := client.GenerateMessageID()

		// store
		message := models.Message{
			JID:         device.JID.String,
			DeviceId:    device.ID,
			MessageId:   newMessageId,
			Destination: request.Destination,
			Body:        request.Message,
		}

		// create message object
		result := app.Models.Create(&message)
		if result.Error != nil {
			zap.S().Debugf("Error creating message: %+v", result)
			apiformattertrait.WriteErrorResponse(w, http.StatusInternalServerError, result.Error.Error())
			return
		}

		// send message
		err = wmeow.ClientPointer[device.ID].SendTextMessage(newMessageId, request.Destination, request.Message)

		returnMessage := &ReturnMessageDTO{
			MessageId:   newMessageId,
			Destination: request.Destination,
			Message:     request.Message,
			Sent:        message.Sent,
			Read:        message.Read,
			Failed:      message.Failed,
		}

		returnMessage.ReadAt = ""

		if message.ReadAt.Valid {
			returnMessage.ReadAt = message.ReadAt.Time.String()
		}

		apiformattertrait.WriteResponse(w, returnMessage)
	}
}
