package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/local/go-eth-demo/config"
	"github.com/local/go-eth-demo/utils"
)

func main() {
	fmt.Println("📝 以太坊交易查询详解")
	fmt.Println("================================")

	// 初始化客户端
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ 配置加载失败: %v", err)
	}

	client, err := utils.NewEthClient(cfg)
	if err != nil {
		log.Fatalf("❌ 连接失败: %v", err)
	}
	defer client.Close()

	ethClient := client.GetClient()
	ctx := context.Background()

	// 1. 从最新区块获取交易哈希
	fmt.Println("🔍 获取最新区块中的交易...")
	latestBlock, err := ethClient.BlockByNumber(ctx, nil)
	if err != nil {
		log.Fatalf("❌ 获取最新区块失败: %v", err)
	}

	transactions := latestBlock.Transactions()
	if len(transactions) == 0 {
		fmt.Println("❌ 最新区块没有交易，尝试获取历史区块...")
		// 尝试获取有交易的历史区块
		for i := 1; i <= 10; i++ {
			blockNumber := new(big.Int).Sub(latestBlock.Number(), big.NewInt(int64(i)))
			block, err := ethClient.BlockByNumber(ctx, blockNumber)
			if err != nil {
				continue
			}
			if len(block.Transactions()) > 0 {
				transactions = block.Transactions()
				latestBlock = block
				fmt.Printf("✅ 找到有交易的区块 #%s，包含 %d 笔交易\n", blockNumber.String(), len(transactions))
				break
			}
		}
	}

	if len(transactions) == 0 {
		log.Fatalf("❌ 未找到包含交易的区块")
	}

	// 选择第一笔交易进行详细分析
	selectedTx := transactions[0]
	txHash := selectedTx.Hash()

	fmt.Printf("\n🎯 选择交易进行详细分析: %s\n", txHash.Hex())

	// 2. 根据交易哈希查询交易详情
	fmt.Println("\n📋 查询交易详情...")
	tx, isPending, err := queryTransactionByHash(ctx, ethClient, txHash)
	if err != nil {
		log.Fatalf("❌ 查询交易失败: %v", err)
	}
	displayTransactionInfo("交易详情", tx, isPending)

	// 3. 查询交易收据
	fmt.Println("\n🧾 查询交易收据...")
	receipt, err := queryTransactionReceipt(ctx, ethClient, txHash)
	if err != nil {
		log.Printf("❌ 查询交易收据失败: %v", err)
	} else {
		displayReceiptInfo("交易收据", receipt)
	}

	// 4. 分析交易类型和数据
	fmt.Println("\n🔬 交易类型分析...")
	analyzeTransactionType(tx)

	// 5. 计算交易费用
	fmt.Println("\n💰 交易费用分析...")
	if receipt != nil {
		calculateTransactionFee(tx, receipt)
	}

	// 6. 查询交易在区块中的位置
	fmt.Println("\n📍 交易位置信息...")
	analyzeTransactionPosition(ctx, ethClient, txHash, latestBlock)

	// 7. 分析更多交易示例
	fmt.Println("\n📊 批量交易分析...")
	analyzeBatchTransactions(ctx, ethClient, transactions[:min(5, len(transactions))])

	fmt.Println("\n✅ 交易查询学习完成!")
}

// queryTransactionByHash 根据交易哈希查询交易
func queryTransactionByHash(ctx context.Context, client *ethclient.Client, txHash common.Hash) (*types.Transaction, bool, error) {
	tx, isPending, err := client.TransactionByHash(ctx, txHash)
	if err != nil {
		return nil, false, fmt.Errorf("查询交易失败: %w", err)
	}
	return tx, isPending, nil
}

// queryTransactionReceipt 查询交易收据
func queryTransactionReceipt(ctx context.Context, client *ethclient.Client, txHash common.Hash) (*types.Receipt, error) {
	receipt, err := client.TransactionReceipt(ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("查询交易收据失败: %w", err)
	}
	return receipt, nil
}

