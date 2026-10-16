// Package model defines domain model conventions, base timestamped models, and metadata reflection helpers.
package model

import "time"

// Model defines the minimal contract for domain models specifying a table name.
type Model interface {
	TableName() string
}

// BaseModel provides standard primary key and lifecycle timestamp fields.
type BaseModel struct {
	ID        string     `json:"id" db:"id,primary"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// TableName returns an empty string fallback to be overridden by domain models.
func (b BaseModel) TableName() string {
	return ""
}
