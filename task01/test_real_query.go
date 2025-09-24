package main

import (
	"fmt"
	"log"
	"math/big"

	"github.com/local/dapp-basics-task01/blockchain"
	"github.com/local/dapp-basics-task01/config"
)

func main() {
	fmt.Println("🔍 真实区块查询测试")
	fmt.Println("==================")

	// 加载配置
	cfg := config.LoadConfig()
	fmt.Printf("📡 网络: %s\n", cfg.NetworkName)
	fmt.Printf("🔗 RPC URL: %s\n", cfg.EthereumRPCURL)

	// 尝试多个RPC端点
	rpcEndpoints := []string{
		"https://eth-sepolia.g.alchemy.com/v2/demo",
		"https://sepolia.infura.io/v3/9aa3d95b3bc440fa88ea12eaa4456161", // 公共端点
		"https://rpc.sepolia.org",
		"https://ethereum-sepolia.blockpi.network/v1/rpc/public",
	}

	var client *blockchain.Client
	var err error

	for i, endpoint := range rpcEndpoints {
		fmt.Printf("\n🔄 尝试连接端点 %d: %s\n", i+1, endpoint)
		client, err = blockchain.NewClient(endpoint)
		if err != nil {
			fmt.Printf("❌ 连接失败: %v\n", err)
			continue
		}

		// 测试连接是否真的可用
		fmt.Println("✅ 连接成功，测试查询...")
		latestBlock, err := client.QueryLatestBlock()
		if err != nil {
			fmt.Printf("❌ 查询失败: %v\n", err)
			client.Close()
			continue
		}

		fmt.Printf("🎉 成功！最新区块号: %s\n", latestBlock.Number.String())
		fmt.Printf("📊 区块哈希: %s\n", latestBlock.Hash)
		fmt.Printf("📊 交易数量: %d\n", latestBlock.TxCount)
		fmt.Printf("📊 时间戳: %s\n", latestBlock.Timestamp.Format("2006-01-02 15:04:05"))
		break
	}

	if client == nil {
		log.Fatal("❌ 所有RPC端点都无法连接")
	}
	defer client.Close()

	// 测试查询历史区块
	fmt.Println("\n📚 测试查询历史区块...")
	historicalBlocks := []*big.Int{
		big.NewInt(5000000),
		big.NewInt(4000000),
		big.NewInt(3000000),
	}

	for _, blockNum := range historicalBlocks {
		fmt.Printf("\n🔍 查询区块 %s...\n", blockNum.String())
		blockInfo, err := client.QueryBlockByNumber(blockNum)
		if err != nil {
			fmt.Printf("❌ 查询失败: %v\n", err)
			continue
		}

		fmt.Printf("✅ 区块 %s 查询成功\n", blockNum.String())
		fmt.Printf("📊 区块哈希: %s\n", blockInfo.Hash)
		fmt.Printf("📊 交易数量: %d\n", blockInfo.TxCount)
		fmt.Printf("📊 时间戳: %s\n", blockInfo.Timestamp.Format("2006-01-02 15:04:05"))
	}

	// 测试余额查询
	fmt.Println("\n💰 测试余额查询...")
	testAddresses := []string{
		"0x742d35Cc6634C0532925a3b8D0C9e3e0C8b0e4c2", // 测试地址1
		"0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045", // Vitalik的地址
		"0x0000000000000000000000000000000000000000", // 零地址
	}

	for _, addr := range testAddresses {
		fmt.Printf("\n🔍 查询地址 %s...\n", addr)
		balance, err := client.GetBalance(addr)
		if err != nil {
			fmt.Printf("❌ 查询失败: %v\n", err)
			continue
		}

		balanceEth := new(big.Float)
		balanceEth.SetString(balance.String())
		balanceEth = balanceEth.Quo(balanceEth, big.NewFloat(1e18))
		fmt.Printf("✅ 余额: %s ETH\n", balanceEth.String())
	}

	fmt.Println("\n🎉 真实区块查询测试完成!")
}
