package kernel

import (
	"testing"

	"nofx/store"
)

func TestNewStrategyEngine_Polymarket(t *testing.T) {
	// Create a mock config with Polymarket enabled
	config := &store.StrategyConfig{
		StrategyType: "polymarket_hybrid",
		PolymarketConfig: &store.PolymarketStrategyConfig{
			MaxPositionUSDC:      1000,
			ProbabilityThreshold: 0.05,
			SubStrategies: []store.PolymarketSubStrategy{
				{Name: "probability_arbitrage", Weight: 1.0, Enabled: true},
			},
		},
		Indicators: store.IndicatorConfig{
			// Minimal valid config
			NofxOSAPIKey: "test_key",
		},
	}

	engine := NewStrategyEngine(config)

	if engine == nil {
		t.Fatal("Expected engine to be created")
	}

	if !engine.IsPolymarketStrategy() {
		t.Error("Expected IsPolymarketStrategy to return true")
	}

	if engine.polymarketEngine == nil {
		t.Error("Expected polymarketEngine to be initialized")
	}

	// Test decision generation
	marketData := &MarketData{
		ID:           "test_market",
		Question:     "Will BTC hit 100k?",
		YesPrice:     0.6, // Implied 60%
		NoPrice:      0.4,
		Liquidity:    10000,
		DaysToExpiry: 10,
	}

	// We didn't set min/max days in config, defaults might be 0.
	// Let's set them in config for the test
	config.PolymarketConfig.MinDaysToExpiry = 1
	config.PolymarketConfig.MaxDaysToExpiry = 30
	config.PolymarketConfig.MinLiquidity = 1000

	// Re-create engine with updated config (passed by reference so it might be updated, but let's be safe)
	engine = NewStrategyEngine(config)

	decision := engine.GetPolymarketDecision(marketData)
	if decision == nil {
		t.Fatal("Expected decision, got nil")
	}

	t.Logf("Decision: Action=%s, Confidence=%.2f, Reasoning=%s",
		decision.Action, decision.Confidence, decision.Reasoning)

	// Test GetFullDecisionWithStrategy rejection
	ctx := &Context{} // Empty context
	_, err := GetFullDecisionWithStrategy(ctx, nil, engine, "")
	if err == nil {
		t.Error("Expected error when calling GetFullDecisionWithStrategy for Polymarket strategy")
	} else {
		t.Logf("Correctly got error: %v", err)
	}
}
