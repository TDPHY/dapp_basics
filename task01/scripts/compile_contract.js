const solc = require('solc');
const fs = require('fs');
const path = require('path');

// 编译Counter合约
function compileContract() {
    console.log('🔨 开始编译Counter智能合约...');
    
    // 读取合约源码
    const contractPath = path.join(__dirname, '../contracts/Counter.sol');
    const source = fs.readFileSync(contractPath, 'utf8');
    
    // 编译输入
    const input = {
        language: 'Solidity',
        sources: {
            'Counter.sol': {
                content: source
            }
        },
        settings: {
            outputSelection: {
                '*': {
                    '*': ['abi', 'evm.bytecode']
                }
            },
            optimizer: {
                enabled: true,
                runs: 200
            }
        }
    };
    
    // 编译合约
    const output = JSON.parse(solc.compile(JSON.stringify(input)));
    
    // 检查编译错误
    if (output.errors) {
        output.errors.forEach(error => {
            if (error.severity === 'error') {
                console.error('❌ 编译错误:', error.formattedMessage);
                process.exit(1);
            } else {
                console.warn('⚠️  编译警告:', error.formattedMessage);
            }
        });
    }
    
    // 获取编译结果
    const contract = output.contracts['Counter.sol']['Counter'];
    
    if (!contract) {
        console.error('❌ 未找到Counter合约');
        process.exit(1);
    }
    
    // 创建输出目录
    const buildDir = path.join(__dirname, '../build');
    if (!fs.existsSync(buildDir)) {
        fs.mkdirSync(buildDir, { recursive: true });
    }
    
    // 保存编译结果
    const contractData = {
        contractName: 'Counter',
        abi: contract.abi,
        bytecode: contract.evm.bytecode.object,
        compiledAt: new Date().toISOString(),
        compiler: {
            version: solc.version(),
            optimizer: true,
            runs: 200
        }
    };
    
    const outputPath = path.join(buildDir, 'Counter.json');
    fs.writeFileSync(outputPath, JSON.stringify(contractData, null, 2));
    
    // 为abigen生成纯ABI文件
    const abiPath = path.join(buildDir, 'Counter.abi');
    fs.writeFileSync(abiPath, JSON.stringify(contract.abi, null, 2));
    
    // 生成字节码文件
    const binPath = path.join(buildDir, 'Counter.bin');
    fs.writeFileSync(binPath, contract.evm.bytecode.object);
    
    console.log('✅ Counter合约编译成功!');
    console.log(`📄 ABI: ${contract.abi.length} 个函数/事件`);
    console.log(`💾 字节码: ${contract.evm.bytecode.object.length / 2} 字节`);
    console.log(`📂 输出文件: ${outputPath}`);
    
    return contractData;
}

// 生成abigen命令
function generateAbigenCommand() {
    const buildDir = path.join(__dirname, '../build');
    const contractsDir = path.join(__dirname, '../contracts');
    
    console.log('\n🔧 生成abigen命令:');
    console.log('请在task01目录下运行以下命令生成Go绑定代码:');
    console.log('');
    console.log('abigen --abi=build/Counter.json --pkg=contracts --out=contracts/Counter.go');
    console.log('');
    console.log('或者使用完整路径:');
    console.log(`abigen --abi="${path.resolve(buildDir, 'Counter.json')}" --pkg=contracts --out="${path.resolve(contractsDir, 'Counter.go')}"`);
}

// 主函数
function main() {
    try {
        const contractData = compileContract();
        generateAbigenCommand();
        
        console.log('\n🎉 编译完成!');
        console.log('📋 下一步:');
        console.log('1. 安装abigen工具: go install github.com/ethereum/go-ethereum/cmd/abigen@latest');
        console.log('2. 运行上面的abigen命令生成Go绑定代码');
        console.log('3. 运行主程序: go run main.go');
        
    } catch (error) {
        console.error('❌ 编译失败:', error.message);
        process.exit(1);
    }
}

if (require.main === module) {
    main();
}

module.exports = { compileContract, generateAbigenCommand };