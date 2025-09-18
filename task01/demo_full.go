package main

import (
	"bufio"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/local/dapp-basics-task01/blockchain"
	"github.com/local/dapp-basics-task01/config"
	"github.com/local/dapp-basics-task01/contracts"
)

func main() {
	fmt.Println("🚀 DApp基础任务演示 - 完整功能测试")
	fmt.Println("=====================================")

	// 加载配置
	cfg := config.LoadConfig()
	fmt.Printf("📡 连接网络: %s\n", cfg.NetworkName)

	// 创建区块链客户端
	client, err := blockchain.NewClient(cfg.EthereumRPCURL)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}
	defer client.Close()

	// 显示演示菜单
	showDemoMenu()

	// 处理用户输入
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\n请选择演示 (输入数字): ")
		if !scanner.Scan() {
			break
		}

		choice := strings.TrimSpace(scanner.Text())
		switch choice {
		case "1":
			demoBlockQuery(client)
		case "2":
			demoTransaction(client, cfg)
		case "3":
			demoContractDeploy(client, cfg)
		case "4":
			demoContractInteraction(client, cfg, scanner)
		case "5":
			demoFullWorkflow(client, cfg)
		case "6":
			showDemoMenu()
		case "0":
			fmt.Println("👋 演示结束！")
			return
		default:
			fmt.Println("❌ 无效选择，请重新输入")
		}
	}
}

func showDemoMenu() {
	fmt.Println("\n📋 演示项目:")
	fmt.Println("1. 区块查询演示")
	fmt.Println("2. 转账交易演示")
	fmt.Println("3. 合约部署演示")
	fmt.Println("4. 合约交互演示")
	fmt.Println("5. 完整工作流演示")
	fmt.Println("6. 显示菜单")
	fmt.Println("0. 退出")
}

func demoBlockQuery(client *blockchain.Client) {
	fmt.Println("\n🔍 === 区块查询演示 ===")

	// 查询最新区块
	fmt.Println("1. 查询最新区块:")
	latestBlock, err := client.QueryLatestBlock()
	if err != nil {
		log.Printf("查询最新区块失败: %v", err)
		return
	}
	latestBlock.PrintBlockInfo()

	// 查询指定区块
	fmt.Println("\n2. 查询指定区块 (最新区块-1):")
	blockNumber := new(big.Int).Sub(latestBlock.Number, big.NewInt(1))
	blockInfo, err := client.QueryBlockByNumber(blockNumber)
	if err != nil {
		log.Printf("查询指定区块失败: %v", err)
		return
	}
	blockInfo.PrintBlockInfo()
}

func demoTransaction(client *blockchain.Client, cfg *config.Config) {
	fmt.Println("\n💸 === 转账交易演示 ===")

	if cfg.PrivateKey == "" {
		fmt.Println("❌ 未配置私钥，跳过转账演示")
		return
	}

	if cfg.ToAddress == "" {
		fmt.Println("❌ 未配置接收方地址，跳过转账演示")
		return
	}

	// 发送小额转账 (0.001 ETH)
	amount := blockchain.EtherToWei(0.001)
	fmt.Printf("发送 0.001 ETH 到 %s\n", cfg.ToAddress)

	txInfo, err := client.SendTransaction(cfg.PrivateKey, cfg.ToAddress, amount)
	if err != nil {
		log.Printf("发送交易失败: %v", err)
		return
	}

	txInfo.PrintTransactionInfo()
	fmt.Printf("🔗 查看交易: https://sepolia.etherscan.io/tx/%s\n", txInfo.Hash)
}

func demoContractDeploy(client *blockchain.Client, cfg *config.Config) {
	fmt.Println("\n🚀 === 合约部署演示 ===")

	if cfg.PrivateKey == "" {
		fmt.Println("❌ 未配置私钥，跳过合约部署演示")
		return
	}

	// 部署Counter合约，初始值为42
	initialValue := big.NewInt(42)
	fmt.Printf("部署Counter合约，初始值: %s\n", initialValue.String())

	address, txHash, err := contracts.DeployCounter(client.GetClient(), cfg.PrivateKey, initialValue)
	if err != nil {
		log.Printf("部署合约失败: %v", err)
		return
	}

	fmt.Printf("✅ 合约部署成功!\n")
	fmt.Printf("合约地址: %s\n", address.Hex())
	fmt.Printf("部署交易: %s\n", txHash)
	fmt.Printf("🔗 查看合约: https://sepolia.etherscan.io/address/%s\n", address.Hex())

	// 保存合约地址到环境变量文件
	saveContractAddress(address.Hex())
}

