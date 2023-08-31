package models

import (
	"gomeow/pkg/helpers"
	"gorm.io/gorm"
	"math/rand"
	"strings"
	"time"
)

type PollDTO struct {
	ID        string          `json:"id"`
	Code      string          `json:"code"`
	Question  string          `json:"question"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	Details   []PollDetailDTO `json:"details"`
}

type Poll struct {
	ID        string    `json:"id" gorm:"type:char(26);primaryKey;autoIncrement:false"`
	Code      string    `json:"code" gorm:"Column:code;type:varchar(5);not null;unique"`
	UserId    int64     `json:"user_id" gorm:"Column:user_id;type:bigint;not null"`
	Question  string    `json:"question" gorm:"Column:question;type:varchar(255);not null"`
	CreatedAt time.Time `json:"created_at" gorm:"type:timestamp"`
	UpdatedAt time.Time `json:"updated_at" gorm:"type:timestamp"`

	// Associations
	Details []PollDetail `gorm:"foreignKey:PollId;references:ID;"`
	User    User         `gorm:"foreignKey:UserId;references:ID;constraint:OnDelete:RESTRICT"`
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

	if p.Code == "" {
		p.Code = RandStringBytesMaskImprSrcSB(5)
	}

	return nil
}

func (p *Poll) ToResponseDTO() *PollDTO {
	return &PollDTO{
		ID:        p.ID,
		Code:      p.Code,
		Question:  p.Question,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func RandStringBytesMaskImprSrcSB(n int) string {
	var src = rand.NewSource(time.Now().UnixNano())

	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const (
		letterIdxBits = 6                    // 6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	)

	sb := strings.Builder{}
	sb.Grow(n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}
