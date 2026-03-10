package polymarket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type GammaClient struct {
	client  *http.Client
	baseURL string
}

func NewGammaClient() *GammaClient {
	return &GammaClient{
		client:  &http.Client{Timeout: 10 * time.Second},
		baseURL: "https://gamma-api.polymarket.com/query",
	}
}

type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

func (g *GammaClient) Query(query string, vars map[string]interface{}) ([]byte, error) {
	reqBody, err := json.Marshal(GraphQLRequest{
		Query:     query,
		Variables: vars,
	})
	if err != nil {
		return nil, err
	}

	resp, err := g.client.Post(g.baseURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// MarketInfo represents simplified market data from Gamma API
type MarketInfo struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	Question  string    `json:"question"`
	EndDate   time.Time `json:"endDate"`
	Liquidity float64   `json:"liquidity"`
	Volume24h float64   `json:"volume24h"`
	Outcomes  []Outcome `json:"outcomes"`
}

type Outcome struct {
	ID      string  `json:"id"`
	TokenID string  `json:"tokenId"`
	Price   float64 `json:"price"`
}

// GetActiveMarkets fetches active markets with liquidity > 1000
func (g *GammaClient) GetActiveMarkets(limit int) ([]MarketInfo, error) {
	query := `
    query GetActiveMarkets($limit: Int!) {
        markets(
            where: {
                closed: false
                liquidity_gt: 1000
            }
            orderBy: liquidity
            orderDirection: desc
            first: $limit
        ) {
            id
            slug
            question
            endDate
            liquidity
            volume24h
            outcomes {
                id
                tokenId
                price
            }
        }
    }`

	vars := map[string]interface{}{"limit": limit}
	data, err := g.Query(query, vars)
	if err != nil {
		return nil, err
	}

	// Parse response
	var response struct {
		Data struct {
			Markets []struct {
				ID        string      `json:"id"`
				Slug      string      `json:"slug"`
				Question  string      `json:"question"`
				EndDate   string      `json:"endDate"` // ISO string
				Liquidity interface{} `json:"liquidity"`
				Volume24h interface{} `json:"volume24h"`
				Outcomes  []struct {
					ID      string      `json:"id"`
					TokenID string      `json:"tokenId"`
					Price   interface{} `json:"price"`
				} `json:"outcomes"`
			} `json:"markets"`
		} `json:"data"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var markets []MarketInfo
	for _, m := range response.Data.Markets {
		// Parse date
		// Note: Gamma API format might vary, handling basic ISO
		// endDate is often string in JSON response
		// We'll use a helper or just skip complex parsing for brevity if needed
		// For now assuming ISO 8601
		// endDateStr := m.EndDate
		// t, _ := time.Parse(time.RFC3339, endDateStr)

		// Parse float fields safely (they might be strings in GraphQL response)
		liq := parseFloat(m.Liquidity)
		vol := parseFloat(m.Volume24h)

		outcomes := make([]Outcome, len(m.Outcomes))
		for i, o := range m.Outcomes {
			outcomes[i] = Outcome{
				ID:      o.ID,
				TokenID: o.TokenID,
				Price:   parseFloat(o.Price),
			}
		}

		markets = append(markets, MarketInfo{
			ID:        m.ID,
			Slug:      m.Slug,
			Question:  m.Question,
			// EndDate:   t,
			Liquidity: liq,
			Volume24h: vol,
			Outcomes:  outcomes,
		})
	}

	return markets, nil
}

// Helper to parse float from interface{} (string or number)
func parseFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case string:
		var f float64
		fmt.Sscanf(val, "%f", &f)
		return f
	default:
		return 0
	}
}
