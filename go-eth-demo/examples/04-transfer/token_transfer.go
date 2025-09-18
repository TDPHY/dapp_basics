package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/local/go-eth-demo/config"
	"github.com/local/go-eth-demo/utils"
)

// ERC20 ABI 定义 (包含转账相关方法)
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
		"constant": false,
		"inputs": [
			{"name": "_to", "type": "address"},
			{"name": "_value", "type": "uint256"}
		],
		"name": "transfer",
		"outputs": [{"name": "", "type": "bool"}],
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [
			{"name": "_owner", "type": "address"},
			{"name": "_spender", "type": "address"}
		],
		"name": "allowance",
		"outputs": [{"name": "", "type": "uint256"}],
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{"name": "_spender", "type": "address"},
			{"name": "_value", "type": "uint256"}
		],
		"name": "approve",
		"outputs": [{"name": "", "type": "bool"}],
		"type": "function"
	},
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

// TokenInfo 代币信息
type TokenInfo struct {
	Address  common.Address
	Name     string
	Symbol   string
	Decimals uint8
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

	fmt.Println("🪙 ERC-20 代币转账演示")
	fmt.Println("================================")

	// 检查是否配置了私钥
	if !cfg.HasPrivateKey() {
		fmt.Println("⚠️  未配置私钥，将演示代币转账流程但不会实际发送交易")
		fmt.Println("如需实际发送交易，请在 .env 文件中配置 PRIVATE_KEY")

		// 演示代币转账流程
		demonstrateTokenTransferProcess(ctx, ethClient)
		return
	}

	fmt.Println("🔑 检测到私钥配置，准备进行实际代币转账演示...")

	// 解析私钥
	privateKey, err := crypto.HexToECDSA(cfg.PrivateKey)
	if err != nil {
		log.Fatalf("解析私钥失败: %v", err)
	}

	// 获取发送方地址
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatalf("无法获取公钥")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	fmt.Printf("发送方地址: %s\n", fromAddress.Hex())

	// 代币合约地址 (Sepolia USDC)
	tokenAddress := common.HexToAddress("0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238")

	// 1. 获取代币信息
	fmt.Println("\n📋 获取代币信息:")
	fmt.Println("--------------------------------")

	tokenInfo, err := getTokenInfo(ctx, ethClient, tokenAddress)
	if err != nil {
		log.Fatalf("获取代币信息失败: %v", err)
	}

	fmt.Printf("代币名称: %s\n", tokenInfo.Name)
	fmt.Printf("代币符号: %s\n", tokenInfo.Symbol)
	fmt.Printf("小数位数: %d\n", tokenInfo.Decimals)
	fmt.Printf("合约地址: %s\n", tokenInfo.Address.Hex())

	// 2. 检查代币余额
	fmt.Println("\n💰 检查代币余额:")
	fmt.Println("--------------------------------")

	balance, err := getTokenBalance(ctx, ethClient, tokenAddress, fromAddress)
	if err != nil {
		log.Fatalf("查询代币余额失败: %v", err)
	}

	balanceFormatted := formatTokenBalance(balance, tokenInfo.Decimals)
	fmt.Printf("当前 %s 余额: %s\n", tokenInfo.Symbol, balanceFormatted)

	// 检查余额是否足够
	minBalance := big.NewInt(1000000) // 1 USDC (6 decimals)
	if balance.Cmp(minBalance) < 0 {
		fmt.Printf("⚠️  %s 余额不足，需要至少 %s 进行转账演示\n",
			tokenInfo.Symbol, formatTokenBalance(minBalance, tokenInfo.Decimals))
		fmt.Printf("请先获取一些测试 %s 到您的地址\n", tokenInfo.Symbol)
		return
	}

	// 3. 检查 ETH 余额 (用于支付 Gas)
	fmt.Println("\n⛽ 检查 ETH 余额 (Gas 费用):")
	fmt.Println("--------------------------------")

	ethBalance, err := ethClient.GetClient().BalanceAt(ctx, fromAddress, nil)
	if err != nil {
		log.Fatalf("查询 ETH 余额失败: %v", err)
	}

	fmt.Printf("ETH 余额: %s ETH\n", utils.WeiToEther(ethBalance))

	// 检查 ETH 余额是否足够支付 Gas
	minETHBalance := big.NewInt(1000000000000000) // 0.001 ETH
	if ethBalance.Cmp(minETHBalance) < 0 {
		fmt.Printf("⚠️  ETH 余额不足支付 Gas 费用，需要至少 %s ETH\n",
			utils.WeiToEther(minETHBalance))
		return
	}

	// 4. 准备转账参数
	fmt.Println("\n📋 准备转账参数:")
	fmt.Println("--------------------------------")

	// 接收方地址
	toAddress := common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6")

	// 转账金额 (0.1 USDC)
	transferAmount := big.NewInt(100000) // 0.1 USDC (6 decimals)

	fmt.Printf("接收方地址: %s\n", toAddress.Hex())
	fmt.Printf("转账金额: %s %s\n",
		formatTokenBalance(transferAmount, tokenInfo.Decimals), tokenInfo.Symbol)

	// 5. 获取 Gas 信息
	fmt.Println("\n⛽ 获取 Gas 信息:")
	fmt.Println("--------------------------------")

	gasPrice, err := ethClient.GetClient().SuggestGasPrice(ctx)
	if err != nil {
		log.Fatalf("获取 Gas 价格失败: %v", err)
	}

	nonce, err := ethClient.GetClient().PendingNonceAt(ctx, fromAddress)
	if err != nil {
		log.Fatalf("获取 Nonce 失败: %v", err)
	}

	fmt.Printf("Gas 价格: %s Gwei\n", utils.WeiToGwei(gasPrice))
	fmt.Printf("Nonce: %d\n", nonce)

	// 6. 估算 Gas 限制
	fmt.Println("\n📊 估算 Gas 限制:")
	fmt.Println("--------------------------------")

	gasLimit, err := estimateTokenTransferGas(ctx, ethClient, tokenAddress, fromAddress, toAddress, transferAmount)
	if err != nil {
		fmt.Printf("⚠️  Gas 估算失败，使用默认值: %v\n", err)
		gasLimit = 60000 // ERC20 转账的典型 Gas 限制
	}

	fmt.Printf("估算 Gas 限制: %s\n", utils.FormatNumber(gasLimit))

	// 计算交易费用
	txFee := new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimit)))
	fmt.Printf("预估交易费用: %s ETH\n", utils.WeiToEther(txFee))

	// 检查 ETH 余额是否足够支付费用
	if ethBalance.Cmp(txFee) < 0 {
		fmt.Printf("❌ ETH 余额不足支付交易费用\n")
		fmt.Printf("需要: %s ETH，当前: %s ETH\n",
			utils.WeiToEther(txFee), utils.WeiToEther(ethBalance))
		return
	}

	// 7. 创建代币转账交易
	fmt.Println("\n📝 创建代币转账交易:")
	fmt.Println("--------------------------------")

	// 解析 ABI
	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		log.Fatalf("解析 ABI 失败: %v", err)
	}

	// 编码 transfer 方法调用
	data, err := parsedABI.Pack("transfer", toAddress, transferAmount)
	if err != nil {
		log.Fatalf("编码 transfer 调用失败: %v", err)
	}

	// 创建交易
	tx := types.NewTransaction(nonce, tokenAddress, big.NewInt(0), gasLimit, gasPrice, data)

	// 获取链 ID
	chainID, err := ethClient.GetClient().ChainID(ctx)
	if err != nil {
		log.Fatalf("获取链 ID 失败: %v", err)
	}

	fmt.Printf("链 ID: %s\n", chainID.String())
	fmt.Printf("交易哈希 (未签名): %s\n", tx.Hash().Hex())

	// 8. 签名交易
	fmt.Println("\n✍️ 签名交易:")
	fmt.Println("--------------------------------")

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatalf("签名交易失败: %v", err)
	}

	fmt.Printf("签名后交易哈希: %s\n", signedTx.Hash().Hex())

	// 9. 交易确认
	fmt.Println("\n🚨 交易确认:")
	fmt.Println("================================")
	fmt.Printf("代币合约: %s (%s)\n", tokenInfo.Address.Hex(), tokenInfo.Symbol)
	fmt.Printf("发送方: %s\n", fromAddress.Hex())
	fmt.Printf("接收方: %s\n", toAddress.Hex())
	fmt.Printf("转账金额: %s %s\n",
		formatTokenBalance(transferAmount, tokenInfo.Decimals), tokenInfo.Symbol)
	fmt.Printf("Gas 费用: %s ETH\n", utils.WeiToEther(txFee))
	fmt.Printf("交易哈希: %s\n", signedTx.Hash().Hex())

	fmt.Println("\n⚠️  这是测试网交易，但仍会消耗真实的测试代币和 ETH")
	fmt.Println("请确认您要继续发送此交易...")

	// 10. 发送交易
	fmt.Println("\n🚀 发送交易:")
	fmt.Println("--------------------------------")

	err = ethClient.GetClient().SendTransaction(ctx, signedTx)
	if err != nil {
		log.Fatalf("发送交易失败: %v", err)
	}

	fmt.Printf("✅ 交易已发送！\n")
	fmt.Printf("交易哈希: %s\n", signedTx.Hash().Hex())
	fmt.Printf("区块浏览器链接: https://sepolia.etherscan.io/tx/%s\n", signedTx.Hash().Hex())

	// 11. 等待交易确认
	fmt.Println("\n⏳ 等待交易确认:")
	fmt.Println("--------------------------------")

	receipt, err := waitForTransactionReceipt(ctx, ethClient, signedTx.Hash())
	if err != nil {
		fmt.Printf("❌ 等待交易确认失败: %v\n", err)
		return
	}

	// 12. 显示交易结果
	displayTokenTransferResult(receipt, signedTx, tokenInfo)

	// 13. 检查余额变化
	fmt.Println("\n💰 检查余额变化:")
	fmt.Println("--------------------------------")

	newBalance, err := getTokenBalance(ctx, ethClient, tokenAddress, fromAddress)
	if err != nil {
		fmt.Printf("❌ 查询新余额失败: %v\n", err)
		return
	}

	balanceChange := new(big.Int).Sub(balance, newBalance)

	fmt.Printf("交易前 %s 余额: %s\n", tokenInfo.Symbol,
		formatTokenBalance(balance, tokenInfo.Decimals))
	fmt.Printf("交易后 %s 余额: %s\n", tokenInfo.Symbol,
		formatTokenBalance(newBalance, tokenInfo.Decimals))
	fmt.Printf("%s 余额变化: -%s\n", tokenInfo.Symbol,
		formatTokenBalance(balanceChange, tokenInfo.Decimals))

	fmt.Println("\n✅ ERC-20 代币转账演示完成！")
}

