package model

import "time"

type StockBatch struct {
	ID                string    `gorm:"type:char(36);primaryKey" json:"id"`
	DrugID            string    `gorm:"type:char(36);not null;index" json:"drug_id"`
	SupplierID        *string   `gorm:"type:char(36);index" json:"supplier_id"`
	BulkOrderID       *string   `gorm:"type:char(36);index" json:"bulk_order_id"`
	ReceivedBy        *string   `gorm:"type:char(36)" json:"received_by"`
	BatchNumber       string    `gorm:"type:varchar(100);not null" json:"batch_number"`
	QuantityReceived  int       `gorm:"not null" json:"quantity_received"`
	QuantityRemaining int       `gorm:"not null" json:"quantity_remaining"`
	ManufactureDate   *time.Time `gorm:"type:date" json:"manufacture_date"`
	ExpiryDate        time.Time `gorm:"type:date;not null;index" json:"expiry_date"`
	CreatedAt         time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt         time.Time `gorm:"not null" json:"updated_at"`
}