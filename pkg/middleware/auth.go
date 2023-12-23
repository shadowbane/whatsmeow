package middleware

import (
	"context"
	"github.com/julienschmidt/httprouter"
	"gomeow/cmd/models"
	"net/http"

	apiformattertrait "gomeow/cmd/api/controllers/traits"
)

func (mw *MwStruct) Auth(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		username, token, ok := r.BasicAuth()
		if !ok {
			apiformattertrait.WriteErrorResponse(w, http.StatusUnauthorized, "Invalid credentials")

			return
		}

		var device models.Device
		result := mw.
			Models.
			Where("code = ? AND token = ?", username, token).
			First(&device)

		if result.RowsAffected == 0 {
			apiformattertrait.WriteErrorResponse(w, http.StatusUnauthorized, "Invalid credentials")

			return
		}

		// add device to context
		var ctx context.Context
		ctx = context.WithValue(r.Context(), "device", device)

		next(w, r.WithContext(ctx), p)
	}
}
