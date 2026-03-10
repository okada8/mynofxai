// /Users/tom/Desktop/nofi/trader/polymarket/trader_fixed.go
package polymarket

import (
	"fmt"
	"nofx/logger"
	"os"
	"path/filepath"
)

type PolymarketTraderFixed struct {
	privateKey    string
	walletAddr    string
	rpcURL        string
	pythonWrapper *PythonWrapper
	initialized   bool
}

func NewPolymarketTraderFixed(privateKey, walletAddr, rpcURL string) (*PolymarketTraderFixed, error) {
	// 查找Python脚本路径
	scriptPath := filepath.Join("polymarket_py", "client.py")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		// 尝试其他路径
		scriptPath = filepath.Join("..", "polymarket_py", "client.py")
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("Python client script not found at %s", scriptPath)
		}
	}

	// 启动Python包装器
	wrapper, err := NewPythonWrapper("python3", scriptPath)
	if err != nil {
		return nil, fmt.Errorf("failed to start Python wrapper: %w", err)
	}

	trader := &PolymarketTraderFixed{
		privateKey:    privateKey,
		walletAddr:    walletAddr,
		rpcURL:        rpcURL,
		pythonWrapper: wrapper,
	}

	// 初始化
	if err := trader.initialize(); err != nil {
		wrapper.Close()
		return nil, err
	}

	return trader, nil
}

func (t *PolymarketTraderFixed) initialize() error {
	response, err := t.pythonWrapper.Call(map[string]interface{}{
		"command":     "initialize",
		"private_key": t.privateKey,
		"chain_id":    137, // Polygon主网
	})

	if err != nil {
		return fmt.Errorf("Python initialization failed: %w", err)
	}

	if status, ok := response["status"].(string); !ok || status != "success" {
		if errMsg, ok := response["error"].(string); ok {
			return fmt.Errorf("Python client error: %s", errMsg)
		}
		return fmt.Errorf("unknown initialization error")
	}

	t.initialized = true
	logger.Infof("✅ Polymarket trader initialized: %v", response["wallet"])
	return nil
}

// 实现Trader接口
func (t *PolymarketTraderFixed) GetBalance() (map[string]interface{}, error) {
	if !t.initialized {
		return nil, fmt.Errorf("trader not initialized")
	}

	response, err := t.pythonWrapper.Call(map[string]interface{}{
		"command": "get_balance",
	})

	if err != nil {
		return nil, err
	}

	if status, ok := response["status"].(string); !ok || status != "success" {
		return nil, fmt.Errorf("balance query failed")
	}

	// 转换格式以匹配NOFX期望
	balanceData := response["balance"].(map[string]interface{})
	return map[string]interface{}{
		"totalWalletBalance":    balanceData["usdc"],
		"totalUnrealizedProfit": 0.0,
		"availableBalance":      balanceData["usdc"],
		"totalEquity":           balanceData["usdc"],
		"network":               "Polygon",
		"wallet_address":        t.walletAddr,
	}, nil
}

func (t *PolymarketTraderFixed) OpenLong(symbol string, quantity float64, leverage int) (map[string]interface{}, error) {
	// 解析symbol格式：conditionId/outcomeIndex
	// 例如：0x123.../0 (NO), 0x123.../1 (YES)

	// 获取token_id（需要从symbol或通过Gamma API查询）
	tokenID := t.getTokenIDFromSymbol(symbol)
	response, err := t.pythonWrapper.Call(map[string]interface{}{
		"command":  "buy_shares",
		"token_id": tokenID,
		"price":    0.50,                // 需要从市场获取当前价格
		"size":     int(quantity * 100), // 转换为份额数量
		"side":     "BUY",
	})

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"orderId": response["order_id"],
		"status":  response["status"],
	}, nil
}

func (t *PolymarketTraderFixed) getTokenIDFromSymbol(symbol string) string {
	// 简化：假设symbol就是token_id
	// 实际需要从Gamma API查询或建立映射
	return symbol
}

// ... 其他Trader接口方法类似实现
