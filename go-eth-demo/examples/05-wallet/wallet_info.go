package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/local/go-eth-demo/config"
	"github.com/local/go-eth-demo/utils"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 创建以太坊客户端
	ethClient, err := utils.NewEthClient(cfg)
	if err != nil {
		log.Fatalf("创建以太坊客户端失败: %v", err)
	}
	defer ethClient.Close()

	ctx := context.Background()

	fmt.Println("💼 钱包信息查看工具")
	fmt.Println("================================")

	// 检查是否配置了私钥
	if !cfg.HasPrivateKey() {
		fmt.Println("⚠️  未配置私钥，将使用示例地址演示")

		// 使用示例地址
		exampleAddress := "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045" // Vitalik's address
		err := displayAddressInfo(ctx, ethClient, exampleAddress)
		if err != nil {
			fmt.Printf("❌ 获取地址信息失败: %v\n", err)
		}
		return
	}

	fmt.Println("🔑 检测到私钥配置，分析当前钱包...")

	// 解析私钥
	privateKey, err := crypto.HexToECDSA(cfg.PrivateKey)
	if err != nil {
		log.Fatalf("解析私钥失败: %v", err)
	}

	// 获取钱包信息
	walletInfo := extractWalletInfo(privateKey)

	// 显示钱包基本信息
	displayWalletBasicInfo(walletInfo)

	// 获取链上信息
	err = displayOnChainInfo(ctx, ethClient, walletInfo.Address)
	if err != nil {
		fmt.Printf("❌ 获取链上信息失败: %v\n", err)
	}

	// 分析钱包安全性
	analyzeWalletSecurity(walletInfo)
}

// WalletInfo 钱包信息结构
type WalletInfo struct {
	Address    string
	PrivateKey string
	PublicKey  string
}

// extractWalletInfo 提取钱包信息
func extractWalletInfo(privateKey *ecdsa.PrivateKey) *WalletInfo {
	// 获取公钥
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

	return &WalletInfo{
		Address:    address.Hex(),
		PrivateKey: privateKeyHex,
		PublicKey:  publicKeyHex,
	}
}

// displayWalletBasicInfo 显示钱包基本信息
func displayWalletBasicInfo(wallet *WalletInfo) {
	fmt.Println("\n📋 钱包基本信息:")
	fmt.Println("================================")
	fmt.Printf("钱包地址: %s\n", wallet.Address)
	fmt.Printf("私钥长度: %d 字符\n", len(wallet.PrivateKey))
	fmt.Printf("公钥长度: %d 字符\n", len(wallet.PublicKey))

	fmt.Println("\n🔐 密钥信息:")
	fmt.Println("--------------------------------")
	fmt.Printf("私钥: %s\n", wallet.PrivateKey)
	fmt.Printf("公钥: %s\n", wallet.PublicKey)

	// 分析地址特征
	analyzeAddressFeatures(wallet.Address)
}

// displayOnChainInfo 显示链上信息
func displayOnChainInfo(ctx context.Context, ethClient *utils.EthClient, address string) error {
	fmt.Println("\n⛓️  链上信息:")
	fmt.Println("================================")

	addr := common.HexToAddress(address)

	// 1. 获取 ETH 余额
	balance, err := ethClient.GetClient().BalanceAt(ctx, addr, nil)
	if err != nil {
		return fmt.Errorf("获取余额失败: %w", err)
	}

	fmt.Printf("ETH 余额: %s ETH\n", utils.WeiToEther(balance))

	// 分析余额等级
	analyzeBalanceLevel(balance)

	// 2. 获取交易计数 (nonce)
	nonce, err := ethClient.GetClient().NonceAt(ctx, addr, nil)
	if err != nil {
		return fmt.Errorf("获取 nonce 失败: %w", err)
	}

	fmt.Printf("交易计数 (Nonce): %d\n", nonce)

	// 分析账户活跃度
	analyzeAccountActivity(nonce)

	// 3. 检查是否为合约地址
	code, err := ethClient.GetClient().CodeAt(ctx, addr, nil)
	if err != nil {
		return fmt.Errorf("获取合约代码失败: %w", err)
	}

	if len(code) > 0 {
		fmt.Printf("账户类型: 智能合约\n")
		fmt.Printf("合约代码长度: %d 字节\n", len(code))
	} else {
		fmt.Printf("账户类型: 外部账户 (EOA)\n")
	}

	// 4. 获取网络信息
	return displayNetworkInfo(ctx, ethClient)
}

