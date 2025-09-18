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

// Uniswap V2 Swap 事件结构
type SwapEvent struct {
	Sender     common.Address
	Amount0In  *big.Int
	Amount1In  *big.Int
	Amount0Out *big.Int
	Amount1Out *big.Int
	To         common.Address
}

// Uniswap V2 Sync 事件结构
type SyncEvent struct {
	Reserve0 *big.Int
	Reserve1 *big.Int
}

// Uniswap V2 Pair ABI (只包含事件定义)
const uniswapV2PairABI = `[
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "sender", "type": "address"},
			{"indexed": false, "name": "amount0In", "type": "uint256"},
			{"indexed": false, "name": "amount1In", "type": "uint256"},
			{"indexed": false, "name": "amount0Out", "type": "uint256"},
			{"indexed": false, "name": "amount1Out", "type": "uint256"},
			{"indexed": true, "name": "to", "type": "address"}
		],
		"name": "Swap",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": false, "name": "reserve0", "type": "uint112"},
			{"indexed": false, "name": "reserve1", "type": "uint112"}
		],
		"name": "Sync",
		"type": "event"
	}
]`

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Printf("警告: 无法加载 .env 文件: %v", err)
	}

	fmt.Println("🦄 Uniswap 交易事件监听")
	fmt.Println("================================")

	// 连接以太坊节点
	wsURL := os.Getenv("ETHEREUM_WS_URL")
	if wsURL == "" {
		log.Fatal("请在 .env 文件中设置 ETHEREUM_WS_URL")
	}

	client, err := ethclient.Dial(wsURL)
	if err != nil {
		log.Fatalf("连接以太坊节点失败: %v", err)
	}
	defer client.Close()

	fmt.Printf("连接到: %s\n", wsURL)
	fmt.Println("✅ WebSocket 连接成功!")

	// 解析 ABI
	contractABI, err := abi.JSON(strings.NewReader(uniswapV2PairABI))
	if err != nil {
		log.Fatalf("解析 ABI 失败: %v", err)
	}

	// 监听知名的 Uniswap V2 交易对 (Sepolia 测试网)
	pairs := map[common.Address]PairInfo{
		// 这些是示例地址，实际使用时需要替换为 Sepolia 测试网的真实地址
		common.HexToAddress("0x1234567890123456789012345678901234567890"): {
			Name:   "WETH/USDC",
			Token0: "WETH",
			Token1: "USDC",
		},
		common.HexToAddress("0x2345678901234567890123456789012345678901"): {
			Name:   "WETH/DAI",
			Token0: "WETH",
			Token1: "DAI",
		},
	}

	// 创建事件过滤器
	query := ethereum.FilterQuery{
		Addresses: getPairAddresses(pairs),
		Topics: [][]common.Hash{
			{
				contractABI.Events["Swap"].ID,
				contractABI.Events["Sync"].ID,
			},
		},
	}

	// 订阅事件日志
	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Fatalf("订阅事件失败: %v", err)
	}
	defer sub.Unsubscribe()

	fmt.Println("\n🔄 开始监听 Uniswap 事件...")
	fmt.Println("监听的交易对:")
	for addr, pair := range pairs {
		fmt.Printf("  📍 %s: %s\n", pair.Name, addr.Hex())
	}
	fmt.Println("\n按 Ctrl+C 停止监听")
	fmt.Println("================================\n")

	// 设置优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 统计信息
	var swapCount, syncCount int
	startTime := time.Now()

	// 事件监听循环
	for {
		select {
		case err := <-sub.Err():
			log.Printf("订阅错误: %v", err)
			return

		case vLog := <-logs:
			// 解析事件
			switch vLog.Topics[0] {
			case contractABI.Events["Swap"].ID:
				handleSwapEvent(vLog, contractABI, pairs)
				swapCount++
			case contractABI.Events["Sync"].ID:
				handleSyncEvent(vLog, contractABI, pairs)
				syncCount++
			}

			// 显示统计信息
			if (swapCount+syncCount)%5 == 0 && (swapCount+syncCount) > 0 {
				duration := time.Since(startTime)
				fmt.Printf("\n📊 统计信息 (运行时间: %s)\n", formatDuration(duration))
				fmt.Printf("  Swap 事件: %d 个\n", swapCount)
				fmt.Printf("  Sync 事件: %d 个\n", syncCount)
				fmt.Printf("  总事件数: %d 个\n", swapCount+syncCount)
				if duration.Minutes() > 0 {
					eventsPerMinute := float64(swapCount+syncCount) / duration.Minutes()
					fmt.Printf("  事件频率: %.2f 个/分钟\n", eventsPerMinute)
				}
				fmt.Println("--------------------------------\n")
			}

		case <-sigChan:
			fmt.Println("\n\n🛑 收到退出信号，停止监听...")
			duration := time.Since(startTime)
			fmt.Printf("总运行时间: %s\n", formatDuration(duration))
			fmt.Printf("总共监听到 %d 个事件\n", swapCount+syncCount)
			fmt.Printf("  Swap: %d 个\n", swapCount)
			fmt.Printf("  Sync: %d 个\n", syncCount)
			return
		}
	}
}

