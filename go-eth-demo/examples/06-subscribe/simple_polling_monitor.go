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
	fmt.Println("🔄 轮询方式区块监控")
	fmt.Println("================================")
	fmt.Println("注意: 这种方式适用于不支持 WebSocket 的 RPC 端点")

	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Printf("警告: 无法加载 .env 文件: %v", err)
	}

	// 获取 HTTP RPC URL
	rpcURL := os.Getenv("ETHEREUM_RPC_URL")
	if rpcURL == "" {
		log.Fatal("请在 .env 文件中设置 ETHEREUM_RPC_URL")
	}

	// 连接到以太坊节点 (HTTP)
	fmt.Printf("连接到: %s\n", rpcURL)
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer client.Close()

	fmt.Println("✅ 连接成功!")

	// 获取当前区块号
	ctx := context.Background()
	latestBlock, err := client.BlockNumber(ctx)
	if err != nil {
		log.Fatalf("获取最新区块号失败: %v", err)
	}
	fmt.Printf("当前区块号: %d\n", latestBlock)

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("\n🔄 开始轮询新区块...")
	fmt.Println("轮询间隔: 10 秒")
	fmt.Println("按 Ctrl+C 停止监控")
	fmt.Println("================================")

	// 轮询参数
	pollInterval := 10 * time.Second
	lastBlockNumber := latestBlock
	blockCount := 0
	startTime := time.Now()

	// 创建定时器
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	// 轮询循环
	for {
		select {
		case <-ticker.C:
			// 检查新区块
			currentBlock, err := client.BlockNumber(ctx)
			if err != nil {
				log.Printf("❌ 获取区块号失败: %v", err)
				continue
			}

			// 如果有新区块
			if currentBlock > lastBlockNumber {
				fmt.Printf("\n🆕 发现 %d 个新区块!\n", currentBlock-lastBlockNumber)

				// 处理所有新区块
				for blockNum := lastBlockNumber + 1; blockNum <= currentBlock; blockNum++ {
					processBlock(ctx, client, blockNum)
					blockCount++
				}
				lastBlockNumber = currentBlock

				// 显示统计信息
				duration := time.Since(startTime)
				if duration.Minutes() >= 1 {
					blocksPerMinute := float64(blockCount) / duration.Minutes()
					fmt.Printf("📊 统计: 已处理 %d 个区块，频率 %.2f 个/分钟\n",
						blockCount, blocksPerMinute)
				}
			} else {
				fmt.Printf("⏳ 等待新区块... (当前: #%d)\n", currentBlock)
			}

		case <-sigChan:
			fmt.Println("\n\n🛑 收到退出信号，停止监控...")

			duration := time.Since(startTime)
			fmt.Printf("总运行时间: %s\n", formatDuration(duration))
			fmt.Printf("总共处理了 %d 个区块\n", blockCount)
			if duration.Minutes() > 0 {
				avgBlocksPerMinute := float64(blockCount) / duration.Minutes()
				fmt.Printf("平均区块频率: %.2f 个/分钟\n", avgBlocksPerMinute)
			}
			fmt.Println("监控已停止!")
			return
		}
	}
}

// processBlock 处理单个区块
func processBlock(ctx context.Context, client *ethclient.Client, blockNumber uint64) {
	// 获取区块信息
	block, err := client.BlockByNumber(ctx, big.NewInt(int64(blockNumber)))
	if err != nil {
		log.Printf("❌ 获取区块 #%d 失败: %v", blockNumber, err)
		return
	}

	fmt.Printf("区块 #%d\n", block.Number().Uint64())
	fmt.Printf("  时间: %s\n", time.Unix(int64(block.Time()), 0).Format("15:04:05"))
	fmt.Printf("  哈希: %s\n", block.Hash().Hex()[:16]+"...")
	fmt.Printf("  交易数: %d 笔\n", len(block.Transactions()))
	fmt.Printf("  Gas 使用: %s/%s (%.1f%%)\n",
		formatGas(block.GasUsed()),
		formatGas(block.GasLimit()),
		float64(block.GasUsed())/float64(block.GasLimit())*100)

	// 分析交易
	if len(block.Transactions()) > 0 {
		analyzeTransactions(block)
	}

	// 检查特殊情况
	checkSpecialConditions(block)
}

// analyzeTransactions 分析交易
func analyzeTransactions(block *types.Block) {
	txs := block.Transactions()
	var transferCount, contractCount int

	for _, tx := range txs {
		if tx.To() == nil {
			contractCount++ // 合约创建
		} else {
			transferCount++ // 转账或合约调用
		}
	}

	if transferCount > 0 {
		fmt.Printf("    转账/调用: %d 笔\n", transferCount)
	}
	if contractCount > 0 {
		fmt.Printf("    合约创建: %d 笔\n", contractCount)
	}
}

// checkSpecialConditions 检查特殊条件
func checkSpecialConditions(block *types.Block) {
	// 检查大区块
	if len(block.Transactions()) > 100 {
		fmt.Printf("  🔥 大区块: 包含 %d 笔交易\n", len(block.Transactions()))
	}

	// 检查高 Gas 使用率
	gasUsagePercent := float64(block.GasUsed()) / float64(block.GasLimit()) * 100
	if gasUsagePercent > 90 {
		fmt.Printf("  ⚡ 高 Gas 使用率: %.2f%%\n", gasUsagePercent)
	}

	// 检查空区块
	if len(block.Transactions()) == 0 {
		fmt.Printf("  📭 空区块\n")
	}
}

// formatGas 格式化 Gas 数量
func formatGas(gas uint64) string {
	if gas >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(gas)/1000000)
	} else if gas >= 1000 {
		return fmt.Sprintf("%.1fK", float64(gas)/1000)
	}
	return fmt.Sprintf("%d", gas)
}

// formatDuration 格式化时间间隔
func formatDuration(d time.Duration) string {
	if d.Hours() >= 1 {
		return fmt.Sprintf("%.1f小时", d.Hours())
	} else if d.Minutes() >= 1 {
		return fmt.Sprintf("%.1f分钟", d.Minutes())
	} else {
		return fmt.Sprintf("%.1f秒", d.Seconds())
	}
}
