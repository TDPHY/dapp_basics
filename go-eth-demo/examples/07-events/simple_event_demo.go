package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

// ERC20 Transfer 事件结构
type Transfer struct {
	From   common.Address
	To     common.Address
	Amount *big.Int
}

// ERC20 ABI
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
	}
]`

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Printf("警告: 无法加载 .env 文件: %v", err)
	}

	fmt.Println("🔍 简单事件查询演示")
	fmt.Println("================================")

	// 连接以太坊节点
	rpcURL := os.Getenv("ETHEREUM_RPC_URL")
	if rpcURL == "" {
		log.Fatal("请在 .env 文件中设置 ETHEREUM_RPC_URL")
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatalf("连接以太坊节点失败: %v", err)
	}
	defer client.Close()

	fmt.Printf("连接到: %s\n", rpcURL)
	fmt.Println("✅ 连接成功!")

	// 解析 ABI
	contractABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		log.Fatalf("解析 ABI 失败: %v", err)
	}

	// 获取当前区块号
	currentBlock, err := client.BlockNumber(context.Background())
	if err != nil {
		log.Fatalf("获取当前区块号失败: %v", err)
	}

	fmt.Printf("当前区块号: #%d\n", currentBlock)
	fmt.Println("================================\n")

	// 示例1: 查询最近10个区块的所有Transfer事件
	fmt.Println("📊 示例1: 查询最近10个区块的Transfer事件")
	queryRecentTransfers(client, contractABI, currentBlock)

	// 示例2: 查询特定区块的事件
	fmt.Println("\n📊 示例2: 查询特定区块的事件")
	querySpecificBlock(client, contractABI, currentBlock)

	// 示例3: 查询知名地址的事件
	fmt.Println("\n📊 示例3: 查询知名地址的事件")
	queryKnownAddresses(client, contractABI, currentBlock)
}

// 示例1: 查询最近10个区块的所有Transfer事件
func queryRecentTransfers(client *ethclient.Client, contractABI abi.ABI, currentBlock uint64) {
	// 查询最近10个区块 (符合免费版限制)
	fromBlock := currentBlock - 9
	if fromBlock > currentBlock {
		fromBlock = 0
	}

	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(currentBlock)),
		Topics: [][]common.Hash{
			{contractABI.Events["Transfer"].ID}, // Transfer 事件
		},
	}

	fmt.Printf("查询区块范围: #%d - #%d (10个区块)\n", fromBlock, currentBlock)

	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		log.Printf("查询事件失败: %v", err)
		return
	}

	fmt.Printf("找到 %d 个Transfer事件:\n", len(logs))

	// 按代币地址分组统计
	tokenStats := make(map[common.Address]int)
	for _, vLog := range logs {
		tokenStats[vLog.Address]++
	}

	fmt.Println("按代币统计:")
	count := 0
	for tokenAddr, eventCount := range tokenStats {
		if count >= 10 {
			fmt.Printf("... 还有 %d 个代币\n", len(tokenStats)-10)
			break
		}
		fmt.Printf("  %s: %d 个事件\n", tokenAddr.Hex(), eventCount)
		count++
	}

	// 显示前几个具体事件
	fmt.Println("\n具体事件详情:")
	for i, vLog := range logs {
		if i >= 5 { // 只显示前5个
			fmt.Printf("... 还有 %d 个事件\n", len(logs)-5)
			break
		}

		var transfer Transfer
		err := contractABI.UnpackIntoInterface(&transfer, "Transfer", vLog.Data)
		if err != nil {
			log.Printf("解析事件失败: %v", err)
			continue
		}

		transfer.From = common.HexToAddress(vLog.Topics[1].Hex())
		transfer.To = common.HexToAddress(vLog.Topics[2].Hex())

		amount := new(big.Float).SetInt(transfer.Amount)
		amount = amount.Quo(amount, big.NewFloat(1e18))

		fmt.Printf("  %d. 代币: %s\n", i+1, vLog.Address.Hex()[:10]+"...")
		fmt.Printf("     从: %s\n", transfer.From.Hex()[:10]+"...")
		fmt.Printf("     到: %s\n", transfer.To.Hex()[:10]+"...")
		fmt.Printf("     金额: %s\n", amount.Text('f', 6))
		fmt.Printf("     区块: #%d\n", vLog.BlockNumber)

		// 检查特殊情况
		checkSpecialTransfer(transfer, amount)
		fmt.Println()
	}
}

// 示例2: 查询特定区块的事件
func querySpecificBlock(client *ethclient.Client, contractABI abi.ABI, currentBlock uint64) {
	// 查询当前区块
	targetBlock := currentBlock

	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(targetBlock)),
		ToBlock:   big.NewInt(int64(targetBlock)),
		Topics: [][]common.Hash{
			{contractABI.Events["Transfer"].ID},
		},
	}

	fmt.Printf("查询区块: #%d\n", targetBlock)

	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		log.Printf("查询事件失败: %v", err)
		return
	}

	fmt.Printf("区块 #%d 中找到 %d 个Transfer事件\n", targetBlock, len(logs))

	if len(logs) == 0 {
		fmt.Println("该区块中没有Transfer事件")
		return
	}

	// 分析事件
	var totalAmount *big.Int = big.NewInt(0)
	uniqueTokens := make(map[common.Address]bool)

	for _, vLog := range logs {
		var transfer Transfer
		err := contractABI.UnpackIntoInterface(&transfer, "Transfer", vLog.Data)
		if err != nil {
			continue
		}

		totalAmount.Add(totalAmount, transfer.Amount)
		uniqueTokens[vLog.Address] = true
	}

	fmt.Printf("统计信息:\n")
	fmt.Printf("  涉及代币数: %d 个\n", len(uniqueTokens))

	totalAmountFloat := new(big.Float).SetInt(totalAmount)
	totalAmountFloat = totalAmountFloat.Quo(totalAmountFloat, big.NewFloat(1e18))
	fmt.Printf("  总转账量: %s (按18位小数计算)\n", totalAmountFloat.Text('f', 2))
}

// 示例3: 查询知名地址的事件
func queryKnownAddresses(client *ethclient.Client, contractABI abi.ABI, currentBlock uint64) {
	// 一些知名地址
	knownAddresses := map[common.Address]string{
		common.HexToAddress("0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045"): "Vitalik Buterin",
		common.HexToAddress("0x3f5CE5FBFe3E9af3971dD833D26bA9b5C936f0bE"): "Binance",
		common.HexToAddress("0x28C6c06298d514Db089934071355E5743bf21d60"): "Binance 2",
		common.HexToAddress("0x21a31Ee1afC51d94C2eFcCAa2092aD1028285549"): "Binance 3",
	}

	// 查询最近5个区块
	fromBlock := currentBlock - 4
	if fromBlock > currentBlock {
		fromBlock = 0
	}

	fmt.Printf("查询知名地址在最近5个区块的活动\n")
	fmt.Printf("区块范围: #%d - #%d\n", fromBlock, currentBlock)

	for addr, name := range knownAddresses {
		fmt.Printf("\n🔍 查询 %s (%s):\n", name, addr.Hex()[:10]+"...")

		// 查询该地址作为发送方的事件
		queryFrom := ethereum.FilterQuery{
			FromBlock: big.NewInt(int64(fromBlock)),
			ToBlock:   big.NewInt(int64(currentBlock)),
			Topics: [][]common.Hash{
				{contractABI.Events["Transfer"].ID},
				{common.BytesToHash(addr.Bytes())}, // from 地址
			},
		}

		logsFrom, err := client.FilterLogs(context.Background(), queryFrom)
		if err != nil {
			log.Printf("查询发送事件失败: %v", err)
			continue
		}

		// 查询该地址作为接收方的事件
		queryTo := ethereum.FilterQuery{
			FromBlock: big.NewInt(int64(fromBlock)),
			ToBlock:   big.NewInt(int64(currentBlock)),
			Topics: [][]common.Hash{
				{contractABI.Events["Transfer"].ID},
				{},                                 // from 可以是任何地址
				{common.BytesToHash(addr.Bytes())}, // to 地址
			},
		}

		logsTo, err := client.FilterLogs(context.Background(), queryTo)
		if err != nil {
			log.Printf("查询接收事件失败: %v", err)
			continue
		}

		fmt.Printf("  发送交易: %d 笔\n", len(logsFrom))
		fmt.Printf("  接收交易: %d 笔\n", len(logsTo))

		if len(logsFrom) == 0 && len(logsTo) == 0 {
			fmt.Printf("  📭 该地址在此期间没有活动\n")
		} else {
			fmt.Printf("  📈 该地址在此期间比较活跃\n")
		}
	}
}

// 检查特殊转账情况
func checkSpecialTransfer(transfer Transfer, amount *big.Float) {
	// 零地址检查 (铸造/销毁)
	zeroAddress := common.HexToAddress("0x0000000000000000000000000000000000000000")

	if transfer.From == zeroAddress {
		fmt.Printf("     🎯 代币铸造 (Mint)\n")
	} else if transfer.To == zeroAddress {
		fmt.Printf("     🔥 代币销毁 (Burn)\n")
	}

	// 大额转账检查
	threshold := big.NewFloat(1000000) // 100万代币
	if amount.Cmp(threshold) > 0 {
		fmt.Printf("     🐋 大额转账: 超过100万代币\n")
	}

	// 小额转账检查
	smallThreshold := big.NewFloat(0.001)
	if amount.Cmp(smallThreshold) < 0 {
		fmt.Printf("     🔍 微小转账: 少于0.001代币\n")
	}
}
