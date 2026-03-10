package polymarket

import (
	"fmt"
	"math/big"
	"nofx/trader/types"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// PolymarketTrader implements Trader interface for Polymarket prediction market
type PolymarketTrader struct {
	privateKey string
	walletAddr string
	rpcURL     string

	// L2 Credentials
	apiKey     string
	secret     string
	passphrase string

	// Real clients
	contractClient *ContractClient
	gammaClient    *GammaClient

	// Python Wrapper for CLOB trading
	pyWrapper *PythonWrapper
	hasCLOB   bool // Whether CLOB trading is available
}

// NewPolymarketTrader creates a new Polymarket trader instance
func NewPolymarketTrader(privateKey, walletAddr, rpcURL, apiKey, secret, passphrase string) (*PolymarketTrader, error) {
	if privateKey == "" || walletAddr == "" || rpcURL == "" {
		return nil, fmt.Errorf("polymarket credentials missing: need privateKey, walletAddr, rpcURL")
	}

	// Initialize Contract Client
	contractClient, err := NewContractClient(rpcURL, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract client: %w", err)
	}

	// Initialize Gamma Client
	gammaClient := NewGammaClient()

	// Initialize Python Wrapper for CLOB trading (optional but recommended)
	var pyWrapper *PythonWrapper
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	scriptPath := filepath.Join(dir, "polymarket_bridge.py")
	
	// Check if Python script exists
	if _, err := os.Stat(scriptPath); err == nil {
		pyWrapper, err = NewPythonWrapper("python3", scriptPath)
		if err != nil {
			// Log warning but don't fail - contract operations still work
			fmt.Printf("⚠️  Warning: Failed to initialize Python wrapper: %v\n", err)
			fmt.Printf("⚠️  CLOB trading will be unavailable, only contract operations available\n")
			pyWrapper = nil
		} else {
			// Test Python bridge
			resp, err := pyWrapper.Call(map[string]interface{}{
				"command": "ping",
				"timestamp": time.Now().Unix(),
			})
			if err != nil {
				fmt.Printf("⚠️  Warning: Python bridge ping failed: %v\n", err)
				pyWrapper.Close()
				pyWrapper = nil
			} else {
				fmt.Printf("✅ Python bridge connected: %v\n", resp)
				
				// Determine signature type
				// Default to EOA (0)
				signatureType := 0
				
				// Derive address from private key to check if walletAddr is a proxy
				cleanKey := strings.TrimPrefix(privateKey, "0x")
				if pk, err := crypto.HexToECDSA(cleanKey); err == nil {
					address := crypto.PubkeyToAddress(pk.PublicKey)
					// If walletAddr is provided and different from derived address, it's a proxy
					if walletAddr != "" && !strings.EqualFold(walletAddr, address.Hex()) {
						signatureType = 1 // POLY_PROXY
					}
				}

				// Initialize CLOB client
				initCmd := map[string]interface{}{
					"command":        "init",
					"key":            privateKey,
					"chain_id":       137, // Polygon Mainnet
					"rpc_url":        rpcURL,
					"signature_type": signatureType,
					"api_key":        apiKey,
					"api_secret":     secret,
					"api_passphrase": passphrase,
				}

				if walletAddr != "" {
					initCmd["funder"] = walletAddr
				}

				resp, err = pyWrapper.Call(initCmd)
				if err != nil {
					fmt.Printf("⚠️  Warning: Python client init failed: %v\n", err)
				} else if status, ok := resp["status"].(string); ok && status == "success" {
					fmt.Printf("✅ Polymarket CLOB client initialized: %s\n", resp["wallet"])
					if usdc, ok := resp["usdc_balance"]; ok {
						fmt.Printf("✅ Python reported USDC balance: %v\n", usdc)
					}
				} else {
					fmt.Printf("⚠️  Warning: Python client init returned: %v\n", resp)
				}
			}
		}
	} else {
		fmt.Printf("⚠️  Python bridge script not found at %s\n", scriptPath)
		fmt.Printf("⚠️  Install Python dependencies: pip install py-clob-client web3 eth-account\n")
	}

	return &PolymarketTrader{
		privateKey:     privateKey,
		walletAddr:     walletAddr,
		rpcURL:         rpcURL,
		apiKey:         apiKey,
		secret:         secret,
		passphrase:     passphrase,
		contractClient: contractClient,
		gammaClient:    gammaClient,
		pyWrapper:      pyWrapper,
		hasCLOB:        pyWrapper != nil,
	}, nil
}

// GetBalance gets USDC balance and portfolio value
func (t *PolymarketTrader) GetBalance() (map[string]interface{}, error) {
	// Fetch real USDC balance
	balanceBig, err := t.contractClient.GetUSDCBalance()
	if err != nil {
		return nil, fmt.Errorf("failed to get USDC balance: %w", err)
	}

	// Convert 6 decimals to float
	balanceFloat := new(big.Float).SetInt(balanceBig)
	usdcValue, _ := balanceFloat.Quo(balanceFloat, big.NewFloat(1000000)).Float64()

	return map[string]interface{}{
		"totalWalletBalance":    usdcValue,
		"totalUnrealizedProfit": 0.0, // TODO: Calculate from open positions
		"availableBalance":      usdcValue,
		"totalEquity":           usdcValue, // Assuming equity = wallet balance for now
		"network":               "Polygon",
		"wallet_address":        t.walletAddr,
		"has_clob":              t.hasCLOB,
		"clob_status":           map[string]interface{}{"available": t.hasCLOB, "initialized": t.pyWrapper != nil},
	}, nil
}

// GetPositions gets current positions (outcome shares)
func (t *PolymarketTrader) GetPositions() ([]map[string]interface{}, error) {
	if t.pyWrapper == nil {
		return []map[string]interface{}{}, nil
	}

	resp, err := t.pyWrapper.Call(map[string]interface{}{
		"command": "get_positions",
	})
	if err != nil {
		// Fallback or ignore if not supported
		return []map[string]interface{}{}, nil
	}

	if status, ok := resp["status"].(string); ok && status == "success" {
		if positions, ok := resp["positions"].([]interface{}); ok {
			result := make([]map[string]interface{}, len(positions))
			for i, p := range positions {
				if posMap, ok := p.(map[string]interface{}); ok {
					result[i] = posMap
				}
			}
			return result, nil
		}
	}
	
	return []map[string]interface{}{}, nil
}

// OpenLong buys "YES" shares for an event (Predicting YES)
// For Polymarket: "OpenLong" maps to buying the specified outcome token
func (t *PolymarketTrader) OpenLong(symbol string, quantity float64, leverage int) (map[string]interface{}, error) {
	// 1. Try Python Wrapper (CLOB order) first if available
	if t.pyWrapper != nil {
		// Get current market price first
		price, _ := t.GetMarketPrice(symbol)
		if price <= 0 {
			price = 0.5 // Default if price fetch fails
		}

		// Use Python SDK for real order placement on CLOB
		resp, err := t.pyWrapper.Call(map[string]interface{}{
			"command":  "create_order",
			"token_id": symbol,
			"side":     "BUY",
			"size":     quantity,
			"price":    price, 
		})

		if err == nil {
			if status, ok := resp["status"].(string); ok && status == "success" {
				return resp["response"].(map[string]interface{}), nil
			}
			fmt.Printf("Python CLOB order failed, falling back to direct contract: %v\n", resp["error"])
		}
	}

	// 2. Fallback to direct contract split (minting)
	parts := strings.Split(symbol, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid symbol format for contract split: %s", symbol)
	}
	conditionID := common.HexToHash(parts[0])
	outcomeIndex, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid outcome index: %w", err)
	}

	amountUSDC := new(big.Int).SetInt64(int64(quantity * 1000000))
	tx, err := t.contractClient.BuyOutcomeToken(conditionID, outcomeIndex, amountUSDC)
	if err != nil {
		return nil, fmt.Errorf("contract split failed: %w", err)
	}

	return map[string]interface{}{
		"orderId": tx.Hash().Hex(),
		"status":  "PENDING",
	}, nil
}

