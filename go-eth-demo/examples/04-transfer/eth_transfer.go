package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
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

	fmt.Println("💸 以太坊 ETH 转账演示")
	fmt.Println("================================")

	// 检查是否配置了私钥
	if !cfg.HasPrivateKey() {
		fmt.Println("⚠️  未配置私钥，将演示转账流程但不会实际发送交易")
		fmt.Println("如需实际发送交易，请在 .env 文件中配置 PRIVATE_KEY")

		// 演示转账流程
		demonstrateTransferProcess(ctx, ethClient)
		return
	}

	fmt.Println("🔑 检测到私钥配置，准备进行实际转账演示...")

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

	// 1. 检查账户余额
	fmt.Println("\n💰 检查账户余额:")
	fmt.Println("--------------------------------")

	balance, err := ethClient.GetClient().BalanceAt(ctx, fromAddress, nil)
	if err != nil {
		log.Fatalf("查询余额失败: %v", err)
	}

	fmt.Printf("当前余额: %s ETH\n", utils.WeiToEther(balance))

	// 检查余额是否足够
	minBalance := big.NewInt(1000000000000000) // 0.001 ETH
	if balance.Cmp(minBalance) < 0 {
		fmt.Printf("⚠️  余额不足，需要至少 %s ETH 进行转账演示\n", utils.WeiToEther(minBalance))
		fmt.Println("请先获取一些测试 ETH 到您的地址")
		return
	}

	// 2. 准备转账参数
	fmt.Println("\n📋 准备转账参数:")
	fmt.Println("--------------------------------")

	// 接收方地址 (使用一个测试地址)
	toAddress := common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6")

	// 转账金额 (0.0001 ETH)
	transferAmount := big.NewInt(100000000000000) // 0.0001 ETH in Wei

	fmt.Printf("接收方地址: %s\n", toAddress.Hex())
	fmt.Printf("转账金额: %s ETH\n", utils.WeiToEther(transferAmount))

	// 3. 获取 Gas 价格和 Nonce
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

	gasLimit := uint64(21000) // ETH 转账的标准 Gas 限制

	fmt.Printf("Gas 价格: %s Gwei\n", utils.WeiToGwei(gasPrice))
	fmt.Printf("Gas 限制: %s\n", utils.FormatNumber(gasLimit))
	fmt.Printf("Nonce: %d\n", nonce)

	// 计算交易费用
	txFee := new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimit)))
	fmt.Printf("预估交易费用: %s ETH\n", utils.WeiToEther(txFee))

	// 计算总成本
	totalCost := new(big.Int).Add(transferAmount, txFee)
	fmt.Printf("总成本: %s ETH\n", utils.WeiToEther(totalCost))

	// 检查余额是否足够支付总成本
	if balance.Cmp(totalCost) < 0 {
		fmt.Printf("❌ 余额不足支付总成本\n")
		fmt.Printf("需要: %s ETH，当前: %s ETH\n",
			utils.WeiToEther(totalCost), utils.WeiToEther(balance))
		return
	}

	// 4. 创建交易
	fmt.Println("\n📝 创建交易:")
	fmt.Println("--------------------------------")

	tx := types.NewTransaction(nonce, toAddress, transferAmount, gasLimit, gasPrice, nil)

	// 获取链 ID
	chainID, err := ethClient.GetClient().ChainID(ctx)
	if err != nil {
		log.Fatalf("获取链 ID 失败: %v", err)
	}

	fmt.Printf("链 ID: %s\n", chainID.String())
	fmt.Printf("交易哈希 (未签名): %s\n", tx.Hash().Hex())

	// 5. 签名交易
	fmt.Println("\n✍️ 签名交易:")
	fmt.Println("--------------------------------")

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatalf("签名交易失败: %v", err)
	}

	fmt.Printf("签名后交易哈希: %s\n", signedTx.Hash().Hex())

	// 6. 发送交易前的最终确认
	fmt.Println("\n🚨 交易确认:")
	fmt.Println("================================")
	fmt.Printf("发送方: %s\n", fromAddress.Hex())
	fmt.Printf("接收方: %s\n", toAddress.Hex())
	fmt.Printf("金额: %s ETH\n", utils.WeiToEther(transferAmount))
	fmt.Printf("Gas 费用: %s ETH\n", utils.WeiToEther(txFee))
	fmt.Printf("总成本: %s ETH\n", utils.WeiToEther(totalCost))
	fmt.Printf("交易哈希: %s\n", signedTx.Hash().Hex())

	fmt.Println("\n⚠️  这是测试网交易，但仍会消耗真实的测试 ETH")
	fmt.Println("请确认您要继续发送此交易...")

	// 在实际应用中，这里应该有用户确认步骤
	// 为了演示，我们直接继续

	// 7. 发送交易
	fmt.Println("\n🚀 发送交易:")
	fmt.Println("--------------------------------")

	err = ethClient.GetClient().SendTransaction(ctx, signedTx)
	if err != nil {
		log.Fatalf("发送交易失败: %v", err)
	}

	fmt.Printf("✅ 交易已发送！\n")
	fmt.Printf("交易哈希: %s\n", signedTx.Hash().Hex())
	fmt.Printf("区块浏览器链接: https://sepolia.etherscan.io/tx/%s\n", signedTx.Hash().Hex())

	// 8. 等待交易确认
	fmt.Println("\n⏳ 等待交易确认:")
	fmt.Println("--------------------------------")

	receipt, err := waitForTransactionReceipt(ctx, ethClient, signedTx.Hash())
	if err != nil {
		fmt.Printf("❌ 等待交易确认失败: %v\n", err)
		return
	}

	// 9. 显示交易结果
	displayTransactionResult(receipt, signedTx)

	// 10. 检查余额变化
	fmt.Println("\n💰 检查余额变化:")
	fmt.Println("--------------------------------")

	newBalance, err := ethClient.GetClient().BalanceAt(ctx, fromAddress, nil)
	if err != nil {
		fmt.Printf("❌ 查询新余额失败: %v\n", err)
		return
	}

	balanceChange := new(big.Int).Sub(balance, newBalance)

	fmt.Printf("交易前余额: %s ETH\n", utils.WeiToEther(balance))
	fmt.Printf("交易后余额: %s ETH\n", utils.WeiToEther(newBalance))
	fmt.Printf("余额变化: -%s ETH\n", utils.WeiToEther(balanceChange))

	fmt.Println("\n✅ ETH 转账演示完成！")
}

