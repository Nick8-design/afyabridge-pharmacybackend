package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Prescription struct {
    // Primary Key
    ID uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`

    // Unique & Foreign Keys
    PrescriptionNumber *string    `gorm:"type:varchar(100);uniqueIndex" json:"prescription_number"`
    PatientID          uuid.UUID  `gorm:"type:char(36);not null;index" json:"patient_id"`
    DoctorID           uuid.UUID  `gorm:"type:char(36);not null;index" json:"doctor_id"`
    PharmacyID         *uuid.UUID `gorm:"type:char(36);index" json:"pharmacy_id"`
    DispensedBy        *uuid.UUID `gorm:"type:char(36)" json:"dispensed_by"`

    // Denormalized Info
    PatientName    string  `gorm:"type:varchar(255);not null" json:"patient_name"`
    PatientPhone   *string `gorm:"type:varchar(20)" json:"patient_phone"`
    PatientAddress *string `gorm:"type:varchar(255)" json:"patient_address"`
    DoctorName     string  `gorm:"type:varchar(255);not null" json:"doctor_name"`

    // Content
    Diagnosis *string `gorm:"type:text" json:"diagnosis"`
    Notes     *string `gorm:"type:text" json:"notes"`

    // Dates
    IssueDate  time.Time  `gorm:"type:date;not null" json:"issue_date"`
    ExpiryDate *time.Time `gorm:"type:date" json:"expiry_date"`

    // Structured Data
    // Using gorm.io/datatypes for native JSON support in Linux/SQL environments
    Items datatypes.JSON `gorm:"type:json;not null" json:"items"`

    // Enums & Status
    Status          string  `gorm:"type:enum('draft','pending','validated','rejected','dispensed','delivered');not null;default:draft;index" json:"status"`
    Priority        string  `gorm:"type:enum('normal','urgent');default:normal" json:"priority"`
    RejectionReason *string `gorm:"type:text" json:"rejection_reason"`
	
		MakeOrder  bool   `json:"make_order" gorm:"default:false"`

    // Timestamps
    DispensedAt *time.Time `gorm:"type:datetime" json:"dispensed_at"`
    CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}
