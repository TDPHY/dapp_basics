package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("🔐 KeyStore 文件使用演示")
	fmt.Println("================================")

	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Printf("警告: 无法加载 .env 文件: %v", err)
	}

	// 1. 显示 KeyStore 文件信息
	fmt.Println("\n📁 KeyStore 文件列表:")
	fmt.Println("--------------------------------")

	keystoreDir := "keystore"
	files, err := os.ReadDir(keystoreDir)
	if err != nil {
		log.Fatalf("读取 keystore 目录失败: %v", err)
	}

	var keystoreFile string
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "UTC--") {
			keystoreFile = filepath.Join(keystoreDir, file.Name())
			fmt.Printf("✅ 找到 KeyStore 文件: %s\n", file.Name())

			// 显示文件详细信息
			if info, err := file.Info(); err == nil {
				fmt.Printf("   文件大小: %d 字节\n", info.Size())
				fmt.Printf("   修改时间: %s\n", info.ModTime().Format("2006-01-02 15:04:05"))
			}
		}
	}

	if keystoreFile == "" {
		log.Fatalf("未找到 KeyStore 文件")
	}

	// 2. 解析 KeyStore 文件内容
	fmt.Println("\n🔍 KeyStore 文件内容分析:")
	fmt.Println("--------------------------------")

	content, err := os.ReadFile(keystoreFile)
	if err != nil {
		log.Fatalf("读取 KeyStore 文件失败: %v", err)
	}

	fmt.Printf("文件内容长度: %d 字节\n", len(content))

	// 检查 JSON 格式
	contentStr := string(content)
	if strings.Contains(contentStr, "\"address\"") {
		fmt.Println("✅ 包含地址字段")
	}
	if strings.Contains(contentStr, "\"crypto\"") || strings.Contains(contentStr, "\"Crypto\"") {
		fmt.Println("✅ 包含加密数据")
	}
	if strings.Contains(contentStr, "\"version\"") {
		fmt.Println("✅ 包含版本信息")
	}

	// 3. 使用 KeyStore 解锁钱包
	fmt.Println("\n🔓 解锁 KeyStore 钱包:")
	fmt.Println("================================")

	password := "demo123456" // 演示密码
	fmt.Printf("使用密码: %s\n", password)

	// 创建 KeyStore 实例
	ks := keystore.NewKeyStore(keystoreDir, keystore.StandardScryptN, keystore.StandardScryptP)

	// 获取账户列表
	accounts := ks.Accounts()
	if len(accounts) == 0 {
		log.Fatalf("KeyStore 中没有找到账户")
	}

	account := accounts[0]
	fmt.Printf("账户地址: %s\n", account.Address.Hex())

	// 解锁账户
	err = ks.Unlock(account, password)
	if err != nil {
		log.Fatalf("解锁账户失败: %v", err)
	}
	fmt.Println("✅ 账户解锁成功!")

	// 4. 验证私钥一致性
	fmt.Println("\n🔑 验证私钥一致性:")
	fmt.Println("--------------------------------")

	// 从环境变量获取私钥
	envPrivateKey := os.Getenv("PRIVATE_KEY")
	if envPrivateKey == "" {
		fmt.Println("⚠️  环境变量中未找到私钥")
	} else {
		fmt.Printf("环境变量私钥长度: %d 字符\n", len(envPrivateKey))

		// 验证私钥对应的地址
		privateKey, err := crypto.HexToECDSA(envPrivateKey)
		if err != nil {
			fmt.Printf("❌ 私钥格式错误: %v\n", err)
		} else {
			address := crypto.PubkeyToAddress(privateKey.PublicKey)
			fmt.Printf("私钥对应地址: %s\n", address.Hex())

			if strings.EqualFold(address.Hex(), account.Address.Hex()) {
				fmt.Println("✅ 私钥与 KeyStore 地址一致!")
			} else {
				fmt.Println("❌ 私钥与 KeyStore 地址不一致!")
			}
		}
	}

	// 5. 演示 KeyStore 的实际用途
	fmt.Println("\n💡 KeyStore 使用场景:")
	fmt.Println("================================")

	demonstrateKeystoreUsage(account.Address)

	// 6. 安全建议
	fmt.Println("\n🛡️  KeyStore 安全建议:")
	fmt.Println("================================")
	displaySecurityRecommendations()

	// 7. 清理：锁定账户
	ks.Lock(account.Address)
	fmt.Println("\n🔒 账户已重新锁定")
	fmt.Println("演示完成!")
}

// demonstrateKeystoreUsage 演示 KeyStore 的使用场景
func demonstrateKeystoreUsage(address common.Address) {
	fmt.Println("1. 🌐 导入到 MetaMask:")
	fmt.Println("   • 打开 MetaMask")
	fmt.Println("   • 选择 '导入账户'")
	fmt.Println("   • 选择 'JSON 文件'")
	fmt.Println("   • 上传 KeyStore 文件")
	fmt.Println("   • 输入密码: demo123456")

	fmt.Println("\n2. 💰 在其他钱包中使用:")
	fmt.Println("   • MyEtherWallet (MEW)")
	fmt.Println("   • MyCrypto")
	fmt.Println("   • Trust Wallet")
	fmt.Println("   • 大部分支持 KeyStore 的钱包")

	fmt.Println("\n3. 🔄 程序中使用:")
	fmt.Println("   • 使用 go-ethereum 的 keystore 包")
	fmt.Println("   • 需要密码解锁")
	fmt.Println("   • 可以签名交易")
	fmt.Println("   • 比直接使用私钥更安全")

	fmt.Printf("\n4. 🎯 当前钱包地址: %s\n", address.Hex())
	fmt.Println("   • 可以接收 ETH 和代币")
	fmt.Println("   • 在 Sepolia 测试网上使用")
	fmt.Println("   • 可以在区块链浏览器查看")
}

// displaySecurityRecommendations 显示安全建议
func displaySecurityRecommendations() {
	fmt.Println("🔐 密码安全:")
	fmt.Println("   • 使用强密码（至少 12 位）")
	fmt.Println("   • 包含大小写字母、数字、特殊字符")
	fmt.Println("   • 不要使用常见密码")
	fmt.Println("   • 定期更换密码")

	fmt.Println("\n📁 文件安全:")
	fmt.Println("   • 备份 KeyStore 文件到安全位置")
	fmt.Println("   • 不要存储在云盘等不安全位置")
	fmt.Println("   • 考虑使用硬件钱包")
	fmt.Println("   • 定期检查文件完整性")

	fmt.Println("\n🌐 网络安全:")
	fmt.Println("   • 不要在不安全的网络环境中使用")
	fmt.Println("   • 避免在公共电脑上操作")
	fmt.Println("   • 使用 HTTPS 连接")
	fmt.Println("   • 注意钓鱼网站")

	fmt.Println("\n💡 最佳实践:")
	fmt.Println("   • 测试网先练习")
	fmt.Println("   • 小额测试后再大额操作")
	fmt.Println("   • 保持软件更新")
	fmt.Println("   • 学习基本的区块链安全知识")
}
