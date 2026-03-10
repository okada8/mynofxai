package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

// BridgeRequest represents the request format for the bridge
type BridgeRequest struct {
	Command   string  `json:"command"`
	Key       string  `json:"key,omitempty"`
	ChainID   int     `json:"chain_id,omitempty"`
	TokenID   string  `json:"token_id,omitempty"`
	Side      string  `json:"side,omitempty"`
	Price     float64 `json:"price,omitempty"`
	Size      float64 `json:"size,omitempty"`
	Timestamp int64   `json:"timestamp,omitempty"`
}

func main() {
	fmt.Println("=== Polymarket Python Bridge Test ===")

	// 1. Check if python script exists
	scriptPath := "trader/polymarket/polymarket_bridge.py"
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		fmt.Printf("❌ Error: Script not found at %s\n", scriptPath)
		return
	}
	fmt.Printf("✅ Script found: %s\n", scriptPath)

	// 2. Start Python process
	cmd := exec.Command("python3", scriptPath)
	
	// Create pipes
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Printf("❌ Failed to create stdin pipe: %v\n", err)
		return
	}
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("❌ Failed to create stdout pipe: %v\n", err)
		return
	}
	
	// Capture stderr for debugging
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		fmt.Printf("❌ Failed to start python process: %v\n", err)
		return
	}
	defer func() {
		cmd.Process.Kill()
	}()

	fmt.Println("✅ Python process started")

	// Helper to send command and read response
	sendAndReceive := func(req BridgeRequest) {
		// Send
		reqBytes, _ := json.Marshal(req)
		fmt.Printf("\n📤 Sending: %s\n", string(reqBytes))
		io.WriteString(stdin, string(reqBytes)+"\n")

		// Read response
		// Use a channel for timeout
		done := make(chan string)
		go func() {
			buf := make([]byte, 1024*10) // 10KB buffer
			n, err := stdout.Read(buf)
			if err != nil {
				fmt.Printf("❌ Read error: %v\n", err)
				return
			}
			done <- string(buf[:n])
		}()

		select {
		case resp := <-done:
			fmt.Printf("📥 Received: %s\n", strings.TrimSpace(resp))
		case <-time.After(5 * time.Second):
			fmt.Println("❌ Timeout waiting for response")
		}
	}

	// 3. Test Ping
	sendAndReceive(BridgeRequest{
		Command:   "ping",
		Timestamp: time.Now().Unix(),
	})

	// 4. Test Init (Mock Key for testing structure, will fail on actual connection if invalid)
	// Using a dummy private key (don't use real funds here!)
	dummyKey := "0000000000000000000000000000000000000000000000000000000000000001"
	sendAndReceive(BridgeRequest{
		Command: "init",
		Key:     dummyKey,
		ChainID: 137,
	})

	// 5. Test Get Price (Mock Token ID)
	sendAndReceive(BridgeRequest{
		Command: "get_price",
		TokenID: "0x1234567890abcdef1234567890abcdef12345678", // Dummy Token ID
	})
	
	fmt.Println("\n=== Test Finished ===")
}
