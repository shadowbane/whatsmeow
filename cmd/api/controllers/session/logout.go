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
		device := r.Context().Value("device").(models.Device)

		if wmeow.ClientPointer[device.ID] == nil {
			apiformattertrait.WriteErrorResponse(w, http.StatusBadRequest, "Device not connected")

			return
		}

		client := wmeow.ClientPointer[device.ID].WAClient
		if client.IsConnected() && client.IsLoggedIn() {
			err := client.Logout()
			if err != nil {
				zap.S().Errorf("Error logging out: %+v", err)
				apiformattertrait.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())

				return
			}

			wmeow.ClientPointer[device.ID].Logout()
			zap.S().Infof("Device %s with JID %s Logged Out", device.Name, device.JID.String)
		} else {
			if client.IsConnected() {
				zap.S().Infof("Device %s is not logged in. Doing logout anyway", device.Name)
				wmeow.KillChannel[device.ID] <- true
			} else {
				apiformattertrait.WriteErrorResponse(w, http.StatusBadRequest, "Device not connected")

				return
			}
		}

		apiformattertrait.WriteResponse(w, map[string]any{"state": "logged out"})
	}
}
