package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/local/go-eth-demo/config"
	"github.com/local/go-eth-demo/utils"
)

func main() {
	fmt.Println("📋 以太坊区块头信息查询")
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

	// 1. 获取最新区块头
	fmt.Println("🔍 查询最新区块头...")
	latestHeader, err := ethClient.HeaderByNumber(ctx, nil)
	if err != nil {
		log.Fatalf("❌ 获取最新区块头失败: %v", err)
	}
	displayHeaderInfo("最新区块头", latestHeader)

	// 2. 获取指定区块号的区块头
	fmt.Println("\n🔍 查询指定区块头...")
	blockNumber := new(big.Int).Sub(latestHeader.Number, big.NewInt(100))
	specificHeader, err := ethClient.HeaderByNumber(ctx, blockNumber)
	if err != nil {
		log.Printf("❌ 获取指定区块头失败: %v", err)
	} else {
		displayHeaderInfo(fmt.Sprintf("区块 #%s 头信息", blockNumber.String()), specificHeader)
	}

	// 3. 根据哈希获取区块头
	fmt.Println("\n🔍 根据哈希查询区块头...")
	headerByHash, err := ethClient.HeaderByHash(ctx, latestHeader.Hash())
	if err != nil {
		log.Printf("❌ 根据哈希获取区块头失败: %v", err)
	} else {
		fmt.Printf("✅ 通过哈希查询成功，区块号: %s\n", headerByHash.Number.String())
	}

	// 4. 比较区块头和完整区块的差异
	fmt.Println("\n📊 区块头 vs 完整区块对比...")
	compareHeaderAndBlock(ctx, ethClient, latestHeader.Number)

	// 5. 分析区块头中的关键字段
	fmt.Println("\n🔬 区块头关键字段分析...")
	analyzeHeaderFields(latestHeader)

	fmt.Println("\n✅ 区块头查询学习完成!")
}

// displayHeaderInfo 显示区块头详细信息
func displayHeaderInfo(title string, header *types.Header) {
	fmt.Printf("\n📋 %s:\n", title)
	fmt.Println("--------------------------------")

	// 基本标识信息
	fmt.Printf("区块号: %s\n", header.Number.String())
	fmt.Printf("区块哈希: %s\n", header.Hash().Hex())
	fmt.Printf("父区块哈希: %s\n", header.ParentHash.Hex())

	// 时间信息
	blockTime := time.Unix(int64(header.Time), 0)
	fmt.Printf("时间戳: %d (%s)\n", header.Time, blockTime.Format("2006-01-02 15:04:05"))

	// 挖矿相关
	fmt.Printf("矿工地址: %s\n", header.Coinbase.Hex())
	fmt.Printf("难度: %s\n", header.Difficulty.String())
	fmt.Printf("Nonce: %d\n", header.Nonce.Uint64())
	fmt.Printf("Mix Hash: %s\n", header.MixDigest.Hex())

	// Gas 信息
	fmt.Printf("Gas 限制: %s\n", formatNumber(header.GasLimit))
	fmt.Printf("Gas 使用: %s\n", formatNumber(header.GasUsed))
	gasUsagePercent := float64(header.GasUsed) / float64(header.GasLimit) * 100
	fmt.Printf("Gas 使用率: %.2f%%\n", gasUsagePercent)

	// Merkle 树根
	fmt.Printf("状态根: %s\n", header.Root.Hex())
	fmt.Printf("交易根: %s\n", header.TxHash.Hex())
	fmt.Printf("收据根: %s\n", header.ReceiptHash.Hex())

	// 其他信息
	fmt.Printf("Bloom 过滤器: %s\n", header.Bloom.Big().String())
	fmt.Printf("Extra Data: %s\n", string(header.Extra))

	// EIP-1559 相关 (如果存在)
	if header.BaseFee != nil {
		fmt.Printf("基础费用: %s Wei (%s Gwei)\n",
			header.BaseFee.String(),
			weiToGwei(header.BaseFee))
	}
}

// compareHeaderAndBlock 比较区块头和完整区块
func compareHeaderAndBlock(ctx context.Context, client *ethclient.Client, blockNumber *big.Int) {
	// 获取区块头
	start := time.Now()
	header, err := client.HeaderByNumber(ctx, blockNumber)
	headerTime := time.Since(start)

	if err != nil {
		log.Printf("❌ 获取区块头失败: %v", err)
		return
	}

	// 获取完整区块
	start = time.Now()
	block, err := client.BlockByNumber(ctx, blockNumber)
	blockTime := time.Since(start)

	if err != nil {
		log.Printf("❌ 获取完整区块失败: %v", err)
		return
	}

	fmt.Printf("📊 性能对比 (区块 #%s):\n", blockNumber.String())
	fmt.Printf("  区块头查询时间: %v\n", headerTime)
	fmt.Printf("  完整区块查询时间: %v\n", blockTime)
	fmt.Printf("  性能提升: %.2fx\n", float64(blockTime.Nanoseconds())/float64(headerTime.Nanoseconds()))

	fmt.Printf("\n📋 数据对比:\n")
	fmt.Printf("  区块头大小: ~500 bytes (估算)\n")
	fmt.Printf("  完整区块大小: %d bytes\n", block.Size())
	fmt.Printf("  完整区块包含交易数: %d\n", len(block.Transactions()))

	// 验证数据一致性
	fmt.Printf("\n🔍 数据一致性验证:\n")
	fmt.Printf("  区块号一致: %v\n", header.Number.Cmp(block.Number()) == 0)
	fmt.Printf("  区块哈希一致: %v\n", header.Hash() == block.Hash())
	fmt.Printf("  Gas 使用一致: %v\n", header.GasUsed == block.GasUsed())
	fmt.Printf("  时间戳一致: %v\n", header.Time == block.Time())
}

