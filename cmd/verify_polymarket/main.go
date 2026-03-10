package main

import (
	"fmt"
	"nofx/kernel"
	"nofx/store"
	"nofx/trader/polymarket"
	"time"
)

func main() {
	fmt.Println("=== 1. 测试 Polymarket 交易器连接 (Trader) ===")
	// 使用 Polygon RPC
	trader, err := polymarket.NewPolymarketTrader(
		"daa5fb99955efc04dcc1cc6527c73e2dc2c0d3dd18cd9d7daf5be887231c1b40",
		"0x1969A5026a5A1e13F2Ee0e4ed69a0f9BA94BC836",
		"https://polygon-rpc.com",
	)

	if err != nil {
		fmt.Printf("❌ 创建交易器失败: %v\n", err)
	} else {
		fmt.Println("✅ Polymarket 交易器创建成功")
		// 测试余额查询 (可能失败，取决于网络)
		balance, err := trader.GetBalance()
		if err != nil {
			fmt.Printf("⚠️ 余额查询失败 (网络问题?): %v\n", err)
		} else {
			fmt.Printf("✅ 余额查询成功: %v\n", balance)
		}
	}

	fmt.Println("\n=== 2. 测试 Polymarket 策略引擎 (Strategy Engine) ===")
	// 配置策略
	strategyConfig := &store.StrategyConfig{
		StrategyType: "polymarket_hybrid",
		PolymarketConfig: &store.PolymarketStrategyConfig{
			MaxPositionUSDC:      1000,
			ProbabilityThreshold: 0.05,
			MinLiquidity:         1000,
			MinDaysToExpiry:      1,
			MaxDaysToExpiry:      30,
			SubStrategies: []store.PolymarketSubStrategy{
				{Name: "probability_arbitrage", Weight: 1.0, Enabled: true},
			},
		},
		Indicators: store.IndicatorConfig{
			NofxOSAPIKey: "test_key",
		},
	}

	// 初始化引擎
	engine := kernel.NewStrategyEngine(strategyConfig)
	if engine.IsPolymarketStrategy() {
		fmt.Println("✅ 策略引擎正确识别为 Polymarket 类型")
	} else {
		fmt.Println("❌ 策略引擎未识别为 Polymarket 类型")
		return
	}

	// 模拟市场数据
	marketData := &kernel.MarketData{
		ID:           "test_market_1",
		Question:     "Will BTC hit 100k by 2025?",
		YesPrice:     0.60, // 市场价格 0.60
		NoPrice:      0.40,
		Liquidity:    50000,
		DaysToExpiry: 15,
		EndDate:      time.Now().AddDate(0, 0, 15),
	}
	
	// 执行策略 (假设公平概率为 0.50)
	// 0.60 (Implied) > 0.50 (Fair) + 0.05 (Threshold) -> 差值 0.10 > 0.05 -> Implied > Fair -> Overvalued -> BUY NO
	// 但是逻辑是: diff = Fair - Implied = 0.50 - 0.60 = -0.10
	// diff < -0.05 -> BUY NO
	decision := engine.GetPolymarketDecision(marketData)
	
	fmt.Printf("输入市场数据: YesPrice=%.2f, Liquidity=%.0f\n", marketData.YesPrice, marketData.Liquidity)
	fmt.Printf("策略决策: Action=%s, Confidence=%.2f\n", decision.Action, decision.Confidence)
	fmt.Printf("决策理由: %s\n", decision.Reasoning)

	if decision.Action == "BUY_NO" {
		fmt.Println("✅ 策略逻辑验证通过: 正确识别高估机会 (BUY_NO)")
	} else {
		fmt.Println("❌ 策略逻辑验证失败: 预期 BUY_NO")
	}
}
