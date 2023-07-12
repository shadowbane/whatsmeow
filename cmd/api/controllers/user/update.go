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

func Update(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Header().Set("Content-Type", "application/json")
		//

		var request models.User
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			apiformattertrait.WriteErrorResponse(w, http.StatusUnprocessableEntity, err.Error())

			return
		}

		user, err := findById(app, p.ByName("id"))
		if err != nil {
			apiformattertrait.WriteErrorResponse(w, http.StatusNotFound, "Not Found")

			return
		}

		// only allow update of token
		user.Token = request.Token

		zap.S().Debugf("Updating user: %+v", user)

		// validating request
		success, err := govalidator.ValidateStruct(user)
		if !success {
			apiformattertrait.WriteErrorResponse(w, http.StatusUnprocessableEntity, err.Error())

			return
		}

		result := app.Models.Save(&user)
		if result.Error != nil {
			zap.S().Debugf("Error updating user: %+v", result)
			apiformattertrait.WriteErrorResponse(w, http.StatusInternalServerError, result.Error.Error())
			return
		}

		apiformattertrait.WriteResponse(w, user)
	}
}
