package models

import "time"

type Message struct {
	ID          int64     `json:"id" gorm:"auto_increment;primary_key"`
	JID         string    `json:"jid" gorm:"Column:jid;type:varchar(255);not null"`
	MessageId   string    `json:"message_id" gorm:"Column:message_id;type:varchar(255);not null;unique"`
	Destination string    `json:"destination" gorm:"not null"`
	Sent        bool      `json:"sent" gorm:"Default:false"`
	Read        bool      `json:"read" gorm:"Default:false"`
	Body        string    `json:"body"`
	CreatedAt   time.Time `json:"created_at" gorm:"type:timestamp"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"type:timestamp"`
}

func (m *Message) TableName() string {
	return "whatsmeow_messages"
}
