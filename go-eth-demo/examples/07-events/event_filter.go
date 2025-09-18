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

	fmt.Println("🔍 事件过滤和历史查询")
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

	// 示例1: 查询特定代币的所有转账事件
	fmt.Println("📊 示例1: 查询特定代币的转账事件")
	queryTokenTransfers(client, contractABI, currentBlock)

	// 示例2: 查询特定地址的转账事件
	fmt.Println("\n📊 示例2: 查询特定地址的转账事件")
	queryAddressTransfers(client, contractABI, currentBlock)

	// 示例3: 查询大额转账事件
	fmt.Println("\n📊 示例3: 查询大额转账事件")
	queryLargeTransfers(client, contractABI, currentBlock)

	// 示例4: 时间范围查询
	fmt.Println("\n📊 示例4: 时间范围查询")
	queryTimeRangeEvents(client, contractABI, currentBlock)

	// 示例5: 多条件组合查询
	fmt.Println("\n📊 示例5: 多条件组合查询")
	queryComplexFilter(client, contractABI, currentBlock)
}

// 示例1: 查询特定代币的所有转账事件
func queryTokenTransfers(client *ethclient.Client, contractABI abi.ABI, currentBlock uint64) {
	// 使用一个示例代币地址 (需要替换为实际的代币地址)
	tokenAddress := common.HexToAddress("0xA0b86a33E6441b8435b662f0E2d0B8A0E4B2B8B0")

	// 查询最近100个区块
	fromBlock := currentBlock - 100
	if fromBlock > currentBlock {
		fromBlock = 0
	}

	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(currentBlock)),
		Addresses: []common.Address{tokenAddress},
		Topics: [][]common.Hash{
			{contractABI.Events["Transfer"].ID}, // Transfer 事件
		},
	}

	fmt.Printf("查询代币 %s 的转账事件\n", tokenAddress.Hex())
	fmt.Printf("区块范围: #%d - #%d\n", fromBlock, currentBlock)

	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		log.Printf("查询事件失败: %v", err)
		return
	}

	fmt.Printf("找到 %d 个转账事件:\n", len(logs))

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

		fmt.Printf("  %d. 从 %s 到 %s, 金额: %s (区块: #%d)\n",
			i+1, transfer.From.Hex()[:10]+"...", transfer.To.Hex()[:10]+"...",
			amount.Text('f', 6), vLog.BlockNumber)
	}
}

// 示例2: 查询特定地址的转账事件
func queryAddressTransfers(client *ethclient.Client, contractABI abi.ABI, currentBlock uint64) {
	// 查询 Vitalik 的地址作为示例
	targetAddress := common.HexToAddress("0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045")

	fromBlock := currentBlock - 200
	if fromBlock > currentBlock {
		fromBlock = 0
	}

	// 查询该地址作为发送方或接收方的转账
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(currentBlock)),
		Topics: [][]common.Hash{
			{contractABI.Events["Transfer"].ID}, // Transfer 事件
			{},                                  // 第二个 topic 可以是任何值
			{},                                  // 第三个 topic 可以是任何值
		},
	}

	fmt.Printf("查询地址 %s 相关的转账事件\n", targetAddress.Hex())
	fmt.Printf("区块范围: #%d - #%d\n", fromBlock, currentBlock)

	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		log.Printf("查询事件失败: %v", err)
		return
	}

	// 过滤出与目标地址相关的事件
	var relevantLogs []types.Log
	for _, vLog := range logs {
		if len(vLog.Topics) >= 3 {
			from := common.HexToAddress(vLog.Topics[1].Hex())
			to := common.HexToAddress(vLog.Topics[2].Hex())

			if from == targetAddress || to == targetAddress {
				relevantLogs = append(relevantLogs, vLog)
			}
		}
	}

	fmt.Printf("找到 %d 个相关转账事件:\n", len(relevantLogs))

	for i, vLog := range relevantLogs {
		if i >= 3 { // 只显示前3个
			fmt.Printf("... 还有 %d 个事件\n", len(relevantLogs)-3)
			break
		}

		var transfer Transfer
		err := contractABI.UnpackIntoInterface(&transfer, "Transfer", vLog.Data)
		if err != nil {
			continue
		}

		transfer.From = common.HexToAddress(vLog.Topics[1].Hex())
		transfer.To = common.HexToAddress(vLog.Topics[2].Hex())

		amount := new(big.Float).SetInt(transfer.Amount)
		amount = amount.Quo(amount, big.NewFloat(1e18))

		direction := "接收"
		if transfer.From == targetAddress {
			direction = "发送"
		}

		fmt.Printf("  %d. %s %s 代币, 对方: %s (区块: #%d)\n",
			i+1, direction, amount.Text('f', 6),
			getOtherAddress(transfer.From, transfer.To, targetAddress).Hex()[:10]+"...",
			vLog.BlockNumber)
	}
}

