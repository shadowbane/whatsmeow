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
		user := r.Context().Value("user").(models.User)

		//check if user is connected
		if wmeow.ClientPointer[user.ID] == nil {
			apiformattertrait.WriteErrorResponse(w, http.StatusBadRequest, "User not connected")
			return
		}

		//check if user is logged in
		client := wmeow.ClientPointer[user.ID].WAClient
		if !client.IsConnected() || !client.IsLoggedIn() {
			apiformattertrait.WriteErrorResponse(w, http.StatusBadRequest, "User is not logged in")
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
			ID:     request.PollId,
			UserId: user.ID,
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
			zap.S().Debugf("Poll not found: %+v", result)
			apiformattertrait.WriteErrorResponse(w, http.StatusNotFound, "Poll not found")
			return
		}

		// Do everything else here

		// generate new messageID
		newMessageId := client.GenerateMessageID()

		// store
		message := models.PollMessage{
			PollId:      request.PollId,
			JID:         user.JID.String,
			UserId:      user.ID,
			MessageId:   newMessageId,
			Destination: request.Destination,
		}

		// create poll message
		createResult := app.Models.Create(&message)
		if createResult.Error != nil {
			zap.S().Debugf("Error creating message: %+v", createResult)
			apiformattertrait.WriteErrorResponse(w, http.StatusInternalServerError, createResult.Error.Error())
			return
		}

		// send message
		err = wmeow.ClientPointer[user.ID].SendPollMessage(newMessageId, request.Destination, pollDTO)

		apiformattertrait.WriteResponse(w, message)
	}
}
