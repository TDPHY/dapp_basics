# Go语言区块链DApp开发项目

## 项目概述

本项目实现了两个主要任务：
1. **区块链读写操作** - 连接Sepolia测试网络，查询区块信息，发送ETH交易
2. **智能合约交互** - 使用abigen工具生成Go绑定代码，实现合约方法调用

## 项目结构

```
task01/
├── main.go                 # 主程序入口（交互式菜单）
├── test_basic.go          # 基础功能测试程序
├── go.mod                 # Go模块定义
├── .env                   # 环境配置文件
├── .env.example          # 配置文件模板
├── README.md             # 项目说明文档
├── blockchain/           # 区块链操作模块
│   ├── client.go         # 以太坊客户端连接
│   ├── query.go          # 区块查询功能
│   └── transaction.go    # 交易发送功能
├── config/               # 配置管理
│   └── config.go         # 环境变量加载
├── contracts/            # 智能合约相关
│   ├── Counter.sol       # Solidity合约源码
│   ├── Counter.go        # abigen生成的Go绑定
│   └── interact.go       # 合约交互封装
└── scripts/              # 编译脚本
    └── compile_contract.js # 合约编译脚本
```

## 功能特性

### 任务1：区块链读写操作
- ✅ 连接Sepolia测试网络
- ✅ 查询最新区块信息
- ✅ 查询指定区块号的区块信息
- ✅ 查询账户ETH余额
- ✅ 发送ETH转账交易

### 任务2：智能合约交互
- ✅ Solidity智能合约开发（Counter合约）
- ✅ 使用abigen生成Go绑定代码
- ✅ 智能合约部署功能
- ✅ 合约方法调用（increment/decrement/getCount）

## 快速开始

### 1. 环境准备

确保已安装：
- Go 1.21+
- Node.js (用于合约编译)
- abigen工具

### 2. 配置环境变量

复制并编辑配置文件：
```bash
cp .env.example .env
```

编辑`.env`文件，填入：
- `ETHEREUM_RPC_URL`: Sepolia RPC端点（可使用Infura或公共端点）
- `PRIVATE_KEY`: 用于发送交易的私钥（可选，仅发送交易时需要）

### 3. 安装依赖

```bash
go mod tidy
npm install
```

### 4. 运行测试

基础功能测试：
```bash
go run test_basic.go
```

完整交互式程序：
```bash
go run main.go
```

## 使用说明

### 基础测试程序
`test_basic.go` 提供了基础功能的自动化测试：
- 配置加载验证
- 区块链连接测试
- 区块查询功能测试
- 余额查询功能测试

### 主程序
`main.go` 提供了完整的交互式菜单：
1. 查询最新区块信息
2. 查询指定区块信息
3. 查询账户余额
4. 发送ETH转账
5. 部署智能合约
6. 调用合约方法
7. 完整工作流演示

## 技术栈

- **语言**: Go 1.21
- **区块链库**: go-ethereum (ethclient)
- **网络**: Ethereum Sepolia测试网
- **合约语言**: Solidity
- **工具**: abigen, solc
- **配置管理**: godotenv

## 注意事项

1. **网络连接**: 确保网络能访问Sepolia测试网RPC端点
2. **私钥安全**: 不要在生产环境中硬编码私钥
3. **测试网ETH**: 发送交易需要Sepolia测试网ETH，可从水龙头获取
4. **Gas费用**: 合约部署和调用需要消耗Gas费用

## 故障排除

### 连接问题
- 检查RPC URL是否正确
- 确认网络连接正常
- 尝试使用不同的RPC端点

### 编译问题
- 运行 `go mod tidy` 更新依赖
- 确保Go版本为1.21+
- 检查GOPATH和GOROOT设置

### 交易问题
- 确认私钥格式正确（不包含0x前缀）
- 检查账户是否有足够的ETH余额
- 确认Gas价格和限制设置合理

## 开发者信息

本项目演示了Go语言在区块链开发中的应用，包括：
- 以太坊网络交互
- 智能合约编译和部署
- Go代码生成和绑定
- 区块链数据查询和交易处理

适合学习区块链开发和Go语言应用的开发者参考。