// demonstrateTransferProcess 演示转账流程（不实际发送）
func demonstrateTransferProcess(ctx context.Context, ethClient *utils.EthClient) {
	fmt.Println("\n📚 ETH 转账流程演示:")
	fmt.Println("================================")

	// 模拟参数
	fromAddress := common.HexToAddress("0x1234567890123456789012345678901234567890")
	toAddress := common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6")
	transferAmount := big.NewInt(100000000000000) // 0.0001 ETH

	fmt.Printf("1. 准备转账参数:\n")
	fmt.Printf("   发送方: %s\n", fromAddress.Hex())
	fmt.Printf("   接收方: %s\n", toAddress.Hex())
	fmt.Printf("   金额: %s ETH\n", utils.WeiToEther(transferAmount))

	fmt.Printf("\n2. 获取网络信息:\n")

	// 获取当前 Gas 价格
	gasPrice, err := ethClient.GetClient().SuggestGasPrice(ctx)
	if err != nil {
		fmt.Printf("   ❌ 获取 Gas 价格失败: %v\n", err)
	} else {
		fmt.Printf("   当前 Gas 价格: %s Gwei\n", utils.WeiToGwei(gasPrice))
	}

	// 获取链 ID
	chainID, err := ethClient.GetClient().ChainID(ctx)
	if err != nil {
		fmt.Printf("   ❌ 获取链 ID 失败: %v\n", err)
	} else {
		fmt.Printf("   链 ID: %s\n", chainID.String())
	}

	gasLimit := uint64(21000)
	fmt.Printf("   Gas 限制: %s\n", utils.FormatNumber(gasLimit))

	// 计算费用
	if gasPrice != nil {
		txFee := new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimit)))
		totalCost := new(big.Int).Add(transferAmount, txFee)

		fmt.Printf("\n3. 费用计算:\n")
		fmt.Printf("   转账金额: %s ETH\n", utils.WeiToEther(transferAmount))
		fmt.Printf("   交易费用: %s ETH\n", utils.WeiToEther(txFee))
		fmt.Printf("   总成本: %s ETH\n", utils.WeiToEther(totalCost))
	}

	fmt.Printf("\n4. 交易流程:\n")
	fmt.Printf("   ✓ 创建交易对象\n")
	fmt.Printf("   ✓ 使用私钥签名\n")
	fmt.Printf("   ✓ 广播到网络\n")
	fmt.Printf("   ✓ 等待矿工打包\n")
	fmt.Printf("   ✓ 获取交易收据\n")

	fmt.Printf("\n💡 要进行实际转账，请:\n")
	fmt.Printf("   1. 在 .env 文件中配置 PRIVATE_KEY\n")
	fmt.Printf("   2. 确保账户有足够的测试 ETH\n")
	fmt.Printf("   3. 重新运行程序\n")
}

