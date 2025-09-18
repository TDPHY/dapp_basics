package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/local/go-eth-demo/config"
	"github.com/local/go-eth-demo/utils"
)

func main() {
	fmt.Println("📦 以太坊区块查询详解")
	fmt.Println("================================")

	// 初始化客户端
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

	// 1. 获取最新区块
	fmt.Println("🔍 查询最新区块...")
	latestBlock, err := queryLatestBlock(ctx, ethClient)
	if err != nil {
		log.Fatalf("❌ 获取最新区块失败: %v", err)
	}
	displayBlockInfo("最新区块", latestBlock)

	// 2. 根据区块号查询历史区块
	fmt.Println("\n🔍 查询历史区块...")
	blockNumber := new(big.Int).Sub(latestBlock.Number(), big.NewInt(10)) // 10个区块前
	historicalBlock, err := queryBlockByNumber(ctx, ethClient, blockNumber)
	if err != nil {
		log.Printf("❌ 获取历史区块失败: %v", err)
	} else {
		displayBlockInfo(fmt.Sprintf("历史区块 #%s", blockNumber.String()), historicalBlock)
	}

	// 3. 根据区块哈希查询区块
	fmt.Println("\n🔍 根据哈希查询区块...")
	blockByHash, err := queryBlockByHash(ctx, ethClient, latestBlock.Hash())
	if err != nil {
		log.Printf("❌ 根据哈希获取区块失败: %v", err)
	} else {
		fmt.Printf("✅ 通过哈希查询成功，区块号: %s\n", blockByHash.Number().String())
	}

	// 4. 分析区块中的交易
	fmt.Println("\n💰 分析区块交易...")
	analyzeBlockTransactions(latestBlock)

	// 5. 区块时间分析
	fmt.Println("\n⏰ 区块时间分析...")
	analyzeBlockTiming(ctx, ethClient, latestBlock.Number())

	// 6. Gas 使用分析
	fmt.Println("\n⛽ Gas 使用分析...")
	analyzeGasUsage(latestBlock)

	fmt.Println("\n✅ 区块查询学习完成!")
}

// queryLatestBlock 查询最新区块
func queryLatestBlock(ctx context.Context, client *ethclient.Client) (*types.Block, error) {
	// 方法1: 使用 nil 获取最新区块
	block, err := client.BlockByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("获取最新区块失败: %w", err)
	}
	return block, nil
}

// queryBlockByNumber 根据区块号查询区块
func queryBlockByNumber(ctx context.Context, client *ethclient.Client, blockNumber *big.Int) (*types.Block, error) {
	block, err := client.BlockByNumber(ctx, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("获取区块 #%s 失败: %w", blockNumber.String(), err)
	}
	return block, nil
}

// queryBlockByHash 根据区块哈希查询区块
func queryBlockByHash(ctx context.Context, client *ethclient.Client, blockHash common.Hash) (*types.Block, error) {
	block, err := client.BlockByHash(ctx, blockHash)
	if err != nil {
		return nil, fmt.Errorf("根据哈希获取区块失败: %w", err)
	}
	return block, nil
}

