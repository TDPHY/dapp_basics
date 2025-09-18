package main

import (
	"bufio"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/local/dapp-basics-task01/blockchain"
	"github.com/local/dapp-basics-task01/config"
)

func main() {
	fmt.Println("🚀 DApp基础任务 - 区块链读写演示")
	fmt.Println("=====================================")

	// 加载配置
	cfg := config.LoadConfig()
	fmt.Printf("📡 连接网络: %s\n", cfg.NetworkName)
	fmt.Printf("🔗 RPC URL: %s\n", cfg.EthereumRPCURL)

	// 创建区块链客户端
	client, err := blockchain.NewClient(cfg.EthereumRPCURL)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}
	defer client.Close()

	// 显示菜单
	showMenu()

	// 处理用户输入
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\n请选择操作 (输入数字): ")
		if !scanner.Scan() {
			break
		}

		choice := strings.TrimSpace(scanner.Text())
		switch choice {
		case "1":
			queryLatestBlock(client)
		case "2":
			queryBlockByNumber(client, scanner)
		case "3":
			queryMultipleBlocks(client, scanner)
		case "4":
			checkBalance(client, scanner)
		case "5":
			sendTransaction(client, cfg, scanner)
		case "6":
			showMenu()
		case "0":
			fmt.Println("👋 再见！")
			return
		default:
			fmt.Println("❌ 无效选择，请重新输入")
		}
	}
}

func showMenu() {
	fmt.Println("\n📋 可用操作:")
	fmt.Println("1. 查询最新区块")
	fmt.Println("2. 查询指定区块")
	fmt.Println("3. 查询多个区块")
	fmt.Println("4. 查询地址余额")
	fmt.Println("5. 发送转账交易")
	fmt.Println("6. 显示菜单")
	fmt.Println("0. 退出")
}

func queryLatestBlock(client *blockchain.Client) {
	fmt.Println("\n🔍 查询最新区块...")
	blockInfo, err := client.QueryLatestBlock()
	if err != nil {
		log.Printf("查询失败: %v", err)
		return
	}
	blockInfo.PrintBlockInfo()
}

func queryBlockByNumber(client *blockchain.Client, scanner *bufio.Scanner) {
	fmt.Print("请输入区块号: ")
	if !scanner.Scan() {
		return
	}

	blockNumberStr := strings.TrimSpace(scanner.Text())
	blockNumber, ok := new(big.Int).SetString(blockNumberStr, 10)
	if !ok {
		fmt.Println("❌ 无效的区块号")
		return
	}

	fmt.Printf("\n🔍 查询区块 %s...\n", blockNumber.String())
	blockInfo, err := client.QueryBlockByNumber(blockNumber)
	if err != nil {
		log.Printf("查询失败: %v", err)
		return
	}
	blockInfo.PrintBlockInfo()
}

func queryMultipleBlocks(client *blockchain.Client, scanner *bufio.Scanner) {
	fmt.Print("请输入起始区块号: ")
	if !scanner.Scan() {
		return
	}
	startBlockStr := strings.TrimSpace(scanner.Text())
	startBlock, err := strconv.ParseInt(startBlockStr, 10, 64)
	if err != nil {
		fmt.Println("❌ 无效的起始区块号")
		return
	}

	fmt.Print("请输入查询数量: ")
	if !scanner.Scan() {
		return
	}
	countStr := strings.TrimSpace(scanner.Text())
	count, err := strconv.ParseInt(countStr, 10, 64)
	if err != nil || count <= 0 || count > 10 {
		fmt.Println("❌ 无效的查询数量 (1-10)")
		return
	}

	fmt.Printf("\n🔍 查询从区块 %d 开始的 %d 个区块...\n", startBlock, count)
	blocks, err := client.QueryMultipleBlocks(startBlock, count)
	if err != nil {
		log.Printf("查询失败: %v", err)
		return
	}

	for i, block := range blocks {
		fmt.Printf("\n--- 区块 %d ---", i+1)
		block.PrintBlockInfo()
	}
}

func checkBalance(client *blockchain.Client, scanner *bufio.Scanner) {
	fmt.Print("请输入地址: ")
	if !scanner.Scan() {
		return
	}

	address := strings.TrimSpace(scanner.Text())
	if address == "" {
		fmt.Println("❌ 地址不能为空")
		return
	}

	fmt.Printf("\n💰 查询地址余额: %s\n", address)
	balance, err := client.GetBalance(address)
	if err != nil {
		log.Printf("查询余额失败: %v", err)
		return
	}

	// 转换为ETH
	balanceEth := new(big.Float)
	balanceEth.SetString(balance.String())
	balanceEth = balanceEth.Quo(balanceEth, big.NewFloat(1e18))

	fmt.Printf("余额: %s Wei\n", balance.String())
	fmt.Printf("余额: %s ETH\n", balanceEth.String())
}

func sendTransaction(client *blockchain.Client, cfg *config.Config, scanner *bufio.Scanner) {
	if cfg.PrivateKey == "" {
		fmt.Println("❌ 未配置私钥，无法发送交易")
		fmt.Println("请在 .env 文件中设置 PRIVATE_KEY")
		return
	}

	fmt.Print("请输入接收方地址 (留空使用默认): ")
	if !scanner.Scan() {
		return
	}

	toAddress := strings.TrimSpace(scanner.Text())
	if toAddress == "" {
		toAddress = cfg.ToAddress
		if toAddress == "" {
			fmt.Println("❌ 未指定接收方地址")
			return
		}
	}

	fmt.Print("请输入转账金额 (ETH): ")
	if !scanner.Scan() {
		return
	}

	amountStr := strings.TrimSpace(scanner.Text())
	amountFloat, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amountFloat <= 0 {
		fmt.Println("❌ 无效的转账金额")
		return
	}

	// 转换为Wei
	amount := blockchain.EtherToWei(amountFloat)

	fmt.Printf("\n💸 发送转账交易...\n")
	fmt.Printf("接收方: %s\n", toAddress)
	fmt.Printf("金额: %s ETH (%s Wei)\n", amountStr, amount.String())

	txInfo, err := client.SendTransaction(cfg.PrivateKey, toAddress, amount)
	if err != nil {
		log.Printf("发送交易失败: %v", err)
		return
	}

	txInfo.PrintTransactionInfo()
	fmt.Printf("🔗 查看交易: https://sepolia.etherscan.io/tx/%s\n", txInfo.Hash)
}