// displayTransactionInfo 显示交易详细信息
func displayTransactionInfo(title string, tx *types.Transaction, isPending bool) {
	fmt.Printf("\n📋 %s:\n", title)
	fmt.Println("--------------------------------")

	// 基本信息
	fmt.Printf("交易哈希: %s\n", tx.Hash().Hex())
	fmt.Printf("状态: %s\n", getTransactionStatus(isPending))

	// 发送方和接收方
	from, err := getTransactionSender(tx)
	if err != nil {
		fmt.Printf("发送方: 无法获取 (%v)\n", err)
	} else {
		fmt.Printf("发送方: %s\n", from.Hex())
	}

	if tx.To() != nil {
		fmt.Printf("接收方: %s\n", tx.To().Hex())
		fmt.Printf("交易类型: 普通转账/合约调用\n")
	} else {
		fmt.Printf("接收方: 合约创建\n")
		fmt.Printf("交易类型: 合约部署\n")
	}

	// 金额和费用信息
	fmt.Printf("转账金额: %s ETH\n", weiToEther(tx.Value()))
	fmt.Printf("Gas 限制: %s\n", formatNumber(tx.Gas()))
	fmt.Printf("Gas 价格: %s Gwei (%s Wei)\n", weiToGwei(tx.GasPrice()), tx.GasPrice().String())

	// 交易数据
	fmt.Printf("Nonce: %d\n", tx.Nonce())
	fmt.Printf("数据大小: %d bytes\n", len(tx.Data()))

	if len(tx.Data()) > 0 {
		fmt.Printf("输入数据 (前64字符): %s...\n", common.Bytes2Hex(tx.Data())[:min(64, len(common.Bytes2Hex(tx.Data())))])
	} else {
		fmt.Printf("输入数据: 空 (简单转账)\n")
	}

	// EIP-155 链 ID
	if chainId := tx.ChainId(); chainId != nil {
		fmt.Printf("链 ID: %s\n", chainId.String())
	}

	// 交易签名信息
	v, r, s := tx.RawSignatureValues()
	fmt.Printf("签名 V: %s\n", v.String())
	fmt.Printf("签名 R: %s\n", r.String())
	fmt.Printf("签名 S: %s\n", s.String())
}

// displayReceiptInfo 显示交易收据信息
func displayReceiptInfo(title string, receipt *types.Receipt) {
	fmt.Printf("\n🧾 %s:\n", title)
	fmt.Println("--------------------------------")

	// 基本信息
	fmt.Printf("交易哈希: %s\n", receipt.TxHash.Hex())
	fmt.Printf("区块哈希: %s\n", receipt.BlockHash.Hex())
	fmt.Printf("区块号: %d\n", receipt.BlockNumber.Uint64())
	fmt.Printf("交易索引: %d\n", receipt.TransactionIndex)

	// 执行结果
	if receipt.Status == types.ReceiptStatusSuccessful {
		fmt.Printf("执行状态: ✅ 成功\n")
	} else {
		fmt.Printf("执行状态: ❌ 失败\n")
	}

	// Gas 使用情况
	fmt.Printf("Gas 使用量: %s\n", formatNumber(receipt.GasUsed))
	fmt.Printf("累计 Gas 使用: %s\n", formatNumber(receipt.CumulativeGasUsed))

	// 合约地址 (如果是合约创建)
	if receipt.ContractAddress != (common.Address{}) {
		fmt.Printf("创建的合约地址: %s\n", receipt.ContractAddress.Hex())
	}

	// Bloom 过滤器
	fmt.Printf("Bloom 过滤器: %s\n", receipt.Bloom.Big().String())

	// 事件日志
	fmt.Printf("事件日志数量: %d\n", len(receipt.Logs))
	if len(receipt.Logs) > 0 {
		fmt.Println("\n📝 事件日志详情:")
		for i, log := range receipt.Logs[:min(3, len(receipt.Logs))] {
			fmt.Printf("  日志 #%d:\n", i+1)
			fmt.Printf("    合约地址: %s\n", log.Address.Hex())
			fmt.Printf("    主题数量: %d\n", len(log.Topics))
			if len(log.Topics) > 0 {
				fmt.Printf("    主题0 (事件签名): %s\n", log.Topics[0].Hex())
			}
			fmt.Printf("    数据长度: %d bytes\n", len(log.Data))
		}
		if len(receipt.Logs) > 3 {
			fmt.Printf("  ... 还有 %d 个日志\n", len(receipt.Logs)-3)
		}
	}
}

