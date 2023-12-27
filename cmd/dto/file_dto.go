package dto

import waProto "go.mau.fi/whatsmeow/binary/proto"

type FileDTO struct {
	Destination string `json:"destination" validate:"required,numeric,min=5,max=20"`
	File        string `json:"file" validate:"required"`
	FileName    string `json:"file_name" validate:"required"`
	ContextInfo waProto.ContextInfo
}
