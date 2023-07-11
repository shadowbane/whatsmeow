package models

import (
	"time"
)

type User struct {
	ID          int64     `json:"id" gorm:"auto_increment;primary_key"`
	Name        string    `json:"name" gorm:"Column:name;type:varchar(255);not null" valid:"required"`
	Token       string    `json:"token" gorm:"Column:token;type:text;not null" valid:"required,stringlength(15|20)"`
	Webhook     string    `json:"webhook" gorm:"Column:webhook;type:text"`
	JID         string    `json:"jid" gorm:"Column:jid;type:varchar(255)"`
	QRCode      string    `json:"qrcode" gorm:"Column:qrcode;type:text"`
	IsConnected bool      `json:"is-connected" gorm:"Column:is_connected;type:tinyint(1);not null;default:0"`
	Expiration  int64     `json:"expiration" gorm:"Column:expiration;type:integer"`
	Events      string    `json:"events" gorm:"Column:events;type:varchar(255);not null;default:'All'"`
	CreatedAt   time.Time `json:"created_at" gorm:"type:timestamp"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"type:timestamp"`
}

func (u *User) TableName() string {
	return "users"
}
