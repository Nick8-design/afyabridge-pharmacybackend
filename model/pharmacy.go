package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Pharmacy struct {
	ID            string         `gorm:"type:char(36);primaryKey" json:"id"`
	Name          string         `gorm:"type:varchar(255);not null" json:"name"`
	Email         string         `gorm:"type:varchar(255);not null;unique" json:"email"`
	Phone         string         `gorm:"type:varchar(20);not null" json:"phone"`
	Logo          *string        `gorm:"type:varchar(500)" json:"logo"`
	AddressLine1  string         `gorm:"type:varchar(255);not null" json:"address_line1"`
	AddressLine2  *string        `gorm:"type:varchar(255)" json:"address_line2"`
	County        string         `gorm:"type:varchar(100);index;not null" json:"county"`
	SubCounty     *string        `gorm:"type:varchar(100)" json:"sub_county"`
	GpsLat        *float64       `gorm:"type:decimal(9,6)" json:"gps_lat"`
	GpsLng        *float64       `gorm:"type:decimal(9,6)" json:"gps_lng"`
	LicenseNumber string         `gorm:"type:varchar(100);index;not null" json:"license_number"`
	LicenseExpiry time.Time      `gorm:"type:date;not null" json:"license_expiry"`
	DeliveryZones datatypes.JSON `gorm:"type:json" json:"delivery_zones"`
	Is_24hr        bool           `gorm:"type:tinyint(1);default:0" json:"is_24hr"`
	IsActive      bool           `gorm:"type:tinyint(1);default:1;index" json:"is_active"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

// BeforeCreate hook to generate UUID
func (p *Pharmacy) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return
}