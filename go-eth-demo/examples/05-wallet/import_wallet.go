package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/term"
)

func main() {
	fmt.Println("📥 以太坊钱包导入工具")
	fmt.Println("================================")

	fmt.Println("请选择导入方式:")
	fmt.Println("1. 从私钥导入")
	fmt.Println("2. 从 KeyStore 文件导入")
	fmt.Print("请输入选择 (1 或 2): ")

	var choice string
	fmt.Scanln(&choice)

	switch choice {
	case "1":
		importFromPrivateKey()
	case "2":
		importFromKeystore()
	default:
		fmt.Println("❌ 无效选择")
		return
	}
}

// importFromPrivateKey 从私钥导入钱包
func importFromPrivateKey() {
	fmt.Println("\n🔑 从私钥导入钱包")
	fmt.Println("================================")

	fmt.Print("请输入私钥 (不包含 0x 前缀): ")
	var privateKeyHex string
	fmt.Scanln(&privateKeyHex)

	// 清理输入
	privateKeyHex = strings.TrimSpace(privateKeyHex)
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")

	// 验证私钥格式
	if len(privateKeyHex) != 64 {
		fmt.Printf("❌ 私钥长度错误，应为64个字符，当前为%d个字符\n", len(privateKeyHex))
		return
	}

	// 解析私钥
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		fmt.Printf("❌ 私钥格式错误: %v\n", err)
		return
	}

	// 获取公钥和地址
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatalf("无法获取公钥")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	publicKeyHex := hex.EncodeToString(crypto.FromECDSAPub(publicKeyECDSA))

	// 显示钱包信息
	fmt.Println("\n✅ 钱包导入成功!")
	fmt.Println("--------------------------------")
	fmt.Printf("钱包地址: %s\n", address.Hex())
	fmt.Printf("私钥: %s\n", privateKeyHex)
	fmt.Printf("公钥: %s\n", publicKeyHex)

	// 验证私钥
	validatePrivateKey(privateKey, address.Hex())

	// 询问是否创建 KeyStore 文件
	fmt.Print("\n是否要为此钱包创建 KeyStore 文件? (y/n): ")
	var createKeystore string
	fmt.Scanln(&createKeystore)

	if strings.ToLower(createKeystore) == "y" || strings.ToLower(createKeystore) == "yes" {
		err := createKeystoreFromPrivateKey(privateKey, address.Hex())
		if err != nil {
			fmt.Printf("❌ 创建 KeyStore 文件失败: %v\n", err)
		}
	}

	// 询问是否保存到 .env 文件
	fmt.Print("\n是否要将私钥保存到 .env 文件? (y/n): ")
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
}

// importFromKeystore 从 KeyStore 文件导入钱包
func importFromKeystore() {
	fmt.Println("\n📁 从 KeyStore 文件导入钱包")
	fmt.Println("================================")

	fmt.Print("请输入 KeyStore 文件路径: ")
	var keystorePath string
	fmt.Scanln(&keystorePath)

	// 检查文件是否存在
	if _, err := os.Stat(keystorePath); os.IsNotExist(err) {
		fmt.Printf("❌ KeyStore 文件不存在: %s\n", keystorePath)
		return
	}

	// 读取 KeyStore 文件
	keystoreData, err := os.ReadFile(keystorePath)
	if err != nil {
		fmt.Printf("❌ 读取 KeyStore 文件失败: %v\n", err)
		return
	}

	// 获取密码
	fmt.Print("请输入 KeyStore 密码: ")
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Printf("❌ 读取密码失败: %v\n", err)
		return
	}
	fmt.Println() // 换行

	// 解密 KeyStore
	key, err := keystore.DecryptKey(keystoreData, string(password))
	if err != nil {
		fmt.Printf("❌ 解密 KeyStore 失败: %v\n", err)
		fmt.Println("请检查密码是否正确")
		return
	}

	// 获取钱包信息
	address := key.Address.Hex()
	privateKeyHex := hex.EncodeToString(crypto.FromECDSA(key.PrivateKey))

	publicKey := key.PrivateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatalf("无法获取公钥")
	}
	publicKeyHex := hex.EncodeToString(crypto.FromECDSAPub(publicKeyECDSA))

	// 显示钱包信息
	fmt.Println("\n✅ KeyStore 解密成功!")
	fmt.Println("--------------------------------")
	fmt.Printf("钱包地址: %s\n", address)
	fmt.Printf("私钥: %s\n", privateKeyHex)
	fmt.Printf("公钥: %s\n", publicKeyHex)

	// 显示 KeyStore 文件信息
	displayKeystoreFileInfo(keystorePath)

	// 验证私钥
	validatePrivateKey(key.PrivateKey, address)

	// 询问是否保存到 .env 文件
	fmt.Print("\n是否要将私钥保存到 .env 文件? (y/n): ")
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
}