// analyzeTransactionType 分析交易类型
func analyzeTransactionType(tx *types.Transaction) {
	fmt.Printf("🔬 交易类型深度分析:\n")
	fmt.Println("--------------------------------")

	// 基本分类
	if tx.To() == nil {
		fmt.Printf("交易类型: 🏗️  合约部署\n")
		fmt.Printf("部署数据大小: %d bytes\n", len(tx.Data()))
	} else if len(tx.Data()) == 0 {
		fmt.Printf("交易类型: 💸 简单 ETH 转账\n")
	} else {
		fmt.Printf("交易类型: 📞 合约调用\n")

		// 尝试解析方法签名
		if len(tx.Data()) >= 4 {
			methodSig := common.Bytes2Hex(tx.Data()[:4])
			fmt.Printf("方法签名: 0x%s\n", methodSig)

			// 识别常见的方法签名
			knownMethods := map[string]string{
				"a9059cbb": "transfer(address,uint256)",
				"095ea7b3": "approve(address,uint256)",
				"23b872dd": "transferFrom(address,address,uint256)",
				"18160ddd": "totalSupply()",
				"70a08231": "balanceOf(address)",
				"dd62ed3e": "allowance(address,address)",
			}

			if methodName, exists := knownMethods[methodSig]; exists {
				fmt.Printf("识别的方法: %s\n", methodName)
			} else {
				fmt.Printf("未知方法签名\n")
			}
		}
	}

	// 金额分析
	if tx.Value().Cmp(big.NewInt(0)) > 0 {
		fmt.Printf("包含 ETH 转账: %s ETH\n", weiToEther(tx.Value()))
	} else {
		fmt.Printf("不包含 ETH 转账 (可能是代币转账或其他操作)\n")
	}

	// Gas 分析
	gasPrice := tx.GasPrice()
	gasLimit := tx.Gas()

	fmt.Printf("Gas 设置分析:\n")
	fmt.Printf("  Gas 限制: %s\n", formatNumber(gasLimit))
	fmt.Printf("  Gas 价格: %s Gwei\n", weiToGwei(gasPrice))

	maxFee := new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimit)))
	fmt.Printf("  最大可能费用: %s ETH\n", weiToEther(maxFee))

	// Gas 价格评估
	gasPriceGwei := new(big.Float).Quo(new(big.Float).SetInt(gasPrice), big.NewFloat(1e9))
	gasPriceFloat, _ := gasPriceGwei.Float64()

	var priceLevel string
	switch {
	case gasPriceFloat < 1:
		priceLevel = "🟢 极低 (测试网或网络空闲)"
	case gasPriceFloat < 10:
		priceLevel = "🟡 较低"
	case gasPriceFloat < 50:
		priceLevel = "🟠 正常"
	case gasPriceFloat < 100:
		priceLevel = "🔴 较高"
	default:
		priceLevel = "🚨 极高 (网络拥堵)"
	}
	fmt.Printf("  Gas 价格水平: %s\n", priceLevel)
}

// calculateTransactionFee 计算交易费用
func calculateTransactionFee(tx *types.Transaction, receipt *types.Receipt) {
	fmt.Printf("💰 交易费用详细计算:\n")
	fmt.Println("--------------------------------")

	gasUsed := receipt.GasUsed
	gasPrice := tx.GasPrice()

	// 实际费用
	actualFee := new(big.Int).Mul(big.NewInt(int64(gasUsed)), gasPrice)
	fmt.Printf("实际交易费用: %s ETH\n", weiToEther(actualFee))
	fmt.Printf("费用计算: %s (Gas使用) × %s (Gas价格) = %s Wei\n",
		formatNumber(gasUsed),
		weiToGwei(gasPrice)+" Gwei",
		actualFee.String())

	// 最大可能费用
	gasLimit := tx.Gas()
	maxPossibleFee := new(big.Int).Mul(big.NewInt(int64(gasLimit)), gasPrice)
	fmt.Printf("最大可能费用: %s ETH\n", weiToEther(maxPossibleFee))

	// 节省的费用
	savedFee := new(big.Int).Sub(maxPossibleFee, actualFee)
	fmt.Printf("节省的费用: %s ETH\n", weiToEther(savedFee))

	// Gas 效率
	gasEfficiency := float64(gasUsed) / float64(gasLimit) * 100
	fmt.Printf("Gas 使用效率: %.2f%%\n", gasEfficiency)

	if gasEfficiency > 95 {
		fmt.Printf("效率评估: 🔴 Gas 限制设置过低，可能导致交易失败\n")
	} else if gasEfficiency > 80 {
		fmt.Printf("效率评估: 🟡 Gas 限制设置合理\n")
	} else if gasEfficiency > 50 {
		fmt.Printf("效率评估: 🟢 Gas 限制设置适中\n")
	} else {
		fmt.Printf("效率评估: 🔵 Gas 限制设置过高，浪费了一些费用\n")
	}
}

