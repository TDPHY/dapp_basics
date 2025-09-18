const fs = require('fs');
const path = require('path');
const solc = require('solc');

/**
 * 编译智能合约
 */
function compileContracts() {
    console.log('🔨 开始编译智能合约...\n');
    
    // 合约文件路径
    const contractsDir = path.join(__dirname, '../contracts');
    const outputDir = path.join(__dirname, '../build');
    
    // 创建输出目录
    if (!fs.existsSync(outputDir)) {
        fs.mkdirSync(outputDir, { recursive: true });
    }
    
    // 读取合约文件
    const contracts = {};
    const contractFiles = fs.readdirSync(contractsDir).filter(file => file.endsWith('.sol'));
    
    console.log('📁 发现合约文件:');
    contractFiles.forEach(file => {
        console.log(`  - ${file}`);
        const contractPath = path.join(contractsDir, file);
        const contractSource = fs.readFileSync(contractPath, 'utf8');
        contracts[file] = {
            content: contractSource
        };
    });
    
    // 编译配置
    const input = {
        language: 'Solidity',
        sources: contracts,
        settings: {
            outputSelection: {
                '*': {
                    '*': ['abi', 'evm.bytecode.object', 'evm.deployedBytecode.object']
                }
            },
            optimizer: {
                enabled: true,
                runs: 200
            }
        }
    };
    
    console.log('\n⚙️  编译设置:');
    console.log('  - 优化器: 启用 (200 runs)');
    console.log('  - 输出: ABI + Bytecode');
    
    try {
        // 编译合约
        console.log('\n🔄 正在编译...');
        const output = JSON.parse(solc.compile(JSON.stringify(input)));
        
        // 检查编译错误
        if (output.errors) {
            console.log('\n⚠️  编译警告/错误:');
            output.errors.forEach(error => {
                const severity = error.severity;
                const message = error.message;
                const formattedMessage = error.formattedMessage;
                
                if (severity === 'error') {
                    console.log(`❌ 错误: ${message}`);
                    console.log(formattedMessage);
                } else {
                    console.log(`⚠️  警告: ${message}`);
                }
            });
            
            // 如果有错误，停止编译
            const hasErrors = output.errors.some(error => error.severity === 'error');
            if (hasErrors) {
                console.log('\n❌ 编译失败，请修复错误后重试');
                return false;
            }
        }
        
        // 保存编译结果
        console.log('\n💾 保存编译结果:');
        let compiledCount = 0;
        
        for (const sourceFile in output.contracts) {
            for (const contractName in output.contracts[sourceFile]) {
                const contract = output.contracts[sourceFile][contractName];
                
                // 创建合约输出对象
                const contractOutput = {
                    contractName: contractName,
                    sourceFile: sourceFile,
                    abi: contract.abi,
                    bytecode: '0x' + contract.evm.bytecode.object,
                    deployedBytecode: '0x' + contract.evm.deployedBytecode.object,
                    compiledAt: new Date().toISOString(),
                    compiler: {
                        version: solc.version(),
                        optimizer: true,
                        runs: 200
                    }
                };
                
                // 保存到文件
                const outputFile = path.join(outputDir, `${contractName}.json`);
                fs.writeFileSync(outputFile, JSON.stringify(contractOutput, null, 2));
                
                console.log(`  ✅ ${contractName} -> ${contractName}.json`);
                console.log(`     - ABI: ${contract.abi.length} 个函数/事件`);
                console.log(`     - Bytecode: ${contract.evm.bytecode.object.length / 2} 字节`);
                
                compiledCount++;
            }
        }
        
        console.log(`\n🎉 编译完成! 成功编译 ${compiledCount} 个合约`);
        console.log(`📂 输出目录: ${outputDir}`);
        
        return true;
        
    } catch (error) {
        console.error('\n❌ 编译过程中发生错误:', error.message);
        return false;
    }
}

// 如果直接运行此脚本
if (require.main === module) {
    const success = compileContracts();
    process.exit(success ? 0 : 1);
}

module.exports = { compileContracts };