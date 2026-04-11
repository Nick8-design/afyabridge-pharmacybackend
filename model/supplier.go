package model

import "time"

type Supplier struct {
	ID          string    `gorm:"type:char(36);primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(255);not null;index" json:"name"`
	ContactName string    `gorm:"type:varchar(255)" json:"contact_name"`
	Email       string    `gorm:"type:varchar(255)" json:"email"`
	Phone       string    `gorm:"type:varchar(20)" json:"phone"`
	Address     string    `gorm:"type:text" json:"address"`
	IsActive    bool      `gorm:"type:tinyint(1);default:1;index" json:"is_active"`
	CreatedAt   time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time `gorm:"not null" json:"updated_at"`
}