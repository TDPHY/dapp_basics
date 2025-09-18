package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/local/go-eth-demo/config"
	"github.com/local/go-eth-demo/utils"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 创建以太坊客户端
	ethClient, err := utils.NewEthClient(cfg)
	if err != nil {
		log.Fatalf("创建以太坊客户端失败: %v", err)
	}
	defer ethClient.Close()

	ctx := context.Background()

	// 要查询的交易哈希 (使用之前查询到的交易)
	txHashStr := "0x123456789abcdef..." // 请替换为实际的交易哈希

	// 如果没有指定交易哈希，使用最新区块中的第一笔交易
	if txHashStr == "0x123456789abcdef..." {
		fmt.Println("🔍 获取最新区块中的交易进行演示...")

		// 获取最新区块
		latestBlock, err := ethClient.GetClient().BlockByNumber(ctx, nil)
		if err != nil {
			log.Fatalf("获取最新区块失败: %v", err)
		}

		transactions := latestBlock.Transactions()
		if len(transactions) == 0 {
			log.Fatalf("最新区块中没有交易")
		}

		// 使用第一笔交易
		txHashStr = transactions[0].Hash().Hex()
		fmt.Printf("使用交易: %s\n\n", txHashStr)
	}

	txHash := common.HexToHash(txHashStr)

	// 查询交易收据
	fmt.Println("📋 查询交易收据...")
	fmt.Println("================================")

	receipt, err := ethClient.GetClient().TransactionReceipt(ctx, txHash)
	if err != nil {
		log.Fatalf("查询交易收据失败: %v", err)
	}

	// 同时获取交易详情用于对比
	tx, isPending, err := ethClient.GetClient().TransactionByHash(ctx, txHash)
	if err != nil {
		log.Fatalf("查询交易详情失败: %v", err)
	}

	if isPending {
		fmt.Println("⚠️  交易仍在等待确认中...")
		return
	}

	// 显示基本收据信息
	displayBasicReceiptInfo(receipt, tx)

	// 显示执行状态
	displayExecutionStatus(receipt)

	// 显示 Gas 使用情况
	displayGasUsage(receipt, tx)

	// 显示事件日志
	displayEventLogs(receipt)

	fmt.Println("\n✅ 交易收据分析完成！")
}

// displayBasicReceiptInfo 显示基本收据信息
func displayBasicReceiptInfo(receipt *types.Receipt, tx *types.Transaction) {
	fmt.Println("\n📄 基本收据信息:")
	fmt.Println("--------------------------------")

	fmt.Printf("交易哈希: %s\n", receipt.TxHash.Hex())
	fmt.Printf("区块号: %d\n", receipt.BlockNumber.Uint64())
	fmt.Printf("区块哈希: %s\n", receipt.BlockHash.Hex())
	fmt.Printf("交易索引: %d\n", receipt.TransactionIndex)

	// 合约地址 (如果是合约部署)
	if receipt.ContractAddress != (common.Address{}) {
		fmt.Printf("🏗️  新部署合约地址: %s\n", receipt.ContractAddress.Hex())
	}
}

// displayExecutionStatus 显示执行状态
func displayExecutionStatus(receipt *types.Receipt) {
	fmt.Println("\n🎯 执行状态:")
	fmt.Println("--------------------------------")

	if receipt.Status == types.ReceiptStatusSuccessful {
		fmt.Printf("✅ 交易执行成功\n")
		fmt.Printf("状态码: 1\n")

		if len(receipt.Logs) > 0 {
			fmt.Printf("📝 产生了 %d 个事件日志\n", len(receipt.Logs))
		} else {
			fmt.Printf("📝 没有产生事件日志\n")
		}
	} else {
		fmt.Printf("❌ 交易执行失败\n")
		fmt.Printf("状态码: 0\n")
		fmt.Printf("⚠️  注意: 失败的交易仍然会消耗 Gas\n")
	}
}

