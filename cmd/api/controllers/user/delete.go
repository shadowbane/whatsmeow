package user

import (
	"github.com/julienschmidt/httprouter"
	"gomeow/pkg/application"
	"net/http"

	apiformattertrait "gomeow/cmd/api/controllers/traits"
)

func Delete(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Header().Set("Content-Type", "application/json")

		// find record in database
		user, err := findById(app, p.ByName("id"))
		if err != nil {
			apiformattertrait.WriteErrorResponse(w, http.StatusNotFound, "Not Found")

			return
		}

		app.Models.Delete(&user)

		apiformattertrait.WriteResponse(w, map[string]string{"message": "User deleted"})
	}
}
