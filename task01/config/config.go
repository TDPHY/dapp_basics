package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config 存储应用程序配置
type Config struct {
	EthereumRPCURL string
	ChainID        string
	NetworkName    string
	PrivateKey     string
	ToAddress      string
}

// LoadConfig 从环境变量加载配置
func LoadConfig() *Config {
	// 尝试加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Println("未找到 .env 文件，使用系统环境变量")
	}

	config := &Config{
		EthereumRPCURL: getEnv("ETHEREUM_RPC_URL", ""),
		ChainID:        getEnv("CHAIN_ID", "11155111"),
		NetworkName:    getEnv("NETWORK_NAME", "sepolia"),
		PrivateKey:     getEnv("PRIVATE_KEY", ""),
		ToAddress:      getEnv("TO_ADDRESS", ""),
	}

	// 验证必需的配置
	if config.EthereumRPCURL == "" {
		log.Fatal("ETHEREUM_RPC_URL 环境变量未设置")
	}

	return config
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
