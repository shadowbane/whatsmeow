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
		user := r.Context().Value("user").(models.User)

		// check if user is connected
		if wmeow.ClientPointer[user.ID] == nil {
			apiformattertrait.WriteErrorResponse(w, http.StatusBadRequest, "User not connected")
			return
		}

		// check if user is logged in
		client := wmeow.ClientPointer[user.ID].WAClient
		if !client.IsConnected() || !client.IsLoggedIn() {
			apiformattertrait.WriteErrorResponse(w, http.StatusBadRequest, "User is not logged in")
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
			JID:         user.JID.String,
			UserId:      user.ID,
			MessageId:   newMessageId,
			Destination: request.Destination,
			Body:        request.Message,
		}

		// create user
		result := app.Models.Create(&message)
		if result.Error != nil {
			zap.S().Debugf("Error creating message: %+v", result)
			apiformattertrait.WriteErrorResponse(w, http.StatusInternalServerError, result.Error.Error())
			return
		}

		// generate JID from message.Desination
		//destinationJID, _ := wmeow.ParseJID(message.Destination)
		//
		//options := []string{
		//	0: "Yes",
		//	1: "No",
		//}
		//
		//// test
		//pollMsg := client.BuildPollCreation(
		//	"Test Polling",
		//	options,
		//	1,
		//)
		//pollData, pollErr := client.SendMessage(context.Background(), destinationJID, pollMsg)
		//
		//zap.S().Debugf("PollData: %+v", pollData)
		//if pollErr != nil {
		//	zap.S().Errorf("Error sending polling: %+v", pollErr)
		//}

		// send message
		err = wmeow.ClientPointer[user.ID].SendTextMessage(newMessageId, request.Destination, request.Message)

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