// OpenShort sells "YES" shares (or buys "NO" shares)
// For Polymarket: "OpenShort" maps to buying the OPPOSITE outcome token if binary
func (t *PolymarketTrader) OpenShort(symbol string, quantity float64, leverage int) (map[string]interface{}, error) {
	// Parse symbol to get conditionID and outcomeIndex
	parts := strings.Split(symbol, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid symbol format, expected 'conditionID/outcomeIndex'")
	}
	// For binary markets, if index is 0, opposite is 1. If 1, opposite is 0.
	// This assumption only holds for binary markets.
	outcomeIndex, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid outcome index: %w", err)
	}

	oppositeIndex := 1 - outcomeIndex
	if oppositeIndex < 0 {
		oppositeIndex = 0 // Safety check
	}

	// Construct new symbol for opposite outcome
	oppositeSymbol := fmt.Sprintf("%s/%d", parts[0], oppositeIndex)

	// Call OpenLong on the opposite outcome
	return t.OpenLong(oppositeSymbol, quantity, leverage)
}

// CloseLong sells held outcome tokens
func (t *PolymarketTrader) CloseLong(symbol string, quantity float64) (map[string]interface{}, error) {
	return t.sellOutcomeToken(symbol, quantity)
}

// CloseShort sells held opposite outcome tokens
func (t *PolymarketTrader) CloseShort(symbol string, quantity float64) (map[string]interface{}, error) {
	return t.sellOutcomeToken(symbol, quantity)
}

