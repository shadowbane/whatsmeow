package user

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
		user := models.User{
			Name:  request.Name,
			Token: request.Token,
		}

		// create user
		result := app.Models.Create(&user)
		if result.Error != nil {
			zap.S().Debugf("Error creating user: %+v", result)
			apiformattertrait.WriteErrorResponse(w, http.StatusInternalServerError, result.Error.Error())
			return
		}

		apiformattertrait.WriteResponse(w, request)
	}
}
