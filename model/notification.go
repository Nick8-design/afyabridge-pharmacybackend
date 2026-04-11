package model

import (
    "time"
    "github.com/google/uuid"
 
    "gorm.io/gorm"
)



type Notification struct {
	ID               string     `gorm:"type:char(36);primaryKey" json:"id"`
	UserID           string     `gorm:"type:char(36);index;not null" json:"user_id"`

	Title            string     `gorm:"type:varchar(255);not null" json:"title"`
	Message          string     `gorm:"type:text;not null" json:"message"`

	NotificationType string     `gorm:"type:enum('appointment','prescription','order','delivery','payment','low_stock','expiry_alert','broadcast','system','chat');index;not null" json:"notification_type"`

	Channel          string     `gorm:"type:enum('sms','email','push','in_app');default:'in_app';index" json:"channel"`

	ReferenceID      *string    `gorm:"type:char(36);index" json:"reference_id,omitempty"`
	ReferenceType    *string    `gorm:"type:varchar(50)" json:"reference_type,omitempty"`

	BroadcastID      *string    `gorm:"type:char(36);index" json:"broadcast_id,omitempty"`

	IsRead           bool       `gorm:"type:tinyint(1);default:0;index" json:"is_read"`
	ReadAt           *time.Time `json:"read_at,omitempty"`

	SentAt           *time.Time `gorm:"index" json:"sent_at,omitempty"`

	CreatedAt        time.Time  `gorm:"not null" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"not null" json:"updated_at"`
}

func (n *Notification) BeforeCreate(tx *gorm.DB) error {
    n.ID = uuid.New().String()
    return nil
}