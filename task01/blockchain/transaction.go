package blockchain

import (
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// TransactionInfo 存储交易信息
type TransactionInfo struct {
	Hash     string
	From     string
	To       string
	Value    *big.Int
	GasLimit uint64
	GasPrice *big.Int
	Nonce    uint64
	Data     []byte
}

// SendTransaction 发送以太币转账交易
func (c *Client) SendTransaction(privateKeyHex, toAddress string, amount *big.Int) (*TransactionInfo, error) {
	// 解析私钥
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("解析私钥失败: %v", err)
	}

	// 获取发送方地址
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("获取公钥失败")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// 获取nonce
	nonce, err := c.client.PendingNonceAt(c.ctx, fromAddress)
	if err != nil {
		return nil, fmt.Errorf("获取nonce失败: %v", err)
	}

	// 获取gas价格
	gasPrice, err := c.client.SuggestGasPrice(c.ctx)
	if err != nil {
		return nil, fmt.Errorf("获取gas价格失败: %v", err)
	}

	// 设置gas限制
	gasLimit := uint64(21000) // 标准转账的gas限制

	// 创建交易
	toAddr := common.HexToAddress(toAddress)
	tx := types.NewTransaction(nonce, toAddr, amount, gasLimit, gasPrice, nil)

	// 获取链ID
	chainID, err := c.client.ChainID(c.ctx)
	if err != nil {
		return nil, fmt.Errorf("获取链ID失败: %v", err)
	}

	// 签名交易
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return nil, fmt.Errorf("签名交易失败: %v", err)
	}

	// 发送交易
	err = c.client.SendTransaction(c.ctx, signedTx)
	if err != nil {
		return nil, fmt.Errorf("发送交易失败: %v", err)
	}

	log.Printf("交易已发送，哈希: %s", signedTx.Hash().Hex())

	// 返回交易信息
	txInfo := &TransactionInfo{
		Hash:     signedTx.Hash().Hex(),
		From:     fromAddress.Hex(),
		To:       toAddress,
		Value:    amount,
		GasLimit: gasLimit,
		GasPrice: gasPrice,
		Nonce:    nonce,
		Data:     signedTx.Data(),
	}

	return txInfo, nil
}

// GetBalance 获取地址余额
func (c *Client) GetBalance(address string) (*big.Int, error) {
	account := common.HexToAddress(address)
	balance, err := c.client.BalanceAt(c.ctx, account, nil)
	if err != nil {
		return nil, fmt.Errorf("获取余额失败: %v", err)
	}
	return balance, nil
}

// PrintTransactionInfo 打印交易信息到控制台
func (info *TransactionInfo) PrintTransactionInfo() {
	fmt.Println("==================== 交易信息 ====================")
	fmt.Printf("交易哈希: %s\n", info.Hash)
	fmt.Printf("发送方: %s\n", info.From)
	fmt.Printf("接收方: %s\n", info.To)
	fmt.Printf("转账金额: %s Wei\n", info.Value.String())
	fmt.Printf("转账金额: %s ETH\n", weiToEther(info.Value).String())
	fmt.Printf("Gas限制: %d\n", info.GasLimit)
	fmt.Printf("Gas价格: %s Wei\n", info.GasPrice.String())
	fmt.Printf("Nonce: %d\n", info.Nonce)
	fmt.Println("================================================")
}

// weiToEther 将Wei转换为Ether
func weiToEther(wei *big.Int) *big.Float {
	ether := new(big.Float)
	ether.SetString(wei.String())
	return ether.Quo(ether, big.NewFloat(1e18))
}

// EtherToWei 将Ether转换为Wei
func EtherToWei(ether float64) *big.Int {
	etherBig := big.NewFloat(ether)
	weiBig := new(big.Float).Mul(etherBig, big.NewFloat(1e18))
	wei := new(big.Int)
	weiBig.Int(wei)
	return wei
}
