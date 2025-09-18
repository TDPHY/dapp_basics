package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

// BlockMonitor 区块监控器
type BlockMonitor struct {
	client       *ethclient.Client
	ctx          context.Context
	cancel       context.CancelFunc
	stats        *MonitorStats
	alertRules   []AlertRule
	blockHistory []*BlockInfo
}

// MonitorStats 监控统计
type MonitorStats struct {
	StartTime        time.Time
	BlockCount       int64
	TotalTxs         int64
	TotalGasUsed     *big.Int
	MaxTxsPerBlock   int
	MinTxsPerBlock   int
	MaxGasUsage      uint64
	MinGasUsage      uint64
	AverageBlockTime time.Duration
	LastBlockTime    time.Time
}

// BlockInfo 区块信息
type BlockInfo struct {
	Number    uint64
	Hash      string
	Timestamp time.Time
	TxCount   int
	GasUsed   uint64
	GasLimit  uint64
	GasPrice  *big.Int
	BlockTime time.Duration
	Miner     string
}

// AlertRule 告警规则
type AlertRule struct {
	Name      string
	Condition func(*BlockInfo, *MonitorStats) bool
	Message   func(*BlockInfo, *MonitorStats) string
}

func main() {
	fmt.Println("🔍 高级区块监控器")
	fmt.Println("================================")

	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Printf("警告: 无法加载 .env 文件: %v", err)
	}

	// 获取 RPC URL
	rpcURL := os.Getenv("ETHEREUM_RPC_URL")
	if rpcURL == "" {
		log.Fatal("请在 .env 文件中设置 ETHEREUM_RPC_URL")
	}

	// 创建监控器
	monitor, err := NewBlockMonitor(rpcURL)
	if err != nil {
		log.Fatalf("创建监控器失败: %v", err)
	}
	defer monitor.Close()

	// 设置告警规则
	monitor.SetupAlertRules()

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("🔔 启动高级区块监控...")
	fmt.Println("按 Ctrl+C 停止监控")
	fmt.Println("================================")

	// 启动监控
	go monitor.Start()

	// 启动定期报告
	go monitor.PeriodicReport()

	// 等待退出信号
	<-sigChan
	fmt.Println("\n\n🛑 停止监控...")
	monitor.Stop()
}

// NewBlockMonitor 创建新的区块监控器
func NewBlockMonitor(rpcURL string) (*BlockMonitor, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &BlockMonitor{
		client: client,
		ctx:    ctx,
		cancel: cancel,
		stats: &MonitorStats{
			StartTime:      time.Now(),
			TotalGasUsed:   big.NewInt(0),
			MinTxsPerBlock: 999999,
			MinGasUsage:    ^uint64(0), // 最大值
		},
		blockHistory: make([]*BlockInfo, 0, 100), // 保留最近100个区块
	}, nil
}

// SetupAlertRules 设置告警规则
func (m *BlockMonitor) SetupAlertRules() {
	m.alertRules = []AlertRule{
		{
			Name: "高交易量区块",
			Condition: func(block *BlockInfo, stats *MonitorStats) bool {
				return block.TxCount > 200
			},
			Message: func(block *BlockInfo, stats *MonitorStats) string {
				return fmt.Sprintf("🔥 高交易量区块 #%d: %d 笔交易", block.Number, block.TxCount)
			},
		},
		{
			Name: "高Gas使用率",
			Condition: func(block *BlockInfo, stats *MonitorStats) bool {
				return float64(block.GasUsed)/float64(block.GasLimit) > 0.95
			},
			Message: func(block *BlockInfo, stats *MonitorStats) string {
				usage := float64(block.GasUsed) / float64(block.GasLimit) * 100
				return fmt.Sprintf("⚡ 高Gas使用率区块 #%d: %.2f%%", block.Number, usage)
			},
		},
		{
			Name: "长区块间隔",
			Condition: func(block *BlockInfo, stats *MonitorStats) bool {
				return block.BlockTime > 20*time.Second
			},
			Message: func(block *BlockInfo, stats *MonitorStats) string {
				return fmt.Sprintf("⏰ 长区块间隔 #%d: %s", block.Number, block.BlockTime)
			},
		},
		{
			Name: "空区块",
			Condition: func(block *BlockInfo, stats *MonitorStats) bool {
				return block.TxCount == 0
			},
			Message: func(block *BlockInfo, stats *MonitorStats) string {
				return fmt.Sprintf("📭 空区块 #%d: 无交易", block.Number)
			},
		},
	}
}

// Start 启动监控
func (m *BlockMonitor) Start() {
	headers := make(chan *types.Header)
	sub, err := m.client.SubscribeNewHead(m.ctx, headers)
	if err != nil {
		log.Fatalf("创建订阅失败: %v", err)
	}
	defer sub.Unsubscribe()

	fmt.Println("✅ 区块监控已启动")

	for {
		select {
		case err := <-sub.Err():
			log.Printf("❌ 订阅错误: %v", err)
			return

		case header := <-headers:
			m.processBlock(header)

		case <-m.ctx.Done():
			fmt.Println("🔔 区块监控已停止")
			return
		}
	}
}

