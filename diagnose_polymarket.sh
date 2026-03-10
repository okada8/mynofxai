#!/bin/bash
echo "🔍 Polymarket机器人启动失败诊断工具"
echo "========================================"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo_step() {
    echo -e "${YELLOW}➤ $1${NC}"
}

echo_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

echo_error() {
    echo -e "${RED}❌ $1${NC}"
}

# 1. 检查Python环境
echo_step "1. 检查Python环境"
if ! command -v python3 &> /dev/null; then
    echo_error "Python3未安装"
    exit 1
fi
echo_success "Python3: $(python3 --version)"

# 2. 检查Python依赖
echo_step "2. 检查Python依赖"
DEPS=("py_clob_client" "web3" "eth_account")
for dep in "${DEPS[@]}"; do
    if ! python3 -c "import $dep" 2>/dev/null; then
        echo_error "$dep 未安装"
        echo "  运行: pip install $dep"
        exit 1
    fi
    echo_success "$dep 已安装"
done

# 3. 检查Python桥接脚本
echo_step "3. 检查Python桥接脚本"
SCRIPT_PATH="trader/polymarket/polymarket_bridge_fixed.py"
if [ ! -f "$SCRIPT_PATH" ]; then
    echo_error "Python桥接脚本不存在: $SCRIPT_PATH"
    exit 1
fi
echo_success "Python桥接脚本存在: $SCRIPT_PATH"

# 4. 测试Python桥接通信
echo_step "4. 测试Python桥接通信"
RESPONSE=$(echo '{"command": "ping", "timestamp": 1234567890}' | python3 "$SCRIPT_PATH" 2>&1)
if echo "$RESPONSE" | grep -q "pong"; then
    echo_success "Python桥接通信正常"
    echo "  响应: $RESPONSE"
else
    echo_error "Python桥接通信失败"
    echo "  响应: $RESPONSE"
    exit 1
fi

# 5. 检查Go模块和文件
echo_step "5. 检查Go模块和文件"
POLY_FILES=(
    "trader/polymarket/trader.go"
    "trader/polymarket/contract.go"
    "trader/polymarket/client.go"
    "trader/polymarket/python_wrapper.go"
)

for file in "${POLY_FILES[@]}"; do
    if [ ! -f "$file" ]; then
        echo_error "文件不存在: $file"
        exit 1
    fi
    echo_success "文件存在: $file"
done

# 6. 检查合约地址
echo_step "6. 检查合约地址"
echo "CTF Exchange地址: 0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E"
echo "USDC地址: 0x3c499c542cEF5E3811e1192ce70d8cC03d5c3359"
echo "CLOB Proxy地址: 0x9A8C4bd4d4259e2f9986f4D8f8eB8E5e7d25dA75"
echo "注意: 这些是Polygon主网地址，测试时请使用Mumbai测试网"

# 7. 模拟交易器初始化
echo_step "7. 模拟交易器初始化流程"
echo "这可能会失败，因为需要真实的私钥和RPC连接"

# 创建一个简单的测试程序
cat > /tmp/test_poly_init.py << 'EOF'
import sys
import os
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)) + "/..")

try:
    from trader.polymarket.polymarket_bridge_fixed import main
    print("✅ 可以导入polymarket_bridge_fixed")
    
    # 测试初始化（使用虚拟私钥）
    import json
    import subprocess
    
    script_path = "trader/polymarket/polymarket_bridge_fixed.py"
    
    # 启动进程
    proc = subprocess.Popen(
        ['python3', script_path],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True
    )
    
    # 发送初始化命令（虚拟私钥）
    init_cmd = json.dumps({
        "command": "init",
        "key": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
        "chain_id": 137,
        "rpc_url": "https://polygon-rpc.com"
    })
    
    stdout, stderr = proc.communicate(input=init_cmd + "\n", timeout=10)
    
    print(f"初始化响应: {stdout.strip()}")
    
    if stderr:
        print(f"标准错误: {stderr}")
        
except Exception as e:
    print(f"❌ 测试失败: {e}")
    import traceback
    traceback.print_exc()
EOF

cd /Users/tom/Desktop/nofi && python3 /tmp/test_poly_init.py

# 8. 检查常见问题
echo_step "8. 常见问题检查"
echo "可能的问题:"
echo "1. RPC URL无法访问 - 尝试使用公共RPC: https://polygon-rpc.com"
echo "2. 私钥格式错误 - 确保是64字符十六进制（可带0x前缀）"
echo "3. 钱包中没有MATIC支付Gas - 需要少量MATIC"
echo "4. 钱包中没有USDC - 需要USDC进行交易"
echo "5. 合约地址过时 - 需要验证当前合约地址"
echo "6. Python进程通信超时 - 检查Python脚本是否有语法错误"

# 9. 推荐测试配置
echo_step "9. 推荐测试配置"
echo "对于开发测试，建议:"
echo "1. 使用Polygon Mumbai测试网"
echo "2. RPC URL: https://rpc-mumbai.maticvigil.com"
echo "3. 从Polygon水龙头获取测试MATIC"
echo "4. 使用测试网USDC合约"
echo "5. 创建新的测试钱包，不要使用主网私钥"

# 10. 检查配置示例
echo_step "10. NOFX配置示例"
cat << 'EOF'
在NOFX配置中添加:
{
  "exchange": "polymarket",
  "polymarketPrivateKey": "0x你的私钥",
  "polymarketWalletAddr": "0x你的地址",
  "polymarketRPCURL": "https://polygon-rpc.com",
  "initialBalance": 100,
  "strategy": "polymarket_basic"
}
EOF

echo ""
echo "========================================"
echo_success "诊断完成"
echo ""
echo "如果机器人仍然启动失败，请检查:"
echo "1. NOFX日志文件中的详细错误信息"
echo "2. Python进程的标准错误输出"
echo "3. 网络连接和RPC可用性"
echo "4. 钱包余额和权限"