// analyzeTransactionPosition 分析交易在区块中的位置
func analyzeTransactionPosition(ctx context.Context, client *ethclient.Client, txHash common.Hash, block *types.Block) {
	fmt.Printf("📍 交易位置分析:\n")
	fmt.Println("--------------------------------")

	transactions := block.Transactions()
	var position int = -1

	// 找到交易在区块中的位置
	for i, tx := range transactions {
		if tx.Hash() == txHash {
			position = i
			break
		}
	}

	if position >= 0 {
		fmt.Printf("区块号: %s\n", block.Number().String())
		fmt.Printf("区块中的位置: %d / %d\n", position+1, len(transactions))
		fmt.Printf("交易索引: %d\n", position)

		positionPercent := float64(position+1) / float64(len(transactions)) * 100
		fmt.Printf("位置百分比: %.2f%%\n", positionPercent)

		if position == 0 {
			fmt.Printf("位置特点: 🥇 区块中的第一笔交易\n")
		} else if position == len(transactions)-1 {
			fmt.Printf("位置特点: 🏁 区块中的最后一笔交易\n")
		} else if positionPercent < 25 {
			fmt.Printf("位置特点: 🟢 靠前位置 (优先级较高)\n")
		} else if positionPercent > 75 {
			fmt.Printf("位置特点: 🔴 靠后位置 (优先级较低)\n")
		} else {
			fmt.Printf("位置特点: 🟡 中间位置\n")
		}
	} else {
		fmt.Printf("❌ 未在指定区块中找到该交易\n")
	}
}

// analyzeBatchTransactions 批量分析交易
func analyzeBatchTransactions(ctx context.Context, client *ethclient.Client, transactions []*types.Transaction) {
	fmt.Printf("📊 批量交易分析 (样本: %d 笔):\n", len(transactions))
	fmt.Println("--------------------------------")

	var totalValue, totalGasPrice, totalGasLimit big.Int
	contractCalls := 0
	contractCreations := 0
	simpleTransfers := 0

	for _, tx := range transactions {
		// 累计统计
		totalValue.Add(&totalValue, tx.Value())
		totalGasPrice.Add(&totalGasPrice, tx.GasPrice())
		totalGasLimit.Add(&totalGasLimit, big.NewInt(int64(tx.Gas())))

		// 分类统计
		if tx.To() == nil {
			contractCreations++
		} else if len(tx.Data()) == 0 {
			simpleTransfers++
		} else {
			contractCalls++
		}
	}

	// 显示统计结果
	fmt.Printf("交易类型分布:\n")
	fmt.Printf("  简单转账: %d 笔 (%.1f%%)\n", simpleTransfers, float64(simpleTransfers)/float64(len(transactions))*100)
	fmt.Printf("  合约调用: %d 笔 (%.1f%%)\n", contractCalls, float64(contractCalls)/float64(len(transactions))*100)
	fmt.Printf("  合约创建: %d 笔 (%.1f%%)\n", contractCreations, float64(contractCreations)/float64(len(transactions))*100)

	fmt.Printf("\n金额统计:\n")
	fmt.Printf("  总转账金额: %s ETH\n", weiToEther(&totalValue))
	avgValue := new(big.Int).Div(&totalValue, big.NewInt(int64(len(transactions))))
	fmt.Printf("  平均转账金额: %s ETH\n", weiToEther(avgValue))

	fmt.Printf("\nGas 统计:\n")
	avgGasPrice := new(big.Int).Div(&totalGasPrice, big.NewInt(int64(len(transactions))))
	fmt.Printf("  平均 Gas 价格: %s Gwei\n", weiToGwei(avgGasPrice))
	avgGasLimit := new(big.Int).Div(&totalGasLimit, big.NewInt(int64(len(transactions))))
	fmt.Printf("  平均 Gas 限制: %s\n", formatNumber(avgGasLimit.Uint64()))
}

// 工具函数

// getTransactionStatus 获取交易状态描述
func getTransactionStatus(isPending bool) string {
	if isPending {
		return "⏳ 待处理"
	}
	return "✅ 已确认"
}

// getTransactionSender 获取交易发送方地址
func getTransactionSender(tx *types.Transaction) (common.Address, error) {
	// 这里需要链 ID 来正确恢复发送方地址
	// 在实际应用中，你可能需要使用正确的 Signer
	chainID := tx.ChainId()
	if chainID == nil {
		return common.Address{}, fmt.Errorf("无法获取链 ID")
	}

	signer := types.NewEIP155Signer(chainID)
	return types.Sender(signer, tx)
}

// weiToEther 将 Wei 转换为 Ether
func weiToEther(wei *big.Int) string {
	ether := new(big.Float).SetInt(wei)
	ether.Quo(ether, big.NewFloat(1e18))
	return ether.Text('f', 6)
}

// weiToGwei 将 Wei 转换为 Gwei
func weiToGwei(wei *big.Int) string {
	gwei := new(big.Float).SetInt(wei)
	gwei.Quo(gwei, big.NewFloat(1e9))
	return gwei.Text('f', 2)
}

// formatNumber 格式化大数字
func formatNumber(n uint64) string {
	str := fmt.Sprintf("%d", n)
	if len(str) <= 3 {
		return str
	}

	result := ""
	for i, char := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result += ","
		}
		result += string(char)
	}
	return result
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
