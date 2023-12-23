package contact

import "strings"

type Contact struct {
	Phone        string `json:"phone"`
	IsOnWhatsApp bool   `json:"is_on_whatsapp"`
	VerifiedName string `json:"verified_name"`
}

func sanitize(phone string) string {
	if phone[0] == '0' {
		phone = phone[1:]
	}

	phone = strings.Split(phone, "@")[0]
	phone = strings.Split(phone, ".")[0]
	phone = strings.Split(phone, ":")[0]

	return phone
}
