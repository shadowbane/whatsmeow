package models

import (
	"database/sql"
	"gomeow/pkg/helpers"
	"gorm.io/gorm"
	"time"
)

type PollHistory struct {
	ID           string       `json:"id" gorm:"type:char(26);primaryKey;autoIncrement:false"`
	PollId       string       `json:"code" gorm:"Column:poll_id;type:char(26);not null"`
	DeviceId     int64        `json:"device_id" gorm:"Column:device_id;type:bigint;not null"`
	PollDetailId string       `json:"poll_detail_id" gorm:"Column:poll_detail_id;type:char(26)"`
	MessageId    string       `json:"message_id" gorm:"Column:message_id;type:varchar(255);not null"`
	Destination  string       `json:"destination" gorm:"Column:destination;not null"`
	AnsweredAt   sql.NullTime `json:"answered_at,omitempty" gorm:"Column:answered_at;type:timestamp;default:null"`
	CreatedAt    time.Time    `json:"created_at" gorm:"type:timestamp"`
	UpdatedAt    time.Time    `json:"updated_at" gorm:"type:timestamp"`

	// Associations
	Poll Poll `gorm:"foreignKey:PollId;references:ID;constraint:OnDelete:RESTRICT"`
}

func (p *PollHistory) TableName() string {
	return "whatsmeow_poll_histories"
}

// BeforeCreate will set a ULID using helpers.NewULID() rather than numeric ID.
// It will check if the ID is already set and if so, it will skip.
// It will also set the sha256 hash of the option.
func (p *PollHistory) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		p.ID = helpers.NewULID()
	}

	return nil
}