// processBlock 处理新区块
func (m *BlockMonitor) processBlock(header *types.Header) {
	// 获取完整区块信息
	block, err := m.client.BlockByHash(m.ctx, header.Hash())
	if err != nil {
		log.Printf("获取区块失败: %v", err)
		return
	}

	// 计算区块时间间隔
	var blockTime time.Duration
	if !m.stats.LastBlockTime.IsZero() {
		blockTime = time.Unix(int64(header.Time), 0).Sub(m.stats.LastBlockTime)
	}
	m.stats.LastBlockTime = time.Unix(int64(header.Time), 0)

	// 计算平均Gas价格
	avgGasPrice := m.calculateAverageGasPrice(block.Transactions())

	// 创建区块信息
	blockInfo := &BlockInfo{
		Number:    header.Number.Uint64(),
		Hash:      header.Hash().Hex(),
		Timestamp: time.Unix(int64(header.Time), 0),
		TxCount:   len(block.Transactions()),
		GasUsed:   header.GasUsed,
		GasLimit:  header.GasLimit,
		GasPrice:  avgGasPrice,
		BlockTime: blockTime,
		Miner:     header.Coinbase.Hex(),
	}

	// 更新统计信息
	m.updateStats(blockInfo)

	// 添加到历史记录
	m.addToHistory(blockInfo)

	// 显示区块信息
	m.displayBlockInfo(blockInfo)

	// 检查告警规则
	m.checkAlerts(blockInfo)
}

// calculateAverageGasPrice 计算平均Gas价格
func (m *BlockMonitor) calculateAverageGasPrice(txs types.Transactions) *big.Int {
	if len(txs) == 0 {
		return big.NewInt(0)
	}

	total := big.NewInt(0)
	count := 0

	for _, tx := range txs {
		if gasPrice := tx.GasPrice(); gasPrice != nil {
			total.Add(total, gasPrice)
			count++
		}
	}

	if count == 0 {
		return big.NewInt(0)
	}

	return total.Div(total, big.NewInt(int64(count)))
}

// updateStats 更新统计信息
func (m *BlockMonitor) updateStats(block *BlockInfo) {
	m.stats.BlockCount++
	m.stats.TotalTxs += int64(block.TxCount)
	m.stats.TotalGasUsed.Add(m.stats.TotalGasUsed, big.NewInt(int64(block.GasUsed)))

	// 更新最大最小值
	if block.TxCount > m.stats.MaxTxsPerBlock {
		m.stats.MaxTxsPerBlock = block.TxCount
	}
	if block.TxCount < m.stats.MinTxsPerBlock {
		m.stats.MinTxsPerBlock = block.TxCount
	}
	if block.GasUsed > m.stats.MaxGasUsage {
		m.stats.MaxGasUsage = block.GasUsed
	}
	if block.GasUsed < m.stats.MinGasUsage {
		m.stats.MinGasUsage = block.GasUsed
	}

	// 计算平均区块时间
	if m.stats.BlockCount > 1 && len(m.blockHistory) > 0 {
		totalTime := time.Duration(0)
		count := 0
		for _, b := range m.blockHistory {
			if b.BlockTime > 0 {
				totalTime += b.BlockTime
				count++
			}
		}
		if count > 0 {
			m.stats.AverageBlockTime = totalTime / time.Duration(count)
		}
	}
}

// addToHistory 添加到历史记录
func (m *BlockMonitor) addToHistory(block *BlockInfo) {
	m.blockHistory = append(m.blockHistory, block)

	// 保持最近100个区块
	if len(m.blockHistory) > 100 {
		m.blockHistory = m.blockHistory[1:]
	}
}

// displayBlockInfo 显示区块信息
func (m *BlockMonitor) displayBlockInfo(block *BlockInfo) {
	fmt.Printf("\n🆕 区块 #%d\n", block.Number)
	fmt.Printf("时间: %s\n", block.Timestamp.Format("15:04:05"))
	fmt.Printf("交易: %d 笔\n", block.TxCount)
	fmt.Printf("Gas: %s/%s (%.1f%%)\n",
		formatGas(block.GasUsed),
		formatGas(block.GasLimit),
		float64(block.GasUsed)/float64(block.GasLimit)*100)

	if block.GasPrice.Cmp(big.NewInt(0)) > 0 {
		fmt.Printf("平均Gas价格: %s Gwei\n", formatGwei(block.GasPrice))
	}

	if block.BlockTime > 0 {
		fmt.Printf("区块间隔: %s\n", block.BlockTime.Round(time.Millisecond))
	}
}

// checkAlerts 检查告警规则
func (m *BlockMonitor) checkAlerts(block *BlockInfo) {
	for _, rule := range m.alertRules {
		if rule.Condition(block, m.stats) {
			fmt.Printf("🚨 %s\n", rule.Message(block, m.stats))
		}
	}
}