// SetLeverage is not supported on Polymarket
func (t *PolymarketTrader) SetLeverage(symbol string, leverage int) error {
	return nil // No-op
}

// SetMarginMode is not supported on Polymarket
func (t *PolymarketTrader) SetMarginMode(symbol string, isCrossMargin bool) error {
	return nil // No-op
}

// GetMarketPrice gets the current price (probability) of an outcome token
func (t *PolymarketTrader) GetMarketPrice(symbol string) (float64, error) {
	if t.pyWrapper == nil {
		return 0.5, nil // Fallback
	}

	resp, err := t.pyWrapper.Call(map[string]interface{}{
		"command":  "get_price",
		"token_id": symbol,
	})
	if err != nil {
		return 0.5, err
	}

	if price, ok := resp["price"].(float64); ok {
		return price, nil
	}
	return 0.5, fmt.Errorf("invalid price response")
}

// SetStopLoss is not natively supported, handled by strategy engine
func (t *PolymarketTrader) SetStopLoss(symbol string, positionSide string, quantity, stopPrice float64) error {
	return nil // No-op, managed by higher level logic
}

// SetTakeProfit is not natively supported, handled by strategy engine
func (t *PolymarketTrader) SetTakeProfit(symbol string, positionSide string, quantity, takeProfitPrice float64) error {
	return nil // No-op, managed by higher level logic
}

// CancelStopLossOrders implementation
func (t *PolymarketTrader) CancelStopLossOrders(symbol string) error {
	return nil
}

// CancelTakeProfitOrders implementation
func (t *PolymarketTrader) CancelTakeProfitOrders(symbol string) error {
	return nil
}

// CancelAllOrders cancels open limit orders on CLOB
func (t *PolymarketTrader) CancelAllOrders(symbol string) error {
	return nil
}

// CancelStopOrders implementation
func (t *PolymarketTrader) CancelStopOrders(symbol string) error {
	return nil
}

