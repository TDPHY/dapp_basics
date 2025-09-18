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
	fmt.Println("🧾 以太坊交易收据深度分析")
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

	// 1. 获取包含事件的交易
	fmt.Println("🔍 寻找包含事件日志的交易...")
	txHash, blockNumber, err := findTransactionWithLogs(ctx, ethClient)
	if err != nil {
		log.Fatalf("❌ 寻找交易失败: %v", err)
	}

	fmt.Printf("✅ 找到交易: %s (区块 #%s)\n", txHash.Hex(), blockNumber.String())

	// 2. 查询交易收据
	fmt.Println("\n📋 查询交易收据...")
	receipt, err := ethClient.TransactionReceipt(ctx, txHash)
	if err != nil {
		log.Fatalf("❌ 查询收据失败: %v", err)
	}

	// 3. 详细分析收据
	analyzeReceiptDetails(receipt)

	// 4. 分析事件日志
	fmt.Println("\n📝 事件日志详细分析...")
	analyzeEventLogs(receipt.Logs)

	// 5. Gas 使用分析
	fmt.Println("\n⛽ Gas 使用详细分析...")
	analyzeGasUsage(ctx, ethClient, txHash, receipt)

	// 6. 收据状态分析
	fmt.Println("\n🔍 收据状态分析...")
	analyzeReceiptStatus(receipt)

	// 7. Bloom 过滤器分析
	fmt.Println("\n🌸 Bloom 过滤器分析...")
	analyzeBloomFilter(receipt)

	fmt.Println("\n✅ 交易收据分析完成!")
}

// findTransactionWithLogs 寻找包含事件日志的交易
func findTransactionWithLogs(ctx context.Context, client *ethclient.Client) (common.Hash, *big.Int, error) {
	// 从最新区块开始向前搜索
	latestBlock, err := client.BlockByNumber(ctx, nil)
	if err != nil {
		return common.Hash{}, nil, err
	}

	fmt.Printf("从区块 #%s 开始搜索...\n", latestBlock.Number().String())

	for i := 0; i < 20; i++ { // 搜索最近20个区块
		blockNumber := new(big.Int).Sub(latestBlock.Number(), big.NewInt(int64(i)))
		block, err := client.BlockByNumber(ctx, blockNumber)
		if err != nil {
			continue
		}

		fmt.Printf("检查区块 #%s (%d 笔交易)...\n", blockNumber.String(), len(block.Transactions()))

		// 检查每笔交易的收据
		for _, tx := range block.Transactions() {
			receipt, err := client.TransactionReceipt(ctx, tx.Hash())
			if err != nil {
				continue
			}

			// 找到包含事件日志的交易
			if len(receipt.Logs) > 0 {
				fmt.Printf("✅ 找到包含 %d 个事件日志的交易\n", len(receipt.Logs))
				return tx.Hash(), blockNumber, nil
			}
		}
	}

	return common.Hash{}, nil, fmt.Errorf("未找到包含事件日志的交易")
}