// displayAddressInfo 显示地址信息（无私钥）
func displayAddressInfo(ctx context.Context, ethClient *utils.EthClient, address string) error {
	fmt.Printf("\n📍 地址信息: %s\n", address)
	fmt.Println("================================")

	return displayOnChainInfo(ctx, ethClient, address)
}

// analyzeAddressFeatures 分析地址特征
func analyzeAddressFeatures(address string) {
	fmt.Println("\n🔍 地址特征分析:")
	fmt.Println("--------------------------------")

	// 移除 0x 前缀
	addr := address[2:]

	// 统计字符类型
	var digits, letters, uppercase, lowercase int
	for _, char := range addr {
		if char >= '0' && char <= '9' {
			digits++
		} else if char >= 'A' && char <= 'F' {
			letters++
			uppercase++
		} else if char >= 'a' && char <= 'f' {
			letters++
			lowercase++
		}
	}

	fmt.Printf("数字字符: %d 个 (%.1f%%)\n", digits, float64(digits)/40*100)
	fmt.Printf("字母字符: %d 个 (%.1f%%)\n", letters, float64(letters)/40*100)
	fmt.Printf("大写字母: %d 个\n", uppercase)
	fmt.Printf("小写字母: %d 个\n", lowercase)

	// 检查校验和格式
	if uppercase > 0 && lowercase > 0 {
		fmt.Println("✅ 使用 EIP-55 校验和格式")
	} else if uppercase == 0 {
		fmt.Println("📝 全小写格式")
	} else {
		fmt.Println("📝 全大写格式")
	}

	// 查找特殊模式
	findSpecialPatterns(addr)

	// 计算地址"美观度"
	calculateAddressBeauty(addr)
}

// findSpecialPatterns 查找特殊模式
func findSpecialPatterns(addr string) {
	fmt.Println("\n🎨 特殊模式:")
	fmt.Println("--------------------------------")

	patterns := map[string]string{
		"000":  "三个连续的0",
		"111":  "三个连续的1",
		"aaa":  "三个连续的a",
		"fff":  "三个连续的f",
		"123":  "连续数字123",
		"abc":  "连续字母abc",
		"dead": "单词dead",
		"beef": "单词beef",
		"cafe": "单词cafe",
		"babe": "单词babe",
		"face": "单词face",
		"deed": "单词deed",
		"feed": "单词feed",
		"fade": "单词fade",
	}

	foundPatterns := 0
	for pattern, description := range patterns {
		if containsPattern(addr, pattern) {
			fmt.Printf("🎯 发现: %s\n", description)
			foundPatterns++
		}
	}

	if foundPatterns == 0 {
		fmt.Println("📝 未发现特殊模式")
	}

	// 检查连续相同字符
	maxConsecutive := findMaxConsecutiveChars(addr)
	if maxConsecutive > 3 {
		fmt.Printf("🔗 最长连续相同字符: %d 个\n", maxConsecutive)
	}
}

// containsPattern 检查是否包含模式
func containsPattern(s, pattern string) bool {
	return len(s) >= len(pattern) &&
		(s[:len(pattern)] == pattern ||
			s[len(s)-len(pattern):] == pattern ||
			containsSubstring(s, pattern))
}

