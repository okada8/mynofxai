package polymarket

import (
	"nofx/trader/polymarket"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	gammaClient *polymarket.GammaClient
}

func NewHandler() *Handler {
	return &Handler{
		gammaClient: polymarket.NewGammaClient(),
	}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/markets", h.GetMarkets)
	r.GET("/markets/:id", h.GetMarketDetail)
	r.POST("/markets/:id/buy", h.BuyShares)
	r.POST("/markets/:id/sell", h.SellShares)
	r.GET("/portfolio", h.GetPortfolio)
	r.POST("/redeem/:conditionID", h.RedeemShares)
}

func (h *Handler) GetMarkets(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	// Call Gamma API
	markets, err := h.gammaClient.GetActiveMarkets(limit)
	if err != nil {
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

func (h *Handler) BuyShares(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		OutcomeIndex int     `json:"outcomeIndex"`
		AmountUSDC   float64 `json:"amountUSDC"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	
	// Interact with Trader/ContractClient here
	c.JSON(200, gin.H{
		"status":  "success",
		"message": "Buy order placed (mock)",
		"market":  id,
		"amount":  req.AmountUSDC,
	})
}

func (h *Handler) SellShares(c *gin.Context) {
	c.JSON(200, gin.H{"status": "success", "message": "Sell order placed (mock)"})
}

func (h *Handler) GetPortfolio(c *gin.Context) {
	c.JSON(200, gin.H{
		"positions": []interface{}{},
		"balance":   1000.0,
	})
}

func (h *Handler) RedeemShares(c *gin.Context) {
	conditionID := c.Param("conditionID")
	c.JSON(200, gin.H{
		"status":      "success",
		"message":     "Redemption triggered (mock)",
		"conditionID": conditionID,
	})
}
