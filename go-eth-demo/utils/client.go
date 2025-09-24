package utils

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/local/go-eth-demo/config"
)

// EthClient 封装以太坊客户端
type EthClient struct {
	client  *ethclient.Client // 底层 go-ethereum 客户端
	config  *config.Config    // 配置信息
	timeout time.Duration     // 操作超时时间
}

// NewEthClient 创建新的以太坊客户端
func NewEthClient(cfg *config.Config) (*EthClient, error) {
	// 1. 建立底层连接
	client, err := ethclient.Dial(cfg.EthereumRPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum client: %w", err)
	}

	// 2. 创建封装实例
	ethClient := &EthClient{
		client:  client,
		config:  cfg,
		timeout: 30 * time.Second, // 默认超时时间
	}

	// 3. 验证连接有效性
	if err := ethClient.ping(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping Ethereum client: %w", err)
	}

	return ethClient, nil
}

// ping 验证客户端连接
func (ec *EthClient) ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), ec.timeout)
	defer cancel()

	// 尝试获取最新区块号来验证连接
	_, err := ec.client.BlockNumber(ctx)
	return err
}

// GetClient 返回底层的 ethclient.Client
func (ec *EthClient) GetClient() *ethclient.Client {
	return ec.client
}

// GetConfig 返回配置
func (ec *EthClient) GetConfig() *config.Config {
	return ec.config
}

// Close 关闭客户端连接
func (ec *EthClient) Close() {
	if ec.client != nil {
		ec.client.Close()
	}
}

// SetTimeout 设置操作超时时间
func (ec *EthClient) SetTimeout(timeout time.Duration) {
	ec.timeout = timeout
}

// GetLatestBlockNumber 获取最新区块号
func (ec *EthClient) GetLatestBlockNumber() (*big.Int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ec.timeout)
	defer cancel()

	blockNumber, err := ec.client.BlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block number: %w", err)
	}

	return big.NewInt(int64(blockNumber)), nil
}

// GetChainID 获取链 ID
func (ec *EthClient) GetChainID() (*big.Int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ec.timeout)
	defer cancel()

	chainID, err := ec.client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	return chainID, nil
}

// GetNetworkID 获取网络 ID
func (ec *EthClient) GetNetworkID() (*big.Int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ec.timeout)
	defer cancel()

	networkID, err := ec.client.NetworkID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get network ID: %w", err)
	}

	return networkID, nil
}

// VerifyNetwork 验证连接的网络是否与配置匹配
func (ec *EthClient) VerifyNetwork() error {
	chainID, err := ec.GetChainID()
	if err != nil {
		return err
	}

	expectedChainID := big.NewInt(ec.config.ChainID)
	if chainID.Cmp(expectedChainID) != 0 {
		return fmt.Errorf("chain ID mismatch: expected %s, got %s",
			expectedChainID.String(), chainID.String())
	}

	return nil
}

// GetConnectionInfo 获取连接信息摘要
func (ec *EthClient) GetConnectionInfo() (map[string]interface{}, error) {
	info := make(map[string]interface{})

	// 获取最新区块号
	if blockNumber, err := ec.GetLatestBlockNumber(); err == nil {
		info["latest_block"] = blockNumber.String()
	} else {
		info["latest_block_error"] = err.Error()
	}

	// 获取链 ID
	if chainID, err := ec.GetChainID(); err == nil {
		info["chain_id"] = chainID.String()
	} else {
		info["chain_id_error"] = err.Error()
	}

	// 获取网络 ID
	if networkID, err := ec.GetNetworkID(); err == nil {
		info["network_id"] = networkID.String()
	} else {
		info["network_id_error"] = err.Error()
	}

	// 添加配置信息
	info["network_name"] = ec.config.NetworkName
	info["rpc_url"] = maskRPCURL(ec.config.EthereumRPCURL)

	return info, nil
}

// maskRPCURL 隐藏 RPC URL 中的敏感信息
func maskRPCURL(url string) string {
	if len(url) > 50 {
		return url[:30] + "***" + url[len(url)-10:]
	}
	return url
}
