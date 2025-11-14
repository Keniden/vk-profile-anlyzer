package domain

import "time"

type AuthUser struct {
	ID           uint      `gorm:"primaryKey"`
	Email        string    `gorm:"uniqueIndex;size:255;not null"`
	PasswordHash string    `gorm:"size:255;not null"`
	VKID         int64     `gorm:"index"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