// getTokenInfo 获取代币信息
func getTokenInfo(ctx context.Context, ethClient *utils.EthClient, tokenAddress common.Address) (*TokenInfo, error) {
	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return nil, fmt.Errorf("解析 ABI 失败: %w", err)
	}

	tokenInfo := &TokenInfo{Address: tokenAddress}

	// 获取代币名称
	if name, err := callContractMethod(ctx, ethClient, tokenAddress, parsedABI, "name"); err == nil {
		if len(name) > 0 {
			tokenInfo.Name = name[0].(string)
		}
	}

	// 获取代币符号
	if symbol, err := callContractMethod(ctx, ethClient, tokenAddress, parsedABI, "symbol"); err == nil {
		if len(symbol) > 0 {
			tokenInfo.Symbol = symbol[0].(string)
		}
	}

	// 获取小数位数
	if decimals, err := callContractMethod(ctx, ethClient, tokenAddress, parsedABI, "decimals"); err == nil {
		if len(decimals) > 0 {
			tokenInfo.Decimals = decimals[0].(uint8)
		}
	}

	return tokenInfo, nil
}

// getTokenBalance 获取代币余额
func getTokenBalance(ctx context.Context, ethClient *utils.EthClient, tokenAddress, userAddress common.Address) (*big.Int, error) {
	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return nil, fmt.Errorf("解析 ABI 失败: %w", err)
	}

	result, err := callContractMethod(ctx, ethClient, tokenAddress, parsedABI, "balanceOf", userAddress)
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
	data, err := parsedABI.Pack(methodName, args...)
	if err != nil {
		return nil, fmt.Errorf("编码方法调用失败: %w", err)
	}

	msg := ethereum.CallMsg{
		To:   &contractAddress,
		Data: data,
	}

	result, err := ethClient.GetClient().CallContract(ctx, msg, nil)
	if err != nil {
		return nil, fmt.Errorf("调用合约失败: %w", err)
	}

	values, err := parsedABI.Unpack(methodName, result)
	if err != nil {
		return nil, fmt.Errorf("解码结果失败: %w", err)
	}

	return values, nil
}

