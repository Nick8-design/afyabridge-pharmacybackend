package model

import (
	"time"

	"gorm.io/datatypes"
)





type User struct {
ID string `gorm:"type:char(36);primaryKey" json:"id"`

	Role string `gorm:"type:enum('patient','doctor','pharmacist','rider','admin');index;not null" json:"role"`

	FullName string `gorm:"type:varchar(255);not null" json:"full_name"`
	Email    string `gorm:"type:varchar(255);unique;not null" json:"email"`

	PasswordHash string `gorm:"type:varchar(255);not null" json:"password_hash"`

	PhoneNumber  *string `gorm:"type:varchar(20);unique" json:"phone_number"`
	ProfileImage *string `gorm:"type:varchar(500)" json:"profile_image"`
	Initials     *string `gorm:"type:varchar(10)" json:"initials"`

	IsActive          bool `gorm:"default:1" json:"is_active"`
	IsVerified        bool `gorm:"default:0" json:"is_verified"`
	TwoFactorEnabled  bool `gorm:"default:0" json:"two_factor_enabled"`
	TwoFactorMethod   string `gorm:"type:enum('sms','email','app');default:sms" json:"two_factor_method"`
	TwoFactorPhone    *string `gorm:"type:varchar(20)" json:"two_factor_phone"`

	LastPasswordChange *time.Time `json:"last_password_change"`
	LastLogin          *time.Time `json:"last_login"`

	AccountStatus string `gorm:"type:enum('active','suspended','locked','disabled');index;default:active" json:"account_status"`
	StatusReason  *string `gorm:"type:varchar(255)" json:"status_reason"`

	Bio        *string `gorm:"type:text" json:"bio"`
	Gender     *string `gorm:"type:varchar(20)" json:"gender"`
	DateOfBirth *string `gorm:"type:varchar(50)" json:"date_of_birth"`
	Age        *int    `json:"age"`
	BloodType  *string `gorm:"type:varchar(10)" json:"blood_type"`
	Address    *string `gorm:"type:varchar(255)" json:"address"`

	ProviderSharing bool `gorm:"default:1" json:"provider_sharing"`
	ResearchOptIn   bool `gorm:"default:0" json:"research_opt_in"`

	EmergencyContacts datatypes.JSON `gorm:"type:json" json:"emergency_contacts"`
	Allergies         datatypes.JSON `gorm:"type:json" json:"allergies"`
	Surgeries         datatypes.JSON `gorm:"type:json" json:"surgeries"`
	Visits            datatypes.JSON `gorm:"type:json" json:"visits"`
	Conditions        datatypes.JSON `gorm:"type:json" json:"conditions"`
	Documents         datatypes.JSON `gorm:"type:json" json:"documents"`

	Specialty       *string  `gorm:"type:varchar(255)" json:"specialty"`
	KMPDCLicense    *string  `gorm:"type:varchar(100)" json:"kmpdc_license"`
	Hospital        *string  `gorm:"type:varchar(255)" json:"hospital"`
	ConsultationFee *float64 `json:"consultation_fee"`

	AllowVideoConsultations    bool `json:"allow_video_consultations"`
	AllowInPersonConsultations bool `json:"allow_in_person_consultations"`

	WorkingHours datatypes.JSON `gorm:"type:json" json:"working_hours"`

	SlotDuration           *int `json:"slot_duration"`
	AutoConfirmAppointments bool `json:"auto_confirm_appointments"`

	Rating       float64 `gorm:"default:0" json:"rating"`
	TotalReviews int     `gorm:"default:0" json:"total_reviews"`

	VerificationStatus *string    `gorm:"type:enum('pending_verification','verified','rejected')" json:"verification_status"`
	VerifiedAt         *time.Time `json:"verified_at"`
	VerifiedBy         *string    `gorm:"type:varchar(100)" json:"verified_by"`

	NationalID       *string    `gorm:"type:varchar(50)" json:"national_id"`
	VehicleType      *string    `gorm:"type:varchar(100)" json:"vehicle_type"`
	PlateNumber      *string    `gorm:"type:varchar(50)" json:"plate_number"`
	DrivingLicenseNo *string    `gorm:"type:varchar(100)" json:"driving_license_no"`
	LicenseExpiry    *time.Time `json:"license_expiry"`

	IDVerified      bool `gorm:"default:0" json:"id_verified"`
	LicenseVerified bool `gorm:"default:0" json:"license_verified"`

	ApprovedStatus *string    `gorm:"type:enum('pending','approved','rejected')" json:"approved_status"`
	DateApproved   *time.Time `json:"date_approved"`

	OnDuty bool `gorm:"default:0" json:"on_duty"`

	EmergencyContact *string `gorm:"type:varchar(100)" json:"emergency_contact"`

	OrdersMade int `gorm:"default:0" json:"orders_made"`

	VerifiedByAdmin bool `gorm:"default:0" json:"verified_by_admin"`

	PharmacyID *string `gorm:"type:char(36);index" json:"pharmacy_id"`

	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`

	GpsLat                     string         `gorm:"type:varchar(255)" json:"gps_lat"`
	GpsLng                     string         `gorm:"type:varchar(255)" json:"gps_lng"`

}


