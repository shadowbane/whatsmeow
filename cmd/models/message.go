package models

import (
	"database/sql"
	"gomeow/pkg/helpers"
	"gorm.io/gorm"
	"time"
)

type Message struct {
	ID           string         `json:"id" gorm:"type:char(26);primaryKey;autoIncrement:false"`
	JID          string         `json:"jid" gorm:"Column:jid;type:varchar(255);not null"`
	DeviceId     int64          `json:"device_id" gorm:"Column:device_id;type:bigint;not null"`
	MessageId    string         `json:"message_id" gorm:"Column:message_id;type:varchar(255);not null;unique"`
	Destination  string         `json:"destination" gorm:"Column:destination;not null"`
	Sent         bool           `json:"sent" gorm:"Column:sent;Default:false"`
	Read         bool           `json:"read" gorm:"Column:read;Default:false"`
	Failed       bool           `json:"failed" gorm:"Column:failed;Default:false"`
	Subject      sql.NullString `json:"subject,omitempty" gorm:"Column:subject;type:varchar(255);default:null"`
	Body         string         `json:"body" gorm:"Column:body;type:text;default:null"`
	File         string         `json:"file;omitempty" gorm:"Column:file;type:text;default:null;comment:For image/file/video message"`
	FileName     string         `json:"file_name;omitempty" gorm:"Column:file_name;type:varchar(255);default:null;comment:File name with extension"`
	PollId       string         `json:"poll_id;omitempty" gorm:"Column:poll_id;type:char(26);default:null;comment:For Poll message"`
	PollDetailId string         `json:"poll_detail_id;omitempty" gorm:"Column:poll_detail_id;type:char(26);default:null;comment:For Poll message"`
	MessageType  string         `json:"message_type;omitempty" gorm:"Column:message_type;type:char(15);comment:Message Types"`

	// Timestamps
	SentAt     sql.NullTime `json:"sent_at,omitempty" gorm:"Column:sent_at;type:timestamp;default:null"`
	ReadAt     sql.NullTime `json:"read_at,omitempty" gorm:"Column:read_at;type:timestamp;default:null"`
	FailedAt   sql.NullTime `json:"failed_at,omitempty" gorm:"Column:failed_at;type:timestamp;default:null"`
	AnsweredAt sql.NullTime `json:"answered_at,omitempty" gorm:"Column:answered_at;type:timestamp;default:null;comment:Poll answer timestamp"`
	CreatedAt  time.Time    `json:"created_at" gorm:"type:timestamp"`
	UpdatedAt  time.Time    `json:"updated_at" gorm:"type:timestamp"`

	// Associations
	Poll   Poll   `json:"poll,omitempty" gorm:"foreignKey:PollId;references:ID;constraint:OnDelete:RESTRICT"`
	Device Device `json:"device,omitempty" gorm:"foreignKey:DeviceId;references:ID;constraint:OnDelete:RESTRICT"`
}

func (m *Message) TableName() string {
	return "whatsmeow_messages"
}

// BeforeCreate will set a ULID using helpers.NewULID() rather than numeric ID.
// It will check if the ID is already set and if so, it will skip
func (m *Message) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = helpers.NewULID()
	}

	return nil
}
