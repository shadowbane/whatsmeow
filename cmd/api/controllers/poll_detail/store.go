package poll_detail

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

		var poll = models.Poll{
			ID:     p.ByName("pollId"),
			UserId: user.ID,
		}
		result := app.Models.First(&poll)

		// count poll detail, if empty, return error
		if result.RowsAffected == 0 {
			apiformattertrait.WriteErrorResponse(w, http.StatusNotFound, "Poll not found")
			return
		}

		// validating request - ensure fields are present
		validationError := app.Validator.Validate(request)
		if validationError != nil {
			zap.S().Debugf("Error validating request: %+v", validationError)
			apiformattertrait.WriteMultipleErrorResponse(w, http.StatusUnprocessableEntity, validationError)

			return
		}

		// create models.PollDetail from request
		pollDetail := models.PollDetail{
			PollId: poll.ID,
			Option: poll.Code + " - " + request.Option,
		}

		// create pollDetail
		createResult := app.Models.Create(&pollDetail)
		if createResult.Error != nil {
			zap.S().Debugf("Error creating poll detail: %+v", createResult)
			apiformattertrait.WriteErrorResponse(w, http.StatusInternalServerError, createResult.Error.Error())
			return
		}

		apiformattertrait.WriteResponse(w, pollDetail.ToResponseDTO())
	}
}
