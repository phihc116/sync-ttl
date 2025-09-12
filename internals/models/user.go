package models

type User struct {
	ID           string `gorm:"primaryKey"`
	ClientID     int64  `gorm:"index"`
	ExternalID   int64
	Email        string
	IsUpdated    bool
	UpdatedCount int64
}