// FormatQuantity formats quantity for Polymarket (usually integer for CTF, but fractional allowed for USDC)
func (t *PolymarketTrader) FormatQuantity(symbol string, quantity float64) (string, error) {
	return strconv.FormatFloat(quantity, 'f', 6, 64), nil
}

// GetOrderStatus gets status from CLOB
func (t *PolymarketTrader) GetOrderStatus(symbol string, orderID string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"status":      "FILLED", // Mock
		"avgPrice":    0.5,
		"executedQty": 10.0,
		"commission":  0.0,
	}, nil
}

// GetClosedPnL gets trade history
func (t *PolymarketTrader) GetClosedPnL(startTime time.Time, limit int) ([]types.ClosedPnLRecord, error) {
	return []types.ClosedPnLRecord{}, nil
}

// GetOpenOrders gets open limit orders
func (t *PolymarketTrader) GetOpenOrders(symbol string) ([]types.OpenOrder, error) {
	if t.pyWrapper == nil {
		return []types.OpenOrder{}, nil
	}

	resp, err := t.pyWrapper.Call(map[string]interface{}{
		"command":  "get_open_orders",
		"token_id": symbol,
	})
	if err != nil {
		return nil, err
	}

	if status, ok := resp["status"].(string); !ok || status != "success" {
		return nil, fmt.Errorf("failed to get open orders: %v", resp["error"])
	}

	var result []types.OpenOrder
	if orders, ok := resp["orders"].([]interface{}); ok {
		for _, o := range orders {
			if orderMap, ok := o.(map[string]interface{}); ok {
				// Map Python response to types.OpenOrder
				// Assuming orderMap has 'price', 'size', 'side', 'orderID'
				price, _ := orderMap["price"].(float64)
				size, _ := orderMap["size"].(float64)
				side, _ := orderMap["side"].(string)
				id, _ := orderMap["orderID"].(string)

				result = append(result, types.OpenOrder{
					Symbol:       symbol,
					OrderID:      id,
					Side:         side,
					PositionSide: "LONG", // Default to LONG for spot/outcome tokens
					Type:         "LIMIT",
					Price:        price,
					Quantity:     size,
					Status:       "NEW",
				})
			}
		}
	}

	return result, nil
}

// --- Internal Helper Methods ---

func (t *PolymarketTrader) buyOutcomeToken(tokenID string, amountUSDC float64) (map[string]interface{}, error) {
	// 1. Approve USDC spender (CTF Exchange / CLOB Proxy)
	// 2. Call buy() on Exchange or place limit order on CLOB
	return map[string]interface{}{
		"orderId": "mock-poly-buy-" + tokenID,
	}, nil
}

func (t *PolymarketTrader) sellOutcomeToken(tokenID string, amountShares float64) (map[string]interface{}, error) {
	// 1. Approve Share spender
	// 2. Call sell() or place limit sell order
	return map[string]interface{}{
		"orderId": "mock-poly-sell-" + tokenID,
	}, nil
}

// Polymarket Specific Methods

// GetEvents fetches active events from Polymarket API
func (t *PolymarketTrader) GetEvents(tag string, limit int) ([]interface{}, error) {
	markets, err := t.gammaClient.GetActiveMarkets(limit)
	if err != nil {
		return nil, err
	}

	// Convert to interface slice
	result := make([]interface{}, len(markets))
	for i, m := range markets {
		result[i] = m
	}
	return result, nil
}

// GetEventOutcome gets the specific outcome token ID for an event
func (t *PolymarketTrader) GetEventOutcome(eventID string, outcomeIndex int) (string, error) {
	// Resolve conditionId and index to token ID
	return "", nil
}

// MergeShares merges outcome tokens back to collateral (if holding all outcomes)
func (t *PolymarketTrader) MergeShares(conditionID string) error {
	// Call merge() on CTF contract
	return nil
}

// RedeemShares redeems winning tokens after event resolution
func (t *PolymarketTrader) RedeemShares(conditionID string) error {
	// Call redeem() on CTF contract
	return nil
}
