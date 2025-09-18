package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("🔔 以太坊区块订阅演示")
	fmt.Println("================================")

	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Printf("警告: 无法加载 .env 文件: %v", err)
	}

	// 获取 RPC URL
	rpcURL := os.Getenv("ETHEREUM_RPC_URL")
	if rpcURL == "" {
		log.Fatal("请在 .env 文件中设置 ETHEREUM_RPC_URL")
	}

	// 连接到以太坊节点
	fmt.Println("\n🌐 连接到以太坊节点...")
	fmt.Println("--------------------------------")
	fmt.Printf("RPC URL: %s\n", rpcURL)

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer client.Close()

	fmt.Println("✅ 连接成功!")

	// 获取当前网络信息
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		log.Fatalf("获取链 ID 失败: %v", err)
	}

	fmt.Printf("链 ID: %s\n", chainID.String())

	// 获取当前区块号
	latestBlock, err := client.BlockNumber(context.Background())
	if err != nil {
		log.Fatalf("获取最新区块号失败: %v", err)
	}

	fmt.Printf("当前区块号: %d\n", latestBlock)

	// 创建上下文和取消函数
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 设置信号处理，优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动区块订阅
	fmt.Println("\n🔔 开始订阅新区块...")
	fmt.Println("================================")
	fmt.Println("按 Ctrl+C 停止订阅")
	fmt.Println()

	// 创建统计信息
	stats := &BlockStats{
		StartTime:    time.Now(),
		BlockCount:   0,
		TotalTxs:     0,
		TotalGasUsed: big.NewInt(0),
	}

	// 启动区块头订阅
	go subscribeNewHeads(ctx, client, stats)

	// 等待退出信号
	<-sigChan
	fmt.Println("\n\n🛑 收到退出信号，正在停止订阅...")
	cancel()

	// 显示最终统计
	displayFinalStats(stats)
	fmt.Println("订阅已停止!")
}

// BlockStats 区块统计信息
type BlockStats struct {
	StartTime    time.Time
	BlockCount   int64
	TotalTxs     int64
	TotalGasUsed *big.Int
	LastBlock    *types.Header
}

// subscribeNewHeads 订阅新区块头
func subscribeNewHeads(ctx context.Context, client *ethclient.Client, stats *BlockStats) {
	// 创建区块头订阅
	headers := make(chan *types.Header)
	sub, err := client.SubscribeNewHead(ctx, headers)
	if err != nil {
		log.Fatalf("创建区块头订阅失败: %v", err)
	}
	defer sub.Unsubscribe()

	fmt.Println("✅ 区块头订阅创建成功!")
	fmt.Println("等待新区块...")
	fmt.Println()

	for {
		select {
		case err := <-sub.Err():
			log.Printf("❌ 订阅错误: %v", err)
			// 尝试重新订阅
			fmt.Println("🔄 尝试重新订阅...")
			time.Sleep(5 * time.Second)
			go subscribeNewHeads(ctx, client, stats)
			return

		case header := <-headers:
			// 处理新区块头
			processNewHeader(ctx, client, header, stats)

		case <-ctx.Done():
			fmt.Println("🔔 区块头订阅已停止")
			return
		}
	}
}

// processNewHeader 处理新区块头
func processNewHeader(ctx context.Context, client *ethclient.Client, header *types.Header, stats *BlockStats) {
	stats.BlockCount++
	stats.LastBlock = header

	fmt.Printf("🆕 新区块 #%d\n", header.Number.Uint64())
	fmt.Println("--------------------------------")

	// 显示区块基本信息
	displayBlockHeader(header)

	// 获取完整区块信息（包含交易）
	block, err := client.BlockByHash(ctx, header.Hash())
	if err != nil {
		fmt.Printf("❌ 获取完整区块失败: %v\n", err)
		return
	}

	// 分析区块内容
	analyzeBlock(block, stats)

	// 显示实时统计
	displayRealtimeStats(stats)

	fmt.Println()
}

