package models

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel contains common fields for all models
// Embed this in your models instead of defining these fields manually
type BaseModel struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
