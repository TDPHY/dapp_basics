#!/usr/bin/env node

/**
 * 智能合约部署脚本
 * 支持编译、部署和测试智能合约
 */

const Web3 = require('web3');
const fs = require('fs');
const path = require('path');
const solc = require('solc');

// 配置
const CONFIG = {
    rpcUrl: 'http://localhost:8545',
    gasPrice: '20000000000', // 20 Gwei
    gasLimit: 3000000,
    contractsDir: path.join(__dirname, '../evidence/contracts'),
    buildDir: path.join(__dirname, '../build')
};

class ContractDeployer {
    constructor() {
        this.web3 = new Web3(CONFIG.rpcUrl);
        this.accounts = [];
        this.deployedContracts = {};
    }

    // 初始化
    async init() {
        try {
            console.log('🚀 初始化合约部署器...');
            
            // 检查网络连接
            const isConnected = await this.web3.eth.net.isListening();
            if (!isConnected) {
                throw new Error('无法连接到Geth节点');
            }
            
            // 获取账户
            this.accounts = await this.web3.eth.getAccounts();
            if (this.accounts.length === 0) {
                throw new Error('没有可用的账户');
            }
            
            console.log(`✅ 连接成功，找到 ${this.accounts.length} 个账户`);
            console.log(`📍 部署账户: ${this.accounts[0]}`);
            
            // 检查账户余额
            const balance = await this.web3.eth.getBalance(this.accounts[0]);
            console.log(`💰 账户余额: ${this.web3.utils.fromWei(balance, 'ether')} ETH`);
            
            if (balance === '0') {
                console.log('⚠️  账户余额为0，请先启动挖矿获取ETH');
            }
            
        } catch (error) {
            console.error('❌ 初始化失败:', error.message);
            process.exit(1);
        }
    }

    // 编译合约
    compileContract(contractName, sourceCode) {
        console.log(`🔨 编译合约: ${contractName}`);
        
        const input = {
            language: 'Solidity',
            sources: {
                [contractName]: {
                    content: sourceCode
                }
            },
            settings: {
                outputSelection: {
                    '*': {
                        '*': ['abi', 'evm.bytecode']
                    }
                }
            }
        };
        
        const output = JSON.parse(solc.compile(JSON.stringify(input)));
        
        if (output.errors) {
            const errors = output.errors.filter(error => error.severity === 'error');
            if (errors.length > 0) {
                console.error('❌ 编译错误:');
                errors.forEach(error => console.error(error.formattedMessage));
                throw new Error('合约编译失败');
            }
        }
        
        const contract = output.contracts[contractName][contractName];
        console.log('✅ 合约编译成功');
        
        return {
            abi: contract.abi,
            bytecode: contract.evm.bytecode.object
        };
    }

    // 部署合约
    async deployContract(contractName, abi, bytecode, constructorArgs = []) {
        try {
            console.log(`🚀 部署合约: ${contractName}`);
            
            const contract = new this.web3.eth.Contract(abi);
            const deployTx = contract.deploy({
                data: '0x' + bytecode,
                arguments: constructorArgs
            });
            
            // 估算Gas
            const gasEstimate = await deployTx.estimateGas({
                from: this.accounts[0]
            });
            
            console.log(`⛽ 估算Gas: ${gasEstimate}`);
            
            // 部署合约
            const deployedContract = await deployTx.send({
                from: this.accounts[0],
                gas: Math.floor(gasEstimate * 1.2), // 增加20%安全边际
                gasPrice: CONFIG.gasPrice
            });
            
            console.log(`✅ 合约部署成功!`);
            console.log(`📍 合约地址: ${deployedContract.options.address}`);
            
            // 保存部署信息
            this.deployedContracts[contractName] = {
                address: deployedContract.options.address,
                abi: abi,
                contract: deployedContract
            };
            
            return deployedContract;
            
        } catch (error) {
            console.error(`❌ 部署失败: ${error.message}`);
            throw error;
        }
    }

    // 测试合约功能
    async testContract(contractName) {
        const contractInfo = this.deployedContracts[contractName];
        if (!contractInfo) {
            throw new Error(`合约 ${contractName} 未部署`);
        }
        
        console.log(`🧪 测试合约: ${contractName}`);
        
        const contract = contractInfo.contract;
        const account = this.accounts[0];
        
        try {
            switch (contractName) {
                case 'SimpleStorage':
                    await this.testSimpleStorage(contract, account);
                    break;
                case 'Counter':
                    await this.testCounter(contract, account);
                    break;
                case 'Token':
                    await this.testToken(contract, account);
                    break;
                default:
                    console.log('⚠️  没有为此合约定义测试用例');
            }
        } catch (error) {
            console.error(`❌ 测试失败: ${error.message}`);
        }
    }