// analyzeReceiptDetails 分析收据详细信息
func analyzeReceiptDetails(receipt *types.Receipt) {
	fmt.Printf("\n🧾 交易收据详细信息:\n")
	fmt.Println("================================")

	// 基本信息
	fmt.Printf("交易哈希: %s\n", receipt.TxHash.Hex())
	fmt.Printf("区块哈希: %s\n", receipt.BlockHash.Hex())
	fmt.Printf("区块号: %d\n", receipt.BlockNumber.Uint64())
	fmt.Printf("交易索引: %d\n", receipt.TransactionIndex)

	// 执行状态
	fmt.Printf("\n📊 执行状态:\n")
	if receipt.Status == types.ReceiptStatusSuccessful {
		fmt.Printf("状态: ✅ 成功 (Status = 1)\n")
		fmt.Printf("说明: 交易成功执行，所有操作完成\n")
	} else {
		fmt.Printf("状态: ❌ 失败 (Status = 0)\n")
		fmt.Printf("说明: 交易执行失败，状态回滚但仍消耗 Gas\n")
	}

	// Gas 信息
	fmt.Printf("\n⛽ Gas 使用信息:\n")
	fmt.Printf("Gas 使用量: %s\n", formatNumber(receipt.GasUsed))
	fmt.Printf("累计 Gas 使用: %s\n", formatNumber(receipt.CumulativeGasUsed))
	fmt.Printf("说明: 累计 Gas 是该交易在区块中的累计使用量\n")

	// 合约地址
	if receipt.ContractAddress != (common.Address{}) {
		fmt.Printf("\n🏗️  合约创建:\n")
		fmt.Printf("新合约地址: %s\n", receipt.ContractAddress.Hex())
		fmt.Printf("说明: 这是一个合约部署交易\n")
	}

	// 事件日志概览
	fmt.Printf("\n📝 事件日志概览:\n")
	fmt.Printf("日志数量: %d\n", len(receipt.Logs))
	if len(receipt.Logs) > 0 {
		fmt.Printf("说明: 交易执行过程中触发了 %d 个事件\n", len(receipt.Logs))
	} else {
		fmt.Printf("说明: 交易没有触发任何事件 (可能是简单转账)\n")
	}

	// Bloom 过滤器
	fmt.Printf("\nBloom 过滤器: %s\n", receipt.Bloom.Big().String())
	if receipt.Bloom.Big().Cmp(big.NewInt(0)) == 0 {
		fmt.Printf("说明: Bloom 过滤器为空，没有事件日志\n")
	} else {
		fmt.Printf("说明: Bloom 过滤器包含事件信息，用于快速检索\n")
	}
}

// analyzeEventLogs 分析事件日志
func analyzeEventLogs(logs []*types.Log) {
	if len(logs) == 0 {
		fmt.Printf("该交易没有产生事件日志\n")
		return
	}

	fmt.Printf("📝 事件日志详细分析 (共 %d 个):\n", len(logs))
	fmt.Println("================================")

	for i, log := range logs {
		fmt.Printf("\n🏷️  日志 #%d:\n", i+1)
		fmt.Println("--------------------------------")

		// 基本信息
		fmt.Printf("合约地址: %s\n", log.Address.Hex())
		fmt.Printf("区块号: %d\n", log.BlockNumber)
		fmt.Printf("交易哈希: %s\n", log.TxHash.Hex())
		fmt.Printf("交易索引: %d\n", log.TxIndex)
		fmt.Printf("日志索引: %d\n", log.Index)
		fmt.Printf("是否已移除: %v\n", log.Removed)

		// 主题分析
		fmt.Printf("\n📋 主题 (Topics) 分析:\n")
		fmt.Printf("主题数量: %d\n", len(log.Topics))

		for j, topic := range log.Topics {
			fmt.Printf("  主题 %d: %s\n", j, topic.Hex())
			if j == 0 {
				fmt.Printf("    说明: 事件签名 (Event Signature)\n")
				// 尝试识别常见事件
				eventName := identifyCommonEvent(topic.Hex())
				if eventName != "" {
					fmt.Printf("    识别: %s\n", eventName)
				}
			} else {
				fmt.Printf("    说明: 索引参数 %d\n", j)
			}
		}

		// 数据分析
		fmt.Printf("\n📊 数据 (Data) 分析:\n")
		fmt.Printf("数据长度: %d bytes\n", len(log.Data))
		if len(log.Data) > 0 {
			fmt.Printf("数据 (hex): %s\n", common.Bytes2Hex(log.Data))
			fmt.Printf("说明: 非索引参数的编码数据\n")

			// 尝试解析数据 (简单示例)
			if len(log.Data) == 32 {
				value := new(big.Int).SetBytes(log.Data)
				fmt.Printf("可能的数值: %s\n", value.String())
			} else if len(log.Data)%32 == 0 {
				chunks := len(log.Data) / 32
				fmt.Printf("数据块数量: %d (每块32字节)\n", chunks)
			}
		} else {
			fmt.Printf("数据: 空\n")
			fmt.Printf("说明: 所有参数都是索引参数\n")
		}
	}

	// 日志统计
	fmt.Printf("\n📈 日志统计摘要:\n")
	fmt.Println("--------------------------------")

	contractMap := make(map[common.Address]int)
	for _, log := range logs {
		contractMap[log.Address]++
	}

	fmt.Printf("涉及合约数量: %d\n", len(contractMap))
	for addr, count := range contractMap {
		fmt.Printf("  %s: %d 个事件\n", addr.Hex(), count)
	}
}

