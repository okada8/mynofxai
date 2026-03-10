package polymarket

import (
	"fmt"
	"net/http"
	"nofx/manager"
	"nofx/store"
	"nofx/trader/polymarket"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	store         *store.Store
	traderManager *manager.TraderManager
	gammaClient   *polymarket.GammaClient
}

func NewHandler(store *store.Store, traderManager *manager.TraderManager) *Handler {
	return &Handler{
		store:         store,
		traderManager: traderManager,
		gammaClient:   polymarket.NewGammaClient(),
	}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/markets", h.GetMarkets)
	r.GET("/markets/:id", h.GetMarketDetail)
	r.GET("/positions", h.GetPositions)
	r.POST("/orders", h.CreateOrder)
	r.DELETE("/orders/:id", h.CancelOrder)
	r.GET("/balance", h.GetBalance)
	r.POST("/redeem/:conditionID", h.RedeemShares)
}

// GetMarkets fetches active markets
func (h *Handler) GetMarkets(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	// Call Gamma API
	markets, err := h.gammaClient.GetActiveMarkets(limit)
	if err != nil {
		fmt.Printf("Error fetching markets: %v\n", err)
		c.JSON(500, gin.H{"error": "Failed to fetch markets: " + err.Error()})
		return
	}

	// Format response for frontend
	var formatted []map[string]interface{}
	for _, m := range markets {
		// Basic normalization
		yesPrice := 0.5
		noPrice := 0.5
		if len(m.Outcomes) >= 2 {
			yesPrice = m.Outcomes[0].Price
			noPrice = m.Outcomes[1].Price
		}

		formatted = append(formatted, map[string]interface{}{
			"id":        m.ID,
			"question":  m.Question,
			"slug":      m.Slug,
			"yesPrice":  yesPrice,
			"noPrice":   noPrice,
			"liquidity": m.Liquidity,
			"volume24h": m.Volume24h,
			"endDate":   m.EndDate,
			"outcomes":  m.Outcomes,
		})
	}

	c.JSON(200, formatted)
}

func (h *Handler) GetMarketDetail(c *gin.Context) {
	id := c.Param("id")
	// For now just return mock or fetch specific
	c.JSON(200, gin.H{
		"id":       id,
		"question": "Detail view not implemented yet",
	})
}

// CreateOrder creates a new order
func (h *Handler) CreateOrder(c *gin.Context) {
	var req struct {
		TraderID string  `json:"trader_id" binding:"required"`
		Symbol   string  `json:"symbol" binding:"required"` // TokenID or ConditionID/Index
		Side     string  `json:"side" binding:"required"`   // BUY/SELL or LONG/SHORT
		Quantity float64 `json:"quantity" binding:"required"`
		Leverage int     `json:"leverage"` // Optional
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get trader
	trader, err := h.traderManager.GetTrader(req.TraderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trader not found"})
		return
	}

	// Check if it's a Polymarket trader
	if trader.GetExchange() != "polymarket" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Trader is not a Polymarket trader"})
		return
	}

	// Execute order
	var result map[string]interface{}
	
	// Normalize Side
	side := req.Side
	if side == "BUY" {
		side = "LONG"
	} else if side == "SELL" {
		side = "SHORT" 
	}

	if side == "LONG" {
		result, err = trader.GetUnderlyingTrader().OpenLong(req.Symbol, req.Quantity, req.Leverage)
	} else if side == "SHORT" {
		// For Polymarket, SHORT typically means buying "NO" shares (if binary) or selling held shares
		// The OpenShort implementation in PolymarketTrader handles "Buy NO" logic
		result, err = trader.GetUnderlyingTrader().OpenShort(req.Symbol, req.Quantity, req.Leverage)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid side"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Order execution failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   result,
	})
}

// CancelOrder cancels an order
func (h *Handler) CancelOrder(c *gin.Context) {
	id := c.Param("id") // Order ID
	traderID := c.Query("trader_id")
	symbol := c.Query("symbol")

	if traderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "trader_id is required"})
		return
	}

	trader, err := h.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trader not found"})
		return
	}

	// PolymarketTrader doesn't support cancelling individual orders by ID yet via interface
	// But we can try CancelOrder if it implements specific interface, or just fail for now
	// The standard Trader interface has CancelAllOrders
	
	if symbol != "" {
		err = trader.GetUnderlyingTrader().CancelAllOrders(symbol)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel orders: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "All orders for symbol cancelled"})
	} else {
		// Try to cancel specific order if supported (needs type assertion)
		if pmTrader, ok := trader.GetUnderlyingTrader().(interface{ CancelOrder(string, string) error }); ok {
             err = pmTrader.CancelOrder(symbol, id) // Symbol might be needed
             if err != nil {
                 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel order: " + err.Error()})
                 return
             }
             c.JSON(http.StatusOK, gin.H{"status": "success"})
        } else {
             c.JSON(http.StatusBadRequest, gin.H{"error": "Cancel by ID not supported, provide symbol to cancel all"})
        }
	}
}

// GetPositions gets positions for a trader
func (h *Handler) GetPositions(c *gin.Context) {
	traderID := c.Query("trader_id")
	if traderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "trader_id is required"})
		return
	}

	trader, err := h.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trader not found"})
		return
	}

	positions, err := trader.GetPositions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get positions: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, positions)
}

// GetBalance gets balance for a trader
func (h *Handler) GetBalance(c *gin.Context) {
	traderID := c.Query("trader_id")
	if traderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "trader_id is required"})
		return
	}

	trader, err := h.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trader not found"})
		return
	}

	balance, err := trader.GetAccountInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get balance: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, balance)
}

func (h *Handler) RedeemShares(c *gin.Context) {
	conditionID := c.Param("conditionID")
	c.JSON(200, gin.H{
		"status":      "success",
		"message":     "Redemption triggered (mock)",
		"conditionID": conditionID,
	})
}