// estimateTokenTransferGas 估算代币转账 Gas
func estimateTokenTransferGas(ctx context.Context, ethClient *utils.EthClient, tokenAddress, from, to common.Address, amount *big.Int) (uint64, error) {
	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return 0, fmt.Errorf("解析 ABI 失败: %w", err)
	}

	data, err := parsedABI.Pack("transfer", to, amount)
	if err != nil {
		return 0, fmt.Errorf("编码 transfer 调用失败: %w", err)
	}

	msg := ethereum.CallMsg{
		From: from,
		To:   &tokenAddress,
		Data: data,
	}

	gasLimit, err := ethClient.GetClient().EstimateGas(ctx, msg)
	if err != nil {
		return 0, fmt.Errorf("估算 Gas 失败: %w", err)
	}

	// 添加一些缓冲
	return gasLimit + 10000, nil
}

// formatTokenBalance 格式化代币余额
func formatTokenBalance(balance *big.Int, decimals uint8) string {
	if balance.Sign() == 0 {
		return "0"
	}

	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	balanceFloat := new(big.Float).SetInt(balance)
	divisorFloat := new(big.Float).SetInt(divisor)

	result := new(big.Float).Quo(balanceFloat, divisorFloat)
	return result.Text('f', 6)
}

// demonstrateTokenTransferProcess 演示代币转账流程
func demonstrateTokenTransferProcess(ctx context.Context, ethClient *utils.EthClient) {
	fmt.Println("\n📚 ERC-20 代币转账流程演示:")
	fmt.Println("================================")

	// 模拟参数
	tokenAddress := common.HexToAddress("0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238")
	fromAddress := common.HexToAddress("0x1234567890123456789012345678901234567890")
	toAddress := common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6")

	fmt.Printf("1. 代币转账与 ETH 转账的区别:\n")
	fmt.Printf("   ETH 转账: 直接在交易中指定接收方和金额\n")
	fmt.Printf("   代币转账: 调用代币合约的 transfer 方法\n")

	fmt.Printf("\n2. 代币转账流程:\n")
	fmt.Printf("   ✓ 获取代币合约信息 (名称、符号、小数位)\n")
	fmt.Printf("   ✓ 检查发送方代币余额\n")
	fmt.Printf("   ✓ 检查发送方 ETH 余额 (支付 Gas)\n")
	fmt.Printf("   ✓ 编码 transfer(to, amount) 方法调用\n")
	fmt.Printf("   ✓ 创建交易 (to=合约地址, value=0, data=方法调用)\n")
	fmt.Printf("   ✓ 签名并发送交易\n")
	fmt.Printf("   ✓ 等待确认并检查 Transfer 事件\n")

	// 获取实际的代币信息进行演示
	tokenInfo, err := getTokenInfo(ctx, ethClient, tokenAddress)
	if err != nil {
		fmt.Printf("   ❌ 获取代币信息失败: %v\n", err)
		return
	}

	fmt.Printf("\n3. 示例代币信息:\n")
	fmt.Printf("   合约地址: %s\n", tokenAddress.Hex())
	fmt.Printf("   代币名称: %s\n", tokenInfo.Name)
	fmt.Printf("   代币符号: %s\n", tokenInfo.Symbol)
	fmt.Printf("   小数位数: %d\n", tokenInfo.Decimals)

	fmt.Printf("\n4. 转账参数示例:\n")
	fmt.Printf("   发送方: %s\n", fromAddress.Hex())
	fmt.Printf("   接收方: %s\n", toAddress.Hex())
	fmt.Printf("   金额: 0.1 %s\n", tokenInfo.Symbol)

	// 获取当前 Gas 价格
	gasPrice, err := ethClient.GetClient().SuggestGasPrice(ctx)
	if err == nil {
		gasLimit := uint64(60000) // 典型的 ERC20 转账 Gas 限制
		txFee := new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimit)))

		fmt.Printf("\n5. Gas 费用估算:\n")
		fmt.Printf("   Gas 价格: %s Gwei\n", utils.WeiToGwei(gasPrice))
		fmt.Printf("   Gas 限制: %s\n", utils.FormatNumber(gasLimit))
		fmt.Printf("   预估费用: %s ETH\n", utils.WeiToEther(txFee))
	}

	fmt.Printf("\n💡 要进行实际代币转账，请:\n")
	fmt.Printf("   1. 在 .env 文件中配置 PRIVATE_KEY\n")
	fmt.Printf("   2. 确保账户有足够的代币和 ETH (支付 Gas)\n")
	fmt.Printf("   3. 重新运行程序\n")
}

