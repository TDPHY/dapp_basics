package main

import (
	"fmt"
	"log"

	"github.com/local/go-eth-demo/config"
	"github.com/local/go-eth-demo/utils"
)

func main() {
	fmt.Println("🚀 以太坊客户端连接示例")
	fmt.Println("================================")

	// 1. 加载配置
	fmt.Println("📋 加载配置...")
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ 配置加载失败: %v", err)
	}
	fmt.Printf("✅ 配置加载成功: %s\n", cfg.GetNetworkInfo())

	// 2. 创建以太坊客户端
	fmt.Println("\n🔗 连接以太坊节点...")
	client, err := utils.NewEthClient(cfg)
	if err != nil {
		log.Fatalf("❌ 连接失败: %v", err)
	}
	defer client.Close()
	fmt.Println("✅ 连接成功!")

	// 3. 验证网络
	fmt.Println("\n🔍 验证网络...")
	if err := client.VerifyNetwork(); err != nil {
		log.Fatalf("❌ 网络验证失败: %v", err)
	}
	fmt.Println("✅ 网络验证通过!")

	// 4. 获取连接信息
	fmt.Println("\n📊 获取网络信息...")
	info, err := client.GetConnectionInfo()
	if err != nil {
		log.Fatalf("❌ 获取网络信息失败: %v", err)
	}

	// 5. 显示详细信息
	fmt.Println("\n🌐 网络详细信息:")
	fmt.Println("--------------------------------")
	for key, value := range info {
		fmt.Printf("%-15s: %v\n", key, value)
	}

	// 6. 获取最新区块号
	fmt.Println("\n📦 最新区块信息:")
	fmt.Println("--------------------------------")
	blockNumber, err := client.GetLatestBlockNumber()
	if err != nil {
		log.Printf("❌ 获取区块号失败: %v", err)
	} else {
		fmt.Printf("最新区块号: %s\n", blockNumber.String())
	}

	fmt.Println("\n🎉 连接测试完成!")
	fmt.Println("您已成功连接到以太坊 Sepolia 测试网!")
}
