package models

type User struct {
	AutoID       int64  `gorm:"primaryKey;autoIncrement"`
	ID           string `gorm:"primaryKey"`
	ClientID     int64  `gorm:"index"`
	ExternalID   int64
	Email        string
	IsUpdated    bool
	UpdatedCount int64
}
