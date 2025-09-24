#!/usr/bin/env node

/**
 * Geth功能测试脚本
 * 全面测试Geth节点的各项功能
 */

const Web3 = require('web3');
const fs = require('fs');
const path = require('path');

// 配置
const CONFIG = {
    rpcUrl: 'http://localhost:8545',
    wsUrl: 'ws://localhost:8546',
    testTimeout: 30000,
    testAccounts: 3
};

class GethTester {
    constructor() {
        this.web3 = new Web3(CONFIG.rpcUrl);
        this.wsWeb3 = null;
        this.testResults = {
            passed: 0,
            failed: 0,
            tests: []
        };
    }

    // 记录测试结果
    recordTest(name, passed, message = '') {
        const result = {
            name,
            passed,
            message,
            timestamp: new Date().toISOString()
        };
        
        this.testResults.tests.push(result);
        
        if (passed) {
            this.testResults.passed++;
            console.log(`✅ ${name}`);
        } else {
            this.testResults.failed++;
            console.log(`❌ ${name}: ${message}`);
        }
    }

    // 基础连接测试
    async testConnection() {
        console.log('\n🔗 测试网络连接...');
        
        try {
            const isListening = await this.web3.eth.net.isListening();
            this.recordTest('网络连接', isListening, isListening ? '' : '节点未响应');
            
            const networkId = await this.web3.eth.net.getId();
            this.recordTest('网络ID获取', networkId > 0, `网络ID: ${networkId}`);
            
            const peerCount = await this.web3.eth.net.getPeerCount();
            this.recordTest('节点连接数查询', true, `连接节点数: ${peerCount}`);
            
        } catch (error) {
            this.recordTest('网络连接', false, error.message);
        }
    }

    // 区块链基础功能测试
    async testBlockchain() {
        console.log('\n⛓️  测试区块链功能...');
        
        try {
            // 获取最新区块号
            const blockNumber = await this.web3.eth.getBlockNumber();
            this.recordTest('获取区块高度', blockNumber >= 0, `当前区块: ${blockNumber}`);
            
            // 获取最新区块
            const latestBlock = await this.web3.eth.getBlock('latest');
            this.recordTest('获取最新区块', !!latestBlock, `区块哈希: ${latestBlock?.hash}`);
            
            // 获取创世区块
            const genesisBlock = await this.web3.eth.getBlock(0);
            this.recordTest('获取创世区块', !!genesisBlock, `创世区块哈希: ${genesisBlock?.hash}`);
            
            // 测试区块信息完整性
            if (latestBlock) {
                const hasRequiredFields = !!(
                    latestBlock.hash &&
                    latestBlock.parentHash &&
                    latestBlock.number !== undefined &&
                    latestBlock.timestamp
                );
                this.recordTest('区块信息完整性', hasRequiredFields);
            }
            
        } catch (error) {
            this.recordTest('区块链功能', false, error.message);
        }
    }

    // 账户管理测试
    async testAccounts() {
        console.log('\n👤 测试账户管理...');
        
        try {
            // 获取账户列表
            const accounts = await this.web3.eth.getAccounts();
            this.recordTest('获取账户列表', accounts.length >= 0, `账户数量: ${accounts.length}`);
            
            if (accounts.length > 0) {
                // 测试余额查询
                for (let i = 0; i < Math.min(accounts.length, 3); i++) {
                    const balance = await this.web3.eth.getBalance(accounts[i]);
                    const ethBalance = this.web3.utils.fromWei(balance, 'ether');
                    this.recordTest(
                        `账户${i}余额查询`, 
                        true, 
                        `${accounts[i]}: ${ethBalance} ETH`
                    );
                }
                
                // 测试交易计数
                const nonce = await this.web3.eth.getTransactionCount(accounts[0]);
                this.recordTest('获取交易计数', nonce >= 0, `Nonce: ${nonce}`);
            }
            
        } catch (error) {
            this.recordTest('账户管理', false, error.message);
        }
    }

