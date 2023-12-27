package dto

import waProto "go.mau.fi/whatsmeow/binary/proto"

type ImageDTO struct {
	Destination string `json:"destination" validate:"required,numeric,min=5,max=20"`
	Message     string `json:"message" validate:"required"`
	Base64Image string `json:"image" validate:"required"`
	ViewOnce    bool   `json:"view_once" validate:""`
	ContextInfo waProto.ContextInfo
}
