package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("🔔 WebSocket 区块订阅演示")
	fmt.Println("================================")

	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Printf("警告: 无法加载 .env 文件: %v", err)
	}

	// 获取 WebSocket URL
	wsURL := os.Getenv("ETHEREUM_WS_URL")
	if wsURL == "" {
		log.Fatal("请在 .env 文件中设置 ETHEREUM_WS_URL")
	}

	// 连接到以太坊节点 (WebSocket)
	fmt.Printf("连接到 WebSocket: %s\n", wsURL)
	client, err := ethclient.Dial(wsURL)
	if err != nil {
		log.Fatalf("WebSocket 连接失败: %v", err)
	}
	defer client.Close()

	fmt.Println("✅ WebSocket 连接成功!")

	// 获取当前区块号
	latestBlock, err := client.BlockNumber(context.Background())
	if err != nil {
		log.Fatalf("获取最新区块号失败: %v", err)
	}
	fmt.Printf("当前区块号: %d\n", latestBlock)

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("\n🔔 开始订阅新区块...")
	fmt.Println("按 Ctrl+C 停止订阅")
	fmt.Println("================================")

	// 创建区块头订阅
	headers := make(chan *types.Header)
	sub, err := client.SubscribeNewHead(ctx, headers)
	if err != nil {
		log.Fatalf("创建订阅失败: %v", err)
	}
	defer sub.Unsubscribe()

	fmt.Println("✅ 区块订阅创建成功!")
	fmt.Println("等待新区块...")

	blockCount := 0
	startTime := time.Now()

	// 订阅循环
	for {
		select {
		case err := <-sub.Err():
			log.Printf("❌ 订阅错误: %v", err)
			fmt.Println("🔄 尝试重新连接...")
			return

		case header := <-headers:
			blockCount++

			fmt.Printf("\n🆕 区块 #%d\n", header.Number.Uint64())
			fmt.Printf("时间: %s\n", time.Unix(int64(header.Time), 0).Format("15:04:05"))
			fmt.Printf("哈希: %s\n", header.Hash().Hex()[:16]+"...")
			fmt.Printf("Gas 使用: %s/%s (%.1f%%)\n",
				formatGas(header.GasUsed),
				formatGas(header.GasLimit),
				float64(header.GasUsed)/float64(header.GasLimit)*100)

			// 获取完整区块信息
			block, err := client.BlockByHash(ctx, header.Hash())
			if err == nil {
				fmt.Printf("交易数: %d 笔\n", len(block.Transactions()))

				// 分析交易类型
				if len(block.Transactions()) > 0 {
					analyzeTransactions(block.Transactions())
				}
			}

			fmt.Printf("已接收: %d 个区块\n", blockCount)

			// 显示运行统计
			duration := time.Since(startTime)
			if duration.Minutes() >= 1 {
				blocksPerMinute := float64(blockCount) / duration.Minutes()
				fmt.Printf("区块频率: %.2f 个/分钟\n", blocksPerMinute)
			}

		case <-sigChan:
			fmt.Println("\n\n🛑 收到退出信号，停止订阅...")
			cancel()

			duration := time.Since(startTime)
			fmt.Printf("总运行时间: %s\n", formatDuration(duration))
			fmt.Printf("总共接收了 %d 个区块\n", blockCount)
			if duration.Minutes() > 0 {
				avgBlocksPerMinute := float64(blockCount) / duration.Minutes()
				fmt.Printf("平均区块频率: %.2f 个/分钟\n", avgBlocksPerMinute)
			}
			fmt.Println("订阅已停止!")
			return
		}
	}
}

// analyzeTransactions 简单分析交易
func analyzeTransactions(txs types.Transactions) {
	var transferCount, contractCount int

	for _, tx := range txs {
		if tx.To() == nil {
			contractCount++ // 合约创建
		} else {
			transferCount++ // 转账或合约调用
		}
	}

	if transferCount > 0 {
		fmt.Printf("  转账/调用: %d 笔", transferCount)
	}
	if contractCount > 0 {
		fmt.Printf("  合约创建: %d 笔", contractCount)
	}
	if transferCount > 0 || contractCount > 0 {
		fmt.Println()
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
