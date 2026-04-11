package model

import (
	"time"
	"gorm.io/datatypes"
)

type Medication struct {
	ID                    string         `gorm:"type:char(36);primaryKey" json:"id"`
	PatientID             string         `gorm:"type:char(36);not null;index" json:"patient_id"`
	PrescriptionID        *string        `gorm:"type:char(36);index" json:"prescription_id"`
	PharmacyID            *string        `gorm:"type:char(36);index" json:"pharmacy_id"`
	PrescribedBy          *string        `gorm:"type:char(36);index" json:"prescribed_by"`
	DispensedBy           *string        `gorm:"type:char(36)" json:"dispensed_by"`
	DrugID                *string        `gorm:"type:char(36);index" json:"drug_id"`
	DrugName              string         `gorm:"type:varchar(255);not null" json:"drug_name"`
	Dosage                string         `gorm:"type:varchar(100)" json:"dosage"`
	DosageForm            string         `gorm:"type:enum('tablet','capsule','syrup','injection','cream','drops','inhaler','patch','other')" json:"dosage_form"`
	Frequency             string         `gorm:"type:varchar(100)" json:"frequency"`
	TimesPerDay           int            `json:"times_per_day"`
	DosageTiming          datatypes.JSON `json:"dosage_timing"`
	WithFood              bool           `gorm:"type:tinyint(1);default:0" json:"with_food"`
	Route                 string         `gorm:"type:varchar(50)" json:"route"`
	Instructions          string         `gorm:"type:text" json:"instructions"`
	DurationDays          int            `json:"duration_days"`
	QuantityDispensed     int            `json:"quantity_dispensed"`
	QuantityRemaining     int            `json:"quantity_remaining"`
	RefillsAllowed        int            `gorm:"default:0" json:"refills_allowed"`
	RefillsUsed           int            `gorm:"default:0" json:"refills_used"`
	StartDate             time.Time      `gorm:"type:date;not null" json:"start_date"`
	EndDate               *time.Time     `gorm:"type:date" json:"end_date"`
	DispensedAt           *time.Time     `json:"dispensed_at"`
	LastTakenAt           *time.Time     `json:"last_taken_at"`
	NextRefillDate        *time.Time     `gorm:"type:date;index" json:"next_refill_date"`
	RefillReminderDays    int            `gorm:"default:3" json:"refill_reminder_days"`
	LastReminderSentAt    *time.Time     `json:"last_reminder_sent_at"`
	Status                string         `gorm:"type:enum('active','completed','discontinued','on_hold','expired');not null;index;default:'active'" json:"status"`
	IsChronic             bool           `gorm:"type:tinyint(1);default:0" json:"is_chronic"`
	IsOtc                 bool           `gorm:"type:tinyint(1);default:0" json:"is_otc"`
	DiscontinuationReason string         `gorm:"type:text" json:"discontinuation_reason"`
	Notes                 string         `gorm:"type:text" json:"notes"`
	Warnings              string         `gorm:"type:text" json:"warnings"`
	CreatedAt             time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt             time.Time      `gorm:"not null" json:"updated_at"`
	DailyLog              datatypes.JSON `json:"daily_log"`
	AdherencePercentage   int            `json:"adherence_percentage"`
}