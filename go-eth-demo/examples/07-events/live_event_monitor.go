package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

// ERC20 Transfer 事件结构
type TransferEvent struct {
	From   common.Address
	To     common.Address
	Amount *big.Int
}

// ERC20 Approval 事件结构
type ApprovalEvent struct {
	Owner   common.Address
	Spender common.Address
	Amount  *big.Int
}

// ERC20 ABI
const erc20EventABI = `[
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "from", "type": "address"},
			{"indexed": true, "name": "to", "type": "address"},
			{"indexed": false, "name": "value", "type": "uint256"}
		],
		"name": "Transfer",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "owner", "type": "address"},
			{"indexed": true, "name": "spender", "type": "address"},
			{"indexed": false, "name": "value", "type": "uint256"}
		],
		"name": "Approval",
		"type": "event"
	}
]`

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Printf("警告: 无法加载 .env 文件: %v", err)
	}

	fmt.Println("🎯 实时事件监听器")
	fmt.Println("================================")

	// 解析 ABI
	contractABI, err := abi.JSON(strings.NewReader(erc20EventABI))
	if err != nil {
		log.Fatalf("解析 ABI 失败: %v", err)
	}

	// 监听知名代币合约 (Sepolia测试网)
	monitoredTokens := map[common.Address]TokenInfo{
		// 这些是示例地址，在实际使用中需要替换为真实的代币地址
		common.HexToAddress("0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984"): {
			Symbol:   "UNI",
			Name:     "Uniswap",
			Decimals: 18,
		},
		common.HexToAddress("0x779877A7B0D9E8603169DdbD7836e478b4624789"): {
			Symbol:   "LINK",
			Name:     "Chainlink",
			Decimals: 18,
		},
	}

	// 尝试WebSocket连接
	wsURL := os.Getenv("ETHEREUM_WS_URL")
	if wsURL != "" {
		fmt.Printf("尝试WebSocket连接: %s\n", wsURL)
		if tryWebSocketMode(wsURL, contractABI, monitoredTokens) {
			return
		}
	}

	// 回退到HTTP轮询模式
	rpcURL := os.Getenv("ETHEREUM_RPC_URL")
	if rpcURL == "" {
		log.Fatal("请在 .env 文件中设置 ETHEREUM_RPC_URL")
	}

	fmt.Printf("使用HTTP轮询模式: %s\n", rpcURL)
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatalf("连接以太坊节点失败: %v", err)
	}
	defer client.Close()

	fmt.Println("✅ HTTP 连接成功!")
	runPollingMode(client, contractABI, monitoredTokens)
}

// 尝试WebSocket模式
func tryWebSocketMode(wsURL string, contractABI abi.ABI, monitoredTokens map[common.Address]TokenInfo) bool {
	client, err := ethclient.Dial(wsURL)
	if err != nil {
		log.Printf("WebSocket连接失败: %v", err)
		return false
	}
	defer client.Close()

	// 简化的事件过滤器 - 只监听Transfer事件
	query := ethereum.FilterQuery{
		Topics: [][]common.Hash{
			{contractABI.Events["Transfer"].ID},
		},
	}

	// 订阅事件日志
	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Printf("WebSocket订阅失败: %v", err)
		return false
	}
	defer sub.Unsubscribe()

	fmt.Println("✅ WebSocket 连接成功!")
	fmt.Println("\n🔄 开始实时监听ERC20 Transfer事件...")
	fmt.Println("监听的代币:")
	for addr, token := range monitoredTokens {
		fmt.Printf("  📍 %s (%s): %s\n", token.Symbol, token.Name, addr.Hex())
	}
	fmt.Println("  🌐 以及所有其他ERC20代币")
	fmt.Println("\n按 Ctrl+C 停止监听")
	fmt.Println("================================\n")

	// 设置优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 统计信息
	stats := EventStats{
		StartTime:    time.Now(),
		UniqueTokens: make(map[common.Address]bool),
		TotalVolume:  big.NewInt(0),
	}

	// 事件监听循环
	for {
		select {
		case err := <-sub.Err():
			log.Printf("订阅错误: %v", err)
			return false

		case vLog := <-logs:
			// 处理Transfer事件
			handleTransferEventWS(vLog, contractABI, monitoredTokens, &stats)

			// 定期显示统计信息
			if stats.TransferCount%10 == 0 && stats.TransferCount > 0 {
				showStatistics(&stats)
			}

		case <-sigChan:
			fmt.Println("\n\n🛑 收到退出信号，停止监听...")
			showFinalStatistics(&stats)
			return true
		}
	}
}

