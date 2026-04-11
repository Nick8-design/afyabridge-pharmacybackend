package model

import "time"

type PharmacyRegistration struct {
	ID                   string     `gorm:"type:char(36);primaryKey" json:"id"`
	PharmacyNameLegal    string     `gorm:"type:varchar(255);not null" json:"pharmacy_name_legal"`
	TradingName          string     `gorm:"type:varchar(255)" json:"trading_name"`
	BusinessRegNo        string     `gorm:"type:varchar(100);not null" json:"business_reg_no"`
	KraPin               string     `gorm:"type:varchar(20);not null" json:"kra_pin"`
	PpbLicenseNo         string     `gorm:"type:varchar(100);not null;index" json:"ppb_license_no"`
	LicenseExpiry        time.Time  `gorm:"type:date;not null" json:"license_expiry"`
	County               string     `gorm:"type:varchar(100);not null" json:"county"`
	SubCounty            string     `gorm:"type:varchar(100)" json:"sub_county"`
	PhysicalAddress      string     `gorm:"type:text;not null" json:"physical_address"`
	GpsLat               float64    `gorm:"type:decimal(9,6)" json:"gps_lat"`
	GpsLng               float64    `gorm:"type:decimal(9,6)" json:"gps_lng"`
	BusinessPhone        string     `gorm:"type:varchar(20);not null" json:"business_phone"`
	BusinessEmail        string     `gorm:"type:varchar(255);not null;index" json:"business_email"`
	PhoneVerified        bool       `gorm:"type:tinyint(1);default:0" json:"phone_verified"`
	EmailVerified        bool       `gorm:"type:tinyint(1);default:0" json:"email_verified"`
	PharmacistName       string     `gorm:"type:varchar(255)" json:"pharmacist_name"`
	IdOrPassportNo       string     `gorm:"type:varchar(50)" json:"id_or_passport_no"`
	PharmacistRegNo      string     `gorm:"type:varchar(100)" json:"pharmacist_reg_no"`
	PracticingLicense    string     `gorm:"type:varchar(100)" json:"practicing_license"`
	PracticingExpiry     *time.Time `gorm:"type:date" json:"practicing_expiry"`
	PharmacistPhone      string     `gorm:"type:varchar(20)" json:"pharmacist_phone"`
	PharmacistEmail      string     `gorm:"type:varchar(255)" json:"pharmacist_email"`
	IdDocument           string     `gorm:"type:varchar(500)" json:"id_document"`
	PracticingLicenseDoc string     `gorm:"type:varchar(500)" json:"practicing_license_doc"`
	OperatingLicenseDoc  string     `gorm:"type:varchar(500)" json:"operating_license_doc"`
	BusinessRegCert      string     `gorm:"type:varchar(500)" json:"business_reg_cert"`
	KraPinCert           string     `gorm:"type:varchar(500)" json:"kra_pin_cert"`
	ProofOfAddressDoc    string     `gorm:"type:varchar(500)" json:"proof_of_address_doc"`
	MpesaMethod          string     `gorm:"type:enum('PAYBILL','TILL')" json:"mpesa_method"`
	ShortCodeName        string     `gorm:"type:varchar(255)" json:"short_code_name"`
	ShortCodeNumber      string     `gorm:"type:varchar(10)" json:"short_code_number"`
	SettlementBank       string     `gorm:"type:varchar(100)" json:"settlement_bank"`
	SettlementFrequency  string     `gorm:"type:enum('DAILY','WEEKLY','MONTHLY');default:'DAILY'" json:"settlement_frequency"`
	Status               string     `gorm:"type:enum('draft','submitted','under_review','approved','rejected');not null;index;default:'draft'" json:"status"`
	CurrentStep          int        `gorm:"default:1" json:"current_step"`
	SubmittedAt          *time.Time `json:"submitted_at"`
	ReviewedBy           *string    `gorm:"type:char(36)" json:"reviewed_by"`
	ReviewNotes          string     `gorm:"type:text" json:"review_notes"`
	CreatedAt            time.Time  `gorm:"not null" json:"created_at"`
	UpdatedAt            time.Time  `gorm:"not null" json:"updated_at"`
}