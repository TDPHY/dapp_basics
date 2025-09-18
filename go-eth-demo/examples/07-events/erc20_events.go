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
type Transfer struct {
	From   common.Address
	To     common.Address
	Amount *big.Int
}

// ERC20 Approval 事件结构
type Approval struct {
	Owner   common.Address
	Spender common.Address
	Amount  *big.Int
}

// ERC20 ABI (只包含事件定义)
const erc20ABI = `[
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

	fmt.Println("🎯 ERC-20 代币事件监听")
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
	contractABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		log.Fatalf("解析 ABI 失败: %v", err)
	}

	// 监听多个知名 ERC-20 代币
	tokens := map[common.Address]string{
		common.HexToAddress("0xA0b86a33E6441b8435b662f0E2d0B8A0E4B2B8B0"): "USDC (Sepolia)",
		common.HexToAddress("0x779877A7B0D9E8603169DdbD7836e478b4624789"): "LINK (Sepolia)",
		common.HexToAddress("0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984"): "UNI (Sepolia)",
	}

	// 创建事件过滤器
	query := ethereum.FilterQuery{
		Addresses: getTokenAddresses(tokens),
		Topics: [][]common.Hash{
			{
				contractABI.Events["Transfer"].ID,
				contractABI.Events["Approval"].ID,
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

	fmt.Println("\n🔄 开始监听 ERC-20 事件...")
	fmt.Println("监听的代币:")
	for addr, name := range tokens {
		fmt.Printf("  📍 %s: %s\n", name, addr.Hex())
	}
	fmt.Println("\n按 Ctrl+C 停止监听")
	fmt.Println("================================\n")

	// 设置优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 统计信息
	var transferCount, approvalCount int
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
			case contractABI.Events["Transfer"].ID:
				handleTransferEvent(vLog, contractABI, tokens)
				transferCount++
			case contractABI.Events["Approval"].ID:
				handleApprovalEvent(vLog, contractABI, tokens)
				approvalCount++
			}

			// 显示统计信息
			if (transferCount+approvalCount)%10 == 0 {
				duration := time.Since(startTime)
				fmt.Printf("\n📊 统计信息 (运行时间: %s)\n", formatDuration(duration))
				fmt.Printf("  Transfer 事件: %d 个\n", transferCount)
				fmt.Printf("  Approval 事件: %d 个\n", approvalCount)
				fmt.Printf("  总事件数: %d 个\n", transferCount+approvalCount)
				fmt.Println("--------------------------------\n")
			}

		case <-sigChan:
			fmt.Println("\n\n🛑 收到退出信号，停止监听...")
			duration := time.Since(startTime)
			fmt.Printf("总运行时间: %s\n", formatDuration(duration))
			fmt.Printf("总共监听到 %d 个事件\n", transferCount+approvalCount)
			fmt.Printf("  Transfer: %d 个\n", transferCount)
			fmt.Printf("  Approval: %d 个\n", approvalCount)
			return
		}
	}
}

// 处理 Transfer 事件
func handleTransferEvent(vLog types.Log, contractABI abi.ABI, tokens map[common.Address]string) {
	var transferEvent Transfer

	// 解析事件数据
	err := contractABI.UnpackIntoInterface(&transferEvent, "Transfer", vLog.Data)
	if err != nil {
		log.Printf("解析 Transfer 事件失败: %v", err)
		return
	}

	// 从 Topics 中获取 indexed 参数
	transferEvent.From = common.HexToAddress(vLog.Topics[1].Hex())
	transferEvent.To = common.HexToAddress(vLog.Topics[2].Hex())

	// 获取代币信息
	tokenName := tokens[vLog.Address]
	if tokenName == "" {
		tokenName = "Unknown Token"
	}

	// 格式化金额 (假设 18 位小数)
	amount := new(big.Float).SetInt(transferEvent.Amount)
	amount = amount.Quo(amount, big.NewFloat(1e18))

	fmt.Printf("💸 Transfer 事件\n")
	fmt.Printf("  代币: %s (%s)\n", tokenName, vLog.Address.Hex())
	fmt.Printf("  从: %s\n", transferEvent.From.Hex())
	fmt.Printf("  到: %s\n", transferEvent.To.Hex())
	fmt.Printf("  金额: %s 代币\n", amount.Text('f', 6))
	fmt.Printf("  区块: #%d\n", vLog.BlockNumber)
	fmt.Printf("  交易: %s\n", vLog.TxHash.Hex())
	fmt.Printf("  时间: %s\n", time.Now().Format("15:04:05"))

	// 检查特殊情况
	checkSpecialTransfer(transferEvent, amount, tokenName)
	fmt.Println()
}

// 处理 Approval 事件
func handleApprovalEvent(vLog types.Log, contractABI abi.ABI, tokens map[common.Address]string) {
	var approvalEvent Approval

	// 解析事件数据
	err := contractABI.UnpackIntoInterface(&approvalEvent, "Approval", vLog.Data)
	if err != nil {
		log.Printf("解析 Approval 事件失败: %v", err)
		return
	}

	// 从 Topics 中获取 indexed 参数
	approvalEvent.Owner = common.HexToAddress(vLog.Topics[1].Hex())
	approvalEvent.Spender = common.HexToAddress(vLog.Topics[2].Hex())

	// 获取代币信息
	tokenName := tokens[vLog.Address]
	if tokenName == "" {
		tokenName = "Unknown Token"
	}

	// 格式化金额
	amount := new(big.Float).SetInt(approvalEvent.Amount)
	amount = amount.Quo(amount, big.NewFloat(1e18))

	fmt.Printf("✅ Approval 事件\n")
	fmt.Printf("  代币: %s (%s)\n", tokenName, vLog.Address.Hex())
	fmt.Printf("  所有者: %s\n", approvalEvent.Owner.Hex())
	fmt.Printf("  被授权者: %s\n", approvalEvent.Spender.Hex())
	fmt.Printf("  授权金额: %s 代币\n", amount.Text('f', 6))
	fmt.Printf("  区块: #%d\n", vLog.BlockNumber)
	fmt.Printf("  交易: %s\n", vLog.TxHash.Hex())
	fmt.Printf("  时间: %s\n", time.Now().Format("15:04:05"))

	// 检查特殊情况
	checkSpecialApproval(approvalEvent, amount, tokenName)
	fmt.Println()
}

// 检查特殊转账情况
func checkSpecialTransfer(transfer Transfer, amount *big.Float, tokenName string) {
	// 零地址检查 (铸造/销毁)
	zeroAddress := common.HexToAddress("0x0000000000000000000000000000000000000000")

	if transfer.From == zeroAddress {
		fmt.Printf("  🎯 特殊事件: 代币铸造 (Mint)\n")
	} else if transfer.To == zeroAddress {
		fmt.Printf("  🔥 特殊事件: 代币销毁 (Burn)\n")
	}

	// 大额转账检查
	threshold := big.NewFloat(1000000) // 100万代币
	if amount.Cmp(threshold) > 0 {
		fmt.Printf("  🐋 大额转账: 超过 100万 %s\n", tokenName)
	}

	// 小额转账检查
	smallThreshold := big.NewFloat(0.001)
	if amount.Cmp(smallThreshold) < 0 {
		fmt.Printf("  🔍 微小转账: 少于 0.001 %s\n", tokenName)
	}
}

// 检查特殊授权情况
func checkSpecialApproval(approval Approval, amount *big.Float, tokenName string) {
	// 无限授权检查
	maxUint256 := new(big.Int)
	maxUint256.SetString("115792089237316195423570985008687907853269984665640564039457584007913129639935", 10)

	if approval.Amount.Cmp(maxUint256) == 0 {
		fmt.Printf("  ♾️  无限授权: 最大 uint256 值\n")
	}

	// 零授权检查 (撤销授权)
	if approval.Amount.Cmp(big.NewInt(0)) == 0 {
		fmt.Printf("  🚫 撤销授权: 授权金额为 0\n")
	}

	// 大额授权检查
	threshold := big.NewFloat(1000000)
	if amount.Cmp(threshold) > 0 {
		fmt.Printf("  ⚠️  大额授权: 超过 100万 %s\n", tokenName)
	}
}

// 获取代币地址列表
func getTokenAddresses(tokens map[common.Address]string) []common.Address {
	addresses := make([]common.Address, 0, len(tokens))
	for addr := range tokens {
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