// displayGasUsage 显示 Gas 使用情况
func displayGasUsage(receipt *types.Receipt, tx *types.Transaction) {
	fmt.Println("\n⛽ Gas 使用分析:")
	fmt.Println("--------------------------------")

	gasUsed := receipt.GasUsed
	gasLimit := tx.Gas()
	gasPrice := tx.GasPrice()

	fmt.Printf("Gas 限制: %s\n", utils.FormatNumber(gasLimit))
	fmt.Printf("Gas 使用: %s\n", utils.FormatNumber(gasUsed))
	fmt.Printf("Gas 价格: %s Gwei\n", utils.WeiToGwei(gasPrice))

	// 计算使用率
	usagePercent := float64(gasUsed) / float64(gasLimit) * 100
	fmt.Printf("使用率: %.2f%%\n", usagePercent)

	// 计算费用
	actualFee := new(big.Int).Mul(big.NewInt(int64(gasUsed)), gasPrice)
	maxFee := new(big.Int).Mul(big.NewInt(int64(gasLimit)), gasPrice)
	savedFee := new(big.Int).Sub(maxFee, actualFee)

	fmt.Printf("实际费用: %s ETH\n", utils.WeiToEther(actualFee))
	fmt.Printf("最大费用: %s ETH\n", utils.WeiToEther(maxFee))
	fmt.Printf("节省费用: %s ETH\n", utils.WeiToEther(savedFee))

	// Gas 使用分析
	baseGas := uint64(21000)
	if gasUsed <= baseGas {
		fmt.Printf("交易类型: 💸 简单 ETH 转账\n")
	} else {
		extraGas := gasUsed - baseGas
		fmt.Printf("交易类型: 📞 合约交互\n")
		fmt.Printf("基础 Gas: %s (转账)\n", utils.FormatNumber(baseGas))
		fmt.Printf("额外 Gas: %s (合约执行)\n", utils.FormatNumber(extraGas))
	}
}

// displayEventLogs 显示事件日志
func displayEventLogs(receipt *types.Receipt) {
	fmt.Println("\n📝 事件日志分析:")
	fmt.Println("--------------------------------")

	logs := receipt.Logs
	if len(logs) == 0 {
		fmt.Printf("没有事件日志\n")
		fmt.Printf("说明: 这是一个简单转账或没有触发事件的合约调用\n")
		return
	}

	fmt.Printf("事件数量: %d\n", len(logs))

	// 统计合约地址
	contractMap := make(map[common.Address]int)
	for _, log := range logs {
		contractMap[log.Address]++
	}

	fmt.Printf("涉及合约: %d 个\n", len(contractMap))
	for addr, count := range contractMap {
		fmt.Printf("  %s: %d 个事件\n", addr.Hex(), count)
	}

	// 显示前几个事件的详细信息
	maxDisplay := 3
	if len(logs) < maxDisplay {
		maxDisplay = len(logs)
	}

	fmt.Printf("\n前 %d 个事件详情:\n", maxDisplay)
	for i := 0; i < maxDisplay; i++ {
		log := logs[i]
		fmt.Printf("\n🏷️  事件 #%d:\n", i+1)
		fmt.Printf("  合约地址: %s\n", log.Address.Hex())
		fmt.Printf("  主题数量: %d\n", len(log.Topics))

		if len(log.Topics) > 0 {
			fmt.Printf("  事件签名: %s\n", log.Topics[0].Hex())

			// 尝试识别常见事件
			eventName := getCommonEventName(log.Topics[0].Hex())
			if eventName != "" {
				fmt.Printf("  识别事件: %s\n", eventName)
			}
		}

		fmt.Printf("  数据长度: %d bytes\n", len(log.Data))

		// 如果数据是32字节，尝试解析为数值
		if len(log.Data) == 32 {
			value := new(big.Int).SetBytes(log.Data)
			fmt.Printf("  可能数值: %s\n", value.String())
		}
	}

	if len(logs) > maxDisplay {
		fmt.Printf("\n... 还有 %d 个事件 (已省略)\n", len(logs)-maxDisplay)
	}
}

// getCommonEventName 获取常见事件名称
func getCommonEventName(signature string) string {
	commonEvents := map[string]string{
		"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef": "Transfer(address,address,uint256)",
		"0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925": "Approval(address,address,uint256)",
		"0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31": "ApprovalForAll(address,address,bool)",
		"0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0": "OwnershipTransferred(address,address)",
	}

	return commonEvents[signature]
}