// 轮询模式
func runPollingMode(client *ethclient.Client, contractABI abi.ABI, monitoredTokens map[common.Address]TokenInfo) {
	fmt.Println("\n🔄 开始轮询模式监听事件...")
	fmt.Println("轮询间隔: 15秒")
	fmt.Println("每次查询最近5个区块")
	fmt.Println("\n按 Ctrl+C 停止监听")
	fmt.Println("================================\n")

	// 设置优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 统计信息
	stats := EventStats{
		StartTime:    time.Now(),
		UniqueTokens: make(map[common.Address]bool),
		TotalVolume:  big.NewInt(0),
	}

	// 获取起始区块
	currentBlock, err := client.BlockNumber(context.Background())
	if err != nil {
		log.Fatalf("获取当前区块号失败: %v", err)
	}
	lastCheckedBlock := currentBlock

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	// 轮询循环
	for {
		select {
		case <-ticker.C:
			// 获取当前区块
			currentBlock, err := client.BlockNumber(context.Background())
			if err != nil {
				log.Printf("获取当前区块号失败: %v", err)
				continue
			}

			// 如果有新区块
			if currentBlock > lastCheckedBlock {
				// 查询最近5个区块，但不超过免费版限制
				fromBlock := lastCheckedBlock + 1
				toBlock := currentBlock
				if toBlock-fromBlock > 5 {
					fromBlock = toBlock - 5
				}

				fmt.Printf("🔍 检查区块 #%d - #%d\n", fromBlock, toBlock)

				// 查询Transfer事件
				query := ethereum.FilterQuery{
					FromBlock: big.NewInt(int64(fromBlock)),
					ToBlock:   big.NewInt(int64(toBlock)),
					Topics: [][]common.Hash{
						{contractABI.Events["Transfer"].ID},
					},
				}

				logs, err := client.FilterLogs(context.Background(), query)
				if err != nil {
					log.Printf("查询事件失败: %v", err)
					continue
				}

				if len(logs) > 0 {
					fmt.Printf("📦 发现 %d 个Transfer事件:\n", len(logs))

					// 处理事件
					for i, vLog := range logs {
						if i >= 5 { // 只显示前5个
							fmt.Printf("... 还有 %d 个事件\n", len(logs)-5)
							break
						}
						handleTransferEventPolling(vLog, contractABI, monitoredTokens, &stats)
					}

					// 显示统计
					if stats.TransferCount%20 == 0 && stats.TransferCount > 0 {
						showStatistics(&stats)
					}
				} else {
					fmt.Printf("⏳ 没有新的Transfer事件\n")
				}

				lastCheckedBlock = currentBlock
			} else {
				fmt.Printf("⏳ 等待新区块... (当前: #%d)\n", currentBlock)
			}

		case <-sigChan:
			fmt.Println("\n\n🛑 收到退出信号，停止监听...")
			showFinalStatistics(&stats)
			return
		}
	}
}

// 代币信息结构
type TokenInfo struct {
	Symbol   string
	Name     string
	Decimals int
}

// 统计信息结构
type EventStats struct {
	StartTime      time.Time
	TransferCount  int
	ApprovalCount  int
	UniqueTokens   map[common.Address]bool
	LargeTransfers int
	TotalVolume    *big.Int
}

// 处理WebSocket Transfer事件
func handleTransferEventWS(vLog types.Log, contractABI abi.ABI, tokens map[common.Address]TokenInfo, stats *EventStats) {
	var transfer TransferEvent

	// 解析事件数据
	err := contractABI.UnpackIntoInterface(&transfer, "Transfer", vLog.Data)
	if err != nil {
		log.Printf("解析Transfer事件失败: %v", err)
		return
	}

	// 从Topics中获取indexed参数
	transfer.From = common.HexToAddress(vLog.Topics[1].Hex())
	transfer.To = common.HexToAddress(vLog.Topics[2].Hex())

	// 获取代币信息
	tokenInfo := tokens[vLog.Address]
	if tokenInfo.Symbol == "" {
		tokenInfo.Symbol = "UNKNOWN"
		tokenInfo.Name = "Unknown Token"
		tokenInfo.Decimals = 18
	}

	// 格式化金额
	decimals := big.NewInt(int64(tokenInfo.Decimals))
	divisor := new(big.Int).Exp(big.NewInt(10), decimals, nil)
	amount := new(big.Float).SetInt(transfer.Amount)
	amount = amount.Quo(amount, new(big.Float).SetInt(divisor))

	// 更新统计
	stats.TransferCount++
	stats.UniqueTokens[vLog.Address] = true
	stats.TotalVolume.Add(stats.TotalVolume, transfer.Amount)

	// 检查是否为大额转账
	threshold := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(tokenInfo.Decimals+6)), nil) // 1M tokens
	if transfer.Amount.Cmp(threshold) > 0 {
		stats.LargeTransfers++
	}

	// 显示事件
	fmt.Printf("💸 Transfer | %s (%s)\n", tokenInfo.Symbol, vLog.Address.Hex()[:10]+"...")
	fmt.Printf("   从: %s\n", transfer.From.Hex()[:10]+"...")
	fmt.Printf("   到: %s\n", transfer.To.Hex()[:10]+"...")
	fmt.Printf("   金额: %s %s\n", amount.Text('f', 6), tokenInfo.Symbol)
	fmt.Printf("   区块: #%d | 时间: %s\n", vLog.BlockNumber, time.Now().Format("15:04:05"))

	// 检查特殊情况
	checkSpecialTransferEvent(transfer, amount, tokenInfo.Symbol)
	fmt.Println()
}

