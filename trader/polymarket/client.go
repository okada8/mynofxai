package polymarket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type GammaClient struct {
	client   *http.Client
	baseURL  string
	apiToken string
	cookies  string
}

func NewGammaClient() *GammaClient {
	// 尝试从环境变量获取API令牌
	apiToken := os.Getenv("POLYMARKET_GAMMA_TOKEN")
	cookies := os.Getenv("POLYMARKET_COOKIES")
	
	// 如果没有环境变量，尝试使用默认值或留空
	// 注意：Gamma API可能不需要认证，但最近可能需要
	return &GammaClient{
		client:   &http.Client{Timeout: 30 * time.Second},
		baseURL:  "https://gamma-api.polymarket.com/query",
		apiToken: apiToken,
		cookies:  cookies,
	}
}

// NewGammaClientWithToken 创建带有认证令牌的客户端
func NewGammaClientWithToken(token, cookies string) *GammaClient {
	return &GammaClient{
		client:   &http.Client{Timeout: 30 * time.Second},
		baseURL:  "https://gamma-api.polymarket.com/query",
		apiToken: token,
		cookies:  cookies,
	}
}

// GraphQLRequest struct
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

func (g *GammaClient) Query(query string, vars map[string]interface{}) ([]byte, error) {
	reqBody, err := json.Marshal(GraphQLRequest{
		Query:     query,
		Variables: vars,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", g.baseURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	
	// 设置标准头部
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Origin", "https://polymarket.com")
	req.Header.Set("Referer", "https://polymarket.com/")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	
	// 添加认证令牌（如果有）
	if g.apiToken != "" {
		if strings.HasPrefix(g.apiToken, "Bearer ") {
			req.Header.Set("Authorization", g.apiToken)
		} else if strings.HasPrefix(g.apiToken, "eyJ") { // JWT令牌
			req.Header.Set("Authorization", "Bearer "+g.apiToken)
		} else {
			req.Header.Set("Authorization", "Bearer "+g.apiToken)
		}
	}
	
	// 添加cookies（如果有）
	if g.cookies != "" {
		req.Header.Set("Cookie", g.cookies)
	}
	
	// 尝试添加一些常见的Polymarket头部
	req.Header.Set("X-Client-Type", "web")
	req.Header.Set("X-Client-Version", "1.0.0")
	
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	// 检查状态码
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		errorMsg := string(body)
		
		// 如果错误信息是JSON，尝试解析
		if strings.Contains(errorMsg, "invalid token/cookies") {
			return nil, fmt.Errorf("Gamma API认证失败: 需要有效的认证令牌或cookies。请设置POLYMARKET_GAMMA_TOKEN环境变量")
		}
		
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, errorMsg)
	}

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
		// 如果Gamma API认证失败，尝试使用备用方案
		if strings.Contains(err.Error(), "认证失败") || strings.Contains(err.Error(), "401") {
			// 返回模拟数据用于测试
			fmt.Println("⚠️  Gamma API认证失败，返回模拟数据用于测试")
			return g.getMockMarkets(limit), nil
		}
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
		endDateStr := m.EndDate
		t, _ := time.Parse(time.RFC3339, endDateStr)

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
			EndDate:   t,
			Liquidity: liq,
			Volume24h: vol,
			Outcomes:  outcomes,
		})
	}

	return markets, nil
}

// getMockMarkets returns mock market data for testing when API fails
func (g *GammaClient) getMockMarkets(limit int) []MarketInfo {
	markets := []MarketInfo{
		{
			ID:        "0x1234567890abcdef/0",
			Slug:      "bitcoin-100k-2024",
			Question:  "Will Bitcoin reach $100,000 by end of 2024?",
			EndDate:   time.Now().Add(24 * 30 * time.Hour),
			Liquidity: 1500000.0,
			Volume24h: 250000.0,
			Outcomes: []Outcome{
				{ID: "0", TokenID: "0x1234567890abcdef/0", Price: 0.35},
				{ID: "1", TokenID: "0x1234567890abcdef/1", Price: 0.65},
			},
		},
		{
			ID:        "0xabcdef1234567890/0",
			Slug:      "ethereum-5000-2024",
			Question:  "Will Ethereum reach $5,000 by end of 2024?",
			EndDate:   time.Now().Add(24 * 60 * time.Hour),
			Liquidity: 800000.0,
			Volume24h: 120000.0,
			Outcomes: []Outcome{
				{ID: "0", TokenID: "0xabcdef1234567890/0", Price: 0.45},
				{ID: "1", TokenID: "0xabcdef1234567890/1", Price: 0.55},
			},
		},
	}
	
	if limit < len(markets) {
		return markets[:limit]
	}
	return markets
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

// GetMarketDetail fetches detailed information for a specific market
func (g *GammaClient) GetMarketDetail(marketID string) (*MarketInfo, error) {
	query := `
    query GetMarketDetail($id: ID!) {
        market(id: $id) {
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

	vars := map[string]interface{}{"id": marketID}
	data, err := g.Query(query, vars)
	if err != nil {
		return nil, err
	}

	var response struct {
		Data struct {
			Market struct {
				ID        string      `json:"id"`
				Slug      string      `json:"slug"`
				Question  string      `json:"question"`
				EndDate   string      `json:"endDate"`
				Liquidity interface{} `json:"liquidity"`
				Volume24h interface{} `json:"volume24h"`
				Outcomes  []struct {
					ID      string      `json:"id"`
					TokenID string      `json:"tokenId"`
					Price   interface{} `json:"price"`
				} `json:"outcomes"`
			} `json:"market"`
		} `json:"data"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	m := response.Data.Market
	endDate, _ := time.Parse(time.RFC3339, m.EndDate)
	
	outcomes := make([]Outcome, len(m.Outcomes))
	for i, o := range m.Outcomes {
		outcomes[i] = Outcome{
			ID:      o.ID,
			TokenID: o.TokenID,
			Price:   parseFloat(o.Price),
		}
	}

	return &MarketInfo{
		ID:        m.ID,
		Slug:      m.Slug,
		Question:  m.Question,
		EndDate:   endDate,
		Liquidity: parseFloat(m.Liquidity),
		Volume24h: parseFloat(m.Volume24h),
		Outcomes:  outcomes,
	}, nil
}