// 示例3: 查询大额转账事件
func queryLargeTransfers(client *ethclient.Client, contractABI abi.ABI, currentBlock uint64) {
	fromBlock := currentBlock - 500
	if fromBlock > currentBlock {
		fromBlock = 0
	}

	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(currentBlock)),
		Topics: [][]common.Hash{
			{contractABI.Events["Transfer"].ID},
		},
	}

	fmt.Printf("查询大额转账事件 (>1000 代币)\n")
	fmt.Printf("区块范围: #%d - #%d\n", fromBlock, currentBlock)

	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		log.Printf("查询事件失败: %v", err)
		return
	}

	var largeTransfers []types.Log
	threshold := new(big.Int)
	threshold.SetString("1000000000000000000000", 10) // 1000 * 10^18

	for _, vLog := range logs {
		var transfer Transfer
		err := contractABI.UnpackIntoInterface(&transfer, "Transfer", vLog.Data)
		if err != nil {
			continue
		}

		if transfer.Amount.Cmp(threshold) > 0 {
			largeTransfers = append(largeTransfers, vLog)
		}
	}

	fmt.Printf("找到 %d 个大额转账事件:\n", len(largeTransfers))

	for i, vLog := range largeTransfers {
		if i >= 5 {
			fmt.Printf("... 还有 %d 个大额转账\n", len(largeTransfers)-5)
			break
		}

		var transfer Transfer
		contractABI.UnpackIntoInterface(&transfer, "Transfer", vLog.Data)
		transfer.From = common.HexToAddress(vLog.Topics[1].Hex())
		transfer.To = common.HexToAddress(vLog.Topics[2].Hex())

		amount := new(big.Float).SetInt(transfer.Amount)
		amount = amount.Quo(amount, big.NewFloat(1e18))

		fmt.Printf("  %d. 🐋 %s 代币, 从 %s 到 %s (区块: #%d)\n",
			i+1, amount.Text('f', 2),
			transfer.From.Hex()[:10]+"...", transfer.To.Hex()[:10]+"...",
			vLog.BlockNumber)
	}
}

// 示例4: 时间范围查询
func queryTimeRangeEvents(client *ethclient.Client, contractABI abi.ABI, currentBlock uint64) {
	// 查询最近1小时的事件 (假设12秒一个区块)
	blocksPerHour := uint64(300) // 3600/12
	fromBlock := currentBlock - blocksPerHour
	if fromBlock > currentBlock {
		fromBlock = 0
	}

	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(currentBlock)),
		Topics: [][]common.Hash{
			{contractABI.Events["Transfer"].ID},
		},
	}

	fmt.Printf("查询最近1小时的转账事件\n")
	fmt.Printf("区块范围: #%d - #%d (约 %d 个区块)\n", fromBlock, currentBlock, blocksPerHour)

	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		log.Printf("查询事件失败: %v", err)
		return
	}

	fmt.Printf("找到 %d 个转账事件\n", len(logs))

	// 按代币地址分组统计
	tokenStats := make(map[common.Address]int)
	for _, vLog := range logs {
		tokenStats[vLog.Address]++
	}

	fmt.Println("按代币统计:")
	count := 0
	for tokenAddr, eventCount := range tokenStats {
		if count >= 5 {
			fmt.Printf("... 还有 %d 个代币\n", len(tokenStats)-5)
			break
		}
		fmt.Printf("  %s: %d 个事件\n", tokenAddr.Hex()[:10]+"...", eventCount)
		count++
	}
}

// 示例5: 多条件组合查询
func queryComplexFilter(client *ethclient.Client, contractABI abi.ABI, currentBlock uint64) {
	// 查询特定代币地址列表
	tokenAddresses := []common.Address{
		common.HexToAddress("0xA0b86a33E6441b8435b662f0E2d0B8A0E4B2B8B0"),
		common.HexToAddress("0x779877A7B0D9E8603169DdbD7836e478b4624789"),
	}

	// 查询特定的发送方地址
	specificSender := common.HexToAddress("0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045")

	fromBlock := currentBlock - 1000
	if fromBlock > currentBlock {
		fromBlock = 0
	}

	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(currentBlock)),
		Addresses: tokenAddresses,
		Topics: [][]common.Hash{
			{contractABI.Events["Transfer"].ID},          // Transfer 事件
			{common.BytesToHash(specificSender.Bytes())}, // 特定发送方
		},
	}

	fmt.Printf("复杂查询: 特定代币 + 特定发送方\n")
	fmt.Printf("代币地址: %d 个\n", len(tokenAddresses))
	fmt.Printf("发送方: %s\n", specificSender.Hex())
	fmt.Printf("区块范围: #%d - #%d\n", fromBlock, currentBlock)

	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		log.Printf("查询事件失败: %v", err)
		return
	}

	fmt.Printf("找到 %d 个匹配的转账事件:\n", len(logs))

	for i, vLog := range logs {
		if i >= 3 {
			fmt.Printf("... 还有 %d 个事件\n", len(logs)-3)
			break
		}

		var transfer Transfer
		err := contractABI.UnpackIntoInterface(&transfer, "Transfer", vLog.Data)
		if err != nil {
			continue
		}

		transfer.From = common.HexToAddress(vLog.Topics[1].Hex())
		transfer.To = common.HexToAddress(vLog.Topics[2].Hex())

		amount := new(big.Float).SetInt(transfer.Amount)
		amount = amount.Quo(amount, big.NewFloat(1e18))

		fmt.Printf("  %d. 代币: %s, 到: %s, 金额: %s (区块: #%d)\n",
			i+1, vLog.Address.Hex()[:10]+"...", transfer.To.Hex()[:10]+"...",
			amount.Text('f', 6), vLog.BlockNumber)
	}
}

// 辅助函数: 获取转账中的另一个地址
func getOtherAddress(from, to, target common.Address) common.Address {
	if from == target {
		return to
	}
	return from
}