// containsSubstring 检查子字符串
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// findMaxConsecutiveChars 找到最长连续相同字符
func findMaxConsecutiveChars(s string) int {
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

// calculateAddressBeauty 计算地址美观度
func calculateAddressBeauty(addr string) {
	fmt.Println("\n✨ 地址美观度评分:")
	fmt.Println("--------------------------------")

	score := 0
	reasons := []string{}

	// 检查开头和结尾
	if addr[:4] == "0000" {
		score += 20
		reasons = append(reasons, "开头四个0 (+20分)")
	} else if addr[:3] == "000" {
		score += 15
		reasons = append(reasons, "开头三个0 (+15分)")
	} else if addr[:2] == "00" {
		score += 10
		reasons = append(reasons, "开头两个0 (+10分)")
	}

	if addr[len(addr)-4:] == "0000" {
		score += 20
		reasons = append(reasons, "结尾四个0 (+20分)")
	} else if addr[len(addr)-3:] == "000" {
		score += 15
		reasons = append(reasons, "结尾三个0 (+15分)")
	} else if addr[len(addr)-2:] == "00" {
		score += 10
		reasons = append(reasons, "结尾两个0 (+10分)")
	}

	// 检查重复模式
	maxConsecutive := findMaxConsecutiveChars(addr)
	if maxConsecutive >= 6 {
		score += 25
		reasons = append(reasons, fmt.Sprintf("连续%d个相同字符 (+25分)", maxConsecutive))
	} else if maxConsecutive >= 4 {
		score += 15
		reasons = append(reasons, fmt.Sprintf("连续%d个相同字符 (+15分)", maxConsecutive))
	}

	// 检查对称性
	if isSymmetric(addr) {
		score += 30
		reasons = append(reasons, "地址对称 (+30分)")
	}

	fmt.Printf("总分: %d/100\n", score)

	if score >= 50 {
		fmt.Println("🌟 这是一个非常美观的地址!")
	} else if score >= 25 {
		fmt.Println("✨ 这是一个比较美观的地址")
	} else if score > 0 {
		fmt.Println("💫 这个地址有一些特色")
	} else {
		fmt.Println("📝 这是一个普通的地址")
	}

	for _, reason := range reasons {
		fmt.Printf("   • %s\n", reason)
	}
}

// isSymmetric 检查是否对称
func isSymmetric(s string) bool {
	length := len(s)
	for i := 0; i < length/2; i++ {
		if s[i] != s[length-1-i] {
			return false
		}
	}
	return true
}

// analyzeBalanceLevel 分析余额等级
func analyzeBalanceLevel(balance *big.Int) {
	fmt.Println("\n💰 余额分析:")
	fmt.Println("--------------------------------")

	// 转换为 ETH
	balanceETH := utils.WeiToEther(balance)

	// 解析为浮点数进行比较
	balanceFloat := new(big.Float)
	balanceFloat.SetString(balanceETH)

	// 定义等级
	levels := []struct {
		threshold float64
		name      string
		emoji     string
	}{
		{100, "鲸鱼级", "🐋"},
		{10, "大户级", "🦈"},
		{1, "中户级", "🐟"},
		{0.1, "小户级", "🐠"},
		{0.01, "微户级", "🦐"},
		{0, "新手级", "🥚"},
	}

	balanceValue, _ := balanceFloat.Float64()

	for _, level := range levels {
		if balanceValue >= level.threshold {
			fmt.Printf("等级: %s %s\n", level.emoji, level.name)
			break
		}
	}

	// 计算美元价值（假设 ETH = $2000）
	ethPrice := 2000.0
	usdValue := balanceValue * ethPrice
	fmt.Printf("估算价值: $%.2f (按 ETH=$%.0f 计算)\n", usdValue, ethPrice)
}

// analyzeAccountActivity 分析账户活跃度
func analyzeAccountActivity(nonce uint64) {
	fmt.Println("\n📊 账户活跃度:")
	fmt.Println("--------------------------------")

	if nonce == 0 {
		fmt.Println("状态: 🆕 全新账户 (未发送过交易)")
	} else if nonce < 10 {
		fmt.Println("状态: 🌱 新手账户 (交易较少)")
	} else if nonce < 100 {
		fmt.Println("状态: 🌿 活跃账户 (有一定交易量)")
	} else if nonce < 1000 {
		fmt.Println("状态: 🌳 高活跃账户 (交易频繁)")
	} else {
		fmt.Println("状态: 🏭 超高活跃账户 (可能是机器人或交易所)")
	}

	fmt.Printf("历史交易数: %d 笔\n", nonce)
}

// displayNetworkInfo 显示网络信息
func displayNetworkInfo(ctx context.Context, ethClient *utils.EthClient) error {
	fmt.Println("\n🌐 网络信息:")
	fmt.Println("--------------------------------")

	// 获取链 ID
	chainID, err := ethClient.GetClient().ChainID(ctx)
	if err != nil {
		return fmt.Errorf("获取链 ID 失败: %w", err)
	}

	// 获取最新区块号
	blockNumber, err := ethClient.GetClient().BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("获取区块号失败: %w", err)
	}

	// 获取 Gas 价格
	gasPrice, err := ethClient.GetClient().SuggestGasPrice(ctx)
	if err != nil {
		return fmt.Errorf("获取 Gas 价格失败: %w", err)
	}

	fmt.Printf("链 ID: %s\n", chainID.String())
	fmt.Printf("最新区块: %d\n", blockNumber)
	fmt.Printf("当前 Gas 价格: %s Gwei\n", utils.WeiToGwei(gasPrice))

	// 识别网络
	networkName := getNetworkName(chainID.Uint64())
	fmt.Printf("网络名称: %s\n", networkName)

	return nil
}

