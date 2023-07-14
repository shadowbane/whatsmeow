package models

import (
	"database/sql"
	"gomeow/pkg/helpers"
	"gorm.io/gorm"
	"time"
)

type Message struct {
	ID          string         `json:"id" gorm:"type:char(26);primaryKey;autoIncrement:false"`
	JID         string         `json:"jid" gorm:"Column:jid;type:varchar(255);not null"`
	UserId      int64          `json:"user_id" gorm:"Column:user_id;type:bigint;not null"`
	MessageId   string         `json:"message_id" gorm:"Column:message_id;type:varchar(255);not null;unique"`
	Destination string         `json:"destination" gorm:"Column:destination;not null"`
	Sent        bool           `json:"sent" gorm:"Column:sent,Default:false"`
	Read        bool           `json:"read" gorm:"Column:read,Default:false"`
	Subject     sql.NullString `json:"subject" gorm:"Column:subject;type:varchar(255);default:null"`
	Body        string         `json:"body" gorm:"Column:body;type:text;"`
	SentAt      sql.NullTime   `json:"sent_at" gorm:"Column:sent_at,type:timestamp;default:null"`
	ReadAt      sql.NullTime   `json:"read_at" gorm:"Column:read_at,type:timestamp;default:null"`
	CreatedAt   time.Time      `json:"created_at" gorm:"type:timestamp"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"type:timestamp"`
}

func (m *Message) TableName() string {
	return "whatsmeow_messages"
}

// BeforeCreate will set a ULID using helper.NewULID() rather than numeric ID.
// It will check if the ID is already set and if so, it will skip
func (m *Message) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = helpers.NewULID()
	}

	return nil
}
