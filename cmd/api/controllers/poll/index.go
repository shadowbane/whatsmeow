package poll

import (
	"github.com/julienschmidt/httprouter"
	"gomeow/cmd/models"
	"gomeow/pkg/application"
	"net/http"

	apiformattertrait "gomeow/cmd/api/controllers/traits"
)

func Index(app *application.Application) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		var polls []models.Poll
		var pollCollection []models.PollDTO

		result := app.Models.
			Where("user_id = ?", r.Context().Value("user").(models.User).ID).
			Preload("Details").
			Find(&polls)

		if result.Error != nil {
			apiformattertrait.WriteErrorResponse(w, http.StatusInternalServerError, result.Error.Error())
			return
		}

		for _, poll := range polls {
			dto := *poll.ToResponseDTO()
			//var pollDetails []models.PollDetail
			var detailDto []models.PollDetailDTO
			app.Models.
				Table("poll_details").
				Where("poll_id = ?", dto.ID).
				Order("`option` ASC").
				Find(&detailDto)
			dto.Details = detailDto
			pollCollection = append(pollCollection, dto)
		}

		apiformattertrait.WriteResponse(w, pollCollection)
	}
}
