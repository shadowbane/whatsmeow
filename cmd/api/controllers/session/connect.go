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
		device := r.Context().Value("device").(models.Device)
		eventstring := ""

		// if device already connected, return error
		if wmeow.ClientPointer[device.ID] != nil {
			apiformattertrait.WriteErrorResponse(w, http.StatusUnprocessableEntity, "Device already connected")

			return
		}

		var subscribedEvents []string
		subscribedEvents = append(subscribedEvents, "All")
		eventstring = strings.Join(subscribedEvents, ",")

		device.Events = eventstring
		result := app.Models.Save(&device)
		if result.Error != nil {
			zap.S().Debugf("Error updating device: %+v", result)
			apiformattertrait.WriteErrorResponse(w, http.StatusInternalServerError, result.Error.Error())
			return
		}

		go func() {
			err := wmeow.StartClient(&device, app, device.JID.String, subscribedEvents)
			if err != nil {
				zap.S().Errorf("Error starting client: %+v", err)
			}
		}()

		apiformattertrait.WriteResponse(w, map[string]any{"id": device.ID, "state": "connected"})
	}
}
