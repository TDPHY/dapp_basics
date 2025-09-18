package blockchain

import (
	"context"
	"log"

	"github.com/ethereum/go-ethereum/ethclient"
)

// Client 封装以太坊客户端
type Client struct {
	client *ethclient.Client
	ctx    context.Context
}

// NewClient 创建新的以太坊客户端连接
func NewClient(rpcURL string) (*Client, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	// 测试连接
	chainID, err := client.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	log.Printf("成功连接到以太坊网络，Chain ID: %d", chainID)

	return &Client{
		client: client,
		ctx:    ctx,
	}, nil
}

// GetClient 获取底层的 ethclient.Client
func (c *Client) GetClient() *ethclient.Client {
	return c.client
}

// GetContext 获取上下文
func (c *Client) GetContext() context.Context {
	return c.ctx
}

// Close 关闭客户端连接
func (c *Client) Close() {
	c.client.Close()
}