// 交易对信息结构
type PairInfo struct {
	Name   string
	Token0 string
	Token1 string
}

// 处理 Swap 事件
func handleSwapEvent(vLog types.Log, contractABI abi.ABI, pairs map[common.Address]PairInfo) {
	var swapEvent SwapEvent

	// 解析事件数据
	err := contractABI.UnpackIntoInterface(&swapEvent, "Swap", vLog.Data)
	if err != nil {
		log.Printf("解析 Swap 事件失败: %v", err)
		return
	}

	// 从 Topics 中获取 indexed 参数
	swapEvent.Sender = common.HexToAddress(vLog.Topics[1].Hex())
	swapEvent.To = common.HexToAddress(vLog.Topics[2].Hex())

	// 获取交易对信息
	pairInfo := pairs[vLog.Address]
	if pairInfo.Name == "" {
		pairInfo.Name = "Unknown Pair"
		pairInfo.Token0 = "Token0"
		pairInfo.Token1 = "Token1"
	}

	fmt.Printf("🔄 Swap 事件\n")
	fmt.Printf("  交易对: %s (%s)\n", pairInfo.Name, vLog.Address.Hex())
	fmt.Printf("  发送者: %s\n", swapEvent.Sender.Hex())
	fmt.Printf("  接收者: %s\n", swapEvent.To.Hex())

	// 分析交易方向
	analyzeSwapDirection(swapEvent, pairInfo)

	fmt.Printf("  区块: #%d\n", vLog.BlockNumber)
	fmt.Printf("  交易: %s\n", vLog.TxHash.Hex())
	fmt.Printf("  时间: %s\n", time.Now().Format("15:04:05"))

	// 检查特殊情况
	checkSpecialSwap(swapEvent, pairInfo)
	fmt.Println()
}