// analyzeGasUsage 分析 Gas 使用情况
func analyzeGasUsage(ctx context.Context, client *ethclient.Client, txHash common.Hash, receipt *types.Receipt) {
	// 获取原始交易
	tx, _, err := client.TransactionByHash(ctx, txHash)
	if err != nil {
		fmt.Printf("❌ 无法获取原始交易: %v\n", err)
		return
	}

	fmt.Printf("⛽ Gas 使用详细分析:\n")
	fmt.Println("================================")

	gasLimit := tx.Gas()
	gasUsed := receipt.GasUsed
	gasPrice := tx.GasPrice()

	fmt.Printf("Gas 限制: %s\n", formatNumber(gasLimit))
	fmt.Printf("Gas 使用: %s\n", formatNumber(gasUsed))
	fmt.Printf("Gas 价格: %s Gwei\n", weiToGwei(gasPrice))

	// 使用效率
	efficiency := float64(gasUsed) / float64(gasLimit) * 100
	fmt.Printf("使用效率: %.2f%%\n", efficiency)

	// 费用计算
	actualFee := new(big.Int).Mul(big.NewInt(int64(gasUsed)), gasPrice)
	maxFee := new(big.Int).Mul(big.NewInt(int64(gasLimit)), gasPrice)
	savedFee := new(big.Int).Sub(maxFee, actualFee)

	fmt.Printf("实际费用: %s ETH\n", weiToEther(actualFee))
	fmt.Printf("最大费用: %s ETH\n", weiToEther(maxFee))
	fmt.Printf("节省费用: %s ETH\n", weiToEther(savedFee))

	// Gas 使用分析
	fmt.Printf("\n📊 Gas 使用分析:\n")
	baseGas := uint64(21000) // 基础交易 Gas

	if gasUsed <= baseGas {
		fmt.Printf("交易类型: 简单 ETH 转账\n")
		fmt.Printf("基础 Gas: %s\n", formatNumber(baseGas))
	} else {
		extraGas := gasUsed - baseGas
		fmt.Printf("基础 Gas: %s (简单转账)\n", formatNumber(baseGas))
		fmt.Printf("额外 Gas: %s (合约执行/数据存储)\n", formatNumber(extraGas))

		// 估算操作复杂度
		if extraGas < 50000 {
			fmt.Printf("复杂度: 🟢 简单合约调用\n")
		} else if extraGas < 200000 {
			fmt.Printf("复杂度: 🟡 中等复杂度操作\n")
		} else if extraGas < 500000 {
			fmt.Printf("复杂度: 🟠 复杂操作\n")
		} else {
			fmt.Printf("复杂度: 🔴 非常复杂的操作\n")
		}
	}

	// 累计 Gas 分析
	fmt.Printf("\n📈 区块中的位置分析:\n")
	fmt.Printf("累计 Gas 使用: %s\n", formatNumber(receipt.CumulativeGasUsed))
	fmt.Printf("交易索引: %d\n", receipt.TransactionIndex)

	if receipt.TransactionIndex == 0 {
		fmt.Printf("位置: 区块中的第一笔交易\n")
	} else {
		prevCumulativeGas := receipt.CumulativeGasUsed - gasUsed
		fmt.Printf("前序交易累计 Gas: %s\n", formatNumber(prevCumulativeGas))
	}
}

