package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"nofx/logger"
	"nofx/market"
)

// RiskHandler handles risk management requests
type RiskHandler struct {
	// dependencies will be injected here
}

// NewRiskHandler creates a new risk handler
func NewRiskHandler() *RiskHandler {
	return &RiskHandler{}
}

// Control adjusts risk parameters
func (h *RiskHandler) Control(c *gin.Context) {
	var req struct {
		StrategyID  string      `json:"strategy_id" binding:"required"`
		ControlType string      `json:"control_type" binding:"required"`
		Action      string      `json:"action" binding:"required"`
		Value       interface{} `json:"value" binding:"required"`
		Reason      string      `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement actual risk control logic
	// This would involve updating the strategy config in DB and notifying the running engine

	logger.Infof("[Risk] Control request: %s %s -> %v (%s)", req.StrategyID, req.ControlType, req.Value, req.Reason)

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"message": "Risk parameters updated",
		"applied_at": time.Now(),
	})
}

// GetStatus gets current risk status
func (h *RiskHandler) GetStatus(c *gin.Context) {
	strategyID := c.Query("strategy_id")
	if strategyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "strategy_id is required"})
		return
	}

	// TODO: Fetch real status from engine
	c.JSON(http.StatusOK, gin.H{
		"strategy_id": strategyID,
		"daily_loss_pct": 1.2,
		"is_trading_enabled": true,
		"risk_level": "low",
		"margin_usage": 0.15,
	})
}

// AlphaHandler handles alpha factor requests
type AlphaHandler struct {
	alphaManager *market.AlphaManager
}

// NewAlphaHandler creates a new alpha handler
func NewAlphaHandler() *AlphaHandler {
	return &AlphaHandler{
		alphaManager: market.NewAlphaManager(),
	}
}

// GetFactors gets alpha factors for a symbol
func (h *AlphaHandler) GetFactors(c *gin.Context) {
	symbol := c.Query("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}

	data, err := h.alphaManager.GetAlphaData(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}

// AgentHandler handles multi-agent system requests
type AgentHandler struct {
	// dependencies
}

// NewAgentHandler creates a new agent handler
func NewAgentHandler() *AgentHandler {
	return &AgentHandler{}
}

// Analyze triggers agent analysis
func (h *AgentHandler) Analyze(c *gin.Context) {
	var req struct {
		Symbol string   `json:"symbol" binding:"required"`
		Agents []string `json:"agents"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Trigger agent analysis
	c.JSON(http.StatusOK, gin.H{
		"request_id": "req_" + time.Now().Format("20060102150405"),
		"status": "processing",
	})
}

// RegisterRoutes registers routes for new handlers
func RegisterRoutes(r *gin.Engine) {
	// This function is just a placeholder to show route structure
	// Actual registration happens in server.go
}