// PeriodicReport 定期报告
func (m *BlockMonitor) PeriodicReport() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.generateReport()
		case <-m.ctx.Done():
			return
		}
	}
}

// generateReport 生成报告
func (m *BlockMonitor) generateReport() {
	fmt.Println("\n📊 定期监控报告")
	fmt.Println("================================")

	duration := time.Since(m.stats.StartTime)
	fmt.Printf("监控时长: %s\n", formatDuration(duration))
	fmt.Printf("处理区块: %d 个\n", m.stats.BlockCount)
	fmt.Printf("总交易数: %d 笔\n", m.stats.TotalTxs)

	if m.stats.BlockCount > 0 {
		avgTxs := float64(m.stats.TotalTxs) / float64(m.stats.BlockCount)
		fmt.Printf("平均交易/区块: %.1f 笔\n", avgTxs)

		fmt.Printf("交易数范围: %d - %d 笔\n", m.stats.MinTxsPerBlock, m.stats.MaxTxsPerBlock)
		fmt.Printf("Gas使用范围: %s - %s\n",
			formatGas(m.stats.MinGasUsage), formatGas(m.stats.MaxGasUsage))
	}

	if m.stats.AverageBlockTime > 0 {
		fmt.Printf("平均区块时间: %s\n", m.stats.AverageBlockTime.Round(time.Millisecond))
	}

	// 显示最近区块的统计
	m.displayRecentBlocksStats()

	fmt.Println("================================")
}

// displayRecentBlocksStats 显示最近区块统计
func (m *BlockMonitor) displayRecentBlocksStats() {
	if len(m.blockHistory) < 10 {
		return
	}

	recent := m.blockHistory[len(m.blockHistory)-10:]

	// 计算最近10个区块的统计
	totalTxs := 0
	totalGas := uint64(0)
	var blockTimes []time.Duration

	for _, block := range recent {
		totalTxs += block.TxCount
		totalGas += block.GasUsed
		if block.BlockTime > 0 {
			blockTimes = append(blockTimes, block.BlockTime)
		}
	}

	fmt.Printf("\n最近10个区块:\n")
	fmt.Printf("  平均交易数: %.1f 笔\n", float64(totalTxs)/10)
	fmt.Printf("  平均Gas使用: %s\n", formatGas(totalGas/10))

	if len(blockTimes) > 0 {
		sort.Slice(blockTimes, func(i, j int) bool {
			return blockTimes[i] < blockTimes[j]
		})

		median := blockTimes[len(blockTimes)/2]
		fmt.Printf("  中位区块时间: %s\n", median.Round(time.Millisecond))
	}
}

// Stop 停止监控
func (m *BlockMonitor) Stop() {
	m.cancel()
	m.generateFinalReport()
}

// generateFinalReport 生成最终报告
func (m *BlockMonitor) generateFinalReport() {
	fmt.Println("\n📈 最终监控报告")
	fmt.Println("================================")

	duration := time.Since(m.stats.StartTime)
	fmt.Printf("总监控时长: %s\n", formatDuration(duration))
	fmt.Printf("处理区块总数: %d 个\n", m.stats.BlockCount)
	fmt.Printf("处理交易总数: %d 笔\n", m.stats.TotalTxs)

	if m.stats.BlockCount > 0 {
		blocksPerHour := float64(m.stats.BlockCount) / duration.Hours()
		fmt.Printf("平均区块频率: %.1f 个/小时\n", blocksPerHour)

		avgTxs := float64(m.stats.TotalTxs) / float64(m.stats.BlockCount)
		fmt.Printf("平均交易/区块: %.1f 笔\n", avgTxs)
	}

	if m.stats.AverageBlockTime > 0 {
		fmt.Printf("平均区块时间: %s\n", m.stats.AverageBlockTime.Round(time.Millisecond))
	}

	fmt.Printf("Gas使用统计: %s (总计)\n", formatGas(m.stats.TotalGasUsed.Uint64()))

	fmt.Println("监控已完成!")
}

// Close 关闭监控器
func (m *BlockMonitor) Close() {
	if m.client != nil {
		m.client.Close()
	}
}

// 格式化函数
func formatGas(gas uint64) string {
	if gas >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(gas)/1000000)
	} else if gas >= 1000 {
		return fmt.Sprintf("%.1fK", float64(gas)/1000)
	}
	return fmt.Sprintf("%d", gas)
}

func formatGwei(wei *big.Int) string {
	gwei := new(big.Float).SetInt(wei)
	gwei.Quo(gwei, big.NewFloat(1e9))
	return fmt.Sprintf("%.2f", gwei)
}

func formatDuration(d time.Duration) string {
	if d.Hours() >= 1 {
		return fmt.Sprintf("%.1f小时", d.Hours())
	} else if d.Minutes() >= 1 {
		return fmt.Sprintf("%.1f分钟", d.Minutes())
	} else {
		return fmt.Sprintf("%.1f秒", d.Seconds())
	}
}
