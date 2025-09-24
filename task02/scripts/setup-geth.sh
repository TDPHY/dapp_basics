#!/bin/bash

# Go-Ethereum 环境搭建脚本
# 适用于 Ubuntu/Debian 系统

set -e

echo "=== Go-Ethereum 环境搭建开始 ==="

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查系统要求
check_system() {
    log_info "检查系统要求..."
    
    # 检查操作系统
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        log_info "操作系统: Linux"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        log_info "操作系统: macOS"
    else
        log_error "不支持的操作系统: $OSTYPE"
        exit 1
    fi
    
    # 检查内存
    total_mem=$(free -m | awk 'NR==2{printf "%.0f", $2/1024}')
    if [ "$total_mem" -lt 4 ]; then
        log_warn "内存不足4GB，可能影响性能"
    else
        log_info "内存: ${total_mem}GB"
    fi
    
    # 检查磁盘空间
    available_space=$(df -BG . | awk 'NR==2 {print $4}' | sed 's/G//')
    if [ "$available_space" -lt 100 ]; then
        log_warn "可用磁盘空间不足100GB"
    else
        log_info "可用磁盘空间: ${available_space}GB"
    fi
}

# 安装依赖
install_dependencies() {
    log_info "安装系统依赖..."
    
    if command -v apt-get &> /dev/null; then
        sudo apt-get update
        sudo apt-get install -y git curl wget build-essential
    elif command -v yum &> /dev/null; then
        sudo yum update -y
        sudo yum install -y git curl wget gcc gcc-c++ make
    elif command -v brew &> /dev/null; then
        brew install git curl wget
    else
        log_error "无法识别的包管理器"
        exit 1
    fi
}

# 安装Go语言
install_go() {
    log_info "检查Go语言环境..."
    
    if command -v go &> /dev/null; then
        go_version=$(go version | awk '{print $3}' | sed 's/go//')
        log_info "已安装Go版本: $go_version"
        
        # 检查版本是否满足要求
        if [[ $(echo "$go_version 1.19" | tr " " "\n" | sort -V | head -n1) != "1.19" ]]; then
            log_warn "Go版本过低，需要1.19+，正在升级..."
            install_go_binary
        fi
    else
        log_info "未检测到Go，正在安装..."
        install_go_binary
    fi
}

install_go_binary() {
    GO_VERSION="1.21.4"
    
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        GO_ARCH="linux-amd64"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        GO_ARCH="darwin-amd64"
    fi
    
    GO_TARBALL="go${GO_VERSION}.${GO_ARCH}.tar.gz"
    
    log_info "下载Go ${GO_VERSION}..."
    wget -q "https://golang.org/dl/${GO_TARBALL}"
    
    log_info "安装Go..."
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf "${GO_TARBALL}"
    rm "${GO_TARBALL}"
    
    # 设置环境变量
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    echo 'export GOPATH=$HOME/go' >> ~/.bashrc
    echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
    
    export PATH=$PATH:/usr/local/go/bin
    export GOPATH=$HOME/go
    
    log_info "Go安装完成: $(go version)"
}

# 克隆并编译Geth
build_geth() {
    log_info "克隆Go-Ethereum源码..."
    
    GETH_DIR="$HOME/go-ethereum"
    
    if [ -d "$GETH_DIR" ]; then
        log_warn "目录已存在，正在更新..."
        cd "$GETH_DIR"
        git pull origin master
    else
        git clone https://github.com/ethereum/go-ethereum.git "$GETH_DIR"
        cd "$GETH_DIR"
    fi
    
    # 切换到稳定版本
    log_info "切换到稳定版本..."
    STABLE_VERSION=$(git tag -l | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | sort -V | tail -1)
    log_info "使用版本: $STABLE_VERSION"
    git checkout "$STABLE_VERSION"
    
    # 编译
    log_info "编译Geth（这可能需要几分钟）..."
    make geth
    
    # 验证编译结果
    if [ -f "build/bin/geth" ]; then
        log_info "Geth编译成功!"
        ./build/bin/geth version
        
        # 添加到PATH
        GETH_BIN_DIR="$GETH_DIR/build/bin"
        if ! echo "$PATH" | grep -q "$GETH_BIN_DIR"; then
            echo "export PATH=\$PATH:$GETH_BIN_DIR" >> ~/.bashrc
            export PATH=$PATH:$GETH_BIN_DIR
        fi
    else
        log_error "Geth编译失败!"
        exit 1
    fi
}

