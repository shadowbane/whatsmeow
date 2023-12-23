package device

import (
	"github.com/julienschmidt/httprouter"
	"gomeow/cmd/models"
	"gomeow/pkg/application"
	"net/http"

	apiformattertrait "gomeow/cmd/api/controllers/traits"
)

func Index(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		var users []models.Device

		result := app.Models.Find(&users)

		if result.Error != nil {
			apiformattertrait.WriteErrorResponse(w, http.StatusInternalServerError, result.Error.Error())
			return
		}

		apiformattertrait.WriteResponse(w, users)
	}
}
