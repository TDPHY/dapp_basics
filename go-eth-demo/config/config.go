package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config 存储应用程序配置
type Config struct {
	// 以太坊网络配置
	EthereumRPCURL string // RPC 节点地址
	ChainID        int64  // 链 ID (Sepolia: 11155111)
	NetworkName    string // 网络名称 (sepolia)

	// 账户配置 (后续阶段使用)
	PrivateKey       string // 私钥 (用于交易签名)
	KeystorePath     string // KeyStore 文件路径
	KeystorePassword string // KeyStore 密码
}

// LoadConfig 从环境变量加载配置
func LoadConfig() (*Config, error) {
	// 尝试加载 .env 文件
	if err := godotenv.Load(); err != nil {
		// .env 文件不存在不是错误，可能使用系统环境变量
		fmt.Println("Warning: .env file not found, using system environment variables")
	}

	config := &Config{
		EthereumRPCURL:   getEnv("ETHEREUM_RPC_URL", ""),
		ChainID:          getEnvAsInt64("CHAIN_ID", 11155111), // Sepolia 默认
		NetworkName:      getEnv("NETWORK_NAME", "sepolia"),
		PrivateKey:       getEnv("PRIVATE_KEY", ""),
		KeystorePath:     getEnv("KEYSTORE_PATH", ""),
		KeystorePassword: getEnv("KEYSTORE_PASSWORD", ""),
	}

	// 验证必需的配置
	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// validate 验证配置的有效性
func (c *Config) validate() error {
	// 1. 检查 RPC URL 是否存在
	if c.EthereumRPCURL == "" {
		return fmt.Errorf("ETHEREUM_RPC_URL is required")
	}

	// 2. 验证 URL 格式
	if !strings.HasPrefix(c.EthereumRPCURL, "http://") &&
		!strings.HasPrefix(c.EthereumRPCURL, "https://") &&
		!strings.HasPrefix(c.EthereumRPCURL, "ws://") &&
		!strings.HasPrefix(c.EthereumRPCURL, "wss://") {
		return fmt.Errorf("invalid RPC URL format: %s", c.EthereumRPCURL)
	}

	// 3. 验证链 ID
	if c.ChainID <= 0 {
		return fmt.Errorf("invalid chain ID: %d", c.ChainID)
	}

	return nil
}

// GetNetworkInfo 返回网络信息摘要
func (c *Config) GetNetworkInfo() string {
	return fmt.Sprintf("Network: %s (Chain ID: %d)", c.NetworkName, c.ChainID)
}

// HasPrivateKey 检查是否配置了私钥
func (c *Config) HasPrivateKey() bool {
	return c.PrivateKey != ""
}

// HasKeystore 检查是否配置了 KeyStore
func (c *Config) HasKeystore() bool {
	return c.KeystorePath != "" && c.KeystorePassword != ""
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt64 获取环境变量并转换为 int64
func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}
