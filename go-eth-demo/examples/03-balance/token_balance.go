package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/local/go-eth-demo/config"
	"github.com/local/go-eth-demo/utils"
)

// ERC20 ABI 定义 (简化版本，只包含必要的方法)
const erc20ABI = `[
	{
		"constant": true,
		"inputs": [{"name": "_owner", "type": "address"}],
		"name": "balanceOf",
		"outputs": [{"name": "balance", "type": "uint256"}],
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "decimals",
		"outputs": [{"name": "", "type": "uint8"}],
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "name",
		"outputs": [{"name": "", "type": "string"}],
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "symbol",
		"outputs": [{"name": "", "type": "string"}],
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "totalSupply",
		"outputs": [{"name": "", "type": "uint256"}],
		"type": "function"
	}
]`

// TokenInfo 代币信息结构
type TokenInfo struct {
	Address     common.Address
	Name        string
	Symbol      string
	Decimals    uint8
	TotalSupply *big.Int
}

// TokenBalance 代币余额结构
type TokenBalance struct {
	Token     TokenInfo
	Balance   *big.Int
	Formatted string
}

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

	fmt.Println("🪙 ERC-20 代币余额查询演示")
	fmt.Println("================================")

	// Sepolia 测试网上的一些代币合约地址
	tokenContracts := []struct {
		name    string
		address string
		desc    string
	}{
		{
			name:    "USDC",
			address: "0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238",
			desc:    "USD Coin (Sepolia 测试网)",
		},
		{
			name:    "USDT",
			address: "0xaA8E23Fb1079EA71e0a56F48a2aA51851D8433D0",
			desc:    "Tether USD (Sepolia 测试网)",
		},
		{
			name:    "WETH",
			address: "0xfFf9976782d46CC05630D1f6eBAb18b2324d6B14",
			desc:    "Wrapped Ether (Sepolia 测试网)",
		},
	}

	// 要查询的地址
	testAddresses := []struct {
		name    string
		address string
	}{
		{
			name:    "Vitalik Buterin",
			address: "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045",
		},
		{
			name:    "Random Address",
			address: "0x1234567890123456789012345678901234567890",
		},
	}

	// 1. 查询代币信息
	fmt.Println("\n📋 代币信息查询:")
	fmt.Println("--------------------------------")

	var tokens []TokenInfo
	for _, tokenContract := range tokenContracts {
		fmt.Printf("\n🪙 %s (%s)\n", tokenContract.name, tokenContract.desc)
		fmt.Printf("合约地址: %s\n", tokenContract.address)

		tokenInfo, err := getTokenInfo(ctx, ethClient, tokenContract.address)
		if err != nil {
			fmt.Printf("❌ 获取代币信息失败: %v\n", err)
			continue
		}

		tokens = append(tokens, *tokenInfo)
		displayTokenInfo(tokenInfo)
	}

	// 2. 批量查询代币余额
	fmt.Println("\n\n💰 代币余额查询:")
	fmt.Println("================================")

	for _, addr := range testAddresses {
		fmt.Printf("\n👤 地址: %s (%s)\n", addr.address, addr.name)
		fmt.Println("--------------------------------")

		for _, token := range tokens {
			balance, err := getTokenBalance(ctx, ethClient, token.Address, addr.address)
			if err != nil {
				fmt.Printf("❌ %s 余额查询失败: %v\n", token.Symbol, err)
				continue
			}

			formatted := formatTokenBalance(balance, token.Decimals)
			fmt.Printf("💎 %s (%s): %s\n", token.Symbol, token.Name, formatted)

			// 分析余额等级
			analyzeTokenBalance(balance, token.Decimals, token.Symbol)
		}
	}

	// 3. 代币持有分析
	fmt.Println("\n\n📊 代币持有分析:")
	fmt.Println("================================")

	analyzeTokenHoldings(ctx, ethClient, tokens, testAddresses)

	fmt.Println("\n✅ ERC-20 代币余额查询演示完成！")
}

