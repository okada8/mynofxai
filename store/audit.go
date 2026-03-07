package store

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// AuditLog represents a system audit log entry
type AuditLog struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    string    `gorm:"index;size:36" json:"user_id"`
	Email     string    `gorm:"index;size:100" json:"email"`   // e.g., "user@example.com"
	Action    string    `gorm:"index;size:50" json:"action"`   // e.g., "LOGIN", "CREATE_TRADER"
	Resource  string    `gorm:"index;size:100" json:"resource"` // e.g., "trader:123"
	ResourceName string `gorm:"size:100" json:"resource_name"` // Human readable name e.g., "BTC_Strategy"
	ResourceType string `gorm:"size:50" json:"resource_type"`  // e.g., "TRADER", "STRATEGY"
	Details   string    `gorm:"type:text" json:"details"`      // JSON details
	IPAddress string    `gorm:"size:50" json:"ip_address"`
	Country   string    `gorm:"size:50" json:"country"`        // IP Country code
	UserAgent string    `gorm:"size:255" json:"user_agent"`
	Status    string    `gorm:"size:20" json:"status"` // SUCCESS, FAILURE

	CreatedAt time.Time `gorm:"index" json:"created_at"`
}

// AuditStore handles audit log operations
type AuditStore struct {
	db *gorm.DB
}

// NewAuditStore creates a new AuditStore
func NewAuditStore(db *gorm.DB) *AuditStore {
	return &AuditStore{db: db}
}

// initTables initializes audit tables
func (s *AuditStore) initTables() error {
	return s.db.AutoMigrate(&AuditLog{})
}

// Create creates a new audit log
func (s *AuditStore) Create(log *AuditLog) error {
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}
	return s.db.Create(log).Error
}

// Log records a new audit log entry conveniently
func (s *AuditStore) Log(userID, email, action, resource, resourceName, resourceType, ip, country, ua, status string, details interface{}) error {
	var detailsStr string
	if details != nil {
		if s, ok := details.(string); ok {
			detailsStr = s
		} else {
			b, _ := json.Marshal(details)
			detailsStr = string(b)
		}
	}

	return s.Create(&AuditLog{
		UserID:    userID,
		Email:     email,
		Action:    action,
		Resource:  resource,
		ResourceName: resourceName,
		ResourceType: resourceType,
		Details:   detailsStr,
		IPAddress: ip,
		Country:   country,
		UserAgent: ua,
		Status:    status,
	})
}

// List returns audit logs with pagination
func (s *AuditStore) List(userID string, limit, offset int) ([]AuditLog, int64, error) {
	var logs []AuditLog
	var total int64

	query := s.db.Model(&AuditLog{})
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error
	return logs, total, err
}
