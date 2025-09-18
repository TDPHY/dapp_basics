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
	fmt.Println("🔔 简单区块订阅演示")
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
	fmt.Printf("连接到: %s\n", rpcURL)
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer client.Close()

	fmt.Println("✅ 连接成功!")

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

	blockCount := 0

	// 订阅循环
	for {
		select {
		case err := <-sub.Err():
			log.Printf("❌ 订阅错误: %v", err)
			return

		case header := <-headers:
			blockCount++

			fmt.Printf("\n🆕 区块 #%d\n", header.Number.Uint64())
			fmt.Printf("时间: %s\n", time.Unix(int64(header.Time), 0).Format("15:04:05"))
			fmt.Printf("哈希: %s\n", header.Hash().Hex()[:10]+"...")
			fmt.Printf("Gas 使用: %s/%s (%.1f%%)\n",
				formatGas(header.GasUsed),
				formatGas(header.GasLimit),
				float64(header.GasUsed)/float64(header.GasLimit)*100)

			// 获取交易数量
			block, err := client.BlockByHash(ctx, header.Hash())
			if err == nil {
				fmt.Printf("交易数: %d 笔\n", len(block.Transactions()))
			}

			fmt.Printf("已接收: %d 个区块\n", blockCount)

		case <-sigChan:
			fmt.Println("\n\n🛑 收到退出信号，停止订阅...")
			cancel()
			fmt.Printf("总共接收了 %d 个区块\n", blockCount)
			fmt.Println("订阅已停止!")
			return
		}
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