// displayBlockHeader 显示区块头信息
func displayBlockHeader(header *types.Header) {
	fmt.Printf("区块哈希: %s\n", header.Hash().Hex())
	fmt.Printf("父区块哈希: %s\n", header.ParentHash.Hex())
	fmt.Printf("时间戳: %s\n", time.Unix(int64(header.Time), 0).Format("2006-01-02 15:04:05"))
	fmt.Printf("Gas 限制: %s\n", formatNumber(header.GasLimit))
	fmt.Printf("Gas 使用: %s (%.2f%%)\n",
		formatNumber(header.GasUsed),
		float64(header.GasUsed)/float64(header.GasLimit)*100)
	fmt.Printf("难度: %s\n", header.Difficulty.String())
	fmt.Printf("矿工: %s\n", header.Coinbase.Hex())
}

// analyzeBlock 分析区块内容
func analyzeBlock(block *types.Block, stats *BlockStats) {
	txCount := len(block.Transactions())
	stats.TotalTxs += int64(txCount)
	stats.TotalGasUsed.Add(stats.TotalGasUsed, new(big.Int).SetUint64(block.GasUsed()))

	fmt.Printf("交易数量: %d\n", txCount)

	if txCount > 0 {
		// 分析交易类型
		analyzeTransactions(block.Transactions())
	}

	// 检查特殊事件
	checkSpecialEvents(block)
}

// analyzeTransactions 分析交易
func analyzeTransactions(txs types.Transactions) {
	var (
		transferCount int
		contractCount int
		totalValue    = big.NewInt(0)
		totalGasFees  = big.NewInt(0)
		maxGasPrice   = big.NewInt(0)
		minGasPrice   *big.Int
	)

	for _, tx := range txs {
		// 统计交易类型
		if tx.To() == nil {
			contractCount++ // 合约创建
		} else {
			transferCount++ // 转账或合约调用
		}

		// 累计交易价值
		totalValue.Add(totalValue, tx.Value())

		// 分析 Gas 价格
		gasPrice := tx.GasPrice()
		if gasPrice != nil {
			if gasPrice.Cmp(maxGasPrice) > 0 {
				maxGasPrice = gasPrice
			}
			if minGasPrice == nil || gasPrice.Cmp(minGasPrice) < 0 {
				minGasPrice = gasPrice
			}

			// 计算 Gas 费用 (gasPrice * gasUsed)
			gasFee := new(big.Int).Mul(gasPrice, big.NewInt(int64(tx.Gas())))
			totalGasFees.Add(totalGasFees, gasFee)
		}
	}

	fmt.Printf("  • 转账/调用: %d 笔\n", transferCount)
	if contractCount > 0 {
		fmt.Printf("  • 合约创建: %d 笔\n", contractCount)
	}

	if totalValue.Cmp(big.NewInt(0)) > 0 {
		fmt.Printf("  • 总价值: %s ETH\n", formatEther(totalValue))
	}

	if minGasPrice != nil && maxGasPrice.Cmp(big.NewInt(0)) > 0 {
		fmt.Printf("  • Gas 价格: %s - %s Gwei\n",
			formatGwei(minGasPrice), formatGwei(maxGasPrice))
	}

	if totalGasFees.Cmp(big.NewInt(0)) > 0 {
		fmt.Printf("  • 总 Gas 费: %s ETH\n", formatEther(totalGasFees))
	}
}

// checkSpecialEvents 检查特殊事件
func checkSpecialEvents(block *types.Block) {
	// 检查大区块
	if len(block.Transactions()) > 100 {
		fmt.Printf("🔥 大区块: 包含 %d 笔交易\n", len(block.Transactions()))
	}

	// 检查高 Gas 使用率
	gasUsagePercent := float64(block.GasUsed()) / float64(block.GasLimit()) * 100
	if gasUsagePercent > 90 {
		fmt.Printf("⚡ 高 Gas 使用率: %.2f%%\n", gasUsagePercent)
	}

	// 检查区块时间间隔
	if block.Number().Uint64() > 0 {
		// 这里可以添加与上一个区块的时间比较
		// 但需要存储上一个区块的时间戳
	}
}