    // 交易功能测试
    async testTransactions() {
        console.log('\n💸 测试交易功能...');
        
        try {
            const accounts = await this.web3.eth.getAccounts();
            
            if (accounts.length < 2) {
                this.recordTest('交易测试', false, '需要至少2个账户进行交易测试');
                return;
            }
            
            const sender = accounts[0];
            const receiver = accounts[1];
            
            // 检查发送者余额
            const senderBalance = await this.web3.eth.getBalance(sender);
            if (senderBalance === '0') {
                this.recordTest('交易测试', false, '发送者账户余额为0');
                return;
            }
            
            // 获取初始余额
            const initialReceiverBalance = await this.web3.eth.getBalance(receiver);
            
            // 发送交易
            const transferAmount = this.web3.utils.toWei('0.1', 'ether');
            const gasPrice = await this.web3.eth.getGasPrice();
            
            console.log('  📤 发送测试交易...');
            const txHash = await this.web3.eth.sendTransaction({
                from: sender,
                to: receiver,
                value: transferAmount,
                gas: 21000,
                gasPrice: gasPrice
            });
            
            this.recordTest('发送交易', !!txHash, `交易哈希: ${txHash}`);
            
            // 等待交易确认
            console.log('  ⏳ 等待交易确认...');
            const receipt = await this.waitForTransaction(txHash);
            
            if (receipt) {
                this.recordTest('交易确认', receipt.status === true, `Gas使用: ${receipt.gasUsed}`);
                
                // 验证余额变化
                const finalReceiverBalance = await this.web3.eth.getBalance(receiver);
                const balanceIncrease = this.web3.utils.toBN(finalReceiverBalance)
                    .sub(this.web3.utils.toBN(initialReceiverBalance));
                
                const expectedIncrease = this.web3.utils.toBN(transferAmount);
                const balanceCorrect = balanceIncrease.eq(expectedIncrease);
                
                this.recordTest('余额更新验证', balanceCorrect, 
                    `余额增加: ${this.web3.utils.fromWei(balanceIncrease, 'ether')} ETH`);
            }
            
        } catch (error) {
            this.recordTest('交易功能', false, error.message);
        }
    }

    // 等待交易确认
    async waitForTransaction(txHash, timeout = 30000) {
        const startTime = Date.now();
        
        while (Date.now() - startTime < timeout) {
            try {
                const receipt = await this.web3.eth.getTransactionReceipt(txHash);
                if (receipt) {
                    return receipt;
                }
                await new Promise(resolve => setTimeout(resolve, 1000));
            } catch (error) {
                console.log('  ⏳ 等待交易确认中...');
                await new Promise(resolve => setTimeout(resolve, 2000));
            }
        }
        
        throw new Error('交易确认超时');
    }

    // Gas相关测试
    async testGas() {
        console.log('\n⛽ 测试Gas机制...');
        
        try {
            // 获取当前Gas价格
            const gasPrice = await this.web3.eth.getGasPrice();
            this.recordTest('获取Gas价格', !!gasPrice, 
                `当前Gas价格: ${this.web3.utils.fromWei(gasPrice, 'gwei')} Gwei`);
            
            // 估算简单转账Gas
            const accounts = await this.web3.eth.getAccounts();
            if (accounts.length >= 2) {
                const gasEstimate = await this.web3.eth.estimateGas({
                    from: accounts[0],
                    to: accounts[1],
                    value: this.web3.utils.toWei('0.01', 'ether')
                });
                
                this.recordTest('Gas估算', gasEstimate > 0, `估算Gas: ${gasEstimate}`);
            }
            
        } catch (error) {
            this.recordTest('Gas机制', false, error.message);
        }
    }

    // 挖矿功能测试
    async testMining() {
        console.log('\n⛏️  测试挖矿功能...');
        
        try {
            // 检查挖矿状态
            const isMining = await this.web3.eth.isMining();
            this.recordTest('挖矿状态查询', true, `挖矿状态: ${isMining ? '进行中' : '已停止'}`);
            
            // 获取算力
            const hashrate = await this.web3.eth.getHashrate();
            this.recordTest('算力查询', hashrate >= 0, `当前算力: ${hashrate} H/s`);
            
            // 获取coinbase地址
            try {
                const coinbase = await this.web3.eth.getCoinbase();
                this.recordTest('Coinbase地址', !!coinbase, `Coinbase: ${coinbase}`);
            } catch (error) {
                this.recordTest('Coinbase地址', false, '未设置coinbase地址');
            }
            
        } catch (error) {
            this.recordTest('挖矿功能', false, error.message);
        }
    }

    // WebSocket连接测试
    async testWebSocket() {
        console.log('\n🔌 测试WebSocket连接...');
        
        try {
            this.wsWeb3 = new Web3(CONFIG.wsUrl);
            
            // 测试连接
            const blockNumber = await this.wsWeb3.eth.getBlockNumber();
            this.recordTest('WebSocket连接', blockNumber >= 0, `通过WS获取区块: ${blockNumber}`);
            
            // 测试事件订阅
            const subscription = this.wsWeb3.eth.subscribe('newBlockHeaders');
            
            let eventReceived = false;
            subscription.on('data', (blockHeader) => {
                if (!eventReceived) {
                    eventReceived = true;
                    this.recordTest('新区块事件订阅', true, `接收到区块: ${blockHeader.number}`);
                    subscription.unsubscribe();
                }
            });
            
            subscription.on('error', (error) => {
                this.recordTest('新区块事件订阅', false, error.message);
            });
            
            // 等待事件或超时
            setTimeout(() => {
                if (!eventReceived) {
                    this.recordTest('新区块事件订阅', false, '未接收到新区块事件');
                    subscription.unsubscribe();
                }
            }, 10000);
            
        } catch (error) {
            this.recordTest('WebSocket连接', false, error.message);
        }
    }

