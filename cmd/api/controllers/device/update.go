package device

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
	"gomeow/pkg/application"
	"net/http"

	apiformattertrait "gomeow/cmd/api/controllers/traits"
)

func Update(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var request UpdateValidator

		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			apiformattertrait.WriteErrorResponse(w, http.StatusUnprocessableEntity, err.Error())

			return
		}

		// ensure device is present
		user, err := findById(app, p.ByName("id"))
		if err != nil {
			apiformattertrait.WriteErrorResponse(w, http.StatusNotFound, "Device Not Found")

			return
		}

		// validating request - ensure fields are present
		validationError := app.Validator.Validate(request)
		if validationError != nil {
			zap.S().Debugf("Error validating request: %+v", validationError)
			apiformattertrait.WriteMultipleErrorResponse(w, http.StatusUnprocessableEntity, validationError)

			return
		}

		// only allow update of token
		user.Token = request.Token

		result := app.Models.Save(&user)
		if result.Error != nil {
			zap.S().Debugf("Error updating device: %+v", result)
			apiformattertrait.WriteErrorResponse(w, http.StatusInternalServerError, result.Error.Error())
			return
		}

		apiformattertrait.WriteResponse(w, user)
	}
}