// getTokenInfo 获取代币信息
func getTokenInfo(ctx context.Context, ethClient *utils.EthClient, tokenAddress string) (*TokenInfo, error) {
	address := common.HexToAddress(tokenAddress)

	// 解析 ABI
	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return nil, fmt.Errorf("解析 ABI 失败: %w", err)
	}

	tokenInfo := &TokenInfo{
		Address: address,
	}

	// 获取代币名称
	if name, err := callContractMethod(ctx, ethClient, address, parsedABI, "name"); err == nil {
		if len(name) > 0 {
			tokenInfo.Name = name[0].(string)
		}
	}

	// 获取代币符号
	if symbol, err := callContractMethod(ctx, ethClient, address, parsedABI, "symbol"); err == nil {
		if len(symbol) > 0 {
			tokenInfo.Symbol = symbol[0].(string)
		}
	}

	// 获取小数位数
	if decimals, err := callContractMethod(ctx, ethClient, address, parsedABI, "decimals"); err == nil {
		if len(decimals) > 0 {
			tokenInfo.Decimals = decimals[0].(uint8)
		}
	}

	// 获取总供应量
	if totalSupply, err := callContractMethod(ctx, ethClient, address, parsedABI, "totalSupply"); err == nil {
		if len(totalSupply) > 0 {
			tokenInfo.TotalSupply = totalSupply[0].(*big.Int)
		}
	}

	return tokenInfo, nil
}

// getTokenBalance 获取代币余额
func getTokenBalance(ctx context.Context, ethClient *utils.EthClient, tokenAddress common.Address, userAddress string) (*big.Int, error) {
	// 解析 ABI
	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return nil, fmt.Errorf("解析 ABI 失败: %w", err)
	}

	// 调用 balanceOf 方法
	userAddr := common.HexToAddress(userAddress)
	result, err := callContractMethod(ctx, ethClient, tokenAddress, parsedABI, "balanceOf", userAddr)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return big.NewInt(0), nil
	}

	balance, ok := result[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("无法解析余额")
	}

	return balance, nil
}

// callContractMethod 调用合约方法
func callContractMethod(ctx context.Context, ethClient *utils.EthClient, contractAddress common.Address, parsedABI abi.ABI, methodName string, args ...interface{}) ([]interface{}, error) {
	// 编码方法调用
	data, err := parsedABI.Pack(methodName, args...)
	if err != nil {
		return nil, fmt.Errorf("编码方法调用失败: %w", err)
	}

	// 创建调用消息
	msg := ethereum.CallMsg{
		To:   &contractAddress,
		Data: data,
	}

	// 执行调用
	result, err := ethClient.GetClient().CallContract(ctx, msg, nil)
	if err != nil {
		return nil, fmt.Errorf("调用合约失败: %w", err)
	}

	// 解码结果
	values, err := parsedABI.Unpack(methodName, result)
	if err != nil {
		return nil, fmt.Errorf("解码结果失败: %w", err)
	}

	return values, nil
}

// displayTokenInfo 显示代币信息
func displayTokenInfo(token *TokenInfo) {
	fmt.Printf("  名称: %s\n", token.Name)
	fmt.Printf("  符号: %s\n", token.Symbol)
	fmt.Printf("  小数位: %d\n", token.Decimals)

	if token.TotalSupply != nil {
		totalSupplyFormatted := formatTokenBalance(token.TotalSupply, token.Decimals)
		fmt.Printf("  总供应量: %s %s\n", totalSupplyFormatted, token.Symbol)
	}
}

// formatTokenBalance 格式化代币余额
func formatTokenBalance(balance *big.Int, decimals uint8) string {
	if balance.Sign() == 0 {
		return "0"
	}

	// 转换为浮点数进行格式化
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	balanceFloat := new(big.Float).SetInt(balance)
	divisorFloat := new(big.Float).SetInt(divisor)

	result := new(big.Float).Quo(balanceFloat, divisorFloat)

	// 格式化为字符串，保留适当的小数位数
	return result.Text('f', 6)
}

