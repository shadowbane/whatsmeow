package message

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"go.mau.fi/whatsmeow"
	"go.uber.org/zap"
	"gomeow/cmd/models"
	"gomeow/pkg/application"
	"net/http"

	apiformattertrait "gomeow/cmd/api/controllers/traits"
)

func SendText(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		user := r.Context().Value("user").(models.User)

		var request TextMessage

		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			apiformattertrait.WriteErrorResponse(w, http.StatusUnprocessableEntity, err.Error())

			return
		}

		// convert phone number to acceptable format
		request.Phone = stripPhoneNumber(request.Phone)

		// validating request
		validationError := app.Validator.Validate(request)
		if validationError != nil {
			zap.S().Debugf("Error validating request: %+v", validationError)
			apiformattertrait.WriteMultipleErrorResponse(w, http.StatusUnprocessableEntity, validationError)

			return
		}

		// Do everything else here

		// generate new messageID
		newMessageId := whatsmeow.GenerateMessageID()

		// store
		message := models.Message{
			JID:         user.JID.String,
			UserId:      user.ID,
			MessageId:   newMessageId,
			Destination: request.Phone,
			Body:        request.Message,
		}

		// create user
		result := app.Models.Create(&message)
		if result.Error != nil {
			zap.S().Debugf("Error creating message: %+v", result)
			apiformattertrait.WriteErrorResponse(w, http.StatusInternalServerError, result.Error.Error())
			return
		}

		apiformattertrait.WriteResponse(w, message)
	}
}
