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

// 自定义合约事件结构
type UserRegistered struct {
	User      common.Address
	Username  string
	Timestamp *big.Int
}

type ItemCreated struct {
	ItemId   *big.Int
	Creator  common.Address
	Name     string
	Price    *big.Int
	Category string
}

type OrderPlaced struct {
	OrderId *big.Int
	Buyer   common.Address
	Seller  common.Address
	ItemId  *big.Int
	Amount  *big.Int
	Status  uint8
}

// 自定义合约 ABI
const customContractABI = `[
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "user", "type": "address"},
			{"indexed": false, "name": "username", "type": "string"},
			{"indexed": false, "name": "timestamp", "type": "uint256"}
		],
		"name": "UserRegistered",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "itemId", "type": "uint256"},
			{"indexed": true, "name": "creator", "type": "address"},
			{"indexed": false, "name": "name", "type": "string"},
			{"indexed": false, "name": "price", "type": "uint256"},
			{"indexed": false, "name": "category", "type": "string"}
		],
		"name": "ItemCreated",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "orderId", "type": "uint256"},
			{"indexed": true, "name": "buyer", "type": "address"},
			{"indexed": true, "name": "seller", "type": "address"},
			{"indexed": false, "name": "itemId", "type": "uint256"},
			{"indexed": false, "name": "amount", "type": "uint256"},
			{"indexed": false, "name": "status", "type": "uint8"}
		],
		"name": "OrderPlaced",
		"type": "event"
	}
]`

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Printf("警告: 无法加载 .env 文件: %v", err)
	}

	fmt.Println("🎨 自定义合约事件监听")
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
	contractABI, err := abi.JSON(strings.NewReader(customContractABI))
	if err != nil {
		log.Fatalf("解析 ABI 失败: %v", err)
	}

	// 监听的自定义合约地址 (示例地址，需要替换为实际部署的合约)
	contracts := map[common.Address]string{
		common.HexToAddress("0x1111111111111111111111111111111111111111"): "用户管理合约",
		common.HexToAddress("0x2222222222222222222222222222222222222222"): "商品管理合约",
		common.HexToAddress("0x3333333333333333333333333333333333333333"): "订单管理合约",
	}

	// 创建事件过滤器 - 监听所有自定义事件
	query := ethereum.FilterQuery{
		Addresses: getContractAddresses(contracts),
		Topics: [][]common.Hash{
			{
				contractABI.Events["UserRegistered"].ID,
				contractABI.Events["ItemCreated"].ID,
				contractABI.Events["OrderPlaced"].ID,
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

	fmt.Println("\n🔄 开始监听自定义合约事件...")
	fmt.Println("监听的合约:")
	for addr, name := range contracts {
		fmt.Printf("  📍 %s: %s\n", name, addr.Hex())
	}
	fmt.Println("\n按 Ctrl+C 停止监听")
	fmt.Println("================================\n")

	// 设置优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 统计信息
	eventCounts := make(map[string]int)
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
			case contractABI.Events["UserRegistered"].ID:
				handleUserRegisteredEvent(vLog, contractABI, contracts)
				eventCounts["UserRegistered"]++
			case contractABI.Events["ItemCreated"].ID:
				handleItemCreatedEvent(vLog, contractABI, contracts)
				eventCounts["ItemCreated"]++
			case contractABI.Events["OrderPlaced"].ID:
				handleOrderPlacedEvent(vLog, contractABI, contracts)
				eventCounts["OrderPlaced"]++
			}

			// 显示统计信息
			totalEvents := getTotalEvents(eventCounts)
			if totalEvents%5 == 0 && totalEvents > 0 {
				duration := time.Since(startTime)
				fmt.Printf("\n📊 统计信息 (运行时间: %s)\n", formatDuration(duration))
				for eventType, count := range eventCounts {
					fmt.Printf("  %s: %d 个\n", eventType, count)
				}
				fmt.Printf("  总事件数: %d 个\n", totalEvents)
				if duration.Minutes() > 0 {
					eventsPerMinute := float64(totalEvents) / duration.Minutes()
					fmt.Printf("  事件频率: %.2f 个/分钟\n", eventsPerMinute)
				}
				fmt.Println("--------------------------------\n")
			}

		case <-sigChan:
			fmt.Println("\n\n🛑 收到退出信号，停止监听...")
			duration := time.Since(startTime)
			fmt.Printf("总运行时间: %s\n", formatDuration(duration))
			totalEvents := getTotalEvents(eventCounts)
			fmt.Printf("总共监听到 %d 个事件\n", totalEvents)
			for eventType, count := range eventCounts {
				fmt.Printf("  %s: %d 个\n", eventType, count)
			}
			return
		}
	}
}

