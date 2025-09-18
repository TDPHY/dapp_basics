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

	// 尝试加载合约地址
	if err := godotenv.Load("contract_address.env"); err != nil {
		log.Printf("警告: 无法加载合约地址文件: %v", err)
	}

	fmt.Println("🔧 SimpleStorage合约交互演示")
	fmt.Println("==============================")

	// 连接以太坊节点
	client, err := ethclient.Dial(os.Getenv("ETHEREUM_RPC_URL"))
	if err != nil {
		log.Fatalf("连接以太坊节点失败: %v", err)
	}
	defer client.Close()

	// 获取合约地址
	contractAddressStr := os.Getenv("CONTRACT_ADDRESS")
	if contractAddressStr == "" {
		fmt.Println("❌ 请先部署合约或设置CONTRACT_ADDRESS环境变量")
		fmt.Println("   运行: go run examples/08-deploy/deploy_contract.go")
		return
	}

	contractAddress := common.HexToAddress(contractAddressStr)
	fmt.Printf("📍 合约地址: %s\n", contractAddress.Hex())

	// 加载合约ABI
	contractABI, err := loadContractABI()
	if err != nil {
		log.Fatalf("加载合约ABI失败: %v", err)
	}

	// 获取私钥和地址
	privateKey, fromAddress, err := getAccountInfo()
	if err != nil {
		log.Fatalf("获取账户信息失败: %v", err)
	}

	fmt.Printf("👤 操作地址: %s\n", fromAddress.Hex())

	// 演示合约交互
	fmt.Println("\n🔍 1. 读取当前存储的值")
	currentValue, err := readStoredValue(client, contractAddress, contractABI)
	if err != nil {
		log.Printf("读取失败: %v", err)
	} else {
		fmt.Printf("   当前值: %s\n", currentValue.String())
	}

	fmt.Println("\n📝 2. 存储新值 (100)")
	txHash, err := storeValue(client, contractAddress, contractABI, privateKey, fromAddress, big.NewInt(100))
	if err != nil {
		log.Printf("存储失败: %v", err)
	} else {
		fmt.Printf("   交易哈希: %s\n", txHash)
		fmt.Printf("   查看交易: https://sepolia.etherscan.io/tx/%s\n", txHash)
	}

	fmt.Println("\n🔍 3. 再次读取存储的值")
	time.Sleep(10 * time.Second) // 等待交易确认
	newValue, err := readStoredValue(client, contractAddress, contractABI)
	if err != nil {
		log.Printf("读取失败: %v", err)
	} else {
		fmt.Printf("   新值: %s\n", newValue.String())
	}

	fmt.Println("\n➕ 4. 增加值 (+50)")
	txHash2, err := incrementValue(client, contractAddress, contractABI, privateKey, fromAddress, big.NewInt(50))
	if err != nil {
		log.Printf("增加失败: %v", err)
	} else {
		fmt.Printf("   交易哈希: %s\n", txHash2)
		fmt.Printf("   查看交易: https://sepolia.etherscan.io/tx/%s\n", txHash2)
	}

	fmt.Println("\n🔍 5. 最终读取存储的值")
	time.Sleep(10 * time.Second) // 等待交易确认
	finalValue, err := readStoredValue(client, contractAddress, contractABI)
	if err != nil {
		log.Printf("读取失败: %v", err)
	} else {
		fmt.Printf("   最终值: %s\n", finalValue.String())
	}

	fmt.Println("\n✅ 合约交互演示完成!")
}

// 读取存储的值
func readStoredValue(client *ethclient.Client, contractAddress common.Address, contractABI *abi.ABI) (*big.Int, error) {
	// 调用retrieve函数
	data, err := contractABI.Pack("retrieve")
	if err != nil {
		return nil, fmt.Errorf("打包函数调用失败: %v", err)
	}

	// 执行调用
	result, err := client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &contractAddress,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("调用合约失败: %v", err)
	}

	// 解析结果
	var value *big.Int
	err = contractABI.UnpackIntoInterface(&value, "retrieve", result)
	if err != nil {
		return nil, fmt.Errorf("解析结果失败: %v", err)
	}

	return value, nil
}

// 存储值
func storeValue(client *ethclient.Client, contractAddress common.Address, contractABI *abi.ABI,
	privateKey *ecdsa.PrivateKey, fromAddress common.Address, value *big.Int) (string, error) {

	// 打包函数调用
	data, err := contractABI.Pack("store", value)
	if err != nil {
		return "", fmt.Errorf("打包函数调用失败: %v", err)
	}

	return sendTransaction(client, contractAddress, data, privateKey, fromAddress)
}

// 增加值
func incrementValue(client *ethclient.Client, contractAddress common.Address, contractABI *abi.ABI,
	privateKey *ecdsa.PrivateKey, fromAddress common.Address, increment *big.Int) (string, error) {

	// 打包函数调用
	data, err := contractABI.Pack("increment", increment)
	if err != nil {
		return "", fmt.Errorf("打包函数调用失败: %v", err)
	}

	return sendTransaction(client, contractAddress, data, privateKey, fromAddress)
}

// 发送交易
func sendTransaction(client *ethclient.Client, contractAddress common.Address, data []byte,
	privateKey *ecdsa.PrivateKey, fromAddress common.Address) (string, error) {

	// 获取nonce
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", fmt.Errorf("获取nonce失败: %v", err)
	}

	// 获取gas价格
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", fmt.Errorf("获取gas价格失败: %v", err)
	}

	// 估算gas
	gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
		From: fromAddress,
		To:   &contractAddress,
		Data: data,
	})
	if err != nil {
		gasLimit = 100000 // 使用默认值
	}

	// 获取网络ID
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return "", fmt.Errorf("获取网络ID失败: %v", err)
	}

	// 创建交易
	tx := types.NewTransaction(nonce, contractAddress, big.NewInt(0), gasLimit, gasPrice, data)

	// 签名交易
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return "", fmt.Errorf("签名交易失败: %v", err)
	}

	// 发送交易
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", fmt.Errorf("发送交易失败: %v", err)
	}

	return signedTx.Hash().Hex(), nil
}

// 加载合约ABI
func loadContractABI() (*abi.ABI, error) {
	data, err := ioutil.ReadFile("build/SimpleStorage.json")
	if err != nil {
		return nil, fmt.Errorf("读取合约文件失败: %v", err)
	}

	var contractData ContractData
	if err := json.Unmarshal(data, &contractData); err != nil {
		return nil, fmt.Errorf("解析合约数据失败: %v", err)
	}

	// 将ABI转换为JSON字符串
	abiBytes, err := json.Marshal(contractData.ABI)
	if err != nil {
		return nil, fmt.Errorf("序列化ABI失败: %v", err)
	}

	contractABI, err := abi.JSON(strings.NewReader(string(abiBytes)))
	if err != nil {
		return nil, fmt.Errorf("解析ABI失败: %v", err)
	}

	return &contractABI, nil
}

// 获取账户信息
func getAccountInfo() (*ecdsa.PrivateKey, common.Address, error) {
	privateKeyHex := os.Getenv("PRIVATE_KEY")
	if privateKeyHex == "" {
		return nil, common.Address{}, fmt.Errorf("请设置PRIVATE_KEY环境变量")
	}

	if strings.HasPrefix(privateKeyHex, "0x") {
		privateKeyHex = privateKeyHex[2:]
	}

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("解析私钥失败: %v", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, common.Address{}, fmt.Errorf("获取公钥失败")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	return privateKey, fromAddress, nil
}