    // 测试SimpleStorage合约
    async testSimpleStorage(contract, account) {
        console.log('📝 测试SimpleStorage合约...');
        
        // 测试初始值
        let value = await contract.methods.get().call();
        console.log(`初始值: ${value}`);
        
        // 测试设置值
        console.log('设置值为 42...');
        const setTx = await contract.methods.set(42).send({
            from: account,
            gas: 100000
        });
        console.log(`交易哈希: ${setTx.transactionHash}`);
        
        // 验证新值
        value = await contract.methods.get().call();
        console.log(`新值: ${value}`);
        
        // 测试所有者
        const owner = await contract.methods.getOwner().call();
        console.log(`合约所有者: ${owner}`);
        
        console.log('✅ SimpleStorage测试完成');
    }

    // 测试Counter合约
    async testCounter(contract, account) {
        console.log('🔢 测试Counter合约...');
        
        // 获取初始计数
        let count = await contract.methods.getCount().call();
        console.log(`初始计数: ${count}`);
        
        // 增加计数
        console.log('增加计数...');
        await contract.methods.increment().send({
            from: account,
            gas: 100000
        });
        
        count = await contract.methods.getCount().call();
        console.log(`增加后计数: ${count}`);
        
        // 减少计数
        console.log('减少计数...');
        await contract.methods.decrement().send({
            from: account,
            gas: 100000
        });
        
        count = await contract.methods.getCount().call();
        console.log(`减少后计数: ${count}`);
        
        console.log('✅ Counter测试完成');
    }

    // 测试Token合约
    async testToken(contract, account) {
        console.log('🪙 测试Token合约...');
        
        // 获取代币信息
        const name = await contract.methods.name().call();
        const symbol = await contract.methods.symbol().call();
        const totalSupply = await contract.methods.totalSupply().call();
        
        console.log(`代币名称: ${name}`);
        console.log(`代币符号: ${symbol}`);
        console.log(`总供应量: ${totalSupply}`);
        
        // 查看余额
        const balance = await contract.methods.balanceOf(account).call();
        console.log(`账户余额: ${balance}`);
        
        // 转账测试（如果有多个账户）
        if (this.accounts.length > 1) {
            const recipient = this.accounts[1];
            const amount = '1000';
            
            console.log(`转账 ${amount} 代币到 ${recipient}...`);
            await contract.methods.transfer(recipient, amount).send({
                from: account,
                gas: 100000
            });
            
            const recipientBalance = await contract.methods.balanceOf(recipient).call();
            console.log(`接收者余额: ${recipientBalance}`);
        }
        
        console.log('✅ Token测试完成');
    }

    // 保存部署结果
    saveDeploymentInfo() {
        const deploymentInfo = {
            timestamp: new Date().toISOString(),
            network: 'private',
            deployer: this.accounts[0],
            contracts: {}
        };
        
        for (const [name, info] of Object.entries(this.deployedContracts)) {
            deploymentInfo.contracts[name] = {
                address: info.address,
                abi: info.abi
            };
        }
        
        // 确保构建目录存在
        if (!fs.existsSync(CONFIG.buildDir)) {
            fs.mkdirSync(CONFIG.buildDir, { recursive: true });
        }
        
        const filePath = path.join(CONFIG.buildDir, 'deployment.json');
        fs.writeFileSync(filePath, JSON.stringify(deploymentInfo, null, 2));
        
        console.log(`💾 部署信息已保存到: ${filePath}`);
    }