// analyzeReceiptStatus 分析收据状态
func analyzeReceiptStatus(receipt *types.Receipt) {
	fmt.Printf("🔍 收据状态深度分析:\n")
	fmt.Println("================================")

	if receipt.Status == types.ReceiptStatusSuccessful {
		fmt.Printf("✅ 交易执行成功\n")
		fmt.Printf("状态码: 1\n")
		fmt.Printf("含义: 所有操作成功完成，状态变更已生效\n")

		if len(receipt.Logs) > 0 {
			fmt.Printf("事件: 产生了 %d 个事件日志\n", len(receipt.Logs))
		}

		if receipt.ContractAddress != (common.Address{}) {
			fmt.Printf("合约: 成功部署到 %s\n", receipt.ContractAddress.Hex())
		}
	} else {
		fmt.Printf("❌ 交易执行失败\n")
		fmt.Printf("状态码: 0\n")
		fmt.Printf("含义: 交易执行过程中发生错误，状态回滚\n")
		fmt.Printf("注意: 尽管失败，仍然消耗了 %s Gas\n", formatNumber(receipt.GasUsed))

		// 失败原因分析
		fmt.Printf("\n🔍 可能的失败原因:\n")
		if receipt.GasUsed == 21000 {
			fmt.Printf("- 可能是发送到不存在的合约地址\n")
		} else if receipt.GasUsed > 21000 {
			fmt.Printf("- 合约执行过程中发生 revert 或 require 失败\n")
			fmt.Printf("- 可能是权限不足或参数错误\n")
		}
	}

	// 区块确认信息
	fmt.Printf("\n📦 区块确认信息:\n")
	fmt.Printf("区块号: %d\n", receipt.BlockNumber.Uint64())
	fmt.Printf("区块哈希: %s\n", receipt.BlockHash.Hex())
	fmt.Printf("交易在区块中的索引: %d\n", receipt.TransactionIndex)
}

// analyzeBloomFilter 分析 Bloom 过滤器
func analyzeBloomFilter(receipt *types.Receipt) {
	fmt.Printf("🌸 Bloom 过滤器详细分析:\n")
	fmt.Println("================================")

	bloomBig := receipt.Bloom.Big()

	if bloomBig.Cmp(big.NewInt(0)) == 0 {
		fmt.Printf("状态: 空 Bloom 过滤器\n")
		fmt.Printf("含义: 该交易没有产生任何事件日志\n")
		fmt.Printf("用途: 可以快速确定交易不包含特定事件\n")
	} else {
		fmt.Printf("状态: 非空 Bloom 过滤器\n")
		fmt.Printf("含义: 该交易产生了事件日志\n")
		fmt.Printf("位长度: %d bits\n", bloomBig.BitLen())

		// 计算设置的位数 (简化计算)
		setBits := 0
		temp := new(big.Int).Set(bloomBig)
		for temp.Cmp(big.NewInt(0)) > 0 {
			if temp.Bit(0) == 1 {
				setBits++
			}
			temp.Rsh(temp, 1)
		}

		fmt.Printf("设置的位数: %d (估算)\n", setBits)
		fmt.Printf("用途: 快速检索包含特定事件的交易\n")

		// Bloom 过滤器的工作原理说明
		fmt.Printf("\n📚 Bloom 过滤器工作原理:\n")
		fmt.Printf("- 每个事件的地址和主题都会在过滤器中设置特定位\n")
		fmt.Printf("- 可以快速判断交易是否可能包含某个事件\n")
		fmt.Printf("- 可能有假阳性，但不会有假阴性\n")
		fmt.Printf("- 用于优化事件日志的查询性能\n")
	}
}

// identifyCommonEvent 识别常见事件
func identifyCommonEvent(signature string) string {
	commonEvents := map[string]string{
		"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef": "Transfer(address,address,uint256)",
		"0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925": "Approval(address,address,uint256)",
		"0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31": "ApprovalForAll(address,address,bool)",
		"0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0": "OwnershipTransferred(address,address)",
	}

	return commonEvents[signature]
}

// 工具函数

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
