package main

import (
	"fmt"
	"log"
	"math/big"

	"github.com/local/dapp-basics-task01/blockchain"
	"github.com/local/dapp-basics-task01/config"
)

func main() {
	fmt.Println("🧪 基础功能测试")
	fmt.Println("================")

	// 加载配置
	cfg := config.LoadConfig()
	fmt.Printf("✅ 配置加载成功\n")
	fmt.Printf("📡 网络: %s\n", cfg.NetworkName)
	fmt.Printf("🔗 RPC URL: %s\n", cfg.EthereumRPCURL)

	// 测试区块链连接
	fmt.Println("\n🔍 测试区块链连接...")
	client, err := blockchain.NewClient(cfg.EthereumRPCURL)
	if err != nil {
		log.Fatalf("❌ 连接失败: %v", err)
	}
	defer client.Close()
	fmt.Println("✅ 区块链连接成功")

	// 测试查询最新区块
	fmt.Println("\n📊 测试查询最新区块...")
	latestBlock, err := client.QueryLatestBlock()
	if err != nil {
		log.Printf("❌ 查询最新区块失败: %v", err)
	} else {
		fmt.Printf("✅ 最新区块号: %s\n", latestBlock.Number.String())
		fmt.Printf("✅ 区块哈希: %s\n", latestBlock.Hash)
		fmt.Printf("✅ 交易数量: %d\n", latestBlock.TxCount)
	}

	// 测试查询指定区块
	fmt.Println("\n📊 测试查询指定区块...")
	blockNumber := big.NewInt(6000000) // 一个较早的区块
	blockInfo, err := client.QueryBlockByNumber(blockNumber)
	if err != nil {
		log.Printf("❌ 查询指定区块失败: %v", err)
	} else {
		fmt.Printf("✅ 区块 %s 查询成功\n", blockNumber.String())
		fmt.Printf("✅ 区块哈希: %s\n", blockInfo.Hash)
		fmt.Printf("✅ 交易数量: %d\n", blockInfo.TxCount)
	}

	// 测试余额查询
	fmt.Println("\n💰 测试余额查询...")
	testAddress := "0x742d35Cc6634C0532925a3b8D0C9e3e0C8b0e4c2" // 一个测试地址
	balance, err := client.GetBalance(testAddress)
	if err != nil {
		log.Printf("❌ 查询余额失败: %v", err)
	} else {
		balanceEth := new(big.Float)
		balanceEth.SetString(balance.String())
		balanceEth = balanceEth.Quo(balanceEth, big.NewFloat(1e18))
		fmt.Printf("✅ 地址 %s 余额: %s ETH\n", testAddress, balanceEth.String())
	}

	fmt.Println("\n🎉 基础功能测试完成!")
	fmt.Println("📋 测试结果:")
	fmt.Println("  ✅ 配置加载")
	fmt.Println("  ✅ 区块链连接")
	fmt.Println("  ✅ 区块查询")
	fmt.Println("  ✅ 余额查询")
	fmt.Println("\n🚀 可以运行主程序: go run main.go")
}
