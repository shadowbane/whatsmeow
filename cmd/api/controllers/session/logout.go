package session

import (
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
	"gomeow/cmd/models"
	"gomeow/pkg/application"
	"gomeow/pkg/wmeow"
	"net/http"

	apiformattertrait "gomeow/cmd/api/controllers/traits"
)

func Logout(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		user := r.Context().Value("user").(models.User)

		if wmeow.ClientPointer[user.ID] == nil {
			apiformattertrait.WriteErrorResponse(w, http.StatusBadRequest, "User not connected")

			return
		}

		client := wmeow.ClientPointer[user.ID].WAClient
		if client.IsConnected() && client.IsLoggedIn() {
			err := client.Logout()
			if err != nil {
				zap.S().Errorf("Error logging out: %+v", err)
				apiformattertrait.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())

				return
			}

			wmeow.ClientPointer[user.ID].Logout()
			zap.S().Infof("User %s with JID %s Logged Out", user.Name, user.JID.String)
		} else {
			if client.IsConnected() {
				zap.S().Infof("User %s is not logged in. Doing logout anyway", user.Name)
				wmeow.KillChannel[user.ID] <- true
			} else {
				apiformattertrait.WriteErrorResponse(w, http.StatusBadRequest, "User not connected")

				return
			}
		}

		apiformattertrait.WriteResponse(w, map[string]any{"state": "logged out"})
	}
}
