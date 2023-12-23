package contact

import (
	"github.com/julienschmidt/httprouter"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	"go.uber.org/zap"
	"gomeow/cmd/models"
	"gomeow/pkg/application"
	"gomeow/pkg/wmeow"
	"net/http"

	apiformattertrait "gomeow/cmd/api/controllers/traits"
)

func Index(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		destination := p.ByName("destination")
		device := r.Context().Value("device").(models.Device)

		// check if device is connected
		if wmeow.ClientPointer[device.ID] == nil {
			apiformattertrait.WriteErrorResponse(w, http.StatusBadRequest, "Device not connected")
			return
		}

		// check if device is logged in
		client := wmeow.ClientPointer[device.ID].WAClient
		if !client.IsConnected() || !client.IsLoggedIn() {
			apiformattertrait.WriteErrorResponse(w, http.StatusBadRequest, "Device is not logged in")
			return
		}

		// Sanitize data
		destination = sanitize(destination)

		// check if contact is on WhatsApp
		contact, err := client.IsOnWhatsApp([]string{destination})
		if err != nil {
			zap.S().Debugf("Error checking if contact is on WhatsApp: %+v", err)
			apiformattertrait.WriteErrorResponse(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		if contact[0].IsIn == false {
			apiformattertrait.WriteErrorResponse(w, http.StatusNotFound, "Contact is not on WhatsApp")
			return
		}

		// Get Device Information
		contactInfo, err := client.GetUserInfo([]types.JID{contact[0].JID})

		// Profile Picture
		picParams := &whatsmeow.GetProfilePictureParams{
			Preview:    false,
			ExistingID: "",
		}
		profPic, err := client.GetProfilePictureInfo(contact[0].JID, picParams)
		zap.S().Debugf("Picture Params: %+v", picParams)
		zap.S().Debugf("Profile Picture: %+v", profPic)
		apiformattertrait.WriteResponse(w, map[string]interface{}{
			"info":    contactInfo,
			"contact": contact,
			"profPic": profPic,
		})
	}
}
