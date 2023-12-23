package session

import (
	"github.com/julienschmidt/httprouter"
	"github.com/skip2/go-qrcode"
	"gomeow/cmd/models"
	"gomeow/pkg/application"
	"gomeow/pkg/wmeow"
	"net/http"
	"strconv"

	apiformattertrait "gomeow/cmd/api/controllers/traits"
)

func GetQRCode(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		device := r.Context().Value("device").(models.Device)

		printQr, _ := strconv.ParseBool(r.URL.Query().Get("print_qr"))

		if wmeow.ClientPointer[device.ID] == nil {
			apiformattertrait.WriteErrorResponse(w, http.StatusBadRequest, "Device not connected")

			return
		}

		client := wmeow.ClientPointer[device.ID].WAClient
		if client.IsConnected() && client.IsLoggedIn() {
			apiformattertrait.WriteErrorResponse(w, http.StatusBadRequest, "Device already logged in")

			return
		}

		if device.QRCode.String == "" {
			apiformattertrait.WriteErrorResponse(w, http.StatusNotFound, "No QR Code found. Please wait a few seconds and try again")

			return
		}

		if !printQr {
			apiformattertrait.WriteResponse(w, map[string]string{"qrcode": device.QRCode.String})

			return
		} else {
			w.Header().Set("Content-Type", "image/png")
			w.Header().Set("Content-Disposition", "attachment; filename=\""+device.Name+".png\"")

			var png []byte
			png, err := qrcode.Encode(device.QRCode.String, qrcode.Medium, 256)
			if err != nil {
				apiformattertrait.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())

				return
			}

			_, err = w.Write(png)
			if err != nil {
				apiformattertrait.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())

				return
			}
		}
	}
}
