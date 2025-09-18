# Go 以太坊客户端学习项目

这是一个系统学习以太坊 Go 客户端 (ethclient) 包的实践项目。

## 🚀 快速开始

### 1. 环境准备

确保您已安装：
- Go 1.19 或更高版本
- Node.js (用于安装 solc)

### 2. 安装依赖

```bash
# 安装 Go 依赖
go mod tidy

# 安装 Solidity 编译器
npm install -g solc

# 安装 abigen 工具
go install github.com/ethereum/go-ethereum/cmd/abigen@latest
```

### 3. 配置环境

1. 复制环境变量模板：
```bash
cp .env.example .env
```

2. 编辑 `.env` 文件，填入您的 Alchemy API Key：
```bash
ETHEREUM_RPC_URL=https://eth-sepolia.g.alchemy.com/v2/YOUR_API_KEY_HERE
```

### 4. 运行第一个示例

```bash
# 测试连接
go run examples/01-basic/connect.go

# 查询网络信息
go run examples/01-basic/network_info.go
```

## 📁 项目结构

```
go-eth-demo/
├── .env.example            # 环境变量模板
├── .gitignore             # Git 忽略文件
├── go.mod                 # Go 模块文件
├── go.sum                 # 依赖版本锁定
├── README.md              # 项目说明
├── config/
│   └── config.go          # 配置管理
├── utils/
│   └── client.go          # 以太坊客户端工具
└── examples/
    └── 01-basic/
        ├── connect.go     # 基础连接示例
        └── network_info.go # 网络信息查询
```

## 🎯 学习阶段

### 阶段一：基础连接 ✅
- [x] 环境配置
- [x] 客户端连接
- [x] 网络信息查询

### 阶段二：数据查询 (进行中)
- [ ] 区块查询
- [ ] 交易查询
- [ ] 账户余额查询

### 阶段三：交易操作
- [ ] ETH 转账
- [ ] Gas 费用管理
- [ ] 交易状态监控

### 阶段四：智能合约
- [ ] 合约部署
- [ ] 合约调用
- [ ] 事件处理

### 阶段五：高级功能
- [ ] 事件订阅
- [ ] 批量操作
- [ ] 性能优化

## 🔒 安全注意事项

- **永远不要**将私钥或 API Key 提交到版本控制系统
- 使用 `.env` 文件存储敏感信息
- 在生产环境中使用更安全的密钥管理方案

## 📚 学习资源

- [以太坊官方文档](https://ethereum.org/developers/)
- [Go-Ethereum 文档](https://geth.ethereum.org/docs/)
- [Solidity 文档](https://docs.soliditylang.org/)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License