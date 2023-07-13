package session

import (
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
	"gomeow/cmd/models"
	"gomeow/pkg/application"
	"gomeow/pkg/wmeow"
	"net/http"
	"strings"

	apiformattertrait "gomeow/cmd/api/controllers/traits"
)

func Connect(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		user := r.Context().Value("user").(models.User)
		eventstring := ""

		// if user already connected, return error
		if wmeow.ClientPointer[user.ID] != nil {
			apiformattertrait.WriteErrorResponse(w, http.StatusUnprocessableEntity, "User already connected")

			return
		}

		var subscribedEvents []string
		subscribedEvents = append(subscribedEvents, "All")
		eventstring = strings.Join(subscribedEvents, ",")

		user.Events = eventstring
		result := app.Models.Save(&user)
		if result.Error != nil {
			zap.S().Debugf("Error updating user: %+v", result)
			apiformattertrait.WriteErrorResponse(w, http.StatusInternalServerError, result.Error.Error())
			return
		}

		go func() {
			err := wmeow.StartClient(&user, app, user.JID, subscribedEvents)
			if err != nil {
				zap.S().Errorf("Error starting client: %+v", err)
			}
		}()

		apiformattertrait.WriteResponse(w, map[string]any{"id": user.ID, "state": "connected"})
	}
}