// 处理轮询 Transfer事件
func handleTransferEventPolling(vLog types.Log, contractABI abi.ABI, tokens map[common.Address]TokenInfo, stats *EventStats) {
	var transfer TransferEvent

	// 解析事件数据
	err := contractABI.UnpackIntoInterface(&transfer, "Transfer", vLog.Data)
	if err != nil {
		// 忽略解析错误，可能是不同的ABI格式
		return
	}

	// 从Topics中获取indexed参数
	transfer.From = common.HexToAddress(vLog.Topics[1].Hex())
	transfer.To = common.HexToAddress(vLog.Topics[2].Hex())

	// 获取代币信息
	tokenInfo := tokens[vLog.Address]
	if tokenInfo.Symbol == "" {
		tokenInfo.Symbol = "TOKEN"
		tokenInfo.Name = "Unknown Token"
		tokenInfo.Decimals = 18
	}

	// 格式化金额
	decimals := big.NewInt(int64(tokenInfo.Decimals))
	divisor := new(big.Int).Exp(big.NewInt(10), decimals, nil)
	amount := new(big.Float).SetInt(transfer.Amount)
	amount = amount.Quo(amount, new(big.Float).SetInt(divisor))

	// 更新统计
	stats.TransferCount++
	stats.UniqueTokens[vLog.Address] = true

	// 显示事件
	fmt.Printf("  💸 %s | 从 %s 到 %s | %s %s | 区块 #%d\n",
		tokenInfo.Symbol,
		transfer.From.Hex()[:8]+"...",
		transfer.To.Hex()[:8]+"...",
		amount.Text('f', 4),
		tokenInfo.Symbol,
		vLog.BlockNumber)
}

// 检查特殊转账情况
func checkSpecialTransferEvent(transfer TransferEvent, amount *big.Float, tokenSymbol string) {
	// 零地址检查
	zeroAddress := common.HexToAddress("0x0000000000000000000000000000000000000000")

	if transfer.From == zeroAddress {
		fmt.Printf("   🎯 代币铸造 (Mint)\n")
	} else if transfer.To == zeroAddress {
		fmt.Printf("   🔥 代币销毁 (Burn)\n")
	}

	// 大额转账检查
	threshold := big.NewFloat(1000000)
	if amount.Cmp(threshold) > 0 {
		fmt.Printf("   🐋 大额转账: 超过100万 %s\n", tokenSymbol)
	}

	// 小额转账检查
	smallThreshold := big.NewFloat(0.001)
	if amount.Cmp(smallThreshold) < 0 {
		fmt.Printf("   🔍 微小转账: 少于0.001 %s\n", tokenSymbol)
	}
}

// 显示统计信息
func showStatistics(stats *EventStats) {
	duration := time.Since(stats.StartTime)
	totalEvents := stats.TransferCount + stats.ApprovalCount

	fmt.Printf("\n📊 实时统计 (运行时间: %s)\n", formatDuration(duration))
	fmt.Printf("  Transfer事件: %d 个\n", stats.TransferCount)
	if stats.ApprovalCount > 0 {
		fmt.Printf("  Approval事件: %d 个\n", stats.ApprovalCount)
	}
	fmt.Printf("  总事件数: %d 个\n", totalEvents)
	fmt.Printf("  涉及代币: %d 个\n", len(stats.UniqueTokens))
	if stats.LargeTransfers > 0 {
		fmt.Printf("  大额转账: %d 个\n", stats.LargeTransfers)
	}

	if duration.Minutes() > 0 {
		eventsPerMinute := float64(totalEvents) / duration.Minutes()
		fmt.Printf("  事件频率: %.2f 个/分钟\n", eventsPerMinute)
	}
	fmt.Println("--------------------------------\n")
}

// 显示最终统计信息
func showFinalStatistics(stats *EventStats) {
	duration := time.Since(stats.StartTime)
	totalEvents := stats.TransferCount + stats.ApprovalCount

	fmt.Printf("📈 最终统计报告\n")
	fmt.Printf("总运行时间: %s\n", formatDuration(duration))
	fmt.Printf("总事件数: %d 个\n", totalEvents)
	fmt.Printf("  Transfer: %d 个\n", stats.TransferCount)
	if stats.ApprovalCount > 0 {
		fmt.Printf("  Approval: %d 个\n", stats.ApprovalCount)
	}
	fmt.Printf("涉及代币数: %d 个\n", len(stats.UniqueTokens))
	if stats.LargeTransfers > 0 {
		fmt.Printf("大额转账: %d 个\n", stats.LargeTransfers)
	}

	if duration.Minutes() > 0 {
		eventsPerMinute := float64(totalEvents) / duration.Minutes()
		fmt.Printf("平均事件频率: %.2f 个/分钟\n", eventsPerMinute)
	}
}

// 格式化持续时间
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1f秒", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.1f分钟", d.Minutes())
	} else {
		return fmt.Sprintf("%.1f小时", d.Hours())
	}
}