// getNetworkName 获取网络名称
func getNetworkName(chainID uint64) string {
	networks := map[uint64]string{
		1:        "以太坊主网",
		11155111: "Sepolia 测试网",
		5:        "Goerli 测试网",
		137:      "Polygon 主网",
		80001:    "Mumbai 测试网",
		56:       "BSC 主网",
		97:       "BSC 测试网",
	}

	if name, exists := networks[chainID]; exists {
		return name
	}
	return fmt.Sprintf("未知网络 (ID: %d)", chainID)
}

// analyzeWalletSecurity 分析钱包安全性
func analyzeWalletSecurity(wallet *WalletInfo) {
	fmt.Println("\n🔒 钱包安全性分析:")
	fmt.Println("================================")

	// 检查私钥强度
	checkPrivateKeyStrength(wallet.PrivateKey)

	// 安全建议
	fmt.Println("\n💡 安全建议:")
	fmt.Println("--------------------------------")
	fmt.Println("✅ 定期备份私钥和助记词")
	fmt.Println("✅ 使用硬件钱包存储大额资产")
	fmt.Println("✅ 不要在不安全的网络环境中使用")
	fmt.Println("✅ 定期检查账户活动")
	fmt.Println("✅ 使用多重签名钱包增加安全性")
	fmt.Println("⚠️  永远不要分享您的私钥")
}

// checkPrivateKeyStrength 检查私钥强度
func checkPrivateKeyStrength(privateKeyHex string) {
	fmt.Println("🔐 私钥强度分析:")
	fmt.Println("--------------------------------")

	// 检查长度
	if len(privateKeyHex) == 64 {
		fmt.Println("✅ 私钥长度正确 (64字符)")
	} else {
		fmt.Printf("❌ 私钥长度异常 (%d字符)\n", len(privateKeyHex))
	}

	// 检查字符分布
	charCount := make(map[rune]int)
	for _, char := range privateKeyHex {
		charCount[char]++
	}

	// 计算熵值
	entropy := 0.0
	length := float64(len(privateKeyHex))
	for _, count := range charCount {
		if count > 0 {
			p := float64(count) / length
			entropy -= p * (3.321928 * p) // 简化的熵计算
		}
	}

	fmt.Printf("字符种类: %d 种\n", len(charCount))
	fmt.Printf("熵值估算: %.2f\n", entropy)

	if entropy > 3.5 {
		fmt.Println("✅ 私钥随机性良好")
	} else if entropy > 3.0 {
		fmt.Println("⚠️  私钥随机性一般")
	} else {
		fmt.Println("❌ 私钥随机性较差，建议重新生成")
	}
}
