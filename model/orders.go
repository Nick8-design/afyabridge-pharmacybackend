package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Order struct {
	ID             string          `gorm:"type:char(36);primaryKey" json:"id"`
	OrderNumber    string          `gorm:"type:varchar(100);uniqueIndex" json:"order_number"`
	PrescriptionID string          `gorm:"type:char(36);index" json:"prescription_id"`
	PharmacyID     string          `gorm:"type:char(36);index" json:"pharmacy_id"`
	Prescription   Prescription    `gorm:"foreignKey:PrescriptionID" json:"prescription"`
	PreparedBy     string          `gorm:"type:char(36)" json:"prepared_by"`
	PatientID      string          `gorm:"type:char(36);index" json:"patient_id"`
	PatientName    string          `gorm:"type:varchar(255)" json:"patient_name"`
	PatientPhone   string          `gorm:"type:varchar(20)" json:"patient_phone"`
	PatientAddress string          `gorm:"type:text" json:"patient_address"`
	
	// Enums handled as strings in Go
	DeliveryType   string          `gorm:"type:enum('pickup','home_delivery');default:'pickup'" json:"delivery_type"`
	Priority       string          `gorm:"type:enum('urgent','normal');default:'normal'" json:"priority"`
	Status         string          `gorm:"type:enum('pending','accepted','processing','ready','dispatched','delivered','cancelled');default:'pending'" json:"status"`
		PatientLat     float64         `gorm:"type:decimal(10,8)" json:"patient_lat"` // 👈 Add this
    PatientLng     float64         `gorm:"type:decimal(11,8)" json:"patient_lng"`
	
	// Financial Data
	TotalAmount    decimal.Decimal `gorm:"type:decimal(10,2);not null;default:0" json:"total_amount"`
	PaymentStatus  string          `gorm:"type:enum('unpaid','paid','refunded');default:'unpaid'" json:"payment_status"`
	PaymentMethod  string          `gorm:"type:enum('mpesa','cash','insurance','nhif')" json:"payment_method"`
	MpesaRef       string          `gorm:"type:varchar(50)" json:"mpesa_ref"`

	CreatedAt      time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

// BeforeCreate hook to automatically generate UUID and Order Number
func (o *Order) BeforeCreate(tx *gorm.DB) (err error) {
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	// Example: Generate a readable order number if not provided
	if o.OrderNumber == "" {
		o.OrderNumber = "ORD-" + time.Now().Format("20060102") + "-" + o.ID[:8]
	}
	return
}