// analyzeTokenBalance 分析代币余额等级
func analyzeTokenBalance(balance *big.Int, decimals uint8, symbol string) {
	if balance.Sign() == 0 {
		fmt.Printf("    等级: 🚫 无持有\n")
		return
	}

	// 转换为浮点数进行分析
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	balanceFloat := new(big.Float).SetInt(balance)
	divisorFloat := new(big.Float).SetInt(divisor)

	result := new(big.Float).Quo(balanceFloat, divisorFloat)
	amount, _ := result.Float64()

	var level, emoji string

	// 根据不同代币类型设置不同的等级标准
	switch {
	case strings.Contains(symbol, "USD") || symbol == "USDT" || symbol == "USDC":
		// 稳定币标准
		switch {
		case amount < 1:
			level, emoji = "尘埃级", "🌫️"
		case amount < 100:
			level, emoji = "小额级", "🪙"
		case amount < 1000:
			level, emoji = "常规级", "💰"
		case amount < 10000:
			level, emoji = "富裕级", "💎"
		default:
			level, emoji = "大户级", "🏆"
		}
	default:
		// 其他代币标准
		switch {
		case amount < 0.01:
			level, emoji = "尘埃级", "🌫️"
		case amount < 1:
			level, emoji = "小额级", "🪙"
		case amount < 100:
			level, emoji = "常规级", "💰"
		case amount < 1000:
			level, emoji = "富裕级", "💎"
		default:
			level, emoji = "大户级", "🏆"
		}
	}

	fmt.Printf("    等级: %s %s\n", emoji, level)
}

// analyzeTokenHoldings 分析代币持有情况
func analyzeTokenHoldings(ctx context.Context, ethClient *utils.EthClient, tokens []TokenInfo, addresses []struct {
	name    string
	address string
}) {
	fmt.Printf("分析 %d 个地址在 %d 种代币上的持有情况...\n", len(addresses), len(tokens))

	// 统计每个地址的持有情况
	for _, addr := range addresses {
		fmt.Printf("\n👤 %s (%s):\n", addr.name, addr.address)

		holdingCount := 0
		var totalValue float64 // 简化的价值计算

		for _, token := range tokens {
			balance, err := getTokenBalance(ctx, ethClient, token.Address, addr.address)
			if err != nil {
				continue
			}

			if balance.Sign() > 0 {
				holdingCount++

				// 简化的价值估算 (假设稳定币价值为1)
				if strings.Contains(token.Symbol, "USD") || token.Symbol == "USDT" || token.Symbol == "USDC" {
					divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(token.Decimals)), nil)
					balanceFloat := new(big.Float).SetInt(balance)
					divisorFloat := new(big.Float).SetInt(divisor)
					result := new(big.Float).Quo(balanceFloat, divisorFloat)
					amount, _ := result.Float64()
					totalValue += amount
				}
			}
		}

		fmt.Printf("  持有代币数量: %d/%d\n", holdingCount, len(tokens))
		if totalValue > 0 {
			fmt.Printf("  估算稳定币价值: $%.2f\n", totalValue)
		}

		// 持有多样性分析
		diversityPercent := float64(holdingCount) / float64(len(tokens)) * 100
		var diversityLevel string
		switch {
		case diversityPercent == 0:
			diversityLevel = "🚫 无持有"
		case diversityPercent < 30:
			diversityLevel = "🔴 低多样性"
		case diversityPercent < 70:
			diversityLevel = "🟡 中等多样性"
		default:
			diversityLevel = "🟢 高多样性"
		}

		fmt.Printf("  多样性: %s (%.1f%%)\n", diversityLevel, diversityPercent)
	}

	// 代币流行度分析
	fmt.Printf("\n📈 代币流行度分析:\n")
	for _, token := range tokens {
		holdersCount := 0

		for _, addr := range addresses {
			balance, err := getTokenBalance(ctx, ethClient, token.Address, addr.address)
			if err != nil {
				continue
			}

			if balance.Sign() > 0 {
				holdersCount++
			}
		}

		popularityPercent := float64(holdersCount) / float64(len(addresses)) * 100
		fmt.Printf("  %s: %d/%d 地址持有 (%.1f%%)\n",
			token.Symbol, holdersCount, len(addresses), popularityPercent)
	}
}