// 处理用户注册事件
func handleUserRegisteredEvent(vLog types.Log, contractABI abi.ABI, contracts map[common.Address]string) {
	var event UserRegistered

	// 解析事件数据
	err := contractABI.UnpackIntoInterface(&event, "UserRegistered", vLog.Data)
	if err != nil {
		log.Printf("解析 UserRegistered 事件失败: %v", err)
		return
	}

	// 从 Topics 中获取 indexed 参数
	event.User = common.HexToAddress(vLog.Topics[1].Hex())

	// 获取合约信息
	contractName := contracts[vLog.Address]
	if contractName == "" {
		contractName = "Unknown Contract"
	}

	// 格式化时间戳
	timestamp := time.Unix(event.Timestamp.Int64(), 0)

	fmt.Printf("👤 UserRegistered 事件\n")
	fmt.Printf("  合约: %s (%s)\n", contractName, vLog.Address.Hex())
	fmt.Printf("  用户地址: %s\n", event.User.Hex())
	fmt.Printf("  用户名: %s\n", event.Username)
	fmt.Printf("  注册时间: %s\n", timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("  区块: #%d\n", vLog.BlockNumber)
	fmt.Printf("  交易: %s\n", vLog.TxHash.Hex())
	fmt.Printf("  当前时间: %s\n", time.Now().Format("15:04:05"))

	// 检查特殊情况
	checkSpecialUser(event)
	fmt.Println()
}

// 处理商品创建事件
func handleItemCreatedEvent(vLog types.Log, contractABI abi.ABI, contracts map[common.Address]string) {
	var event ItemCreated

	// 解析事件数据
	err := contractABI.UnpackIntoInterface(&event, "ItemCreated", vLog.Data)
	if err != nil {
		log.Printf("解析 ItemCreated 事件失败: %v", err)
		return
	}

	// 从 Topics 中获取 indexed 参数
	event.ItemId = vLog.Topics[1].Big()
	event.Creator = common.HexToAddress(vLog.Topics[2].Hex())

	// 获取合约信息
	contractName := contracts[vLog.Address]
	if contractName == "" {
		contractName = "Unknown Contract"
	}

	// 格式化价格
	price := new(big.Float).SetInt(event.Price)
	price = price.Quo(price, big.NewFloat(1e18))

	fmt.Printf("🛍️ ItemCreated 事件\n")
	fmt.Printf("  合约: %s (%s)\n", contractName, vLog.Address.Hex())
	fmt.Printf("  商品ID: %s\n", event.ItemId.String())
	fmt.Printf("  创建者: %s\n", event.Creator.Hex())
	fmt.Printf("  商品名称: %s\n", event.Name)
	fmt.Printf("  价格: %s ETH\n", price.Text('f', 6))
	fmt.Printf("  分类: %s\n", event.Category)
	fmt.Printf("  区块: #%d\n", vLog.BlockNumber)
	fmt.Printf("  交易: %s\n", vLog.TxHash.Hex())
	fmt.Printf("  当前时间: %s\n", time.Now().Format("15:04:05"))

	// 检查特殊情况
	checkSpecialItem(event, price)
	fmt.Println()
}

// 处理订单创建事件
func handleOrderPlacedEvent(vLog types.Log, contractABI abi.ABI, contracts map[common.Address]string) {
	var event OrderPlaced

	// 解析事件数据
	err := contractABI.UnpackIntoInterface(&event, "OrderPlaced", vLog.Data)
	if err != nil {
		log.Printf("解析 OrderPlaced 事件失败: %v", err)
		return
	}

	// 从 Topics 中获取 indexed 参数
	event.OrderId = vLog.Topics[1].Big()
	event.Buyer = common.HexToAddress(vLog.Topics[2].Hex())
	event.Seller = common.HexToAddress(vLog.Topics[3].Hex())

	// 获取合约信息
	contractName := contracts[vLog.Address]
	if contractName == "" {
		contractName = "Unknown Contract"
	}

	// 格式化金额
	amount := new(big.Float).SetInt(event.Amount)
	amount = amount.Quo(amount, big.NewFloat(1e18))

	// 订单状态
	statusNames := []string{"待付款", "已付款", "已发货", "已完成", "已取消"}
	statusName := "未知状态"
	if int(event.Status) < len(statusNames) {
		statusName = statusNames[event.Status]
	}

	fmt.Printf("📦 OrderPlaced 事件\n")
	fmt.Printf("  合约: %s (%s)\n", contractName, vLog.Address.Hex())
	fmt.Printf("  订单ID: %s\n", event.OrderId.String())
	fmt.Printf("  买家: %s\n", event.Buyer.Hex())
	fmt.Printf("  卖家: %s\n", event.Seller.Hex())
	fmt.Printf("  商品ID: %s\n", event.ItemId.String())
	fmt.Printf("  金额: %s ETH\n", amount.Text('f', 6))
	fmt.Printf("  状态: %s (%d)\n", statusName, event.Status)
	fmt.Printf("  区块: #%d\n", vLog.BlockNumber)
	fmt.Printf("  交易: %s\n", vLog.TxHash.Hex())
	fmt.Printf("  当前时间: %s\n", time.Now().Format("15:04:05"))

	// 检查特殊情况
	checkSpecialOrder(event, amount)
	fmt.Println()
}

// 检查特殊用户情况
func checkSpecialUser(event UserRegistered) {
	// 检查用户名长度
	if len(event.Username) < 3 {
		fmt.Printf("  ⚠️  短用户名: 少于3个字符\n")
	} else if len(event.Username) > 20 {
		fmt.Printf("  📏 长用户名: 超过20个字符\n")
	}

	// 检查是否包含特殊字符
	if strings.ContainsAny(event.Username, "!@#$%^&*()") {
		fmt.Printf("  ✨ 特殊字符: 用户名包含特殊字符\n")
	}

	// 检查注册时间
	now := time.Now()
	regTime := time.Unix(event.Timestamp.Int64(), 0)
	if now.Sub(regTime) < time.Minute {
		fmt.Printf("  🆕 新注册: 刚刚注册的用户\n")
	}
}

// 检查特殊商品情况
func checkSpecialItem(event ItemCreated, price *big.Float) {
	// 高价商品检查
	highPrice := big.NewFloat(10) // 10 ETH
	if price.Cmp(highPrice) > 0 {
		fmt.Printf("  💎 高价商品: 价格超过 10 ETH\n")
	}

	// 免费商品检查
	if price.Cmp(big.NewFloat(0)) == 0 {
		fmt.Printf("  🆓 免费商品: 价格为 0\n")
	}

	// 商品名称检查
	if len(event.Name) > 50 {
		fmt.Printf("  📝 长名称: 商品名称超过50个字符\n")
	}

	// 分类检查
	popularCategories := []string{"电子产品", "服装", "书籍", "家居", "运动"}
	isPopular := false
	for _, cat := range popularCategories {
		if event.Category == cat {
			isPopular = true
			break
		}
	}
	if !isPopular {
		fmt.Printf("  🔍 特殊分类: %s\n", event.Category)
	}
}

// 检查特殊订单情况
func checkSpecialOrder(event OrderPlaced, amount *big.Float) {
	// 大额订单检查
	bigAmount := big.NewFloat(5) // 5 ETH
	if amount.Cmp(bigAmount) > 0 {
		fmt.Printf("  🐋 大额订单: 金额超过 5 ETH\n")
	}

	// 自买自卖检查
	if event.Buyer == event.Seller {
		fmt.Printf("  🔄 自交易: 买家和卖家是同一人\n")
	}

	// 订单状态检查
	if event.Status == 4 { // 已取消
		fmt.Printf("  ❌ 已取消订单\n")
	} else if event.Status == 3 { // 已完成
		fmt.Printf("  ✅ 已完成订单\n")
	}
}

// 获取合约地址列表
func getContractAddresses(contracts map[common.Address]string) []common.Address {
	addresses := make([]common.Address, 0, len(contracts))
	for addr := range contracts {
		addresses = append(addresses, addr)
	}
	return addresses
}

// 计算总事件数
func getTotalEvents(eventCounts map[string]int) int {
	total := 0
	for _, count := range eventCounts {
		total += count
	}
	return total
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
