package message

type TextMessage struct {
	Phone   string `json:"phone" validate:"required,numeric,min=5,max=20"`
	Message string `json:"message" validate:"required"`
}

// stripPhoneNumber removes the first character of a phone number if it is a "+" or "0".
func stripPhoneNumber(phoneNumber string) string {
	firstCharacter := phoneNumber[0:1]
	if firstCharacter == "+" || firstCharacter == "0" {
		phoneNumber = phoneNumber[1:]
	}

	return phoneNumber
}