// displayBlockInfo 显示区块详细信息
func displayBlockInfo(title string, block *types.Block) {
	fmt.Printf("\n📋 %s 详细信息:\n", title)
	fmt.Println("--------------------------------")

	// 基本信息
	fmt.Printf("区块号: %s\n", block.Number().String())
	fmt.Printf("区块哈希: %s\n", block.Hash().Hex())
	fmt.Printf("父区块哈希: %s\n", block.ParentHash().Hex())

	// 时间信息
	blockTime := time.Unix(int64(block.Time()), 0)
	fmt.Printf("区块时间: %s (%d)\n", blockTime.Format("2006-01-02 15:04:05"), block.Time())

	// 挖矿信息
	fmt.Printf("矿工地址: %s\n", block.Coinbase().Hex())
	fmt.Printf("难度: %s\n", block.Difficulty().String())

	// 交易信息
	fmt.Printf("交易数量: %d\n", len(block.Transactions()))
	fmt.Printf("叔块数量: %d\n", len(block.Uncles()))

	// Gas 信息
	fmt.Printf("Gas 限制: %s\n", formatNumber(block.GasLimit()))
	fmt.Printf("Gas 使用: %s\n", formatNumber(block.GasUsed()))
	gasUsagePercent := float64(block.GasUsed()) / float64(block.GasLimit()) * 100
	fmt.Printf("Gas 使用率: %.2f%%\n", gasUsagePercent)

	// 其他信息
	fmt.Printf("区块大小: %s bytes\n", formatNumber(uint64(block.Size())))
	fmt.Printf("Nonce: %d\n", block.Nonce())
	fmt.Printf("Extra Data: %s\n", string(block.Extra()))

	// Merkle 根
	fmt.Printf("交易根: %s\n", block.TxHash().Hex())
	fmt.Printf("状态根: %s\n", block.Root().Hex())
	fmt.Printf("收据根: %s\n", block.ReceiptHash().Hex())
}

// analyzeBlockTransactions 分析区块中的交易
func analyzeBlockTransactions(block *types.Block) {
	transactions := block.Transactions()

	if len(transactions) == 0 {
		fmt.Println("该区块没有交易")
		return
	}

	fmt.Printf("📊 交易统计 (总计: %d 笔):\n", len(transactions))
	fmt.Println("--------------------------------")

	var totalValue, totalGasUsed, totalGasPrice big.Int
	contractCreations := 0

	// 分析前5笔交易的详细信息
	displayCount := 5
	if len(transactions) < displayCount {
		displayCount = len(transactions)
	}

	fmt.Printf("🔍 前 %d 笔交易详情:\n", displayCount)
	for i := 0; i < displayCount; i++ {
		tx := transactions[i]
		fmt.Printf("\n交易 #%d:\n", i+1)
		fmt.Printf("  哈希: %s\n", tx.Hash().Hex())
		fmt.Printf("  发送方: %s\n", "需要签名恢复") // 简化显示
		if tx.To() != nil {
			fmt.Printf("  接收方: %s\n", tx.To().Hex())
		} else {
			fmt.Printf("  接收方: 合约创建\n")
			contractCreations++
		}
		fmt.Printf("  金额: %s ETH\n", weiToEther(tx.Value()))
		fmt.Printf("  Gas 限制: %s\n", formatNumber(tx.Gas()))
		fmt.Printf("  Gas 价格: %s Gwei\n", weiToGwei(tx.GasPrice()))
		fmt.Printf("  Nonce: %d\n", tx.Nonce())
	}

	// 统计所有交易
	for _, tx := range transactions {
		totalValue.Add(&totalValue, tx.Value())
		totalGasUsed.Add(&totalGasUsed, big.NewInt(int64(tx.Gas())))
		totalGasPrice.Add(&totalGasPrice, tx.GasPrice())

		if tx.To() == nil {
			contractCreations++
		}
	}

	// 显示统计信息
	fmt.Printf("\n📈 交易统计摘要:\n")
	fmt.Printf("  总转账金额: %s ETH\n", weiToEther(&totalValue))
	fmt.Printf("  平均 Gas 价格: %s Gwei\n", weiToGwei(new(big.Int).Div(&totalGasPrice, big.NewInt(int64(len(transactions))))))
	fmt.Printf("  合约创建交易: %d 笔\n", contractCreations)
	fmt.Printf("  普通转账交易: %d 笔\n", len(transactions)-contractCreations)
}

