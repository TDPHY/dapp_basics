package blockchain

import (
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
)

// BlockInfo 存储区块信息
type BlockInfo struct {
	Number          *big.Int
	Hash            string
	ParentHash      string
	Timestamp       uint64
	TxCount         int
	GasUsed         uint64
	GasLimit        uint64
	Miner           string
	Difficulty      *big.Int
	TotalDifficulty *big.Int
	Size            uint64
}

// QueryBlockByNumber 根据区块号查询区块信息
func (c *Client) QueryBlockByNumber(blockNumber *big.Int) (*BlockInfo, error) {
	block, err := c.client.BlockByNumber(c.ctx, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("查询区块失败: %v", err)
	}

	return c.extractBlockInfo(block), nil
}

// QueryLatestBlock 查询最新区块信息
func (c *Client) QueryLatestBlock() (*BlockInfo, error) {
	block, err := c.client.BlockByNumber(c.ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("查询最新区块失败: %v", err)
	}

	return c.extractBlockInfo(block), nil
}

// QueryBlockByHash 根据区块哈希查询区块信息
func (c *Client) QueryBlockByHash(blockHash string) (*BlockInfo, error) {
	// 这里可以实现根据哈希查询区块的逻辑
	// 为了简化，暂时不实现
	return nil, fmt.Errorf("根据哈希查询区块功能暂未实现")
}

// extractBlockInfo 从区块中提取信息
func (c *Client) extractBlockInfo(block *types.Block) *BlockInfo {
	return &BlockInfo{
		Number:          block.Number(),
		Hash:            block.Hash().Hex(),
		ParentHash:      block.ParentHash().Hex(),
		Timestamp:       block.Time(),
		TxCount:         len(block.Transactions()),
		GasUsed:         block.GasUsed(),
		GasLimit:        block.GasLimit(),
		Miner:           block.Coinbase().Hex(),
		Difficulty:      block.Difficulty(),
		TotalDifficulty: block.Difficulty(), // 注意：这里简化了，实际应该是累计难度
		Size:            block.Size(),
	}
}

// PrintBlockInfo 打印区块信息到控制台
func (info *BlockInfo) PrintBlockInfo() {
	fmt.Println("==================== 区块信息 ====================")
	fmt.Printf("区块号: %s\n", info.Number.String())
	fmt.Printf("区块哈希: %s\n", info.Hash)
	fmt.Printf("父区块哈希: %s\n", info.ParentHash)
	fmt.Printf("时间戳: %d (%s)\n", info.Timestamp, time.Unix(int64(info.Timestamp), 0).Format("2006-01-02 15:04:05"))
	fmt.Printf("交易数量: %d\n", info.TxCount)
	fmt.Printf("Gas 使用量: %d\n", info.GasUsed)
	fmt.Printf("Gas 限制: %d\n", info.GasLimit)
	fmt.Printf("矿工地址: %s\n", info.Miner)
	fmt.Printf("难度: %s\n", info.Difficulty.String())
	fmt.Printf("区块大小: %d bytes\n", info.Size)
	fmt.Println("================================================")
}

// QueryMultipleBlocks 查询多个区块信息
func (c *Client) QueryMultipleBlocks(startBlock, count int64) ([]*BlockInfo, error) {
	var blocks []*BlockInfo

	log.Printf("开始查询从区块 %d 开始的 %d 个区块...", startBlock, count)

	for i := int64(0); i < count; i++ {
		blockNumber := big.NewInt(startBlock + i)
		blockInfo, err := c.QueryBlockByNumber(blockNumber)
		if err != nil {
			log.Printf("查询区块 %d 失败: %v", startBlock+i, err)
			continue
		}
		blocks = append(blocks, blockInfo)
		log.Printf("成功查询区块 %d", startBlock+i)
	}

	return blocks, nil
}
