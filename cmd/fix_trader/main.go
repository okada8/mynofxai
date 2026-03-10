package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Trader struct {
	ID                string `gorm:"primaryKey"`
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

	// Update the existing trader to show in competition
	result := db.Model(&Trader{}).Where("id = ?", "e824725b_562011f9-7d1d-4b2e-8c38-41693f8eec20_deepseek_1772984606").Update("show_in_competition", true)
	if result.Error != nil {
		log.Fatalf("Failed to update trader: %v", result.Error)
	}
	
	fmt.Printf("Updated %d trader(s) to show in competition\n", result.RowsAffected)
}
