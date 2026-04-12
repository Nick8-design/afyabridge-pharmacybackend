package model

import "time"

/*
i want all pharmarcies to receive all orders no matter the id based on status pending  , but to receive only if they have the medicen drug items in the prescriptions in their inventory , by default to check first if the pharmacy that was sent to has the medicine all the medicine in his inventory if yes then automatically mark the status accepted ,but id not let it remain pending ,  and when the other pharmacy which has the drugs in prescription receives it and let him mark accepted manually so that the other dont see it , and when serving (making delivery item to go to delivery table) it let same amount of drugs be removed from his inventory and be placed in the patients inventory and fill the needed details including the capacity included in the prescription,and as this  make this delivery item is being added to the deliveries table you can assign it to specific rider who is not on duty or lets just live  it unssigned so that any rider can see it

*/

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
