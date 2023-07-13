package session

import (
	"github.com/julienschmidt/httprouter"
	"gomeow/cmd/models"
	"gomeow/pkg/application"
	"gomeow/pkg/wmeow"
	"net/http"

	apiformattertrait "gomeow/cmd/api/controllers/traits"
)

func Disconnect(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		user := r.Context().Value("user").(models.User)

		if wmeow.ClientPointer[user.ID] == nil {
			apiformattertrait.WriteErrorResponse(w, http.StatusBadRequest, "User not connected")

			return
		}

		wmeow.KillChannel[user.ID] <- true

		apiformattertrait.WriteResponse(w, map[string]any{"state": "disconnected"})
	}
}
