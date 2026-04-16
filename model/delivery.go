package model

import "time"



type Delivery struct {
	ID                    string     `gorm:"type:char(36);primaryKey" json:"id"`
	PackageNumber         string     `gorm:"type:varchar(100);not null;unique" json:"package_number"`
	OrderID               string     `gorm:"type:char(36);not null;unique" json:"order_id"`
	RiderID               *string    `gorm:"type:char(36);index" json:"rider_id"`
	Status                string     `gorm:"type:enum('pending','assigned','accepted','picked_up','out_for_delivery','delivered','failed','cancelled');not null;index;default:'pending'" json:"status"`
	AcceptStatus          bool       `gorm:"type:tinyint(1);default:0" json:"accept_status"`
	PickupLocation        string     `gorm:"type:varchar(255)" json:"pickup_location"`
	PickupLat             float64    `gorm:"type:decimal(9,6)" json:"pickup_lat"`
	PickupLng             float64    `gorm:"type:decimal(9,6)" json:"pickup_lng"`
	PickupContact         string     `gorm:"type:varchar(20)" json:"pickup_contact"`
	PickupTime            *time.Time `json:"pickup_time"`
	DropoffLocation       string     `gorm:"type:varchar(255)" json:"dropoff_location"`
	DropoffLat            float64    `gorm:"type:decimal(9,6)" json:"dropoff_lat"`
	DropoffLng            float64    `gorm:"type:decimal(9,6)" json:"dropoff_lng"`
	ReceiverContact       string     `gorm:"type:varchar(20)" json:"receiver_contact"`
	Requirement           string     `gorm:"type:varchar(255)" json:"requirement"`
	EstimatedDeliveryTime string     `gorm:"type:varchar(100)" json:"estimated_delivery_time"`
	Distance              float32    `json:"distance"`
	Charges               float64    `gorm:"type:decimal(10,2)" json:"charges"`
	DeliveryZone          string     `gorm:"type:varchar(100)" json:"delivery_zone"`
	DeliveryNotes         string     `gorm:"type:text" json:"delivery_notes"`
	OtpCode               string     `gorm:"type:varchar(6)" json:"otp_code"`
	DeliveredAt           *time.Time `json:"delivered_at"`
	DateApproved          *time.Time `json:"date_approved"`
	CreatedAt             time.Time  `gorm:"not null" json:"created_at"`
	UpdatedAt             time.Time  `gorm:"not null" json:"updated_at"`
	PackageSealed         bool       `gorm:"type:tinyint(1)" json:"package_sealed"`
	LabeledCorrectly      bool       `gorm:"type:tinyint(1)" json:"labeled_correctly"`
	VerifiedWithPharmacy  bool       `gorm:"type:tinyint(1)" json:"verified_with_pharmacy"`
}
