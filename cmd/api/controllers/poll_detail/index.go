package poll_detail

import (
	"github.com/julienschmidt/httprouter"
	"gomeow/cmd/models"
	"gomeow/pkg/application"
	"net/http"

	apiformattertrait "gomeow/cmd/api/controllers/traits"
)

func Index(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var pollDetails []models.PollDetailDTO
		device := r.Context().Value("device").(models.Device)

		var poll = models.Poll{
			ID:       p.ByName("pollId"),
			DeviceId: device.ID,
		}
		pollResult := app.Models.First(&poll)

		// count poll detail, if empty, return error
		if pollResult.RowsAffected == 0 {
			apiformattertrait.WriteErrorResponse(w, http.StatusNotFound, "Poll not found")
			return
		}

		result := app.Models.
			Where("poll_id = ?", poll.ID).
			Order("`option` ASC").
			Find(&[]models.PollDetail{}).
			Scan(&pollDetails)

		if result.Error != nil {
			apiformattertrait.WriteErrorResponse(w, http.StatusInternalServerError, result.Error.Error())
			return
		}

		apiformattertrait.WriteResponse(w, pollDetails)
	}
}
