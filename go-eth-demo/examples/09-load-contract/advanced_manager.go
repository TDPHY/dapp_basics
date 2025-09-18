package main

import (
	"context"
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

// 合约管理器
type ContractManager struct {
	client    *ethclient.Client
	contracts map[string]*ContractInstance
}

// 合约配置
type ContractConfig struct {
	Name       string `json:"name"`
	Address    string `json:"address"`
	ABIFile    string `json:"abiFile"`
	Network    string `json:"network"`
	DeployedAt string `json:"deployedAt"`
}

// 合约注册表
type ContractRegistry struct {
	Contracts []ContractConfig  `json:"contracts"`
	Networks  map[string]string `json:"networks"`
}

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Printf("警告: 无法加载 .env 文件: %v", err)
	}

	fmt.Println("🏗️  智能合约管理器演示")
	fmt.Println("========================")

	// 创建合约管理器
	manager, err := NewContractManager(os.Getenv("ETHEREUM_RPC_URL"))
	if err != nil {
		log.Fatalf("创建合约管理器失败: %v", err)
	}
	defer manager.Close()

	// 演示1: 创建合约注册表
	fmt.Println("\n📋 1. 创建合约注册表")
	registry := createSampleRegistry()
	err = saveContractRegistry(registry, "contracts_registry.json")
	if err != nil {
		fmt.Printf("❌ 保存注册表失败: %v\n", err)
	} else {
		fmt.Println("✅ 合约注册表已创建")
	}

	// 演示2: 从注册表加载合约
	fmt.Println("\n📋 2. 从注册表加载合约")
	err = manager.LoadFromRegistry("contracts_registry.json")
	if err != nil {
		fmt.Printf("❌ 从注册表加载失败: %v\n", err)
	} else {
		fmt.Printf("✅ 成功加载 %d 个合约\n", len(manager.contracts))
	}

	// 演示3: 列出所有已加载的合约
	fmt.Println("\n📋 3. 列出已加载的合约")
	manager.ListContracts()

	// 演示4: 获取特定合约
	fmt.Println("\n📋 4. 获取特定合约")
	contract := manager.GetContract("SimpleStorage")
	if contract != nil {
		fmt.Printf("✅ 找到合约: %s\n", contract.Address.Hex())

		// 调用合约方法
		value, err := callRetrieveMethod(contract)
		if err != nil {
			fmt.Printf("❌ 调用失败: %v\n", err)
		} else {
			fmt.Printf("📊 当前存储值: %s\n", value.String())
		}
	} else {
		fmt.Println("❌ 未找到合约")
	}

	// 演示5: 监控合约事件
	fmt.Println("\n📋 5. 监控合约事件")
	if contract != nil {
		fmt.Println("🔍 开始监控DataStored事件...")
		go monitorContractEvents(contract)

		// 等待一段时间以观察事件
		time.Sleep(5 * time.Second)
	}

	// 演示6: 合约健康检查
	fmt.Println("\n📋 6. 合约健康检查")
	healthReport := manager.HealthCheck()
	fmt.Printf("📊 健康检查报告:\n")
	for name, status := range healthReport {
		if status {
			fmt.Printf("   ✅ %s: 健康\n", name)
		} else {
			fmt.Printf("   ❌ %s: 异常\n", name)
		}
	}

	fmt.Println("\n🎉 合约管理器演示完成!")
}

// 创建合约管理器
func NewContractManager(rpcURL string) (*ContractManager, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("连接以太坊节点失败: %v", err)
	}

	return &ContractManager{
		client:    client,
		contracts: make(map[string]*ContractInstance),
	}, nil
}

// 关闭管理器
func (cm *ContractManager) Close() {
	if cm.client != nil {
		cm.client.Close()
	}
}

// 从注册表加载合约
func (cm *ContractManager) LoadFromRegistry(registryFile string) error {
	data, err := ioutil.ReadFile(registryFile)
	if err != nil {
		return fmt.Errorf("读取注册表文件失败: %v", err)
	}

	var registry ContractRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		return fmt.Errorf("解析注册表失败: %v", err)
	}

	for _, config := range registry.Contracts {
		contract, err := cm.loadContractFromConfig(config)
		if err != nil {
			fmt.Printf("   ⚠️  加载合约 %s 失败: %v\n", config.Name, err)
			continue
		}
		cm.contracts[config.Name] = contract
		fmt.Printf("   ✅ 加载合约: %s (%s)\n", config.Name, config.Address)
	}

	return nil
}

