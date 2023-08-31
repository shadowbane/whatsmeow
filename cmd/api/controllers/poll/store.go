package poll

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
	apiformattertrait "gomeow/cmd/api/controllers/traits"
	"gomeow/cmd/models"
	"gomeow/pkg/application"
	"net/http"
)

func Store(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var request CreateValidator
		user := r.Context().Value("user").(models.User)

		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			apiformattertrait.WriteErrorResponse(w, http.StatusUnprocessableEntity, err.Error())

			return
		}

		// validating request - ensure fields are present
		validationError := app.Validator.Validate(request)
		if validationError != nil {
			zap.S().Debugf("Error validating request: %+v", validationError)
			apiformattertrait.WriteMultipleErrorResponse(w, http.StatusUnprocessableEntity, validationError)

			return
		}

		// create models.User from request
		poll := models.Poll{
			UserId:   user.ID,
			Question: request.Question,
		}

		// create poll
		result := app.Models.Create(&poll)
		if result.Error != nil {
			zap.S().Debugf("Error creating poll: %+v", result)
			apiformattertrait.WriteErrorResponse(w, http.StatusInternalServerError, result.Error.Error())
			return
		}

		apiformattertrait.WriteResponse(w, poll.ToResponseDTO())
	}
}
