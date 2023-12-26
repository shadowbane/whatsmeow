package pollmessage

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
	"gomeow/cmd/models"
	"gomeow/pkg/application"
	"gomeow/pkg/wmeow"
	"net/http"
	"sort"

	apiformattertrait "gomeow/cmd/api/controllers/traits"
)

func SendPoll(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		device := r.Context().Value("device").(models.Device)

		//check if device is connected
		if wmeow.ClientPointer[device.ID] == nil {
			apiformattertrait.WriteErrorResponse(w, http.StatusBadRequest, "Device not connected")
			return
		}

		//check if device is logged in
		client := wmeow.ClientPointer[device.ID].WAClient
		if !client.IsConnected() || !client.IsLoggedIn() {
			apiformattertrait.WriteErrorResponse(w, http.StatusBadRequest, "Device is not logged in")
			return
		}

		// create request instance
		var request CreateValidator
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

		// Load Poll
		poll := models.Poll{
			ID:       request.PollId,
			DeviceId: device.ID,
		}
		//var pollDTO models.PollDTO
		pollDTO := models.PollDTO{
			Details: make([]models.PollDetailDTO, 0),
		}
		result := app.Models.
			Preload("Details").
			First(&poll).
			Scan(&pollDTO)

		// convert []PollDetail to []PollDetailDTO
		for _, detail := range poll.Details {
			pollDTO.Details = append(pollDTO.Details, *detail.ToResponseDTO())
		}

		// sort pollDTO.Details ascending by Option
		sort.Slice(pollDTO.Details, func(i, j int) bool {
			return pollDTO.Details[i].Option < pollDTO.Details[j].Option
		})

		if result.RowsAffected == 0 {
			zap.S().Errorf("Poll not found: %+v", result)
			apiformattertrait.WriteErrorResponse(w, http.StatusNotFound, "Poll not found")
			return
		}

		// Do everything else here

		// generate new messageID
		newMessageId := client.GenerateMessageID()

		// store
		//message := models.PollMessage{
		//	PollId:      request.PollId,
		//	JID:         device.JID.String,
		//	DeviceId:    device.ID,
		//	MessageId:   newMessageId,
		//	Destination: request.Destination,
		//}
		message := models.Message{
			PollId:      request.PollId,
			JID:         device.JID.String,
			DeviceId:    device.ID,
			MessageId:   newMessageId,
			Destination: request.Destination,
			MessageType: "poll",
		}

		// create poll message
		createResult := app.Models.Create(&message)
		if createResult.Error != nil {
			zap.S().Debugf("Error creating message: %+v", createResult)
			apiformattertrait.WriteErrorResponse(w, http.StatusInternalServerError, createResult.Error.Error())
			return
		}

		// send message
		err = wmeow.ClientPointer[device.ID].SendPollMessage(newMessageId, request.Destination, pollDTO)

		apiformattertrait.WriteResponse(w, &ReturnMessageDTO{
			ID:          message.ID,
			MessageId:   message.MessageId,
			Destination: message.Destination,
			Poll:        pollDTO,
		})
	}
}
