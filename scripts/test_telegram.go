package main

import (
	"log"
	"nofx/config"
	"nofx/notification"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	// Try loading from parent directory (if running from scripts/)
	if err := godotenv.Load("../.env"); err != nil {
		// Try loading from current directory (if running from root)
		if err := godotenv.Load(); err != nil {
             log.Printf("Warning: Error loading .env file: %v", err)
        }
	}

	// Initialize config
	config.Init()

	log.Println("Testing Telegram Notification...")
    // Mask token for security in logs
    token := config.Get().TelegramBotToken
    if len(token) > 10 {
        token = token[:5] + "..." + token[len(token)-5:]
    }
	log.Printf("Bot Token: %s", token)
	log.Printf("Chat ID: %s", config.Get().TelegramChatID)

	err := notification.SendTelegramMessage("🔔 Hello! This is a test message from NOFX Terminal script.")
	if err != nil {
		log.Fatalf("❌ Failed to send message: %v", err)
	}

	log.Println("✅ Message sent successfully!")
}
