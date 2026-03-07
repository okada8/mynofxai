package api

import (
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/oschwald/geoip2-golang"
	"nofx/logger"
)

// Global GeoIP reader
var geoIPReader *geoip2.Reader

func init() {
	var err error
	// You need to download GeoLite2-Country.mmdb and place it in the project root or specify path
	// For now we will try to open it, if fails, country will be empty
	geoIPReader, err = geoip2.Open("GeoLite2-Country.mmdb")
	if err != nil {
		// Log error but don't panic, just continue without GeoIP
		// logger.Warnf("Failed to open GeoIP database: %v", err)
	}
}

// AuditLogEntry represents a single audit log event
type AuditLogEntry struct {
	UserID       string
	UserEmail    string
	Action       string
	Resource     string
	ResourceName string
	ResourceType string
	IP           string
	Country      string
	UA           string
	Status       string
	Details      interface{}
}

// auditWorker processes audit logs from the channel
func (s *Server) auditWorker() {
	for entry := range s.auditChan {
		err := s.store.Audit().Log(
			entry.UserID,
			entry.UserEmail,
			entry.Action,
			entry.Resource,
			entry.ResourceName,
			entry.ResourceType,
			entry.IP,
			entry.Country,
			entry.UA,
			entry.Status,
			entry.Details,
		)
		if err != nil {
			logger.Errorf("Failed to write audit log: %v", err)
		}
	}
}

// handleListAuditLogs List audit logs
func (s *Server) handleListAuditLogs(c *gin.Context) {
	userID := c.GetString("user_id")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	logs, total, err := s.store.Audit().List(userID, limit, offset)
	if err != nil {
		SafeInternalError(c, "Failed to get audit logs", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   logs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// logAudit Helper to record audit log asynchronously
func (s *Server) logAudit(c *gin.Context, action, resource, status string, details interface{}) {
	userID := c.GetString("user_id")
	// Try to get user email if available (from context or store)
	var userEmail string
	if val, exists := c.Get("email"); exists {
		userEmail = val.(string)
	} else if userID != "" {
		if user, err := s.store.User().GetByID(userID); err == nil {
			userEmail = user.Email
		}
	}

	// Copy context values needed
	ip := c.ClientIP()
	ua := c.Request.UserAgent()

	// Resolve country
	var country string
	if geoIPReader != nil && ip != "" {
		if netIP := net.ParseIP(ip); netIP != nil {
			if record, err := geoIPReader.Country(netIP); err == nil {
				country = record.Country.IsoCode
			}
			// Handle Local/Private IPs if GeoIP fails or returns empty
			if country == "" {
				if netIP.IsLoopback() {
					country = "Local"
				} else if isPrivateIP(netIP) {
					country = "LAN"
				}
			}
		}
	} else if ip == "::1" || ip == "127.0.0.1" {
		country = "Local"
	}

	// Resolve resource details
	var resourceName, resourceType string

	// Try to parse resource type from resource string or details
	if resource != "" {
		if strings.Contains(resource, "_") {
			parts := strings.Split(resource, "_")
			if len(parts) > 0 {
				// Simple heuristic for resource ID prefixes
			}
		}
	}

	// Try to extract name from details map using type assertion instead of JSON round-trip
	if details != nil {
		// Optimization: Use type assertion for common types
		if m, ok := details.(map[string]interface{}); ok {
			if name, ok := m["name"].(string); ok {
				resourceName = name
			}
			// Infer type
			if _, ok := m["ai_model_id"]; ok {
				resourceType = "TRADER"
			} else if _, ok := m["api_key"]; ok {
				resourceType = "EXCHANGE"
			} else if _, ok := m["config"]; ok {
				resourceType = "STRATEGY"
			}
		} else {
			// Optimization: Skip complex reflection or JSON round-trip for structs.
			// Handlers should prefer passing maps if name extraction is required,
			// or we can implement reflection-based extraction in the future.
		}
	}

	// Fallback for resource type based on action
	if resourceType == "" {
		if strings.Contains(action, "TRADER") {
			resourceType = "TRADER"
		} else if strings.Contains(action, "STRATEGY") {
			resourceType = "STRATEGY"
		} else if strings.Contains(action, "EXCHANGE") {
			resourceType = "EXCHANGE"
		} else if strings.Contains(action, "MODEL") {
			resourceType = "AI_MODEL"
		} else if strings.Contains(action, "LOGIN") || strings.Contains(action, "LOGOUT") {
			resourceType = "AUTH"
		}
	}

	// Create entry and push to channel
	entry := &AuditLogEntry{
		UserID:       userID,
		UserEmail:    userEmail,
		Action:       action,
		Resource:     resource,
		ResourceName: resourceName,
		ResourceType: resourceType,
		IP:           ip,
		Country:      country,
		UA:           ua,
		Status:       status,
		Details:      details,
	}

	// Non-blocking send (drop if full to prevent blocking main request)
	select {
	case s.auditChan <- entry:
	default:
		logger.Warnf("Audit log channel full, dropping log: %s %s", action, resource)
	}
}
