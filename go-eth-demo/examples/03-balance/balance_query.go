package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
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

	fmt.Println("💰 以太坊账户余额查询演示")
	fmt.Println("================================")

	// 测试地址列表 (一些知名地址)
	testAddresses := []struct {
		name    string
		address string
		desc    string
	}{
		{
			name:    "Vitalik Buterin",
			address: "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045",
			desc:    "以太坊创始人地址",
		},
		{
			name:    "Uniswap V3 Router",
			address: "0xE592427A0AEce92De3Edee1F18E0157C05861564",
			desc:    "Uniswap V3 路由合约",
		},
		{
			name:    "USDC Contract",
			address: "0xA0b86a33E6441b8C4505B4afDcA7FBf074d9eeE4",
			desc:    "USDC 代币合约 (Sepolia)",
		},
		{
			name:    "Random Address",
			address: "0x1234567890123456789012345678901234567890",
			desc:    "随机测试地址",
		},
	}

	// 1. 批量查询 ETH 余额
	fmt.Println("\n🔍 批量查询 ETH 余额:")
	fmt.Println("--------------------------------")

	for i, addr := range testAddresses {
		fmt.Printf("\n📍 地址 #%d: %s\n", i+1, addr.name)
		fmt.Printf("描述: %s\n", addr.desc)
		fmt.Printf("地址: %s\n", addr.address)

		balance, err := queryETHBalance(ctx, ethClient, addr.address)
		if err != nil {
			fmt.Printf("❌ 查询失败: %v\n", err)
			continue
		}

		displayETHBalance(balance)

		// 查询历史余额 (前一个区块)
		if err := queryHistoricalBalance(ctx, ethClient, addr.address); err != nil {
			fmt.Printf("⚠️  历史余额查询失败: %v\n", err)
		}
	}

	// 2. 余额变化分析
	fmt.Println("\n\n📊 余额变化分析:")
	fmt.Println("================================")

	// 选择一个活跃地址进行分析
	activeAddress := testAddresses[0].address
	fmt.Printf("分析地址: %s (%s)\n", activeAddress, testAddresses[0].name)

	if err := analyzeBalanceHistory(ctx, ethClient, activeAddress); err != nil {
		fmt.Printf("❌ 余额分析失败: %v\n", err)
	}

	// 3. 多地址余额对比
	fmt.Println("\n\n🔄 多地址余额对比:")
	fmt.Println("================================")

	compareAddresses(ctx, ethClient, testAddresses)

	fmt.Println("\n✅ 账户余额查询演示完成！")
}

// queryETHBalance 查询 ETH 余额
func queryETHBalance(ctx context.Context, ethClient *utils.EthClient, addressStr string) (*big.Int, error) {
	address := common.HexToAddress(addressStr)

	balance, err := ethClient.GetClient().BalanceAt(ctx, address, nil)
	if err != nil {
		return nil, fmt.Errorf("查询余额失败: %w", err)
	}

	return balance, nil
}

// displayETHBalance 显示 ETH 余额信息
func displayETHBalance(balance *big.Int) {
	// 转换为不同单位
	weiStr := balance.String()
	etherStr := utils.WeiToEther(balance)
	gweiStr := utils.WeiToGwei(balance)

	fmt.Printf("💎 ETH 余额:\n")
	fmt.Printf("  Wei:   %s\n", weiStr)
	fmt.Printf("  Gwei:  %s\n", gweiStr)
	fmt.Printf("  Ether: %s ETH\n", etherStr)

	// 余额等级分析
	analyzeBalanceLevel(balance)
}

// analyzeBalanceLevel 分析余额等级
func analyzeBalanceLevel(balance *big.Int) {
	// 转换为 Ether 进行比较
	etherFloat := new(big.Float).SetInt(balance)
	etherFloat.Quo(etherFloat, big.NewFloat(1e18))

	ether, _ := etherFloat.Float64()

	var level, emoji, desc string

	switch {
	case ether == 0:
		level = "空账户"
		emoji = "🚫"
		desc = "没有 ETH 余额"
	case ether < 0.001:
		level = "尘埃级"
		emoji = "🌫️"
		desc = "极少量 ETH，可能是测试或空投残留"
	case ether < 0.01:
		level = "微量级"
		emoji = "💧"
		desc = "少量 ETH，适合小额交易"
	case ether < 0.1:
		level = "小额级"
		emoji = "🪙"
		desc = "小额 ETH，适合日常使用"
	case ether < 1:
		level = "常规级"
		emoji = "💰"
		desc = "常规 ETH 余额，适合多数操作"
	case ether < 10:
		level = "富裕级"
		emoji = "💎"
		desc = "较多 ETH，可进行大额操作"
	case ether < 100:
		level = "大户级"
		emoji = "🏆"
		desc = "大量 ETH，属于大户范畴"
	default:
		level = "巨鲸级"
		emoji = "🐋"
		desc = "巨量 ETH，属于巨鲸级别"
	}

	fmt.Printf("  等级: %s %s\n", emoji, level)
	fmt.Printf("  说明: %s\n", desc)
}

