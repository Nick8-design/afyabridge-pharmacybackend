package model

type PharmacyHour struct {
	ID         string `gorm:"type:char(36);primaryKey" json:"id"`
	PharmacyID string `gorm:"type:char(36);not null;index" json:"pharmacy_id"`
	DayOfWeek  string `gorm:"type:enum('MON','TUE','WED','THU','FRI','SAT','SUN');not null" json:"day_of_week"`
	OpenTime   string `gorm:"type:time" json:"open_time"`
	CloseTime  string `gorm:"type:time" json:"close_time"`
	IsClosed   bool   `gorm:"type:tinyint(1);default:0" json:"is_closed"`
}