# 创建工作目录
setup_workspace() {
    log_info "创建工作目录..."
    
    ETHEREUM_DIR="$HOME/ethereum"
    mkdir -p "$ETHEREUM_DIR"/{data,keystore,logs,contracts}
    
    log_info "工作目录创建完成: $ETHEREUM_DIR"
    
    # 创建创世区块配置
    cat > "$ETHEREUM_DIR/genesis.json" << EOF
{
  "config": {
    "chainId": 12345,
    "homesteadBlock": 0,
    "eip150Block": 0,
    "eip155Block": 0,
    "eip158Block": 0,
    "byzantiumBlock": 0,
    "constantinopleBlock": 0,
    "petersburgBlock": 0,
    "istanbulBlock": 0,
    "berlinBlock": 0,
    "londonBlock": 0,
    "ethash": {}
  },
  "difficulty": "0x4000",
  "gasLimit": "0x8000000",
  "alloc": {
    "0x7df9a875a174b3bc565e6424a0050ebc1b2d1d82": {
      "balance": "0x200000000000000000000"
    }
  }
}
EOF
    
    log_info "创世区块配置已创建: $ETHEREUM_DIR/genesis.json"
}

# 初始化私有链
init_private_chain() {
    log_info "初始化私有链..."
    
    ETHEREUM_DIR="$HOME/ethereum"
    cd "$ETHEREUM_DIR"
    
    # 初始化创世区块
    geth --datadir data init genesis.json
    
    log_info "私有链初始化完成!"
}

# 创建启动脚本
create_start_script() {
    log_info "创建启动脚本..."
    
    ETHEREUM_DIR="$HOME/ethereum"
    
    cat > "$ETHEREUM_DIR/start-geth.sh" << 'EOF'
#!/bin/bash

# Geth私有链启动脚本

ETHEREUM_DIR="$HOME/ethereum"
cd "$ETHEREUM_DIR"

echo "启动Geth私有链节点..."

geth --datadir data \
     --networkid 12345 \
     --http \
     --http.addr "0.0.0.0" \
     --http.port 8545 \
     --http.api "eth,net,web3,personal,miner,admin,debug" \
     --http.corsdomain "*" \
     --ws \
     --ws.addr "0.0.0.0" \
     --ws.port 8546 \
     --ws.api "eth,net,web3,personal,miner,admin,debug" \
     --ws.origins "*" \
     --allow-insecure-unlock \
     --mine \
     --miner.threads 1 \
     --verbosity 3 \
     --log.file logs/geth.log \
     --console
EOF
    
    chmod +x "$ETHEREUM_DIR/start-geth.sh"
    
    log_info "启动脚本已创建: $ETHEREUM_DIR/start-geth.sh"
}

# 安装Node.js和npm（用于智能合约开发）
install_nodejs() {
    log_info "检查Node.js环境..."
    
    if command -v node &> /dev/null; then
        node_version=$(node --version)
        log_info "已安装Node.js版本: $node_version"
    else
        log_info "安装Node.js..."
        
        # 安装Node.js 18 LTS
        curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
        sudo apt-get install -y nodejs
        
        log_info "Node.js安装完成: $(node --version)"
    fi
    
    # 安装全局包
    log_info "安装开发工具..."
    npm install -g solc web3 truffle ganache-cli
}

