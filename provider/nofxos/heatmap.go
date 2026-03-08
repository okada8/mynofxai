package nofxos

import (
	"encoding/json"
	"fmt"
	"log"
)

// HeatmapItem represents market depth data for a single coin
type HeatmapItem struct {
	Rank      int     `json:"rank"`
	Symbol    string  `json:"symbol"`
	BidVolume float64 `json:"bid_volume"`
	AskVolume float64 `json:"ask_volume"`
	Delta     float64 `json:"delta"`
}

// HeatmapResponse is the API response structure for heatmap data
type HeatmapResponse struct {
	Success bool `json:"success"`
	Code    int  `json:"code"`
	Data    struct {
		Heatmaps []HeatmapItem `json:"heatmaps"`
		Trade    string        `json:"trade"`
		Limit    int           `json:"limit"`
	} `json:"data"`
}

// GetHeatmap retrieves market heatmap data
func (c *Client) GetHeatmap(tradeType string, limit int) ([]HeatmapItem, error) {
	if tradeType == "" {
		tradeType = "future"
	}
	if limit <= 0 {
		limit = 20
	}

	endpoint := fmt.Sprintf("/api/heatmap/list?trade=%s&limit=%d", tradeType, limit)

	body, err := c.doRequest(endpoint)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	var response HeatmapResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("JSON parsing failed: %w", err)
	}

	if !response.Success && response.Code != 0 {
		return nil, fmt.Errorf("API returned error code: %d", response.Code)
	}

	log.Printf("✓ Fetched %s heatmap data: %d items", tradeType, len(response.Data.Heatmaps))

	return response.Data.Heatmaps, nil
}
