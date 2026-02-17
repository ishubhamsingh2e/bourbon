package logging

import (
	"time"

	"gorm.io/gorm"
)

// ErrorLog represents a server error stored in database
type ErrorLog struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Timestamp time.Time `gorm:"index" json:"timestamp"`
	Level     string    `gorm:"index;size:20" json:"level"`
	Message   string    `gorm:"type:text" json:"message"`
	Method    string    `gorm:"size:10" json:"method"`
	Path      string    `gorm:"size:500" json:"path"`
	Status    int       `gorm:"index" json:"status"`
	IP        string    `gorm:"size:45" json:"ip"`
	UserAgent string    `gorm:"size:500" json:"user_agent"`
	Stack     string    `gorm:"type:text" json:"stack,omitempty"`
	Extra     string    `gorm:"type:json" json:"extra,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// ErrorStore handles storing errors to database
type ErrorStore struct {
	db      *gorm.DB
	enabled bool
}

// NewErrorStore creates a new error store
func NewErrorStore(db *gorm.DB, enabled bool) *ErrorStore {
	return &ErrorStore{
		db:      db,
		enabled: enabled,
	}
}

// Store saves an error log to database
func (s *ErrorStore) Store(log *ErrorLog) error {
	if !s.enabled || s.db == nil {
		return nil
	}

	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}

	return s.db.Create(log).Error
}

// Migrate runs the migration for error logs table
func (s *ErrorStore) Migrate() error {
	if s.db == nil {
		return nil
	}
	return s.db.AutoMigrate(&ErrorLog{})
}

// Clean removes old error logs based on retention policy
func (s *ErrorStore) Clean(olderThan time.Duration) error {
	if !s.enabled || s.db == nil {
		return nil
	}

	cutoff := time.Now().Add(-olderThan)
	return s.db.Where("timestamp < ?", cutoff).Delete(&ErrorLog{}).Error
}

// GetRecent retrieves recent error logs
func (s *ErrorStore) GetRecent(limit int) ([]ErrorLog, error) {
	if !s.enabled || s.db == nil {
		return nil, nil
	}

	var logs []ErrorLog
	err := s.db.Order("timestamp DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

// GetByStatus retrieves error logs by HTTP status code
func (s *ErrorStore) GetByStatus(status int, limit int) ([]ErrorLog, error) {
	if !s.enabled || s.db == nil {
		return nil, nil
	}

	var logs []ErrorLog
	err := s.db.Where("status = ?", status).Order("timestamp DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

// GetServerErrors retrieves 5xx server errors
func (s *ErrorStore) GetServerErrors(limit int) ([]ErrorLog, error) {
	if !s.enabled || s.db == nil {
		return nil, nil
	}

	var logs []ErrorLog
	err := s.db.Where("status >= ? AND status < ?", 500, 600).Order("timestamp DESC").Limit(limit).Find(&logs).Error
	return logs, err
}