// waitForTransactionReceipt 等待交易确认
func waitForTransactionReceipt(ctx context.Context, ethClient *utils.EthClient, txHash common.Hash) (*types.Receipt, error) {
	fmt.Printf("等待交易 %s 确认...\n", txHash.Hex())

	// 在实际应用中，应该设置超时和重试机制
	for i := 0; i < 60; i++ { // 最多等待 60 次，每次间隔可以根据网络调整
		receipt, err := ethClient.GetClient().TransactionReceipt(ctx, txHash)
		if err == nil {
			fmt.Printf("✅ 交易已确认！(尝试 %d 次)\n", i+1)
			return receipt, nil
		}

		// 如果是 "not found" 错误，继续等待
		if err.Error() == "not found" {
			fmt.Printf("⏳ 等待确认... (尝试 %d/60)\n", i+1)
			// 在实际应用中应该使用 time.Sleep 或更好的等待机制
			continue
		}

		// 其他错误直接返回
		return nil, err
	}

	return nil, fmt.Errorf("交易确认超时")
}

// displayTransactionResult 显示交易结果
func displayTransactionResult(receipt *types.Receipt, tx *types.Transaction) {
	fmt.Println("📋 交易结果:")
	fmt.Println("--------------------------------")

	if receipt.Status == types.ReceiptStatusSuccessful {
		fmt.Printf("✅ 交易执行成功\n")
	} else {
		fmt.Printf("❌ 交易执行失败\n")
	}

	fmt.Printf("区块号: %d\n", receipt.BlockNumber.Uint64())
	fmt.Printf("区块哈希: %s\n", receipt.BlockHash.Hex())
	fmt.Printf("交易索引: %d\n", receipt.TransactionIndex)
	fmt.Printf("Gas 使用: %s\n", utils.FormatNumber(receipt.GasUsed))

	// 计算实际费用
	gasPrice := tx.GasPrice()
	actualFee := new(big.Int).Mul(big.NewInt(int64(receipt.GasUsed)), gasPrice)
	fmt.Printf("实际费用: %s ETH\n", utils.WeiToEther(actualFee))

	// Gas 使用效率
	gasLimit := tx.Gas()
	efficiency := float64(receipt.GasUsed) / float64(gasLimit) * 100
	fmt.Printf("Gas 使用效率: %.2f%%\n", efficiency)
}
