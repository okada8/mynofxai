package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Exchange struct {
	ID           string `gorm:"primaryKey"`
	ExchangeType string
	AccountName  string
	UserID       string
	Enabled      bool
}

type Trader struct {
	ID                string `gorm:"primaryKey"`
	Name              string
	UserID            string
	ExchangeID        string
	AIModelID         string
	ShowInCompetition bool
}

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSLMODE"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	var traders []Trader
	result := db.Find(&traders)
	if result.Error != nil {
		log.Printf("Error querying traders: %v", result.Error)
	}
	fmt.Printf("Traders Count: %d\n", len(traders))
	for _, t := range traders {
		fmt.Printf("Trader: %s (ID: %s), UserID: %s, ExchangeID: %s, AIModelID: %s, ShowInCompetition: %v\n",
			t.Name, t.ID, t.UserID, t.ExchangeID, t.AIModelID, t.ShowInCompetition)
	}

	var exchanges []Exchange
	result = db.Find(&exchanges)
	if result.Error != nil {
		log.Printf("Error querying exchanges: %v", result.Error)
	}
	fmt.Printf("\nExchanges Count: %d\n", len(exchanges))
	for _, e := range exchanges {
		fmt.Printf("Exchange: %s (Type: %s), ID: %s, UserID: %s, AccountName: %s, Enabled: %v\n",
			e.ExchangeType, e.ExchangeType, e.ID, e.UserID, e.AccountName, e.Enabled)
	}
}