// displayRealtimeStats 显示实时统计
func displayRealtimeStats(stats *BlockStats) {
	duration := time.Since(stats.StartTime)

	fmt.Println("\n📊 实时统计:")
	fmt.Printf("  • 运行时间: %s\n", formatDuration(duration))
	fmt.Printf("  • 接收区块: %d 个\n", stats.BlockCount)
	fmt.Printf("  • 总交易数: %d 笔\n", stats.TotalTxs)

	if stats.BlockCount > 0 {
		avgTxPerBlock := float64(stats.TotalTxs) / float64(stats.BlockCount)
		fmt.Printf("  • 平均每区块交易: %.1f 笔\n", avgTxPerBlock)

		blocksPerMinute := float64(stats.BlockCount) / duration.Minutes()
		if blocksPerMinute > 0 {
			fmt.Printf("  • 区块频率: %.2f 个/分钟\n", blocksPerMinute)
		}
	}

	if stats.TotalGasUsed.Cmp(big.NewInt(0)) > 0 {
		fmt.Printf("  • 总 Gas 使用: %s\n", formatNumber(stats.TotalGasUsed.Uint64()))
	}
}

// displayFinalStats 显示最终统计
func displayFinalStats(stats *BlockStats) {
	fmt.Println("\n📈 最终统计报告:")
	fmt.Println("================================")

	duration := time.Since(stats.StartTime)
	fmt.Printf("总运行时间: %s\n", formatDuration(duration))
	fmt.Printf("接收区块总数: %d 个\n", stats.BlockCount)
	fmt.Printf("处理交易总数: %d 笔\n", stats.TotalTxs)

	if stats.BlockCount > 0 {
		avgTxPerBlock := float64(stats.TotalTxs) / float64(stats.BlockCount)
		fmt.Printf("平均每区块交易数: %.1f 笔\n", avgTxPerBlock)

		blocksPerHour := float64(stats.BlockCount) / duration.Hours()
		fmt.Printf("平均区块频率: %.1f 个/小时\n", blocksPerHour)
	}

	if stats.LastBlock != nil {
		fmt.Printf("最后处理区块: #%d\n", stats.LastBlock.Number.Uint64())
		fmt.Printf("最后区块时间: %s\n",
			time.Unix(int64(stats.LastBlock.Time), 0).Format("2006-01-02 15:04:05"))
	}

	if stats.TotalGasUsed.Cmp(big.NewInt(0)) > 0 {
		fmt.Printf("总 Gas 消耗: %s\n", formatNumber(stats.TotalGasUsed.Uint64()))
	}
}

// 格式化函数
func formatNumber(n uint64) string {
	if n >= 1000000000 {
		return fmt.Sprintf("%.2fB", float64(n)/1000000000)
	} else if n >= 1000000 {
		return fmt.Sprintf("%.2fM", float64(n)/1000000)
	} else if n >= 1000 {
		return fmt.Sprintf("%.2fK", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}

func formatEther(wei *big.Int) string {
	ether := new(big.Float).SetInt(wei)
	ether.Quo(ether, big.NewFloat(1e18))
	return fmt.Sprintf("%.6f", ether)
}

func formatGwei(wei *big.Int) string {
	gwei := new(big.Float).SetInt(wei)
	gwei.Quo(gwei, big.NewFloat(1e9))
	return fmt.Sprintf("%.2f", gwei)
}

func formatDuration(d time.Duration) string {
	if d.Hours() >= 1 {
		return fmt.Sprintf("%.1f小时", d.Hours())
	} else if d.Minutes() >= 1 {
		return fmt.Sprintf("%.1f分钟", d.Minutes())
	} else {
		return fmt.Sprintf("%.1f秒", d.Seconds())
	}
}
