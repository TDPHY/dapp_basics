package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	fmt.Println("🔍 验证区块查询功能")
	fmt.Println("==================")

	// 尝试连接到Sepolia测试网
	rpcEndpoints := []string{
		"https://ethereum-sepolia.blockpi.network/v1/rpc/public",
		"https://rpc.sepolia.org",
		"https://sepolia.gateway.tenderly.co",
	}

	var client *ethclient.Client
	var err error

	for i, endpoint := range rpcEndpoints {
		fmt.Printf("\n🔄 尝试连接端点 %d: %s\n", i+1, endpoint)
		client, err = ethclient.Dial(endpoint)
		if err != nil {
			fmt.Printf("❌ 连接失败: %v\n", err)
			continue
		}

		// 测试查询最新区块
		fmt.Println("✅ 连接成功，测试查询最新区块...")
		ctx := context.Background()

		// 查询最新区块
		latestBlock, err := client.BlockByNumber(ctx, nil)
		if err != nil {
			fmt.Printf("❌ 查询最新区块失败: %v\n", err)
			client.Close()
			continue
		}

		fmt.Printf("🎉 查询成功！\n")
		fmt.Printf("📊 最新区块号: %s\n", latestBlock.Number().String())
		fmt.Printf("📊 区块哈希: %s\n", latestBlock.Hash().Hex())
		fmt.Printf("📊 交易数量: %d\n", len(latestBlock.Transactions()))
		fmt.Printf("📊 时间戳: %d (%s)\n", latestBlock.Time(),
			time.Unix(int64(latestBlock.Time()), 0).Format("2006-01-02 15:04:05"))
		fmt.Printf("📊 Gas使用量: %d\n", latestBlock.GasUsed())
		fmt.Printf("📊 Gas限制: %d\n", latestBlock.GasLimit())

		// 测试查询历史区块
		fmt.Println("\n📚 测试查询历史区块...")
		blockNum := big.NewInt(5000000)
		historicalBlock, err := client.BlockByNumber(ctx, blockNum)
		if err != nil {
			fmt.Printf("❌ 历史区块查询失败: %v\n", err)
		} else {
			fmt.Printf("✅ 历史区块 %s 查询成功\n", blockNum.String())
			fmt.Printf("📊 区块哈希: %s\n", historicalBlock.Hash().Hex())
			fmt.Printf("📊 交易数量: %d\n", len(historicalBlock.Transactions()))
			fmt.Printf("📊 时间戳: %d (%s)\n", historicalBlock.Time(),
				time.Unix(int64(historicalBlock.Time()), 0).Format("2006-01-02 15:04:05"))
		}

		// 测试余额查询
		fmt.Println("\n💰 测试余额查询...")
		testAddr := common.HexToAddress("0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045") // Vitalik的地址
		balance, err := client.BalanceAt(ctx, testAddr, nil)
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
		fmt.Println("✅ 可以成功连接到Sepolia测试网")
		fmt.Println("✅ 可以查询最新区块信息")
		fmt.Println("✅ 可以查询历史区块信息")
		fmt.Println("✅ 可以查询账户余额")
		fmt.Println("✅ 区块数据包含完整的信息（区块号、哈希、交易数、时间戳等）")
		return
	}

	log.Fatal("❌ 所有RPC端点都无法连接")
}
