package main

import (
	"fmt"
	"nofx/trader/polymarket"
)

func main() {
	fmt.Println("🧪 测试Polymarket交易器初始化...")
	
	// 注意：使用虚拟私钥进行测试，不会实际连接网络
	// 在实际使用中，需要真实的私钥、地址和RPC URL
	testPrivateKey := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef" // 64字符虚拟私钥
	testWalletAddr := "0x742d35Cc6634C0532925a3b844Bc9e90F4b4e1e1" // 虚拟地址
	testRPCURL := "https://polygon-rpc.com" // 公共RPC，可能速率受限
	
	fmt.Printf("使用配置:\n")
	fmt.Printf("  私钥: %s...\n", testPrivateKey[:20])
	fmt.Printf("  地址: %s\n", testWalletAddr)
	fmt.Printf("  RPC: %s\n", testRPCURL)
	
	// 尝试创建交易器
	trader, err := polymarket.NewPolymarketTrader(testPrivateKey, testWalletAddr, testRPCURL)
	if err != nil {
		fmt.Printf("❌ 创建交易器失败: %v\n", err)
		fmt.Println("\n可能的原因:")
		fmt.Println("1. Python依赖未安装: pip install py-clob-client web3 eth-account")
		fmt.Println("2. RPC URL无法访问")
		fmt.Println("3. 私钥格式错误")
		fmt.Println("4. 合约地址可能过时")
		return
	}
	
	fmt.Println("✅ Polymarket交易器创建成功")
	
	// 尝试获取余额（可能会失败，因为使用虚拟私钥）
	fmt.Println("\n🧪 测试余额查询...")
	balance, err := trader.GetBalance()
	if err != nil {
		fmt.Printf("⚠️  余额查询失败（预期中，因为使用虚拟私钥）: %v\n", err)
		fmt.Println("   这通常是正常的，因为虚拟私钥无法访问真实钱包")
	} else {
		fmt.Printf("✅ 余额查询成功:\n")
		for k, v := range balance {
			fmt.Printf("  %s: %v\n", k, v)
		}
	}
	
	fmt.Println("\n🎯 初始化测试完成")
	fmt.Println("\n下一步:")
	fmt.Println("1. 使用真实的Polygon钱包私钥和地址")
	fmt.Println("2. 考虑使用Polygon Mumbai测试网进行开发:")
	fmt.Println("   RPC URL: https://rpc-mumbai.maticvigil.com")
	fmt.Println("3. 从Polygon水龙头获取测试网MATIC")
	fmt.Println("4. 使用测试网USDC进行交易测试")
}