// validatePrivateKey 验证私钥
func validatePrivateKey(privateKey *ecdsa.PrivateKey, expectedAddress string) {
	fmt.Println("\n🔍 验证私钥...")
	fmt.Println("--------------------------------")

	// 从私钥重新生成地址
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		fmt.Println("❌ 无法获取公钥")
		return
	}

	derivedAddress := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()

	if strings.EqualFold(derivedAddress, expectedAddress) {
		fmt.Println("✅ 私钥验证成功")
		fmt.Printf("地址匹配: %s\n", derivedAddress)
	} else {
		fmt.Println("❌ 私钥验证失败")
		fmt.Printf("期望地址: %s\n", expectedAddress)
		fmt.Printf("实际地址: %s\n", derivedAddress)
	}

	// 检查私钥强度
	checkPrivateKeyStrength(privateKey)
}

// checkPrivateKeyStrength 检查私钥强度
func checkPrivateKeyStrength(privateKey *ecdsa.PrivateKey) {
	fmt.Println("\n🔒 私钥安全性分析:")
	fmt.Println("--------------------------------")

	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex := hex.EncodeToString(privateKeyBytes)

	// 检查是否包含过多的0
	zeroCount := strings.Count(privateKeyHex, "0")
	if zeroCount > 20 {
		fmt.Printf("⚠️  私钥包含较多的0字符 (%d个)，可能安全性较低\n", zeroCount)
	} else {
		fmt.Printf("✅ 私钥0字符数量正常 (%d个)\n", zeroCount)
	}

	// 检查是否有重复模式
	if hasRepeatingPattern(privateKeyHex) {
		fmt.Println("⚠️  私钥可能包含重复模式")
	} else {
		fmt.Println("✅ 私钥无明显重复模式")
	}

	// 检查熵值
	entropy := calculateEntropy(privateKeyHex)
	fmt.Printf("私钥熵值: %.2f (满分4.0)\n", entropy)

	if entropy > 3.5 {
		fmt.Println("✅ 私钥熵值良好")
	} else if entropy > 3.0 {
		fmt.Println("⚠️  私钥熵值一般")
	} else {
		fmt.Println("❌ 私钥熵值较低，建议重新生成")
	}
}

// hasRepeatingPattern 检查是否有重复模式
func hasRepeatingPattern(s string) bool {
	// 检查连续重复的字符
	for i := 0; i < len(s)-3; i++ {
		if s[i] == s[i+1] && s[i+1] == s[i+2] && s[i+2] == s[i+3] {
			return true
		}
	}

	// 检查简单的重复模式
	patterns := []string{"0123", "abcd", "1234", "0000", "1111", "aaaa", "ffff"}
	for _, pattern := range patterns {
		if strings.Contains(s, pattern) {
			return true
		}
	}

	return false
}

// calculateEntropy 计算熵值
func calculateEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}

	// 统计字符频率
	freq := make(map[rune]int)
	for _, char := range s {
		freq[char]++
	}

	// 计算熵值
	var entropy float64
	length := float64(len(s))

	for _, count := range freq {
		if count > 0 {
			p := float64(count) / length
			entropy -= p * logBase2(p)
		}
	}

	return entropy
}

// logBase2 计算以2为底的对数
func logBase2(x float64) float64 {
	return 0.693147180559945309417 * 1.4426950408889634074 *
		(x - 1) / (x + 1) // 简化的对数计算
}

// displayKeystoreFileInfo 显示 KeyStore 文件信息
func displayKeystoreFileInfo(keystorePath string) {
	fmt.Println("\n📁 KeyStore 文件信息:")
	fmt.Println("--------------------------------")

	fileInfo, err := os.Stat(keystorePath)
	if err != nil {
		fmt.Printf("❌ 获取文件信息失败: %v\n", err)
		return
	}

	fmt.Printf("文件路径: %s\n", keystorePath)
	fmt.Printf("文件大小: %d 字节\n", fileInfo.Size())
	fmt.Printf("修改时间: %s\n", fileInfo.ModTime().Format("2006-01-02 15:04:05"))

	// 尝试读取并分析 KeyStore 内容
	data, err := os.ReadFile(keystorePath)
	if err == nil {
		fmt.Printf("文件内容长度: %d 字符\n", len(data))

		// 检查是否包含标准 KeyStore 字段
		content := string(data)
		if strings.Contains(content, "\"crypto\"") || strings.Contains(content, "\"Crypto\"") {
			fmt.Println("✅ 标准 KeyStore 格式")
		}
		if strings.Contains(content, "\"version\"") {
			fmt.Println("✅ 包含版本信息")
		}
		if strings.Contains(content, "\"address\"") {
			fmt.Println("✅ 包含地址信息")
		}
	}
}

// createKeystoreFromPrivateKey 从私钥创建 KeyStore 文件
func createKeystoreFromPrivateKey(privateKey *ecdsa.PrivateKey, address string) error {
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

	return nil
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
