package user

import (
	"github.com/julienschmidt/httprouter"
	"gomeow/cmd/models"
	"gomeow/pkg/application"
	"net/http"

	apiformattertrait "gomeow/cmd/api/controllers/traits"
)

func Show(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		// find record in database
		user, err := findById(app, p.ByName("id"))
		if err != nil {
			apiformattertrait.WriteErrorResponse(w, http.StatusNotFound, "Not Found")

			return
		}

		apiformattertrait.WriteResponse(w, user)
	}
}

func findById(app *application.Application, id string) (*models.User, error) {
	var user models.User
	if err := app.Models.
		Where("id = ?", id).
		First(&user).
		Error; err != nil {
		return &user, err
	}

	return &user, nil
}
