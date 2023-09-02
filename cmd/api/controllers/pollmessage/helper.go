package pollmessage

import "gomeow/cmd/models"

type TextMessage struct {
	Destination string `json:"destination" validate:"required,numeric,min=5,max=20"`
	Message     string `json:"message" validate:"required"`
}

type ReturnMessageDTO struct {
	ID          string         `json:"id"`
	MessageId   string         `json:"message_id"`
	Destination string         `json:"destination"`
	Poll        models.PollDTO `json:"poll"`
}

// stripPhoneNumber removes the first character of a phone number if it is a "+" or "0".
func stripPhoneNumber(phoneNumber string) string {
	firstCharacter := phoneNumber[0:1]
	if firstCharacter == "+" || firstCharacter == "0" {
		phoneNumber = phoneNumber[1:]
	}

	return phoneNumber
}
