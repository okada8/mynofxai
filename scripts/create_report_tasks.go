package main

import (
	"fmt"
	"log"
	"nofx/store"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	var db *gorm.DB
	var err error

	dbType := os.Getenv("DB_TYPE")
	if dbType == "postgres" {
		dsn := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_SSLMODE"),
		)
		fmt.Printf("Connecting to Postgres: host=%s port=%s dbname=%s\n",
			os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))

		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
	} else {
		// Fallback to SQLite logic (not used here but kept for reference)
		log.Fatal("Script only configured for Postgres based on .env analysis")
	}

	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Read Traders
	var traders []store.Trader
	if err := db.Find(&traders).Error; err != nil {
		log.Fatalf("failed to list traders: %v", err)
	}

	fmt.Printf("Found %d traders\n", len(traders))

	// Create Task for each trader
	taskStore := store.NewTaskStore(db)

	// Ensure table exists
	// Note: We access the unexported initTables via reflection or just trust it exists if we used AutoMigrate
	// Since we are in main package and cannot access unexported methods easily without modification
	// We will rely on GORM AutoMigrate here directly for the Task model
	if err := db.AutoMigrate(&store.Task{}); err != nil {
		log.Printf("Init tables error: %v", err)
	}

	for _, t := range traders {
		// Check if task already exists
		var existingTask store.Task
		err := db.Where("trader_id = ? AND type = ?", t.ID, "report").First(&existingTask).Error
		if err == nil {
			fmt.Printf("Task already exists for trader %s (%s), skipping...\n", t.Name, t.ID)
			continue
		}

		task := &store.Task{
			ID:             uuid.New().String(),
			Type:           "report",
			Name:           fmt.Sprintf("%s Hourly Report", t.Name),
			Description:    "Auto-generated hourly position report",
			TraderID:       t.ID,
			CronExpression: "@hourly",
			Enabled:        true,
			Params:         "{}",
			CreatedAt:      time.Now().UnixMilli(),
			UpdatedAt:      time.Now().UnixMilli(),
		}

		if err := taskStore.Create(task); err != nil {
			fmt.Printf("Failed to create task for %s: %v\n", t.Name, err)
		} else {
			fmt.Printf("✅ Created hourly report task for trader: %s\n", t.Name)
		}
	}
}