// analyzeHeaderFields 分析区块头关键字段
func analyzeHeaderFields(header *types.Header) {
	fmt.Printf("🔬 关键字段深度分析:\n")
	fmt.Println("--------------------------------")

	// 1. Bloom 过滤器分析
	fmt.Printf("📊 Bloom 过滤器:\n")
	bloomBits := header.Bloom.Big()
	if bloomBits.Cmp(big.NewInt(0)) == 0 {
		fmt.Printf("  状态: 空 (该区块没有日志事件)\n")
	} else {
		fmt.Printf("  状态: 非空 (该区块包含日志事件)\n")
		fmt.Printf("  位数: %d\n", bloomBits.BitLen())
	}

	// 2. 难度分析
	fmt.Printf("\n⚡ 挖矿难度:\n")
	difficulty := header.Difficulty
	fmt.Printf("  当前难度: %s\n", difficulty.String())

	// 估算挖矿时间 (基于难度)
	if difficulty.Cmp(big.NewInt(0)) > 0 {
		// 这是一个简化的估算，实际情况更复杂
		fmt.Printf("  难度级别: %s\n", getDifficultyLevel(difficulty))
	}

	// 3. Gas 分析
	fmt.Printf("\n⛽ Gas 详细分析:\n")
	gasLimit := header.GasLimit
	gasUsed := header.GasUsed

	fmt.Printf("  Gas 限制: %s\n", formatNumber(gasLimit))
	fmt.Printf("  Gas 使用: %s\n", formatNumber(gasUsed))
	fmt.Printf("  剩余 Gas: %s\n", formatNumber(gasLimit-gasUsed))

	utilization := float64(gasUsed) / float64(gasLimit) * 100
	fmt.Printf("  利用率: %.2f%%\n", utilization)

	// Gas 利用率评估
	var status string
	switch {
	case utilization > 95:
		status = "🔴 极度拥堵"
	case utilization > 80:
		status = "🟡 拥堵"
	case utilization > 50:
		status = "🟢 正常"
	default:
		status = "🔵 空闲"
	}
	fmt.Printf("  网络状态: %s\n", status)

	// 4. 时间分析
	fmt.Printf("\n⏰ 时间信息:\n")
	blockTime := time.Unix(int64(header.Time), 0)
	now := time.Now()
	age := now.Sub(blockTime)

	fmt.Printf("  区块时间: %s\n", blockTime.Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("  区块年龄: %v\n", age.Truncate(time.Second))

	if age < time.Minute {
		fmt.Printf("  状态: 🟢 最新区块\n")
	} else if age < time.Hour {
		fmt.Printf("  状态: 🟡 较新区块\n")
	} else {
		fmt.Printf("  状态: 🔴 历史区块\n")
	}

	// 5. Extra Data 分析
	fmt.Printf("\n📝 Extra Data 分析:\n")
	extraData := header.Extra
	if len(extraData) == 0 {
		fmt.Printf("  内容: 空\n")
	} else {
		fmt.Printf("  长度: %d bytes\n", len(extraData))
		fmt.Printf("  内容 (hex): %x\n", extraData)
		fmt.Printf("  内容 (string): %s\n", string(extraData))

		// 尝试识别常见的矿池标识
		extraStr := string(extraData)
		if len(extraStr) > 0 {
			fmt.Printf("  可能的矿池: %s\n", identifyMiningPool(extraStr))
		}
	}
}

// getDifficultyLevel 获取难度级别描述
func getDifficultyLevel(difficulty *big.Int) string {
	// 这是一个简化的分类，实际的难度评估更复杂
	diffFloat := new(big.Float).SetInt(difficulty)

	// 使用科学记数法表示
	return fmt.Sprintf("%.2e", diffFloat)
}

// identifyMiningPool 识别矿池
func identifyMiningPool(extraData string) string {
	// 简化的矿池识别逻辑
	poolMap := map[string]string{
		"Ethermine": "Ethermine",
		"f2pool":    "F2Pool",
		"SparkPool": "SparkPool",
		"Hiveon":    "Hiveon Pool",
		"2miners":   "2Miners",
		"Nanopool":  "Nanopool",
		"Flexpool":  "Flexpool",
	}

	for key, pool := range poolMap {
		if contains(extraData, key) {
			return pool
		}
	}

	return "未知矿池"
}

// 工具函数

// contains 检查字符串是否包含子字符串（忽略大小写）
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
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
