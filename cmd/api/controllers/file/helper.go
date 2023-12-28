package file

type ReturnMessageDTO struct {
	ID          string `json:"id"`
	MessageId   string `json:"message_id"`
	Destination string `json:"destination"`
	Sent        bool   `json:"sent"`
	Read        bool   `json:"read"`
	Failed      bool   `json:"failed"`
	ReadAt      string `json:"read_at"`
}

// stripPhoneNumber removes the first character of a phone number if it is a "+" or "0".
// ToDo: Refactor this
func stripPhoneNumber(phoneNumber string) string {
	firstCharacter := phoneNumber[0:1]
	if firstCharacter == "+" || firstCharacter == "0" {
		phoneNumber = phoneNumber[1:]
	}

	return phoneNumber
}