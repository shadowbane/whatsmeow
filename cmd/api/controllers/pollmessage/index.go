package pollmessage

import (
	"github.com/julienschmidt/httprouter"
	"gomeow/cmd/models"
	"gomeow/pkg/application"
	"net/http"

	apiformattertrait "gomeow/cmd/api/controllers/traits"
)

func Index(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		device := r.Context().Value("device").(models.Device)

		apiformattertrait.WriteResponse(w, device)
	}
}
