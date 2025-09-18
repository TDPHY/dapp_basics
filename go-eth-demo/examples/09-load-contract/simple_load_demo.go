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

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

// 合约数据结构
type ContractData struct {
	ContractName string      `json:"contractName"`
	SourceFile   string      `json:"sourceFile"`
	ABI          interface{} `json:"abi"`
	Bytecode     string      `json:"bytecode"`
	CompiledAt   string      `json:"compiledAt"`
}

// 合约实例结构
type ContractInstance struct {
	Address common.Address
	ABI     *abi.ABI
	Client  *ethclient.Client
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

	fmt.Println("🔧 智能合约加载演示")
	fmt.Println("====================")

	// 连接以太坊节点
	client, err := ethclient.Dial(os.Getenv("ETHEREUM_RPC_URL"))
	if err != nil {
		log.Fatalf("连接以太坊节点失败: %v", err)
	}
	defer client.Close()

	fmt.Println("✅ 以太坊节点连接成功")

	// 演示1: 从环境变量加载合约
	fmt.Println("\n📋 方法1: 从环境变量加载合约")
	contract1, err := loadContractFromEnv(client)
	if err != nil {
		fmt.Printf("❌ 从环境变量加载失败: %v\n", err)
	} else {
		fmt.Printf("✅ 合约地址: %s\n", contract1.Address.Hex())
		demonstrateContractInfo(contract1)
	}

	// 演示2: 手动指定地址加载合约
	fmt.Println("\n📋 方法2: 手动指定地址加载合约")
	// 使用我们刚才部署的合约地址
	contractAddress := "0xd737368D3bB49649CC4Fc098b93B7c2D6Af9E3B5"
	contract2, err := loadContractByAddress(client, contractAddress)
	if err != nil {
		fmt.Printf("❌ 手动加载失败: %v\n", err)
	} else {
		fmt.Printf("✅ 合约地址: %s\n", contract2.Address.Hex())
		demonstrateContractInfo(contract2)
	}

	// 演示3: 验证合约是否存在
	fmt.Println("\n📋 方法3: 验证合约存在性")
	exists, err := verifyContractExists(client, contractAddress)
	if err != nil {
		fmt.Printf("❌ 验证失败: %v\n", err)
	} else if exists {
		fmt.Println("✅ 合约存在且有代码")
	} else {
		fmt.Println("❌ 合约不存在或无代码")
	}

	// 演示4: 获取合约基本信息
	fmt.Println("\n📋 方法4: 获取合约基本信息")
	if contract2 != nil {
		err = getContractBasicInfo(contract2)
		if err != nil {
			fmt.Printf("❌ 获取信息失败: %v\n", err)
		}
	}

	// 演示5: 批量加载多个合约
	fmt.Println("\n📋 方法5: 批量加载多个合约")
	addresses := []string{
		"0xd737368D3bB49649CC4Fc098b93B7c2D6Af9E3B5", // 我们的合约
		// 可以添加更多合约地址
	}
	contracts, err := loadMultipleContracts(client, addresses)
	if err != nil {
		fmt.Printf("❌ 批量加载失败: %v\n", err)
	} else {
		fmt.Printf("✅ 成功加载 %d 个合约\n", len(contracts))
		for i, contract := range contracts {
			fmt.Printf("   合约%d: %s\n", i+1, contract.Address.Hex())
		}
	}

	fmt.Println("\n🎉 合约加载演示完成!")
}

// 方法1: 从环境变量加载合约
func loadContractFromEnv(client *ethclient.Client) (*ContractInstance, error) {
	contractAddressStr := os.Getenv("CONTRACT_ADDRESS")
	if contractAddressStr == "" {
		return nil, fmt.Errorf("环境变量 CONTRACT_ADDRESS 未设置")
	}

	return loadContractByAddress(client, contractAddressStr)
}

