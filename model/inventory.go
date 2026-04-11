package model

import "time"

type Inventory struct {
	ID              string  `gorm:"type:char(36);primaryKey" json:"id"`
	PharmacyID      string  `gorm:"type:char(36);not null;index" json:"pharmacy_id"`
	DrugName        string  `gorm:"type:varchar(255);not null;index" json:"drug_name"`
	GenericName     string  `gorm:"type:varchar(255)" json:"generic_name"`
	Category        string  `gorm:"type:enum('antibiotic','analgesic','chronic','vitamin','antifungal','other');not null;index" json:"category"`
	Unit            string  `gorm:"type:enum('tablet','capsule','bottle','vial','sachet','tube');default:'tablet'" json:"unit"`
	UnitPrice       float64 `gorm:"type:decimal(10,2);not null" json:"unit_price"`
	QuantityInStock int     `gorm:"index;default:0" json:"quantity_in_stock"`
	ReorderLevel    int     `gorm:"default:20" json:"reorder_level"`
	CriticalLevel   int     `gorm:"default:5" json:"critical_level"`
	RequiresRx      bool    `gorm:"type:tinyint(1);default:1" json:"requires_rx"`
	IsActive        bool    `gorm:"type:tinyint(1);default:1;index" json:"is_active"`
	CreatedAt       time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt       time.Time `gorm:"not null" json:"updated_at"`
}