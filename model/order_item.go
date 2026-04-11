package model

import "time"

type OrderItem struct {
	ID        string    `gorm:"type:char(36);primaryKey" json:"id"`
	OrderID   string    `gorm:"type:char(36);not null;index" json:"order_id"`
	DrugID    string    `gorm:"type:char(36);not null;index" json:"drug_id"`
	DrugName  string    `gorm:"type:varchar(255);not null" json:"drug_name"`
	Dosage    string    `gorm:"type:varchar(100)" json:"dosage"`
	Frequency string    `gorm:"type:varchar(100)" json:"frequency"`
	Quantity  int       `gorm:"not null;default:1" json:"quantity"`
	UnitPrice float64   `gorm:"type:decimal(10,2);not null;default:0" json:"unit_price"`
	CreatedAt time.Time `gorm:"type:datetime(6);not null;default:CURRENT_TIMESTAMP(6)" json:"created_at"`
}