# 获取Sepolia测试网ETH指南

## 🚰 测试网水龙头 (Faucets)

由于您的账户余额为0，需要先获取一些测试网ETH才能部署合约。以下是几个可用的Sepolia测试网水龙头：

### 1. Alchemy Sepolia Faucet ⭐ (推荐)
- **网址**: https://www.alchemy.com/faucets/ethereum-sepolia
- **要求**: 需要Alchemy账号
- **每日限额**: 0.5 ETH
- **到账时间**: 几分钟内

### 2. QuickNode Faucet
- **网址**: https://faucet.quicknode.com/ethereum/sepolia
- **要求**: 需要QuickNode账号
- **每日限额**: 0.1 ETH
- **到账时间**: 几分钟内

### 3. Infura Faucet
- **网址**: https://www.infura.io/faucet/sepolia
- **要求**: 需要Infura账号
- **每日限额**: 0.5 ETH
- **到账时间**: 几分钟内

### 4. 公共水龙头 (无需注册)
- **网址**: https://sepolia-faucet.pk910.de/
- **要求**: 无需注册，但可能需要完成简单任务
- **限额**: 变动
- **到账时间**: 几分钟到几小时

## 📋 获取步骤

1. **复制您的钱包地址**:
   ```
   0xC52d29F4273e9BFE2C4E76B4A684Dab18D0F0191
   ```

2. **访问任一水龙头网站**

3. **粘贴地址并申请**

4. **等待到账** (通常几分钟内)

5. **验证余额**:
   ```bash
   cd go-eth-demo
   go run examples/03-balance/eth_balance.go
   ```

## 💡 小贴士

- 建议使用Alchemy水龙头，因为您已经在使用Alchemy的RPC服务
- 如果一个水龙头暂时不可用，可以尝试其他的
- 获取到0.1 ETH就足够部署和测试合约了
- 测试网ETH没有实际价值，仅用于开发测试

## 🔄 检查余额

获取测试网ETH后，可以运行以下命令检查余额：

```bash
cd go-eth-demo
go run examples/03-balance/eth_balance.go
```

当余额大于0.01 ETH时，就可以继续部署合约了：

```bash
go run examples/08-deploy/deploy_simple_storage.go