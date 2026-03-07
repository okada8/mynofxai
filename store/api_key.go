package store

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"time"

	"gorm.io/gorm"
)

// APIKey API key model
type APIKey struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	UserID      string    `gorm:"index;not null" json:"user_id"`
	KeyHash     string    `gorm:"uniqueIndex;not null" json:"-"`
	Name        string    `gorm:"not null" json:"name"`
	Permissions string    `gorm:"type:text" json:"permissions"` // JSON or comma-separated list
	LastUsedAt  time.Time `json:"last_used_at"`
	ExpiresAt   *time.Time `json:"expires_at"` // Nullable
	CreatedAt   time.Time `json:"created_at"`
}

func (APIKey) TableName() string { return "api_keys" }

// APIKeyStore storage for API keys
type APIKeyStore struct {
	db *gorm.DB
}

func NewAPIKeyStore(db *gorm.DB) *APIKeyStore {
	return &APIKeyStore{db: db}
}

// GenerateKey generates a new random API key and its hash
// Returns: key (plain text, to show user once), hash (to save to db), error
func GenerateKey() (string, string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", "", err
		}
		b[i] = charset[n.Int64()]
	}
	
	key := "nk_" + string(b)
	
	// Hash the key
	hasher := sha256.New()
	hasher.Write([]byte(key))
	hash := hex.EncodeToString(hasher.Sum(nil))
	
	return key, hash, nil
}

// HashKey hashes an API key for comparison
func HashKey(key string) string {
	hasher := sha256.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

// Create creates a new API key
func (s *APIKeyStore) Create(apiKey *APIKey) error {
	return s.db.Create(apiKey).Error
}

// List gets all API keys for a user
func (s *APIKeyStore) List(userID string) ([]APIKey, error) {
	var keys []APIKey
	err := s.db.Where("user_id = ?", userID).Order("created_at desc").Find(&keys).Error
	return keys, err
}

// Delete deletes an API key
func (s *APIKeyStore) Delete(userID, id string) error {
	return s.db.Where("user_id = ? AND id = ?", userID, id).Delete(&APIKey{}).Error
}

// GetByHash gets an API key by its hash
func (s *APIKeyStore) GetByHash(hash string) (*APIKey, error) {
	var apiKey APIKey
	err := s.db.Where("key_hash = ?", hash).First(&apiKey).Error
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

// UpdateLastUsed updates the last used timestamp
func (s *APIKeyStore) UpdateLastUsed(id string) error {
	return s.db.Model(&APIKey{}).Where("id = ?", id).Update("last_used_at", time.Now()).Error
}
