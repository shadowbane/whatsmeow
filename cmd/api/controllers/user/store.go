package user

import (
	"encoding/json"
	"github.com/asaskevich/govalidator"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
	apiformattertrait "gomeow/cmd/api/controllers/traits"
	"gomeow/cmd/models"
	"gomeow/pkg/application"
	"net/http"
)

func Store(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var request models.User

		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			apiformattertrait.WriteErrorResponse(w, http.StatusUnprocessableEntity, err.Error())

			return
		}

		// validating request
		success, err := govalidator.ValidateStruct(request)
		if !success {
			apiformattertrait.WriteErrorResponse(w, http.StatusUnprocessableEntity, err.Error())

			return
		}

		result := app.Models.Create(&request)
		if result.Error != nil {
			zap.S().Debugf("Error creating user: %+v", result)
			apiformattertrait.WriteErrorResponse(w, http.StatusInternalServerError, result.Error.Error())
			return
		}

		apiformattertrait.WriteResponse(w, request)
	}
}
