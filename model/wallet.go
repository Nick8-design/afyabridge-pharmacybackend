package model



import (
	"time"
	"github.com/shopspring/decimal"
)

// Wallet represents the user balance and history table
type Wallet struct {
	ID                 string          `gorm:"type:char(36);primaryKey" json:"id"`
	UserID             string          `gorm:"type:char(36);unique;not null" json:"user_id"`
	Balance decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"balance"` // Changed from float64
	Currency           string          `gorm:"type:varchar(10);default:KES" json:"currency"`
	IsActive           bool            `gorm:"type:tinyint(1);index;default:1" json:"is_active"`
	PayoutMethod       *string         `gorm:"type:enum('mpesa','bank')" json:"payout_method"`
	PayoutAccount      string          `gorm:"type:varchar(100)" json:"payout_account"`
	CreatedAt          time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
	// JSON types require a custom scanner/valuer to work with Go
	Trend              JSONMap         `gorm:"type:json" json:"trend"`
	RecentPayouts      JSONMap         `gorm:"type:json" json:"recent_payouts"`
	TransactionHistory JSONMap         `gorm:"type:json" json:"transaction_history"`
}

func (Wallet) TableName() string {
	return "wallets"
}

type JSONMap map[string]interface{}