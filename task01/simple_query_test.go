package main

import (
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/local/dapp-basics-task01/blockchain"
	"github.com/local/dapp-basics-task01/config"
)

func main() {
	fmt.Println("🔍 简单区块查询测试")
	fmt.Println("==================")

	// 加载配置
	cfg := config.LoadConfig()
	fmt.Printf("📡 网络: %s\n", cfg.NetworkName)

	// 尝试多个可靠的RPC端点
	rpcEndpoints := []string{
		"https://ethereum-sepolia.blockpi.network/v1/rpc/public",
		"https://rpc.sepolia.org",
		"https://sepolia.gateway.tenderly.co",
	}

	var client *blockchain.Client
	var err error

	for i, endpoint := range rpcEndpoints {
		fmt.Printf("\n🔄 尝试端点 %d: %s\n", i+1, endpoint)
		client, err = blockchain.NewClient(endpoint)
		if err != nil {
			fmt.Printf("❌ 连接失败: %v\n", err)
			continue
		}

		// 测试查询最新区块
		fmt.Println("✅ 连接成功，测试查询最新区块...")
		latestBlock, err := client.QueryLatestBlock()
		if err != nil {
			fmt.Printf("❌ 查询失败: %v\n", err)
			client.Close()
			continue
		}

		fmt.Printf("🎉 查询成功！\n")
		fmt.Printf("📊 最新区块号: %s\n", latestBlock.Number.String())
		fmt.Printf("📊 区块哈希: %s\n", latestBlock.Hash)
		fmt.Printf("📊 交易数量: %d\n", latestBlock.TxCount)
		fmt.Printf("📊 时间戳: %d (%s)\n", latestBlock.Timestamp,
			time.Unix(int64(latestBlock.Timestamp), 0).Format("2006-01-02 15:04:05"))

		// 测试查询历史区块
		fmt.Println("\n📚 测试查询历史区块...")
		blockNum := big.NewInt(5000000)
		blockInfo, err := client.QueryBlockByNumber(blockNum)
		if err != nil {
			fmt.Printf("❌ 历史区块查询失败: %v\n", err)
		} else {
			fmt.Printf("✅ 历史区块 %s 查询成功\n", blockNum.String())
			fmt.Printf("📊 区块哈希: %s\n", blockInfo.Hash)
			fmt.Printf("📊 交易数量: %d\n", blockInfo.TxCount)
			fmt.Printf("📊 时间戳: %d (%s)\n", blockInfo.Timestamp,
				time.Unix(int64(blockInfo.Timestamp), 0).Format("2006-01-02 15:04:05"))
		}

		// 测试余额查询
		fmt.Println("\n💰 测试余额查询...")
		testAddr := "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045" // Vitalik的地址
		balance, err := client.GetBalance(testAddr)
		if err != nil {
			fmt.Printf("❌ 余额查询失败: %v\n", err)
		} else {
			balanceEth := new(big.Float)
			balanceEth.SetString(balance.String())
			balanceEth = balanceEth.Quo(balanceEth, big.NewFloat(1e18))
			fmt.Printf("✅ 地址余额: %s ETH\n", balanceEth.String())
		}

		client.Close()
		fmt.Println("\n🎉 区块查询功能验证成功！")
		return
	}

	log.Fatal("❌ 所有RPC端点都无法连接")
}
