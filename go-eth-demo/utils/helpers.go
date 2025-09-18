package utils

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// WeiToEther 将 Wei 转换为 Ether
func WeiToEther(wei *big.Int) string {
	ether := new(big.Float).SetInt(wei)
	ether.Quo(ether, big.NewFloat(1e18))
	return ether.Text('f', 6)
}

// WeiToGwei 将 Wei 转换为 Gwei
func WeiToGwei(wei *big.Int) string {
	gwei := new(big.Float).SetInt(wei)
	gwei.Quo(gwei, big.NewFloat(1e9))
	return gwei.Text('f', 2)
}

// FormatNumber 格式化大数字
func FormatNumber(n uint64) string {
	str := fmt.Sprintf("%d", n)
	if len(str) <= 3 {
		return str
	}

	result := ""
	for i, char := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result += ","
		}
		result += string(char)
	}
	return result
}

// GetTransactionSender 获取交易发送方地址
func GetTransactionSender(tx *types.Transaction) (common.Address, error) {
	chainID := tx.ChainId()
	if chainID == nil {
		return common.Address{}, fmt.Errorf("无法获取链 ID")
	}

	signer := types.NewEIP155Signer(chainID)
	return types.Sender(signer, tx)
}

// Min 返回两个整数中的较小值
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
