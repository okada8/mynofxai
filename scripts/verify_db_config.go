package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"nofx/crypto"
	"nofx/store"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// 1. 加载 .env
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// 2. 初始化 Crypto 服务
	cs, err := crypto.NewCryptoService()
	if err != nil {
		log.Fatalf("❌ 初始化加密服务失败: %v", err)
	}
	crypto.SetGlobalCryptoService(cs)
	fmt.Println("🔐 加密服务已初始化")

	// 3. 连接数据库
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSLMODE"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("❌ 连接数据库失败: %v", err)
	}
	fmt.Println("💾 数据库连接成功")

	// 4. 查询交易所配置
	var exchanges []store.Exchange
	if err := db.Find(&exchanges).Error; err != nil {
		log.Fatalf("❌ 查询交易所配置失败: %v", err)
	}

	fmt.Printf("🔍 找到 %d 个交易所配置\n", len(exchanges))

	foundBinance := false
	for _, ex := range exchanges {
		if ex.ExchangeType == "binance" {
			foundBinance = true
			fmt.Printf("\n--------------------------------------------------\n")
			fmt.Printf("🏦 正在测试交易所: %s (账户: %s)\n", ex.Name, ex.AccountName)
			fmt.Printf("   Testnet: %v\n", ex.Testnet)
			
			// GORM 的 EncryptedString 已经自动解密了
			apiKey := string(ex.APIKey)
			secretKey := string(ex.SecretKey)

			if apiKey == "" || secretKey == "" {
				fmt.Println("⚠️  API Key 或 Secret Key 为空，跳过")
				continue
			}

			// 掩码显示 Key
			maskedKey := apiKey
			if len(apiKey) > 8 {
				maskedKey = apiKey[:4] + "..." + apiKey[len(apiKey)-4:]
			}
			fmt.Printf("   API Key: %s\n", maskedKey)

			// 5. 连接币安
			client := futures.NewClient(apiKey, secretKey)
			if ex.Testnet {
				client.BaseURL = "https://testnet.binancefuture.com"
				fmt.Println("   🌐 使用测试网地址")
			} else {
				fmt.Println("   🌐 使用实盘地址")
			}

			// 6. 获取余额
			acc, err := client.NewGetAccountService().Do(context.Background())
			if err != nil {
				fmt.Printf("❌ 获取账户信息失败: %v\n", err)
			} else {
				fmt.Printf("✅ 连接成功！\n")
				fmt.Printf("   💰 总钱包余额: %s USDT (仅展示总额)\n", acc.TotalWalletBalance)
				fmt.Printf("   💰 可用余额: %s USDT\n", acc.AvailableBalance)
			}
		}
	}

	if !foundBinance {
		fmt.Println("\n⚠️  未找到币安 (Binance) 交易所配置")
	}
	fmt.Printf("\n--------------------------------------------------\n")
}
