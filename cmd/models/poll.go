package models

import (
	"database/sql"
	"gomeow/pkg/helpers"
	"gorm.io/gorm"
	"time"
)

type Poll struct {
	ID          string       `json:"id" gorm:"type:char(26);primaryKey;autoIncrement:false"`
	Code        string       `json:"code" gorm:"Column:code;type:varchar(5);not null;unique"`
	JID         string       `json:"jid" gorm:"Column:jid;type:varchar(255);not null"`
	UserId      int64        `json:"user_id" gorm:"Column:user_id;type:bigint;not null"`
	MessageId   string       `json:"message_id" gorm:"Column:message_id;type:varchar(255);not null;unique"`
	Question    string       `json:"question" gorm:"Column:question;type:varchar(255);not null"`
	Destination string       `json:"destination" gorm:"Column:destination;not null"`
	Sent        bool         `json:"sent" gorm:"Column:sent;Default:false"`
	Read        bool         `json:"read" gorm:"Column:read;Default:false"`
	Failed      bool         `json:"failed" gorm:"Column:failed;Default:false"`
	SentAt      sql.NullTime `json:"sent_at,omitempty" gorm:"Column:sent_at;type:timestamp;default:null"`
	ReadAt      sql.NullTime `json:"read_at,omitempty" gorm:"Column:read_at;type:timestamp;default:null"`
	FailedAt    sql.NullTime `json:"failed_at,omitempty" gorm:"Column:failed_at;type:timestamp;default:null"`
	CreatedAt   time.Time    `json:"created_at" gorm:"type:timestamp"`
	UpdatedAt   time.Time    `json:"updated_at" gorm:"type:timestamp"`

	// Associations
	Details []PollDetail `gorm:"foreignKey:PollId;references:ID;"`
}

func (p *Poll) TableName() string {
	return "polls"
}

// BeforeCreate will set a ULID using helper.NewULID() rather than numeric ID.
// It will check if the ID is already set and if so, it will skip
func (p *Poll) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		p.ID = helpers.NewULID()
	}

	return nil
}
