package poll

import "gomeow/cmd/models"

type createResponseDTO struct {
	ID       string              `json:"id"`
	Code     string              `json:"code"`
	Question string              `json:"question"`
	Details  []models.PollDetail `json:"details"`
}