// 处理 Sync 事件
func handleSyncEvent(vLog types.Log, contractABI abi.ABI, pairs map[common.Address]PairInfo) {
	var syncEvent SyncEvent

	// 解析事件数据
	err := contractABI.UnpackIntoInterface(&syncEvent, "Sync", vLog.Data)
	if err != nil {
		log.Printf("解析 Sync 事件失败: %v", err)
		return
	}

	// 获取交易对信息
	pairInfo := pairs[vLog.Address]
	if pairInfo.Name == "" {
		pairInfo.Name = "Unknown Pair"
		pairInfo.Token0 = "Token0"
		pairInfo.Token1 = "Token1"
	}

	// 格式化储备量
	reserve0 := new(big.Float).SetInt(syncEvent.Reserve0)
	reserve0 = reserve0.Quo(reserve0, big.NewFloat(1e18))

	reserve1 := new(big.Float).SetInt(syncEvent.Reserve1)
	reserve1 = reserve1.Quo(reserve1, big.NewFloat(1e18))

	fmt.Printf("⚖️  Sync 事件\n")
	fmt.Printf("  交易对: %s (%s)\n", pairInfo.Name, vLog.Address.Hex())
	fmt.Printf("  %s 储备: %s\n", pairInfo.Token0, reserve0.Text('f', 6))
	fmt.Printf("  %s 储备: %s\n", pairInfo.Token1, reserve1.Text('f', 6))

	// 计算价格比率
	if reserve0.Cmp(big.NewFloat(0)) > 0 && reserve1.Cmp(big.NewFloat(0)) > 0 {
		price := new(big.Float).Quo(reserve1, reserve0)
		fmt.Printf("  价格: 1 %s = %s %s\n", pairInfo.Token0, price.Text('f', 6), pairInfo.Token1)
	}

	fmt.Printf("  区块: #%d\n", vLog.BlockNumber)
	fmt.Printf("  时间: %s\n", time.Now().Format("15:04:05"))
	fmt.Println()
}

// 分析交易方向
func analyzeSwapDirection(swap SwapEvent, pair PairInfo) {
	// 检查输入和输出
	if swap.Amount0In.Cmp(big.NewInt(0)) > 0 {
		// Token0 输入，Token1 输出
		amount0In := new(big.Float).SetInt(swap.Amount0In)
		amount0In = amount0In.Quo(amount0In, big.NewFloat(1e18))

		amount1Out := new(big.Float).SetInt(swap.Amount1Out)
		amount1Out = amount1Out.Quo(amount1Out, big.NewFloat(1e18))

		fmt.Printf("  交易: %s %s → %s %s\n",
			amount0In.Text('f', 6), pair.Token0,
			amount1Out.Text('f', 6), pair.Token1)
	} else if swap.Amount1In.Cmp(big.NewInt(0)) > 0 {
		// Token1 输入，Token0 输出
		amount1In := new(big.Float).SetInt(swap.Amount1In)
		amount1In = amount1In.Quo(amount1In, big.NewFloat(1e18))

		amount0Out := new(big.Float).SetInt(swap.Amount0Out)
		amount0Out = amount0Out.Quo(amount0Out, big.NewFloat(1e18))

		fmt.Printf("  交易: %s %s → %s %s\n",
			amount1In.Text('f', 6), pair.Token1,
			amount0Out.Text('f', 6), pair.Token0)
	}
}

// 检查特殊交易情况
func checkSpecialSwap(swap SwapEvent, pair PairInfo) {
	// 计算总交易量
	totalIn := new(big.Int).Add(swap.Amount0In, swap.Amount1In)
	totalOut := new(big.Int).Add(swap.Amount0Out, swap.Amount1Out)

	// 大额交易检查
	threshold := new(big.Int)
	threshold.SetString("1000000000000000000000", 10) // 1000 tokens

	if totalIn.Cmp(threshold) > 0 || totalOut.Cmp(threshold) > 0 {
		fmt.Printf("  🐋 大额交易: 超过 1000 代币\n")
	}

	// 检查是否为套利交易 (同时有输入和输出)
	if swap.Amount0In.Cmp(big.NewInt(0)) > 0 && swap.Amount1In.Cmp(big.NewInt(0)) > 0 {
		fmt.Printf("  🔄 复杂交易: 同时输入两种代币\n")
	}

	// 检查接收者是否与发送者不同 (可能是代理交易)
	if swap.Sender != swap.To {
		fmt.Printf("  🤝 代理交易: 发送者与接收者不同\n")
	}
}

// 获取交易对地址列表
func getPairAddresses(pairs map[common.Address]PairInfo) []common.Address {
	addresses := make([]common.Address, 0, len(pairs))
	for addr := range pairs {
		addresses = append(addresses, addr)
	}
	return addresses
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
