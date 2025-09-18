package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

// 合约编译输出结构
type ContractData struct {
	ContractName string      `json:"contractName"`
	SourceFile   string      `json:"sourceFile"`
	ABI          interface{} `json:"abi"`
	Bytecode     string      `json:"bytecode"`
	CompiledAt   string      `json:"compiledAt"`
}

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Printf("警告: 无法加载 .env 文件: %v", err)
	}

	fmt.Println("🚀 开始部署SimpleStorage智能合约")
	fmt.Println("=====================================")

	// 连接以太坊节点
	client, err := ethclient.Dial(os.Getenv("ETHEREUM_RPC_URL"))
	if err != nil {
		log.Fatalf("连接以太坊节点失败: %v", err)
	}
	defer client.Close()

	// 获取私钥
	privateKeyHex := os.Getenv("PRIVATE_KEY")
	if strings.HasPrefix(privateKeyHex, "0x") {
		privateKeyHex = privateKeyHex[2:]
	}

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatalf("解析私钥失败: %v", err)
	}

	// 获取部署地址
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("获取公钥失败")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	fmt.Printf("📍 部署地址: %s\n", fromAddress.Hex())

	// 检查余额
	balance, err := client.BalanceAt(context.Background(), fromAddress, nil)
	if err != nil {
		log.Fatalf("获取余额失败: %v", err)
	}

	balanceEth := new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(1e18))
	fmt.Printf("💰 当前余额: %s ETH\n", balanceEth.Text('f', 6))

	// 检查余额是否充足
	minBalance := big.NewFloat(0.005)
	if balanceEth.Cmp(minBalance) < 0 {
		fmt.Printf("❌ 余额不足，至少需要 %s ETH\n", minBalance.Text('f', 3))
		fmt.Println("请先获取测试网ETH后再部署")
		return
	}

	// 加载合约数据
	contractData, contractABI, err := loadContractData()
	if err != nil {
		log.Fatalf("加载合约数据失败: %v", err)
	}

	fmt.Println("✅ 合约数据加载成功")

	// 获取网络ID
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatalf("获取网络ID失败: %v", err)
	}

	// 获取nonce
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatalf("获取nonce失败: %v", err)
	}

	// 获取gas价格
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf("获取gas价格失败: %v", err)
	}

	fmt.Printf("⛽ Gas价格: %s Gwei\n",
		new(big.Float).Quo(new(big.Float).SetInt(gasPrice), big.NewFloat(1e9)).Text('f', 2))

	// 准备构造函数参数（初始值设为42）
	initialValue := big.NewInt(42)
	input, err := contractABI.Pack("", initialValue)
	if err != nil {
		log.Fatalf("打包构造函数参数失败: %v", err)
	}

	// 准备部署数据
	bytecode := common.FromHex(contractData.Bytecode)
	deployData := append(bytecode, input...)

	// 估算gas
	gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
		From: fromAddress,
		Data: deployData,
	})
	if err != nil {
		log.Printf("警告: gas估算失败，使用默认值: %v", err)
		gasLimit = 500000
	}

	// 增加20%的gas缓冲
	gasLimit = gasLimit * 120 / 100

	fmt.Printf("⛽ Gas限制: %d\n", gasLimit)

	// 计算总成本
	totalCost := new(big.Int).Mul(big.NewInt(int64(gasLimit)), gasPrice)
	totalCostEth := new(big.Float).Quo(new(big.Float).SetInt(totalCost), big.NewFloat(1e18))
	fmt.Printf("💸 预估成本: %s ETH\n", totalCostEth.Text('f', 6))

	// 创建部署交易
	tx := types.NewContractCreation(nonce, big.NewInt(0), gasLimit, gasPrice, deployData)

	// 签名交易
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatalf("签名交易失败: %v", err)
	}

	fmt.Println("📝 交易已签名，开始发送...")

	// 发送交易
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatalf("发送交易失败: %v", err)
	}

	txHash := signedTx.Hash().Hex()
	fmt.Printf("✅ 交易已发送!\n")
	fmt.Printf("📋 交易哈希: %s\n", txHash)
	fmt.Printf("🔗 查看交易: https://sepolia.etherscan.io/tx/%s\n", txHash)

	// 等待交易确认
	fmt.Println("⏳ 等待交易确认...")
	receipt, err := waitForTransaction(client, signedTx.Hash())
	if err != nil {
		log.Fatalf("等待交易确认失败: %v", err)
	}

	if receipt.Status == 1 {
		fmt.Println("🎉 合约部署成功!")
		fmt.Printf("📍 合约地址: %s\n", receipt.ContractAddress.Hex())
		fmt.Printf("⛽ 实际Gas使用: %d\n", receipt.GasUsed)
		fmt.Printf("🔗 查看合约: https://sepolia.etherscan.io/address/%s\n", receipt.ContractAddress.Hex())

		// 保存合约地址到文件
		saveContractAddress(receipt.ContractAddress.Hex())

		fmt.Println("\n🔧 下一步可以运行交互程序:")
		fmt.Println("   go run examples/08-deploy/interact_contract.go")
	} else {
		fmt.Println("❌ 合约部署失败")
		fmt.Printf("交易状态: %d\n", receipt.Status)
	}
}

// 加载合约编译数据
func loadContractData() (*ContractData, *abi.ABI, error) {
	// 读取编译输出
	data, err := ioutil.ReadFile("build/SimpleStorage.json")
	if err != nil {
		return nil, nil, fmt.Errorf("读取合约文件失败: %v", err)
	}

	var contractData ContractData
	if err := json.Unmarshal(data, &contractData); err != nil {
		return nil, nil, fmt.Errorf("解析合约数据失败: %v", err)
	}

	// 将ABI转换为JSON字符串
	abiBytes, err := json.Marshal(contractData.ABI)
	if err != nil {
		return nil, nil, fmt.Errorf("序列化ABI失败: %v", err)
	}

	// 解析ABI
	contractABI, err := abi.JSON(strings.NewReader(string(abiBytes)))
	if err != nil {
		return nil, nil, fmt.Errorf("解析ABI失败: %v", err)
	}

	return &contractData, &contractABI, nil
}

// 等待交易确认
func waitForTransaction(client *ethclient.Client, txHash common.Hash) (*types.Receipt, error) {
	for i := 0; i < 60; i++ { // 最多等待5分钟
		receipt, err := client.TransactionReceipt(context.Background(), txHash)
		if err == nil {
			return receipt, nil
		}

		if i%10 == 0 {
			fmt.Printf("⏳ 等待确认中... (%d/60)\n", i+1)
		}
		time.Sleep(5 * time.Second)
	}
	return nil, fmt.Errorf("交易确认超时")
}

// 保存合约地址
func saveContractAddress(address string) {
	content := fmt.Sprintf("CONTRACT_ADDRESS=%s\n", address)
	err := ioutil.WriteFile("contract_address.env", []byte(content), 0644)
	if err != nil {
		log.Printf("保存合约地址失败: %v", err)
	} else {
		fmt.Printf("✅ 合约地址已保存到 contract_address.env\n")
	}
}