// queryHistoricalBalance 查询历史余额
func queryHistoricalBalance(ctx context.Context, ethClient *utils.EthClient, addressStr string) error {
	address := common.HexToAddress(addressStr)

	// 获取当前区块号
	currentBlock, err := ethClient.GetLatestBlockNumber()
	if err != nil {
		return err
	}

	// 查询前一个区块的余额
	prevBlock := new(big.Int).Sub(currentBlock, big.NewInt(1))

	prevBalance, err := ethClient.GetClient().BalanceAt(ctx, address, prevBlock)
	if err != nil {
		return err
	}

	// 获取当前余额
	currentBalance, err := ethClient.GetClient().BalanceAt(ctx, address, nil)
	if err != nil {
		return err
	}

	// 计算变化
	change := new(big.Int).Sub(currentBalance, prevBalance)

	fmt.Printf("📈 余额变化 (最近一个区块):\n")
	fmt.Printf("  前一区块 (#%s): %s ETH\n", prevBlock.String(), utils.WeiToEther(prevBalance))
	fmt.Printf("  当前区块 (#%s): %s ETH\n", currentBlock.String(), utils.WeiToEther(currentBalance))

	if change.Sign() == 0 {
		fmt.Printf("  变化: 无变化 ⚪\n")
	} else if change.Sign() > 0 {
		fmt.Printf("  变化: +%s ETH 📈\n", utils.WeiToEther(change))
	} else {
		absChange := new(big.Int).Abs(change)
		fmt.Printf("  变化: -%s ETH 📉\n", utils.WeiToEther(absChange))
	}

	return nil
}

// analyzeBalanceHistory 分析余额历史
func analyzeBalanceHistory(ctx context.Context, ethClient *utils.EthClient, addressStr string) error {
	address := common.HexToAddress(addressStr)

	// 获取当前区块号
	currentBlock, err := ethClient.GetLatestBlockNumber()
	if err != nil {
		return err
	}

	fmt.Printf("分析最近 5 个区块的余额变化...\n")

	var balances []*big.Int
	var blockNumbers []*big.Int

	// 查询最近 5 个区块的余额
	for i := 4; i >= 0; i-- {
		blockNum := new(big.Int).Sub(currentBlock, big.NewInt(int64(i)))
		balance, err := ethClient.GetClient().BalanceAt(ctx, address, blockNum)
		if err != nil {
			return err
		}

		balances = append(balances, balance)
		blockNumbers = append(blockNumbers, blockNum)
	}

	// 显示余额历史
	fmt.Printf("\n📊 余额历史记录:\n")
	for i, balance := range balances {
		blockNum := blockNumbers[i]
		etherStr := utils.WeiToEther(balance)

		var indicator string
		if i > 0 {
			prev := balances[i-1]
			if balance.Cmp(prev) > 0 {
				indicator = "📈"
			} else if balance.Cmp(prev) < 0 {
				indicator = "📉"
			} else {
				indicator = "⚪"
			}
		} else {
			indicator = "🔵"
		}

		fmt.Printf("  区块 #%s: %s ETH %s\n", blockNum.String(), etherStr, indicator)
	}

	// 计算总变化
	totalChange := new(big.Int).Sub(balances[len(balances)-1], balances[0])
	fmt.Printf("\n📈 总变化 (5个区块): ")
	if totalChange.Sign() == 0 {
		fmt.Printf("无变化\n")
	} else if totalChange.Sign() > 0 {
		fmt.Printf("+%s ETH\n", utils.WeiToEther(totalChange))
	} else {
		absChange := new(big.Int).Abs(totalChange)
		fmt.Printf("-%s ETH\n", utils.WeiToEther(absChange))
	}

	return nil
}

// compareAddresses 对比多个地址的余额
func compareAddresses(ctx context.Context, ethClient *utils.EthClient, addresses []struct {
	name    string
	address string
	desc    string
}) {
	type AddressBalance struct {
		name    string
		address string
		balance *big.Int
		ether   float64
	}

	var balances []AddressBalance

	// 查询所有地址的余额
	for _, addr := range addresses {
		balance, err := queryETHBalance(ctx, ethClient, addr.address)
		if err != nil {
			fmt.Printf("❌ %s 查询失败: %v\n", addr.name, err)
			continue
		}

		// 转换为 Ether
		etherFloat := new(big.Float).SetInt(balance)
		etherFloat.Quo(etherFloat, big.NewFloat(1e18))
		ether, _ := etherFloat.Float64()

		balances = append(balances, AddressBalance{
			name:    addr.name,
			address: addr.address,
			balance: balance,
			ether:   ether,
		})
	}

	// 按余额排序 (简单冒泡排序)
	for i := 0; i < len(balances)-1; i++ {
		for j := 0; j < len(balances)-1-i; j++ {
			if balances[j].ether < balances[j+1].ether {
				balances[j], balances[j+1] = balances[j+1], balances[j]
			}
		}
	}

	// 显示排序结果
	fmt.Printf("💰 余额排行榜 (从高到低):\n")
	for i, bal := range balances {
		var medal string
		switch i {
		case 0:
			medal = "🥇"
		case 1:
			medal = "🥈"
		case 2:
			medal = "🥉"
		default:
			medal = fmt.Sprintf("#%d", i+1)
		}

		fmt.Printf("  %s %s: %s ETH\n", medal, bal.name, utils.WeiToEther(bal.balance))
	}

	// 计算统计信息
	if len(balances) > 0 {
		var total big.Int
		for _, bal := range balances {
			total.Add(&total, bal.balance)
		}

		avgBalance := new(big.Int).Div(&total, big.NewInt(int64(len(balances))))

		fmt.Printf("\n📊 统计信息:\n")
		fmt.Printf("  总余额: %s ETH\n", utils.WeiToEther(&total))
		fmt.Printf("  平均余额: %s ETH\n", utils.WeiToEther(avgBalance))
		fmt.Printf("  最高余额: %s ETH (%s)\n", utils.WeiToEther(balances[0].balance), balances[0].name)
		fmt.Printf("  最低余额: %s ETH (%s)\n", utils.WeiToEther(balances[len(balances)-1].balance), balances[len(balances)-1].name)
	}
}
