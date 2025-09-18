package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/local/go-eth-demo/config"
	"github.com/local/go-eth-demo/utils"
)

func main() {
	fmt.Println("🌐 以太坊网络信息查询")
	fmt.Println("================================")

	// 加载配置并创建客户端
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ 配置加载失败: %v", err)
	}

	client, err := utils.NewEthClient(cfg)
	if err != nil {
		log.Fatalf("❌ 连接失败: %v", err)
	}
	defer client.Close()

	ethClient := client.GetClient()
	ctx := context.Background()

	fmt.Println("📊 正在获取网络详细信息...")
	fmt.Println()

	// 1. 基础网络信息
	fmt.Println("🔗 基础网络信息:")
	fmt.Println("--------------------------------")

	chainID, err := ethClient.ChainID(ctx)
	if err != nil {
		log.Printf("❌ 获取 Chain ID 失败: %v", err)
	} else {
		fmt.Printf("Chain ID: %s\n", chainID.String())
	}

	networkID, err := ethClient.NetworkID(ctx)
	if err != nil {
		log.Printf("❌ 获取 Network ID 失败: %v", err)
	} else {
		fmt.Printf("Network ID: %s\n", networkID.String())
	}

	// 2. 区块信息
	fmt.Println("\n📦 区块信息:")
	fmt.Println("--------------------------------")

	blockNumber, err := ethClient.BlockNumber(ctx)
	if err != nil {
		log.Printf("❌ 获取最新区块号失败: %v", err)
	} else {
		fmt.Printf("最新区块号: %d\n", blockNumber)
	}

	// 获取最新区块详细信息
	if blockNumber > 0 {
		block, err := ethClient.BlockByNumber(ctx, big.NewInt(int64(blockNumber)))
		if err != nil {
			log.Printf("❌ 获取区块详情失败: %v", err)
		} else {
			fmt.Printf("区块哈希: %s\n", block.Hash().Hex())
			fmt.Printf("父区块哈希: %s\n", block.ParentHash().Hex())
			fmt.Printf("区块时间: %d\n", block.Time())
			fmt.Printf("交易数量: %d\n", len(block.Transactions()))
			fmt.Printf("Gas 使用量: %d\n", block.GasUsed())
			fmt.Printf("Gas 限制: %d\n", block.GasLimit())
		}
	}

	// 3. Gas 价格信息
	fmt.Println("\n⛽ Gas 价格信息:")
	fmt.Println("--------------------------------")

	gasPrice, err := ethClient.SuggestGasPrice(ctx)
	if err != nil {
		log.Printf("❌ 获取 Gas 价格失败: %v", err)
	} else {
		// 转换为 Gwei (1 Gwei = 10^9 Wei)
		gwei := new(big.Int).Div(gasPrice, big.NewInt(1000000000))
		fmt.Printf("建议 Gas 价格: %s Wei (%s Gwei)\n", gasPrice.String(), gwei.String())
	}

	// 4. 网络同步状态
	fmt.Println("\n🔄 同步状态:")
	fmt.Println("--------------------------------")

	syncProgress, err := ethClient.SyncProgress(ctx)
	if err != nil {
		log.Printf("❌ 获取同步状态失败: %v", err)
	} else {
		if syncProgress == nil {
			fmt.Println("节点已完全同步")
		} else {
			fmt.Printf("正在同步: %d/%d (%.2f%%)\n",
				syncProgress.CurrentBlock,
				syncProgress.HighestBlock,
				float64(syncProgress.CurrentBlock)/float64(syncProgress.HighestBlock)*100)
		}
	}

	// 5. 节点信息
	fmt.Println("\n🖥️  节点信息:")
	fmt.Println("--------------------------------")

	// 尝试获取节点版本（某些 RPC 提供商可能不支持）
	var nodeInfo string
	err = ethClient.Client().Call(&nodeInfo, "web3_clientVersion")
	if err != nil {
		fmt.Printf("节点版本: 无法获取 (%v)\n", err)
	} else {
		fmt.Printf("节点版本: %s\n", nodeInfo)
	}

	// 6. 网络统计
	fmt.Println("\n📈 网络统计:")
	fmt.Println("--------------------------------")

	// 获取待处理交易数量
	pendingCount, err := ethClient.PendingTransactionCount(ctx)
	if err != nil {
		log.Printf("❌ 获取待处理交易数量失败: %v", err)
	} else {
		fmt.Printf("待处理交易数量: %d\n", pendingCount)
	}

	fmt.Println("\n✅ 网络信息查询完成!")
}
