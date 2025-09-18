package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Printf("警告: 无法加载 .env 文件: %v", err)
	}

	fmt.Println("🔍 智能合约部署准备检查")
	fmt.Println("========================")

	// 连接以太坊节点
	rpcURL := os.Getenv("ETHEREUM_RPC_URL")
	if rpcURL == "" {
		fmt.Println("❌ 请在 .env 文件中设置 ETHEREUM_RPC_URL")
		return
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		fmt.Printf("❌ 连接以太坊节点失败: %v\n", err)
		return
	}
	defer client.Close()

	fmt.Printf("✅ RPC连接成功\n")

	// 获取网络信息
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		fmt.Printf("❌ 获取网络ID失败: %v\n", err)
		return
	}

	var networkName string
	switch chainID.String() {
	case "11155111":
		networkName = "Sepolia Testnet"
	case "1":
		networkName = "Ethereum Mainnet"
	default:
		networkName = "Unknown Network"
	}

	fmt.Printf("✅ 网络: %s (ID: %s)\n", networkName, chainID.String())

	// 检查私钥
	privateKeyHex := os.Getenv("PRIVATE_KEY")
	if privateKeyHex == "" {
		fmt.Println("❌ 请在 .env 文件中设置 PRIVATE_KEY")
		return
	}

	if strings.HasPrefix(privateKeyHex, "0x") {
		privateKeyHex = privateKeyHex[2:]
	}

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		fmt.Printf("❌ 解析私钥失败: %v\n", err)
		return
	}

	// 获取地址
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		fmt.Println("❌ 获取公钥失败")
		return
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	fmt.Printf("✅ 部署地址: %s\n", fromAddress.Hex())

	// 检查余额
	balance, err := client.BalanceAt(context.Background(), fromAddress, nil)
	if err != nil {
		fmt.Printf("❌ 获取余额失败: %v\n", err)
		return
	}

	balanceEth := new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(1e18))
	fmt.Printf("💰 当前余额: %s ETH\n", balanceEth.Text('f', 6))

	// 检查余额状态
	minBalance := big.NewFloat(0.01)
	if balanceEth.Cmp(minBalance) >= 0 {
		fmt.Println("✅ 余额充足，可以部署合约!")

		// 估算Gas成本
		gasPrice, err := client.SuggestGasPrice(context.Background())
		if err == nil {
			estimatedGas := uint64(500000)
			estimatedCost := new(big.Int).Mul(big.NewInt(int64(estimatedGas)), gasPrice)
			estimatedCostEth := new(big.Float).Quo(new(big.Float).SetInt(estimatedCost), big.NewFloat(1e18))

			fmt.Printf("⛽ Gas价格: %s Gwei\n",
				new(big.Float).Quo(new(big.Float).SetInt(gasPrice), big.NewFloat(1e9)).Text('f', 2))
			fmt.Printf("💸 预估成本: %s ETH\n", estimatedCostEth.Text('f', 6))
		}

		fmt.Println("\n🚀 可以运行部署命令:")
		fmt.Println("   go run examples/08-deploy/deploy_simple_storage.go")

	} else {
		fmt.Println("❌ 余额不足，需要获取测试网ETH")
		needed := new(big.Float).Sub(minBalance, balanceEth)
		fmt.Printf("   还需要: %s ETH\n", needed.Text('f', 6))
		fmt.Println("\n🚰 获取测试网ETH:")
		fmt.Println("   1. 访问: https://www.alchemy.com/faucets/ethereum-sepolia")
		fmt.Printf("   2. 输入地址: %s\n", fromAddress.Hex())
		fmt.Println("   3. 申请测试网ETH")
		fmt.Println("   4. 等待到账后重新检查")
	}

	// 检查合约文件
	fmt.Println("\n📄 检查合约编译文件:")
	if _, err := os.Stat("build/SimpleStorage.json"); err == nil {
		fmt.Println("✅ SimpleStorage.json 存在")
	} else {
		fmt.Println("❌ SimpleStorage.json 不存在")
		fmt.Println("   请运行: node scripts/compile_contracts.js")
	}

	fmt.Println("\n📊 检查完成!")
}