func demoContractInteraction(client *blockchain.Client, cfg *config.Config, scanner *bufio.Scanner) {
	fmt.Println("\n🔧 === 合约交互演示 ===")

	fmt.Print("请输入合约地址: ")
	if !scanner.Scan() {
		return
	}

	contractAddress := strings.TrimSpace(scanner.Text())
	if contractAddress == "" {
		fmt.Println("❌ 合约地址不能为空")
		return
	}

	// 创建合约管理器
	counterManager, err := contracts.NewCounterManager(client.GetClient(), contractAddress, cfg.PrivateKey)
	if err != nil {
		log.Printf("创建合约管理器失败: %v", err)
		return
	}

	// 显示合约信息
	counterManager.PrintContractInfo()

	// 如果有私钥，演示写操作
	if cfg.PrivateKey != "" {
		fmt.Println("\n执行合约操作:")

		// 增加计数器
		fmt.Println("1. 增加计数器...")
		txHash, err := counterManager.Increment()
		if err != nil {
			log.Printf("增加计数器失败: %v", err)
		} else {
			fmt.Printf("✅ 增加计数器成功，交易: %s\n", txHash)
		}

		// 等待一下，然后查询新值
		fmt.Println("2. 查询新的计数值...")
		count, err := counterManager.GetCount()
		if err != nil {
			log.Printf("查询计数值失败: %v", err)
		} else {
			fmt.Printf("✅ 当前计数值: %s\n", count.String())
		}

		// 增加指定数量
		fmt.Println("3. 增加10...")
		txHash, err = counterManager.Add(big.NewInt(10))
		if err != nil {
			log.Printf("增加数量失败: %v", err)
		} else {
			fmt.Printf("✅ 增加10成功，交易: %s\n", txHash)
		}
	}
}

func demoFullWorkflow(client *blockchain.Client, cfg *config.Config) {
	fmt.Println("\n🎯 === 完整工作流演示 ===")

	if cfg.PrivateKey == "" {
		fmt.Println("❌ 未配置私钥，无法执行完整工作流")
		return
	}

	fmt.Println("步骤1: 查询最新区块")
	latestBlock, err := client.QueryLatestBlock()
	if err != nil {
		log.Printf("查询最新区块失败: %v", err)
		return
	}
	fmt.Printf("✅ 最新区块: %s\n", latestBlock.Number.String())

	fmt.Println("\n步骤2: 部署Counter合约")
	initialValue := big.NewInt(100)
	address, txHash, err := contracts.DeployCounter(client.GetClient(), cfg.PrivateKey, initialValue)
	if err != nil {
		log.Printf("部署合约失败: %v", err)
		return
	}
	fmt.Printf("✅ 合约地址: %s\n", address.Hex())
	fmt.Printf("✅ 部署交易: %s\n", txHash)

	fmt.Println("\n步骤3: 与合约交互")
	counterManager, err := contracts.NewCounterManager(client.GetClient(), address.Hex(), cfg.PrivateKey)
	if err != nil {
		log.Printf("创建合约管理器失败: %v", err)
		return
	}

	// 显示初始状态
	counterManager.PrintContractInfo()

	// 执行一系列操作
	operations := []struct {
		name string
		fn   func() (string, error)
	}{
		{"增加计数器", counterManager.Increment},
		{"减少计数器", counterManager.Decrement},
		{"增加5", func() (string, error) { return counterManager.Add(big.NewInt(5)) }},
		{"减去3", func() (string, error) { return counterManager.Subtract(big.NewInt(3)) }},
	}

	for i, op := range operations {
		fmt.Printf("\n步骤%d: %s\n", i+4, op.name)
		txHash, err := op.fn()
		if err != nil {
			log.Printf("%s失败: %v", op.name, err)
			continue
		}
		fmt.Printf("✅ %s成功，交易: %s\n", op.name, txHash)

		// 查询当前值
		count, err := counterManager.GetCount()
		if err != nil {
			log.Printf("查询计数值失败: %v", err)
		} else {
			fmt.Printf("📊 当前计数值: %s\n", count.String())
		}
	}

	fmt.Println("\n🎉 完整工作流演示完成!")
	fmt.Printf("🔗 查看合约: https://sepolia.etherscan.io/address/%s\n", address.Hex())
}

func saveContractAddress(address string) {
	// 简单地打印到控制台，实际项目中可以保存到文件
	fmt.Printf("\n💾 合约地址已记录: %s\n", address)
	fmt.Println("可以将此地址保存到 .env 文件中的 CONTRACT_ADDRESS 变量")
}