# 创建测试脚本
create_test_scripts() {
    log_info "创建测试脚本..."
    
    ETHEREUM_DIR="$HOME/ethereum"
    
    # 创建账户测试脚本
    cat > "$ETHEREUM_DIR/test-accounts.js" << 'EOF'
// 账户管理测试脚本
const Web3 = require('web3');

async function testAccounts() {
    const web3 = new Web3('http://localhost:8545');
    
    try {
        console.log('=== 账户管理测试 ===');
        
        // 获取账户列表
        const accounts = await web3.eth.getAccounts();
        console.log('账户列表:', accounts);
        
        if (accounts.length > 0) {
            // 查看余额
            for (let i = 0; i < accounts.length; i++) {
                const balance = await web3.eth.getBalance(accounts[i]);
                console.log(`账户 ${i}: ${accounts[i]}`);
                console.log(`余额: ${web3.utils.fromWei(balance, 'ether')} ETH`);
            }
        }
        
        // 查看区块信息
        const blockNumber = await web3.eth.getBlockNumber();
        console.log('当前区块高度:', blockNumber);
        
        const latestBlock = await web3.eth.getBlock('latest');
        console.log('最新区块信息:', {
            number: latestBlock.number,
            hash: latestBlock.hash,
            timestamp: new Date(latestBlock.timestamp * 1000),
            transactions: latestBlock.transactions.length
        });
        
    } catch (error) {
        console.error('测试失败:', error.message);
    }
}

testAccounts();
EOF
    
    # 创建网络测试脚本
    cat > "$ETHEREUM_DIR/test-network.js" << 'EOF'
// 网络连接测试脚本
const Web3 = require('web3');

async function testNetwork() {
    const web3 = new Web3('http://localhost:8545');
    
    try {
        console.log('=== 网络连接测试 ===');
        
        // 检查连接
        const isConnected = await web3.eth.net.isListening();
        console.log('网络连接状态:', isConnected);
        
        // 获取网络ID
        const networkId = await web3.eth.net.getId();
        console.log('网络ID:', networkId);
        
        // 获取节点信息
        const nodeInfo = await web3.eth.getNodeInfo();
        console.log('节点信息:', nodeInfo);
        
        // 检查挖矿状态
        const isMining = await web3.eth.isMining();
        console.log('挖矿状态:', isMining);
        
        // 获取Gas价格
        const gasPrice = await web3.eth.getGasPrice();
        console.log('当前Gas价格:', web3.utils.fromWei(gasPrice, 'gwei'), 'Gwei');
        
    } catch (error) {
        console.error('网络测试失败:', error.message);
    }
}

testNetwork();
EOF
    
    chmod +x "$ETHEREUM_DIR/test-accounts.js"
    chmod +x "$ETHEREUM_DIR/test-network.js"
    
    log_info "测试脚本已创建"
}

# 显示使用说明
show_usage() {
    log_info "=== 安装完成! ==="
    echo
    echo "工作目录: $HOME/ethereum"
    echo "Geth二进制文件: $HOME/go-ethereum/build/bin/geth"
    echo
    echo "使用方法:"
    echo "1. 启动私有链节点:"
    echo "   cd $HOME/ethereum && ./start-geth.sh"
    echo
    echo "2. 测试账户功能:"
    echo "   cd $HOME/ethereum && node test-accounts.js"
    echo
    echo "3. 测试网络连接:"
    echo "   cd $HOME/ethereum && node test-network.js"
    echo
    echo "4. 连接到Geth控制台:"
    echo "   geth attach $HOME/ethereum/data/geth.ipc"
    echo
    echo "5. 通过HTTP连接:"
    echo "   curl -X POST -H 'Content-Type: application/json' \\"
    echo "        --data '{\"jsonrpc\":\"2.0\",\"method\":\"eth_blockNumber\",\"params\":[],\"id\":1}' \\"
    echo "        http://localhost:8545"
    echo
    log_warn "注意: 请重新加载shell环境变量: source ~/.bashrc"
}

# 主函数
main() {
    echo "Go-Ethereum 自动化安装脚本"
    echo "================================"
    
    check_system
    install_dependencies
    install_go
    build_geth
    setup_workspace
    init_private_chain
    create_start_script
    install_nodejs
    create_test_scripts
    show_usage
    
    log_info "所有步骤完成!"
}

# 执行主函数
main "$@"