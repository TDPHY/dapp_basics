package contracts

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// CounterManager 管理Counter合约交互
type CounterManager struct {
	client   *ethclient.Client
	contract *Contracts
	auth     *bind.TransactOpts
	address  common.Address
	ctx      context.Context
}

// NewCounterManager 创建新的Counter合约管理器
func NewCounterManager(client *ethclient.Client, contractAddress string, privateKeyHex string) (*CounterManager, error) {
	ctx := context.Background()

	// 解析合约地址
	address := common.HexToAddress(contractAddress)

	// 创建合约实例
	contract, err := NewContracts(address, client)
	if err != nil {
		return nil, fmt.Errorf("创建合约实例失败: %v", err)
	}

	// 创建交易授权
	var auth *bind.TransactOpts
	if privateKeyHex != "" {
		privateKey, err := crypto.HexToECDSA(privateKeyHex)
		if err != nil {
			return nil, fmt.Errorf("解析私钥失败: %v", err)
		}

		chainID, err := client.ChainID(ctx)
		if err != nil {
			return nil, fmt.Errorf("获取链ID失败: %v", err)
		}

		auth, err = bind.NewKeyedTransactorWithChainID(privateKey, chainID)
		if err != nil {
			return nil, fmt.Errorf("创建交易授权失败: %v", err)
		}
	}

	return &CounterManager{
		client:   client,
		contract: contract,
		auth:     auth,
		address:  address,
		ctx:      ctx,
	}, nil
}

// GetCount 获取当前计数值
func (cm *CounterManager) GetCount() (*big.Int, error) {
	count, err := cm.contract.GetCount(&bind.CallOpts{Context: cm.ctx})
	if err != nil {
		return nil, fmt.Errorf("获取计数值失败: %v", err)
	}
	return count, nil
}

// GetInfo 获取合约信息
func (cm *CounterManager) GetInfo() (*big.Int, common.Address, error) {
	info, err := cm.contract.GetInfo(&bind.CallOpts{Context: cm.ctx})
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("获取合约信息失败: %v", err)
	}
	return info.Count, info.Owner, nil
}

// GetOwner 获取合约所有者
func (cm *CounterManager) GetOwner() (common.Address, error) {
	owner, err := cm.contract.Owner(&bind.CallOpts{Context: cm.ctx})
	if err != nil {
		return common.Address{}, fmt.Errorf("获取所有者失败: %v", err)
	}
	return owner, nil
}

// Increment 增加计数器
func (cm *CounterManager) Increment() (string, error) {
	if cm.auth == nil {
		return "", fmt.Errorf("未配置私钥，无法发送交易")
	}

	tx, err := cm.contract.Increment(cm.auth)
	if err != nil {
		return "", fmt.Errorf("增加计数器失败: %v", err)
	}

	log.Printf("增加计数器交易已发送: %s", tx.Hash().Hex())
	return tx.Hash().Hex(), nil
}

// Add 增加指定数量
func (cm *CounterManager) Add(value *big.Int) (string, error) {
	if cm.auth == nil {
		return "", fmt.Errorf("未配置私钥，无法发送交易")
	}

	tx, err := cm.contract.Add(cm.auth, value)
	if err != nil {
		return "", fmt.Errorf("增加数量失败: %v", err)
	}

	log.Printf("增加数量 %s 交易已发送: %s", value.String(), tx.Hash().Hex())
	return tx.Hash().Hex(), nil
}

// PrintContractInfo 打印合约信息
func (cm *CounterManager) PrintContractInfo() {
	fmt.Println("==================== 合约信息 ====================")
	fmt.Printf("合约地址: %s\n", cm.address.Hex())

	count, owner, err := cm.GetInfo()
	if err != nil {
		fmt.Printf("获取合约信息失败: %v\n", err)
		return
	}

	fmt.Printf("当前计数值: %s\n", count.String())
	fmt.Printf("合约所有者: %s\n", owner.Hex())
	fmt.Println("================================================")
}

// DeployCounter 部署Counter合约
func DeployCounter(client *ethclient.Client, privateKeyHex string, initialValue *big.Int) (common.Address, string, error) {
	ctx := context.Background()

	// 解析私钥
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return common.Address{}, "", fmt.Errorf("解析私钥失败: %v", err)
	}

	// 获取链ID
	chainID, err := client.ChainID(ctx)
	if err != nil {
		return common.Address{}, "", fmt.Errorf("获取链ID失败: %v", err)
	}

	// 创建交易授权
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return common.Address{}, "", fmt.Errorf("创建交易授权失败: %v", err)
	}

	// 部署合约
	address, tx, _, err := DeployContracts(auth, client, initialValue)
	if err != nil {
		return common.Address{}, "", fmt.Errorf("部署合约失败: %v", err)
	}

	log.Printf("合约部署交易已发送: %s", tx.Hash().Hex())
	log.Printf("合约地址: %s", address.Hex())

	return address, tx.Hash().Hex(), nil
}
