# 🔧 智能合约加载指南

本目录演示了如何在Go中加载和管理已部署的智能合约。

## 📁 文件说明

### 1. `load_contract_demo.go`
基础合约加载演示，包含5种不同的加载方法：

- **方法1**: 从环境变量加载合约
- **方法2**: 手动指定地址加载合约  
- **方法3**: 验证合约存在性
- **方法4**: 获取合约基本信息
- **方法5**: 批量加载多个合约

### 2. `contract_manager.go`
高级合约管理器，提供企业级合约管理功能：

- 合约注册表管理
- 批量合约加载
- 合约健康检查
- 事件监控
- 配置文件管理

## 🚀 运行演示

### 基础加载演示
```bash
cd go-eth-demo
go run examples/09-load-contract/load_contract_demo.go
```

### 合约管理器演示
```bash
cd go-eth-demo
go run examples/09-load-contract/contract_manager.go
```

## 📋 预期输出

### 基础演示输出
```
🔧 智能合约加载演示
====================
✅ 以太坊节点连接成功

📋 方法1: 从环境变量加载合约
✅ 合约地址: 0xd737368D3bB49649CC4Fc098b93B7c2D6Af9E3B5
   🔍 尝试读取合约状态...
   ✅ 当前存储值: 150
   👤 合约所有者: 0xC52d29F4273e9BFE2C4E76B4A684Dab18D0F0191

📋 方法2: 手动指定地址加载合约
✅ 合约地址: 0xd737368D3bB49649CC4Fc098b93B7c2D6Af9E3B5
   🔍 尝试读取合约状态...
   ✅ 当前存储值: 150

📋 方法3: 验证合约存在性
✅ 合约存在且有代码

📋 方法4: 获取合约基本信息
   📏 合约代码大小: 1234 字节
   💰 合约余额: 0.000000 ETH
   🔢 合约nonce: 1

📋 方法5: 批量加载多个合约
✅ 成功加载 1 个合约
   合约1: 0xd737368D3bB49649CC4Fc098b93B7c2D6Af9E3B5

🎉 合约加载演示完成!
```

### 管理器演示输出
```
🏗️  智能合约管理器演示
========================

📋 1. 创建合约注册表
✅ 合约注册表已创建

📋 2. 从注册表加载合约
   ✅ 加载合约: SimpleStorage (0xd737368D3bB49649CC4Fc098b93B7c2D6Af9E3B5)
✅ 成功加载 1 个合约

📋 3. 列出已加载的合约
   📄 SimpleStorage: 0xd737368D3bB49649CC4Fc098b93B7c2D6Af9E3B5

📋 4. 获取特定合约
✅ 找到合约: 0xd737368D3bB49649CC4Fc098b93B7c2D6Af9E3B5
📊 当前存储值: 150

📋 5. 监控合约事件
🔍 开始监控DataStored事件...
🔍 正在监听合约事件...
⏰ 监听超时，停止监控

📋 6. 合约健康检查
📊 健康检查报告:
   ✅ SimpleStorage: 健康

🎉 合约管理器演示完成!
```

## 🔑 核心概念

### 1. 合约实例结构
```go
type ContractInstance struct {
    Address common.Address  // 合约地址
    ABI     *abi.ABI       // 合约ABI
    Client  *ethclient.Client // 以太坊客户端
}
```

### 2. 合约验证
- 检查地址格式有效性
- 验证合约代码存在
- 测试基本方法调用

### 3. 批量管理
- 合约注册表配置
- 批量加载和验证
- 健康状态监控

### 4. 事件监控
- 实时事件订阅
- 日志过滤和解析
- 异步事件处理

## 📚 学习要点

### 地址验证
```go
if !common.IsHexAddress(addressStr) {
    return nil, fmt.Errorf("无效的合约地址: %s", addressStr)
}
```

### 合约存在性检查
```go
code, err := client.CodeAt(context.Background(), contractAddress, nil)
return len(code) > 0, err
```

### ABI加载和解析
```go
contractABI, err := abi.JSON(strings.NewReader(abiString))
```

### 只读方法调用
```go
result, err := client.CallContract(context.Background(), ethereum.CallMsg{
    To:   &contractAddress,
    Data: data,
}, nil)
```

## 🛠️ 最佳实践

1. **地址验证**: 始终验证合约地址格式
2. **存在性检查**: 确认合约已部署且有代码
3. **错误处理**: 优雅处理网络和合约错误
4. **配置管理**: 使用配置文件管理多个合约
5. **健康监控**: 定期检查合约状态
6. **事件监听**: 实现异步事件处理

## 🔗 相关链接

- [以太坊合约地址](https://sepolia.etherscan.io/address/0xd737368D3bB49649CC4Fc098b93B7c2D6Af9E3B5)
- [Go-Ethereum文档](https://geth.ethereum.org/docs/developers/dapp-developer/native-bindings)
- [ABI规范](https://docs.soliditylang.org/en/latest/abi-spec.html)