    // JSON-RPC API测试
    async testJsonRpc() {
        console.log('\n🌐 测试JSON-RPC API...');
        
        const testCases = [
            { method: 'web3_clientVersion', params: [] },
            { method: 'net_version', params: [] },
            { method: 'eth_protocolVersion', params: [] },
            { method: 'eth_syncing', params: [] },
            { method: 'eth_chainId', params: [] }
        ];
        
        for (const testCase of testCases) {
            try {
                const response = await this.makeRpcCall(testCase.method, testCase.params);
                this.recordTest(`RPC ${testCase.method}`, !!response, 
                    `响应: ${JSON.stringify(response).substring(0, 100)}`);
            } catch (error) {
                this.recordTest(`RPC ${testCase.method}`, false, error.message);
            }
        }
    }

    // 发送RPC调用
    async makeRpcCall(method, params) {
        const response = await fetch(CONFIG.rpcUrl, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                jsonrpc: '2.0',
                method: method,
                params: params,
                id: 1
            })
        });
        
        const data = await response.json();
        
        if (data.error) {
            throw new Error(data.error.message);
        }
        
        return data.result;
    }

    // 性能测试
    async testPerformance() {
        console.log('\n🚀 测试性能指标...');
        
        try {
            // 测试批量查询性能
            const startTime = Date.now();
            const batchSize = 10;
            const promises = [];
            
            for (let i = 0; i < batchSize; i++) {
                promises.push(this.web3.eth.getBlockNumber());
            }
            
            await Promise.all(promises);
            const endTime = Date.now();
            const duration = endTime - startTime;
            const rps = (batchSize / duration) * 1000;
            
            this.recordTest('批量查询性能', true, 
                `${batchSize}个请求耗时${duration}ms，RPS: ${rps.toFixed(2)}`);
            
            // 测试单个查询延迟
            const latencyStart = Date.now();
            await this.web3.eth.getBlockNumber();
            const latency = Date.now() - latencyStart;
            
            this.recordTest('查询延迟', latency < 1000, `单次查询延迟: ${latency}ms`);
            
        } catch (error) {
            this.recordTest('性能测试', false, error.message);
        }
    }

    // 运行所有测试
    async runAllTests() {
        console.log('🧪 开始Geth功能全面测试...\n');
        
        const tests = [
            () => this.testConnection(),
            () => this.testBlockchain(),
            () => this.testAccounts(),
            () => this.testGas(),
            () => this.testMining(),
            () => this.testJsonRpc(),
            () => this.testWebSocket(),
            () => this.testTransactions(),
            () => this.testPerformance()
        ];
        
        for (const test of tests) {
            try {
                await test();
            } catch (error) {
                console.error(`测试执行错误: ${error.message}`);
            }
        }
        
        this.generateReport();
    }

    // 生成测试报告
    generateReport() {
        console.log('\n📊 测试报告');
        console.log('='.repeat(50));
        
        const total = this.testResults.passed + this.testResults.failed;
        const passRate = total > 0 ? (this.testResults.passed / total * 100).toFixed(2) : 0;
        
        console.log(`总测试数: ${total}`);
        console.log(`通过: ${this.testResults.passed}`);
        console.log(`失败: ${this.testResults.failed}`);
        console.log(`通过率: ${passRate}%`);
        
        if (this.testResults.failed > 0) {
            console.log('\n❌ 失败的测试:');
            this.testResults.tests
                .filter(test => !test.passed)
                .forEach(test => {
                    console.log(`  - ${test.name}: ${test.message}`);
                });
        }
        
        // 保存详细报告
        const reportPath = path.join(__dirname, '../evidence/test-report.json');
        const reportDir = path.dirname(reportPath);
        
        if (!fs.existsSync(reportDir)) {
            fs.mkdirSync(reportDir, { recursive: true });
        }
        
        fs.writeFileSync(reportPath, JSON.stringify({
            summary: {
                total,
                passed: this.testResults.passed,
                failed: this.testResults.failed,
                passRate: `${passRate}%`,
                timestamp: new Date().toISOString()
            },
            tests: this.testResults.tests
        }, null, 2));
        
        console.log(`\n📄 详细报告已保存到: ${reportPath}`);
        
        // 清理WebSocket连接
        if (this.wsWeb3 && this.wsWeb3.currentProvider) {
            this.wsWeb3.currentProvider.disconnect();
        }
    }
}

// 主函数
async function main() {
    const tester = new GethTester();
    
    try {
        await tester.runAllTests();
        
        const exitCode = tester.testResults.failed > 0 ? 1 : 0;
        process.exit(exitCode);
        
    } catch (error) {
        console.error('❌ 测试运行失败:', error.message);
        process.exit(1);
    }
}

// 如果直接运行此脚本
if (require.main === module) {
    main();
}

module.exports = GethTester;