// analyzeBlockTiming 分析区块时间
func analyzeBlockTiming(ctx context.Context, client *ethclient.Client, currentBlockNumber *big.Int) {
	if currentBlockNumber.Cmp(big.NewInt(1)) <= 0 {
		fmt.Println("无法分析创世区块的时间")
		return
	}

	// 获取前一个区块
	prevBlockNumber := new(big.Int).Sub(currentBlockNumber, big.NewInt(1))
	prevBlock, err := client.BlockByNumber(ctx, prevBlockNumber)
	if err != nil {
		fmt.Printf("❌ 获取前一个区块失败: %v\n", err)
		return
	}

	currentBlock, err := client.BlockByNumber(ctx, currentBlockNumber)
	if err != nil {
		fmt.Printf("❌ 获取当前区块失败: %v\n", err)
		return
	}

	// 计算区块间隔
	timeDiff := currentBlock.Time() - prevBlock.Time()
	fmt.Printf("与前一区块的时间间隔: %d 秒\n", timeDiff)

	// 分析最近10个区块的平均出块时间
	analyzeAverageBlockTime(ctx, client, currentBlockNumber, 10)
}

// analyzeAverageBlockTime 分析平均出块时间
func analyzeAverageBlockTime(ctx context.Context, client *ethclient.Client, latestBlockNumber *big.Int, count int) {
	if latestBlockNumber.Cmp(big.NewInt(int64(count))) < 0 {
		fmt.Printf("区块数量不足，无法分析最近 %d 个区块\n", count)
		return
	}

	startBlockNumber := new(big.Int).Sub(latestBlockNumber, big.NewInt(int64(count-1)))

	startBlock, err := client.BlockByNumber(ctx, startBlockNumber)
	if err != nil {
		fmt.Printf("❌ 获取起始区块失败: %v\n", err)
		return
	}

	endBlock, err := client.BlockByNumber(ctx, latestBlockNumber)
	if err != nil {
		fmt.Printf("❌ 获取结束区块失败: %v\n", err)
		return
	}

	totalTime := endBlock.Time() - startBlock.Time()
	averageTime := float64(totalTime) / float64(count-1)

	fmt.Printf("最近 %d 个区块的平均出块时间: %.2f 秒\n", count, averageTime)
}

// analyzeGasUsage 分析 Gas 使用情况
func analyzeGasUsage(block *types.Block) {
	gasLimit := block.GasLimit()
	gasUsed := block.GasUsed()
	gasUsagePercent := float64(gasUsed) / float64(gasLimit) * 100

	fmt.Printf("Gas 限制: %s\n", formatNumber(gasLimit))
	fmt.Printf("Gas 使用: %s\n", formatNumber(gasUsed))
	fmt.Printf("Gas 使用率: %.2f%%\n", gasUsagePercent)
	fmt.Printf("剩余 Gas: %s\n", formatNumber(gasLimit-gasUsed))

	// Gas 使用率分析
	if gasUsagePercent > 95 {
		fmt.Println("🔴 Gas 使用率很高，网络拥堵")
	} else if gasUsagePercent > 80 {
		fmt.Println("🟡 Gas 使用率较高，网络繁忙")
	} else if gasUsagePercent > 50 {
		fmt.Println("🟢 Gas 使用率正常")
	} else {
		fmt.Println("🔵 Gas 使用率较低，网络空闲")
	}
}

// 工具函数

// weiToEther 将 Wei 转换为 Ether
func weiToEther(wei *big.Int) string {
	ether := new(big.Float).SetInt(wei)
	ether.Quo(ether, big.NewFloat(1e18))
	return ether.Text('f', 6)
}

// weiToGwei 将 Wei 转换为 Gwei
func weiToGwei(wei *big.Int) string {
	gwei := new(big.Float).SetInt(wei)
	gwei.Quo(gwei, big.NewFloat(1e9))
	return gwei.Text('f', 2)
}

// formatNumber 格式化大数字，添加千位分隔符
func formatNumber(n uint64) string {
	str := fmt.Sprintf("%d", n)
	if len(str) <= 3 {
		return str
	}

	result := ""
	for i, char := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result += ","
		}
		result += string(char)
	}
	return result
}
