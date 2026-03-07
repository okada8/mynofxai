package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"nofx/store"
)

// handleCreateAPIKey creates a new API key
func (s *Server) handleCreateAPIKey(c *gin.Context) {
	userID := c.GetString("user_id")
	
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		SafeBadRequest(c, "Invalid request parameters")
		return
	}

	key, hash, err := store.GenerateKey()
	if err != nil {
		SafeInternalError(c, "Failed to generate key", err)
		return
	}

	apiKey := &store.APIKey{
		ID:        uuid.New().String(),
		UserID:    userID,
		KeyHash:   hash,
		Name:      req.Name,
		CreatedAt: time.Now(),
	}

	if err := s.store.APIKey().Create(apiKey); err != nil {
		SafeInternalError(c, "Failed to create API key", err)
		return
	}

	// Return key only once!
	c.JSON(http.StatusCreated, gin.H{
		"id":   apiKey.ID,
		"name": apiKey.Name,
		"key":  key, // Show only once
	})
}

// handleListAPIKeys lists API keys
func (s *Server) handleListAPIKeys(c *gin.Context) {
	userID := c.GetString("user_id")
	
	keys, err := s.store.APIKey().List(userID)
	if err != nil {
		SafeInternalError(c, "Failed to list API keys", err)
		return
	}

	var result []gin.H
	for _, k := range keys {
		result = append(result, gin.H{
			"id":           k.ID,
			"name":         k.Name,
			"created_at":   k.CreatedAt,
			"last_used_at": k.LastUsedAt,
			"prefix":       "nk_", // Hint for user
		})
	}

	c.JSON(http.StatusOK, result)
}

// handleDeleteAPIKey deletes an API key
func (s *Server) handleDeleteAPIKey(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	if err := s.store.APIKey().Delete(userID, id); err != nil {
		SafeInternalError(c, "Failed to delete API key", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API key deleted"})
}