// 从配置加载合约
func (cm *ContractManager) loadContractFromConfig(config ContractConfig) (*ContractInstance, error) {
	// 验证地址
	if !common.IsHexAddress(config.Address) {
		return nil, fmt.Errorf("无效的合约地址: %s", config.Address)
	}

	contractAddress := common.HexToAddress(config.Address)

	// 加载ABI
	contractABI, err := loadABIFromFile(config.ABIFile)
	if err != nil {
		return nil, fmt.Errorf("加载ABI失败: %v", err)
	}

	// 验证合约存在
	code, err := cm.client.CodeAt(context.Background(), contractAddress, nil)
	if err != nil {
		return nil, fmt.Errorf("获取合约代码失败: %v", err)
	}
	if len(code) == 0 {
		return nil, fmt.Errorf("合约不存在或无代码")
	}

	return &ContractInstance{
		Address: contractAddress,
		ABI:     contractABI,
		Client:  cm.client,
	}, nil
}

// 获取合约
func (cm *ContractManager) GetContract(name string) *ContractInstance {
	return cm.contracts[name]
}

// 列出所有合约
func (cm *ContractManager) ListContracts() {
	if len(cm.contracts) == 0 {
		fmt.Println("   📭 没有已加载的合约")
		return
	}

	for name, contract := range cm.contracts {
		fmt.Printf("   📄 %s: %s\n", name, contract.Address.Hex())
	}
}

// 健康检查
func (cm *ContractManager) HealthCheck() map[string]bool {
	healthReport := make(map[string]bool)

	for name, contract := range cm.contracts {
		// 检查合约代码是否存在
		code, err := cm.client.CodeAt(context.Background(), contract.Address, nil)
		if err != nil || len(code) == 0 {
			healthReport[name] = false
			continue
		}

		// 尝试调用一个简单的方法
		healthy := true
		if _, exists := contract.ABI.Methods["retrieve"]; exists {
			_, err := callRetrieveMethod(contract)
			if err != nil {
				healthy = false
			}
		}

		healthReport[name] = healthy
	}

	return healthReport
}

// 创建示例注册表
func createSampleRegistry() *ContractRegistry {
	return &ContractRegistry{
		Contracts: []ContractConfig{
			{
				Name:       "SimpleStorage",
				Address:    "0xd737368D3bB49649CC4Fc098b93B7c2D6Af9E3B5",
				ABIFile:    "build/SimpleStorage.json",
				Network:    "sepolia",
				DeployedAt: time.Now().Format(time.RFC3339),
			},
		},
		Networks: map[string]string{
			"sepolia": "https://ethereum-sepolia.publicnode.com",
			"mainnet": "https://ethereum.publicnode.com",
		},
	}
}

// 保存合约注册表
func saveContractRegistry(registry *ContractRegistry, filename string) error {
	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化注册表失败: %v", err)
	}

	return ioutil.WriteFile(filename, data, 0644)
}

// 从文件加载ABI
func loadABIFromFile(filename string) (*abi.ABI, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("读取ABI文件失败: %v", err)
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

// 调用retrieve方法
func callRetrieveMethod(contract *ContractInstance) (*big.Int, error) {
	data, err := contract.ABI.Pack("retrieve")
	if err != nil {
		return nil, fmt.Errorf("打包函数调用失败: %v", err)
	}

	result, err := contract.Client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &contract.Address,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("调用合约失败: %v", err)
	}

	var value *big.Int
	err = contract.ABI.UnpackIntoInterface(&value, "retrieve", result)
	if err != nil {
		return nil, fmt.Errorf("解析结果失败: %v", err)
	}

	return value, nil
}

// 监控合约事件
func monitorContractEvents(contract *ContractInstance) {
	// 创建事件过滤器
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contract.Address},
		Topics:    [][]common.Hash{},
	}

	// 订阅日志
	logs := make(chan types.Log)
	sub, err := contract.Client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		fmt.Printf("❌ 订阅事件失败: %v\n", err)
		return
	}
	defer sub.Unsubscribe()

	fmt.Println("🔍 正在监听合约事件...")

	for {
		select {
		case err := <-sub.Err():
			fmt.Printf("❌ 事件订阅错误: %v\n", err)
			return
		case vLog := <-logs:
			fmt.Printf("📨 收到事件: 区块 %d, 交易 %s\n",
				vLog.BlockNumber, vLog.TxHash.Hex())
		case <-time.After(10 * time.Second):
			fmt.Println("⏰ 监听超时，停止监控")
			return
		}
	}
}
