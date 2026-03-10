package main

import (
	"fmt"
	"log"
	"os"

	"nofx/trader/polymarket"
)

func main() {
	// Credentials provided by user
	walletAddr := "0x1969A5026a5A1e13F2Ee0e4ed69a0f9BA94BC836"
	privateKey := "daa5fb99955efc04dcc1cc6527c73e2dc2c0d3dd18cd9d7daf5be887231c1b40"
	// rpcURL := "https://polygon-rpc.com"
	rpcURL := "https://polygon-bor-rpc.publicnode.com"

    // L2 Credentials
    apiKey := "019cd7ab-3aee-7341-8494-57901e39e8c9"
    secret := "M1nydm87ipheIdeK95uXT0wk63gy6D9pslNw7GeOOEQ="
    passphrase := "a3ff10ef9658550e5718389d8f8be3107902e5ebd1f7c8904ff0242e5dfc04c4"

	fmt.Printf("🚀 Starting Polymarket Trader Test\n")
	fmt.Printf("Wallet Address: %s\n", walletAddr)
	
	// Create trader instance
	trader, err := polymarket.NewPolymarketTrader(privateKey, walletAddr, rpcURL, apiKey, secret, passphrase)
	if err != nil {
		log.Fatalf("❌ Failed to create trader: %v", err)
	}
	fmt.Printf("✅ Trader instance created successfully\n")

	// Verify Python Bridge connection
	// Access private field via reflection or check output logs from NewPolymarketTrader
	// Since we can't easily access private fields in main, we rely on the console output of NewPolymarketTrader

	// Test GetBalance
	fmt.Println("\n💰 Testing GetBalance...")
	balance, err := trader.GetBalance()
	if err != nil {
		fmt.Printf("❌ GetBalance failed: %v\n", err)
	} else {
		fmt.Printf("✅ Balance: %+v\n", balance)
	}

	// Test GetPositions
	fmt.Println("\n📊 Testing GetPositions...")
	positions, err := trader.GetPositions()
	if err != nil {
		fmt.Printf("❌ GetPositions failed: %v\n", err)
	} else {
		fmt.Printf("✅ Positions: %d found\n", len(positions))
		for i, pos := range positions {
			fmt.Printf("  %d. %+v\n", i+1, pos)
		}
	}

    // Test GetEvents (Market Data)
    fmt.Println("\n📈 Testing GetEvents (Active Markets)...")
    events, err := trader.GetEvents("", 5)
    if err != nil {
        fmt.Printf("❌ GetEvents failed: %v\n", err)
    } else {
        fmt.Printf("✅ Active Markets: %d found\n", len(events))
    }

	fmt.Println("\n✨ Test Complete")
    
    // Check for python script location
    // _, filename, _, _ := runtime.Caller(0)
    // dir := filepath.Dir(filename)
    // Adjust path to point to trader/polymarket/polymarket_bridge.py
    // Assuming we run this from project root
    scriptPath := "trader/polymarket/polymarket_bridge.py"
    if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
        fmt.Printf("⚠️  Warning: Python bridge script not found at %s. Please ensure you are running from project root.\n", scriptPath)
    }
}
