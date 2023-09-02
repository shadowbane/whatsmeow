package models

import (
	"crypto/sha256"
	"fmt"
	"gomeow/pkg/helpers"
	"gorm.io/gorm"
	"time"
)

type PollDetailDTO struct {
	ID        string    `json:"id"`
	PollId    string    `json:"poll_id"`
	Option    string    `json:"option"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PollDetail struct {
	ID           string    `json:"id" gorm:"type:char(26);primaryKey;autoIncrement:false"`
	PollId       string    `json:"poll_id" gorm:"Column:poll_id;type:char(26);not null"`
	Option       string    `json:"option" gorm:"Column:option;type:varchar(255);not null"`
	OptionSha256 string    `json:"option_sha256" gorm:"Column:option_sha256;type:text;not null"`
	CreatedAt    time.Time `json:"created_at" gorm:"type:timestamp"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"type:timestamp"`

	// Associations
	Poll Poll `gorm:"foreignKey:PollId;references:ID;constraint:OnDelete:RESTRICT"`
}

func (p *PollDetail) TableName() string {
	return "whatsmeow_poll_details"
}

// BeforeCreate will set a ULID using helpers.NewULID() rather than numeric ID.
// It will check if the ID is already set and if so, it will skip.
// It will also set the sha256 hash of the option.
func (p *PollDetail) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		p.ID = helpers.NewULID()
	}

	if p.OptionSha256 == "" {
		sha256val := sha256.Sum256([]byte(p.Option))
		p.OptionSha256 = fmt.Sprintf("%x", sha256val)
	}

	return nil
}

func (p *PollDetail) ToResponseDTO() *PollDetailDTO {
	return &PollDetailDTO{
		ID:        p.ID,
		PollId:    p.PollId,
		Option:    p.Option,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}