// 方法2: 通过地址加载合约
func loadContractByAddress(client *ethclient.Client, addressStr string) (*ContractInstance, error) {
	// 验证地址格式
	if !common.IsHexAddress(addressStr) {
		return nil, fmt.Errorf("无效的合约地址: %s", addressStr)
	}

	contractAddress := common.HexToAddress(addressStr)

	// 加载合约ABI
	contractABI, err := loadContractABI()
	if err != nil {
		return nil, fmt.Errorf("加载ABI失败: %v", err)
	}

	// 验证合约是否存在
	exists, err := verifyContractExists(client, addressStr)
	if err != nil {
		return nil, fmt.Errorf("验证合约存在性失败: %v", err)
	}
	if !exists {
		return nil, fmt.Errorf("合约不存在或无代码")
	}

	return &ContractInstance{
		Address: contractAddress,
		ABI:     contractABI,
		Client:  client,
	}, nil
}

// 方法3: 验证合约是否存在
func verifyContractExists(client *ethclient.Client, addressStr string) (bool, error) {
	contractAddress := common.HexToAddress(addressStr)

	// 获取合约代码
	code, err := client.CodeAt(context.Background(), contractAddress, nil)
	if err != nil {
		return false, err
	}

	// 如果代码长度大于0，说明合约存在
	return len(code) > 0, nil
}

// 方法4: 获取合约基本信息
func getContractBasicInfo(contract *ContractInstance) error {
	ctx := context.Background()

	// 获取合约代码大小
	code, err := contract.Client.CodeAt(ctx, contract.Address, nil)
	if err != nil {
		return fmt.Errorf("获取合约代码失败: %v", err)
	}

	fmt.Printf("   📏 合约代码大小: %d 字节\n", len(code))

	// 获取合约余额
	balance, err := contract.Client.BalanceAt(ctx, contract.Address, nil)
	if err != nil {
		return fmt.Errorf("获取合约余额失败: %v", err)
	}

	balanceEth := new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(1e18))
	fmt.Printf("   💰 合约余额: %s ETH\n", balanceEth.Text('f', 6))

	// 获取合约nonce（如果合约能发送交易）
	nonce, err := contract.Client.NonceAt(ctx, contract.Address, nil)
	if err != nil {
		return fmt.Errorf("获取合约nonce失败: %v", err)
	}
	fmt.Printf("   🔢 合约nonce: %d\n", nonce)

	return nil
}

// 方法5: 批量加载多个合约
func loadMultipleContracts(client *ethclient.Client, addresses []string) ([]*ContractInstance, error) {
	var contracts []*ContractInstance

	for _, addr := range addresses {
		contract, err := loadContractByAddress(client, addr)
		if err != nil {
			fmt.Printf("   ⚠️  加载合约 %s 失败: %v\n", addr, err)
			continue
		}
		contracts = append(contracts, contract)
	}

	if len(contracts) == 0 {
		return nil, fmt.Errorf("没有成功加载任何合约")
	}

	return contracts, nil
}

// 演示合约信息
func demonstrateContractInfo(contract *ContractInstance) {
	// 尝试调用合约的只读方法
	fmt.Println("   🔍 尝试读取合约状态...")

	// 调用retrieve方法
	data, err := contract.ABI.Pack("retrieve")
	if err != nil {
		fmt.Printf("   ❌ 打包函数调用失败: %v\n", err)
		return
	}

	result, err := contract.Client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &contract.Address,
		Data: data,
	}, nil)
	if err != nil {
		fmt.Printf("   ❌ 调用合约失败: %v\n", err)
		return
	}

	// 解析结果
	var value *big.Int
	err = contract.ABI.UnpackIntoInterface(&value, "retrieve", result)
	if err != nil {
		fmt.Printf("   ❌ 解析结果失败: %v\n", err)
		return
	}

	fmt.Printf("   ✅ 当前存储值: %s\n", value.String())

	// 尝试获取owner地址
	ownerData, err := contract.ABI.Pack("owner")
	if err == nil {
		ownerResult, err := contract.Client.CallContract(context.Background(), ethereum.CallMsg{
			To:   &contract.Address,
			Data: ownerData,
		}, nil)
		if err == nil {
			var owner common.Address
			err = contract.ABI.UnpackIntoInterface(&owner, "owner", ownerResult)
			if err == nil {
				fmt.Printf("   👤 合约所有者: %s\n", owner.Hex())
			}
		}
	}
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

// 辅助函数：获取账户信息
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