// waitForTransactionReceipt 等待交易确认
func waitForTransactionReceipt(ctx context.Context, ethClient *utils.EthClient, txHash common.Hash) (*types.Receipt, error) {
	fmt.Printf("等待交易 %s 确认...\n", txHash.Hex())

	for i := 0; i < 60; i++ {
		receipt, err := ethClient.GetClient().TransactionReceipt(ctx, txHash)
		if err == nil {
			fmt.Printf("✅ 交易已确认！(尝试 %d 次)\n", i+1)
			return receipt, nil
		}

		if err.Error() == "not found" {
			fmt.Printf("⏳ 等待确认... (尝试 %d/60)\n", i+1)
			continue
		}

		return nil, err
	}

	return nil, fmt.Errorf("交易确认超时")
}

// displayTokenTransferResult 显示代币转账结果
func displayTokenTransferResult(receipt *types.Receipt, tx *types.Transaction, tokenInfo *TokenInfo) {
	fmt.Println("📋 代币转账结果:")
	fmt.Println("--------------------------------")

	if receipt.Status == types.ReceiptStatusSuccessful {
		fmt.Printf("✅ 代币转账成功\n")
	} else {
		fmt.Printf("❌ 代币转账失败\n")
	}

	fmt.Printf("区块号: %d\n", receipt.BlockNumber.Uint64())
	fmt.Printf("区块哈希: %s\n", receipt.BlockHash.Hex())
	fmt.Printf("交易索引: %d\n", receipt.TransactionIndex)
	fmt.Printf("Gas 使用: %s\n", utils.FormatNumber(receipt.GasUsed))

	// 计算实际费用
	gasPrice := tx.GasPrice()
	actualFee := new(big.Int).Mul(big.NewInt(int64(receipt.GasUsed)), gasPrice)
	fmt.Printf("实际费用: %s ETH\n", utils.WeiToEther(actualFee))

	// 分析事件日志
	fmt.Printf("事件日志数量: %d\n", len(receipt.Logs))

	// 查找 Transfer 事件
	transferEventSignature := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

	for i, log := range receipt.Logs {
		if len(log.Topics) > 0 && log.Topics[0] == transferEventSignature {
			fmt.Printf("🎉 发现 Transfer 事件 (日志 #%d):\n", i+1)

			if len(log.Topics) >= 3 {
				from := common.HexToAddress(log.Topics[1].Hex())
				to := common.HexToAddress(log.Topics[2].Hex())

				fmt.Printf("   发送方: %s\n", from.Hex())
				fmt.Printf("   接收方: %s\n", to.Hex())

				if len(log.Data) >= 32 {
					amount := new(big.Int).SetBytes(log.Data[:32])
					fmt.Printf("   金额: %s %s\n",
						formatTokenBalance(amount, tokenInfo.Decimals), tokenInfo.Symbol)
				}
			}
		}
	}
}
