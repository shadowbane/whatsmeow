package models

import (
	"database/sql"
	"gomeow/pkg/helpers"
	"gorm.io/gorm"
	"time"
)

type PollMessage struct {
	ID           string       `json:"id" gorm:"type:char(26);primaryKey;autoIncrement:false"`
	JID          string       `json:"jid" gorm:"Column:jid;type:varchar(255);not null"`
	PollId       string       `json:"code" gorm:"Column:poll_id;type:char(26);not null"`
	UserId       int64        `json:"user_id" gorm:"Column:user_id;type:bigint;not null"`
	PollDetailId string       `json:"poll_detail_id" gorm:"Column:poll_detail_id;type:char(26)"`
	MessageId    string       `json:"message_id" gorm:"Column:message_id;type:varchar(255);not null;unique"`
	Destination  string       `json:"destination" gorm:"Column:destination;not null"`
	Sent         bool         `json:"sent" gorm:"Column:sent;Default:false"`
	Read         bool         `json:"read" gorm:"Column:read;Default:false"`
	Failed       bool         `json:"failed" gorm:"Column:failed;Default:false"`
	SentAt       sql.NullTime `json:"sent_at,omitempty" gorm:"Column:sent_at;type:timestamp;default:null"`
	ReadAt       sql.NullTime `json:"read_at,omitempty" gorm:"Column:read_at;type:timestamp;default:null"`
	FailedAt     sql.NullTime `json:"failed_at,omitempty" gorm:"Column:failed_at;type:timestamp;default:null"`
	AnsweredAt   sql.NullTime `json:"answered_at,omitempty" gorm:"Column:answered_at;type:timestamp;default:null"`
	CreatedAt    time.Time    `json:"created_at" gorm:"type:timestamp"`
	UpdatedAt    time.Time    `json:"updated_at" gorm:"type:timestamp"`

	// Associations
	Poll Poll `json:"poll,omitempty" gorm:"foreignKey:PollId;references:ID;constraint:OnDelete:RESTRICT"`
	User User `json:"user,omitempty" gorm:"foreignKey:UserId;references:ID;constraint:OnDelete:RESTRICT"`
}

func (p *PollMessage) TableName() string {
	return "whatsmeow_poll_messages"
}

// BeforeCreate will set a ULID using helper.NewULID() rather than numeric ID.
// It will check if the ID is already set and if so, it will skip
func (p *PollMessage) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		p.ID = helpers.NewULID()
	}

	return nil
}
