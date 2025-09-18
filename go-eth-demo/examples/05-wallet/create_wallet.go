package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/term"
)

func main() {
	fmt.Println("🔐 以太坊钱包创建工具")
	fmt.Println("================================")

	// 1. 生成新的私钥
	fmt.Println("\n🎲 生成新的私钥...")
	fmt.Println("--------------------------------")

	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatalf("生成私钥失败: %v", err)
	}

	// 2. 从私钥导出公钥和地址
	fmt.Println("\n🔑 导出钱包信息...")
	fmt.Println("--------------------------------")

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatalf("无法获取公钥")
	}

	// 获取地址
	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	// 转换为十六进制字符串
	privateKeyHex := hex.EncodeToString(crypto.FromECDSA(privateKey))
	publicKeyHex := hex.EncodeToString(crypto.FromECDSAPub(publicKeyECDSA))

	// 3. 显示钱包信息
	displayWalletInfo(address.Hex(), privateKeyHex, publicKeyHex)

	// 4. 安全提示
	displaySecurityWarning()

	// 5. 询问是否创建 KeyStore 文件
	fmt.Println("\n💾 KeyStore 文件选项:")
	fmt.Println("================================")
	fmt.Println("KeyStore 文件是加密存储私钥的安全方式")
	fmt.Print("是否要创建 KeyStore 文件? (y/n): ")

	var createKeystore string
	fmt.Scanln(&createKeystore)

	if strings.ToLower(createKeystore) == "y" || strings.ToLower(createKeystore) == "yes" {
		err := createKeystoreFile(privateKey, address.Hex())
		if err != nil {
			fmt.Printf("❌ 创建 KeyStore 文件失败: %v\n", err)
		}
	}

	// 6. 询问是否保存到 .env 文件
	fmt.Println("\n📝 环境配置选项:")
	fmt.Println("================================")
	fmt.Print("是否要将私钥保存到 .env 文件? (y/n): ")

	var saveToEnv string
	fmt.Scanln(&saveToEnv)

	if strings.ToLower(saveToEnv) == "y" || strings.ToLower(saveToEnv) == "yes" {
		err := updateEnvFile(privateKeyHex)
		if err != nil {
			fmt.Printf("❌ 更新 .env 文件失败: %v\n", err)
		} else {
			fmt.Println("✅ 私钥已保存到 .env 文件")
		}
	}

	// 7. 最终提示
	fmt.Println("\n🎉 钱包创建完成!")
	fmt.Println("================================")
	fmt.Println("请务必安全保存您的私钥和助记词!")
	fmt.Println("丢失私钥将无法恢复钱包中的资产!")
}

// displayWalletInfo 显示钱包信息
func displayWalletInfo(address, privateKey, publicKey string) {
	fmt.Println("✅ 新钱包创建成功!")
	fmt.Println("--------------------------------")
	fmt.Printf("钱包地址: %s\n", address)
	fmt.Printf("私钥长度: %d 字符\n", len(privateKey))
	fmt.Printf("公钥长度: %d 字符\n", len(publicKey))

	fmt.Println("\n🔐 私钥信息:")
	fmt.Println("--------------------------------")
	fmt.Printf("私钥 (Hex): %s\n", privateKey)

	fmt.Println("\n🔓 公钥信息:")
	fmt.Println("--------------------------------")
	fmt.Printf("公钥 (Hex): %s\n", publicKey)

	fmt.Println("\n📍 地址信息:")
	fmt.Println("--------------------------------")
	fmt.Printf("地址: %s\n", address)
	fmt.Printf("地址 (小写): %s\n", strings.ToLower(address))
	fmt.Printf("地址校验和: %s\n", crypto.PubkeyToAddress(*getPublicKeyFromHex(publicKey)).Hex())

	// 显示地址的一些特征
	analyzeAddress(address)
}

// analyzeAddress 分析地址特征
func analyzeAddress(address string) {
	fmt.Println("\n🔍 地址特征分析:")
	fmt.Println("--------------------------------")

	// 统计数字和字母
	var digits, letters int
	for _, char := range strings.ToLower(address[2:]) { // 跳过 "0x"
		if char >= '0' && char <= '9' {
			digits++
		} else if char >= 'a' && char <= 'f' {
			letters++
		}
	}

	fmt.Printf("数字字符: %d 个\n", digits)
	fmt.Printf("字母字符: %d 个\n", letters)

	// 检查是否有连续的相同字符
	consecutiveCount := findMaxConsecutive(address[2:])
	if consecutiveCount > 2 {
		fmt.Printf("最长连续相同字符: %d 个\n", consecutiveCount)
	}

	// 检查是否包含常见模式
	checkCommonPatterns(address)
}

// findMaxConsecutive 找到最长连续相同字符
func findMaxConsecutive(s string) int {
	if len(s) == 0 {
		return 0
	}

	maxCount := 1
	currentCount := 1

	for i := 1; i < len(s); i++ {
		if s[i] == s[i-1] {
			currentCount++
		} else {
			if currentCount > maxCount {
				maxCount = currentCount
			}
			currentCount = 1
		}
	}

	if currentCount > maxCount {
		maxCount = currentCount
	}

	return maxCount
}