    // 创建示例合约
    createSampleContracts() {
        console.log('📝 创建示例合约...');
        
        // 确保合约目录存在
        if (!fs.existsSync(CONFIG.contractsDir)) {
            fs.mkdirSync(CONFIG.contractsDir, { recursive: true });
        }
        
        // SimpleStorage合约
        const simpleStorageCode = `
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract SimpleStorage {
    uint256 private storedData;
    address public owner;
    
    event DataStored(uint256 indexed newValue, address indexed setter);
    
    constructor() {
        owner = msg.sender;
        storedData = 0;
    }
    
    function set(uint256 x) public {
        storedData = x;
        emit DataStored(x, msg.sender);
    }
    
    function get() public view returns (uint256) {
        return storedData;
    }
    
    function getOwner() public view returns (address) {
        return owner;
    }
}`;
        
        // Counter合约
        const counterCode = `
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract Counter {
    uint256 private count;
    address public owner;
    
    event CountChanged(uint256 newCount, address changer);
    
    constructor() {
        owner = msg.sender;
        count = 0;
    }
    
    function increment() public {
        count += 1;
        emit CountChanged(count, msg.sender);
    }
    
    function decrement() public {
        require(count > 0, "Counter cannot be negative");
        count -= 1;
        emit CountChanged(count, msg.sender);
    }
    
    function getCount() public view returns (uint256) {
        return count;
    }
    
    function reset() public {
        require(msg.sender == owner, "Only owner can reset");
        count = 0;
        emit CountChanged(count, msg.sender);
    }
}`;
        
        // Token合约
        const tokenCode = `
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract Token {
    string public name;
    string public symbol;
    uint8 public decimals;
    uint256 public totalSupply;
    
    mapping(address => uint256) public balanceOf;
    mapping(address => mapping(address => uint256)) public allowance;
    
    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);
    
    constructor(
        string memory _name,
        string memory _symbol,
        uint8 _decimals,
        uint256 _totalSupply
    ) {
        name = _name;
        symbol = _symbol;
        decimals = _decimals;
        totalSupply = _totalSupply * 10**_decimals;
        balanceOf[msg.sender] = totalSupply;
    }
    
    function transfer(address to, uint256 value) public returns (bool) {
        require(balanceOf[msg.sender] >= value, "Insufficient balance");
        balanceOf[msg.sender] -= value;
        balanceOf[to] += value;
        emit Transfer(msg.sender, to, value);
        return true;
    }
    
    function approve(address spender, uint256 value) public returns (bool) {
        allowance[msg.sender][spender] = value;
        emit Approval(msg.sender, spender, value);
        return true;
    }
    
    function transferFrom(address from, address to, uint256 value) public returns (bool) {
        require(balanceOf[from] >= value, "Insufficient balance");
        require(allowance[from][msg.sender] >= value, "Insufficient allowance");
        
        balanceOf[from] -= value;
        balanceOf[to] += value;
        allowance[from][msg.sender] -= value;
        
        emit Transfer(from, to, value);
        return true;
    }
}`;
        
        // 保存合约文件
        fs.writeFileSync(path.join(CONFIG.contractsDir, 'SimpleStorage.sol'), simpleStorageCode);
        fs.writeFileSync(path.join(CONFIG.contractsDir, 'Counter.sol'), counterCode);
        fs.writeFileSync(path.join(CONFIG.contractsDir, 'Token.sol'), tokenCode);
        
        console.log('✅ 示例合约已创建');
    }

    // 部署所有示例合约
    async deployAllSamples() {
        console.log('🚀 部署所有示例合约...');
        
        try {
            // 部署SimpleStorage
            const simpleStorageCode = fs.readFileSync(
                path.join(CONFIG.contractsDir, 'SimpleStorage.sol'), 'utf8'
            );
            const simpleStorage = this.compileContract('SimpleStorage.sol', simpleStorageCode);
            await this.deployContract('SimpleStorage', simpleStorage.abi, simpleStorage.bytecode);
            await this.testContract('SimpleStorage');
            
            // 部署Counter
            const counterCode = fs.readFileSync(
                path.join(CONFIG.contractsDir, 'Counter.sol'), 'utf8'
            );
            const counter = this.compileContract('Counter.sol', counterCode);
            await this.deployContract('Counter', counter.abi, counter.bytecode);
            await this.testContract('Counter');
            
            // 部署Token
            const tokenCode = fs.readFileSync(
                path.join(CONFIG.contractsDir, 'Token.sol'), 'utf8'
            );
            const token = this.compileContract('Token.sol', tokenCode);
            await this.deployContract('Token', token.abi, token.bytecode, [
                'TestToken', 'TTK', 18, 1000000
            ]);
            await this.testContract('Token');
            
            console.log('🎉 所有合约部署完成!');
            
        } catch (error) {
            console.error('❌ 部署过程中出现错误:', error.message);
        }
    }
}

// 主函数
async function main() {
    console.log('=== 智能合约部署工具 ===\n');
    
    const deployer = new ContractDeployer();
    
    try {
        await deployer.init();
        deployer.createSampleContracts();
        await deployer.deployAllSamples();
        deployer.saveDeploymentInfo();
        
        console.log('\n🎉 部署流程完成!');
        console.log('📄 查看部署信息: cat ../build/deployment.json');
        
    } catch (error) {
        console.error('\n❌ 部署失败:', error.message);
        process.exit(1);
    }
}

// 如果直接运行此脚本
if (require.main === module) {
    main();
}

module.exports = ContractDeployer;