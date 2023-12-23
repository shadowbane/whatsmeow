package session

import (
	"github.com/julienschmidt/httprouter"
	"gomeow/cmd/models"
	"gomeow/pkg/application"
	"gomeow/pkg/wmeow"
	"net/http"

	apiformattertrait "gomeow/cmd/api/controllers/traits"
)

// Disconnect device from server.
// This is like closing your WhatsApp web, and user don't need to relog
// (scan qr code) when reopening / reconnecting.
func Disconnect(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		device := r.Context().Value("device").(models.Device)

		if wmeow.ClientPointer[device.ID] == nil {
			apiformattertrait.WriteErrorResponse(w, http.StatusBadRequest, "Device not connected")

			return
		}

		wmeow.KillChannel[device.ID] <- true

		apiformattertrait.WriteResponse(w, map[string]any{"state": "disconnected"})
	}
}