// checkCommonPatterns 检查常见模式
func checkCommonPatterns(address string) {
	addr := strings.ToLower(address)

	patterns := map[string]string{
		"000":  "包含三个连续的0",
		"111":  "包含三个连续的1",
		"abc":  "包含连续字母abc",
		"123":  "包含连续数字123",
		"dead": "包含单词dead",
		"beef": "包含单词beef",
		"cafe": "包含单词cafe",
		"babe": "包含单词babe",
	}

	for pattern, description := range patterns {
		if strings.Contains(addr, pattern) {
			fmt.Printf("特殊模式: %s\n", description)
		}
	}
}

// getPublicKeyFromHex 从十六进制字符串获取公钥
func getPublicKeyFromHex(publicKeyHex string) *ecdsa.PublicKey {
	publicKeyBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return nil
	}

	publicKey, err := crypto.UnmarshalPubkey(publicKeyBytes)
	if err != nil {
		return nil
	}

	return publicKey
}

// displaySecurityWarning 显示安全警告
func displaySecurityWarning() {
	fmt.Println("\n⚠️  重要安全提示:")
	fmt.Println("================================")
	fmt.Println("🔒 私钥是您钱包的唯一凭证，请务必:")
	fmt.Println("   • 安全保存私钥，不要泄露给任何人")
	fmt.Println("   • 建议使用硬件钱包或冷存储")
	fmt.Println("   • 不要在不安全的网络环境中使用")
	fmt.Println("   • 定期备份，防止数据丢失")
	fmt.Println("   • 这是测试网钱包，不要用于主网大额资产")

	fmt.Println("\n💡 使用建议:")
	fmt.Println("--------------------------------")
	fmt.Println("   • 可以将地址分享给他人接收转账")
	fmt.Println("   • 私钥只有您自己知道")
	fmt.Println("   • 使用 KeyStore 文件可以增加安全性")
	fmt.Println("   • 定期更换钱包以提高安全性")
}

// createKeystoreFile 创建 KeyStore 文件
func createKeystoreFile(privateKey *ecdsa.PrivateKey, address string) error {
	fmt.Println("\n🔐 创建 KeyStore 文件...")
	fmt.Println("--------------------------------")

	// 获取密码
	fmt.Print("请输入 KeyStore 密码: ")
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("读取密码失败: %w", err)
	}
	fmt.Println() // 换行

	fmt.Print("请再次输入密码确认: ")
	confirmPassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("读取确认密码失败: %w", err)
	}
	fmt.Println() // 换行

	if string(password) != string(confirmPassword) {
		return fmt.Errorf("两次输入的密码不一致")
	}

	if len(password) < 8 {
		return fmt.Errorf("密码长度至少需要8个字符")
	}

	// 创建 keystore 目录
	keystoreDir := "keystore"
	if err := os.MkdirAll(keystoreDir, 0755); err != nil {
		return fmt.Errorf("创建 keystore 目录失败: %w", err)
	}

	// 创建 KeyStore
	ks := keystore.NewKeyStore(keystoreDir, keystore.StandardScryptN, keystore.StandardScryptP)

	// 导入私钥
	account, err := ks.ImportECDSA(privateKey, string(password))
	if err != nil {
		return fmt.Errorf("创建 KeyStore 失败: %w", err)
	}

	fmt.Printf("✅ KeyStore 文件创建成功!\n")
	fmt.Printf("文件路径: %s\n", account.URL.Path)
	fmt.Printf("账户地址: %s\n", account.Address.Hex())

	// 显示 KeyStore 文件信息
	displayKeystoreInfo(account.URL.Path)

	return nil
}

// displayKeystoreInfo 显示 KeyStore 文件信息
func displayKeystoreInfo(keystorePath string) {
	fmt.Println("\n📁 KeyStore 文件信息:")
	fmt.Println("--------------------------------")

	// 获取文件信息
	fileInfo, err := os.Stat(keystorePath)
	if err != nil {
		fmt.Printf("❌ 获取文件信息失败: %v\n", err)
		return
	}

	fmt.Printf("文件名: %s\n", filepath.Base(keystorePath))
	fmt.Printf("文件大小: %d 字节\n", fileInfo.Size())
	fmt.Printf("创建时间: %s\n", fileInfo.ModTime().Format("2006-01-02 15:04:05"))

	fmt.Println("\n💡 KeyStore 使用说明:")
	fmt.Println("--------------------------------")
	fmt.Println("   • KeyStore 文件包含加密的私钥")
	fmt.Println("   • 需要密码才能解锁使用")
	fmt.Println("   • 可以导入到 MetaMask 等钱包")
	fmt.Println("   • 请安全保存文件和密码")
}

// updateEnvFile 更新 .env 文件
func updateEnvFile(privateKey string) error {
	envPath := ".env"

	// 读取现有的 .env 文件
	var lines []string
	if data, err := os.ReadFile(envPath); err == nil {
		lines = strings.Split(string(data), "\n")
	}

	// 查找并更新 PRIVATE_KEY 行
	privateKeyUpdated := false
	for i, line := range lines {
		if strings.HasPrefix(line, "PRIVATE_KEY=") {
			lines[i] = fmt.Sprintf("PRIVATE_KEY=%s", privateKey)
			privateKeyUpdated = true
			break
		}
	}

	// 如果没有找到 PRIVATE_KEY 行，添加一行
	if !privateKeyUpdated {
		lines = append(lines, fmt.Sprintf("PRIVATE_KEY=%s", privateKey))
	}

	// 写回文件
	content := strings.Join(lines, "\n")
	return os.WriteFile(envPath, []byte(content), 0644)
}
