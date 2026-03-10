#!/bin/bash
echo "🔍 检查 Polymarket 集成环境..."

# 1. 检查Python
if ! command -v python3 &> /dev/null; then
    echo "❌ Python3 未安装"
    exit 1
fi
echo "✅ Python3: $(python3 --version)"

# 2. 检查py-clob-client
echo "检查 py-clob-client 依赖..."
if ! python3 -c "import py_clob_client" 2>/dev/null; then
    echo "❌ py-clob-client 未安装"
    echo "安装命令: pip install py-clob-client"
    read -p "是否现在安装？[y/N] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        pip install py-clob-client
    else
        echo "请手动安装: pip install py-clob-client"
        exit 1
    fi
fi

# 3. 检查web3和eth-account
echo "检查 web3 和 eth-account..."
if ! python3 -c "import web3, eth_account" 2>/dev/null; then
    echo "❌ web3 或 eth-account 未安装"
    echo "安装命令: pip install web3 eth-account"
    read -p "是否现在安装？[y/N] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        pip install web3 eth-account
    else
        echo "请手动安装: pip install web3 eth-account"
        exit 1
    fi
fi

echo "✅ 所有Python依赖已安装"

# 4. 测试Python桥接
echo "🧪 测试Python桥接..."
cd "$(dirname "$0")"
RESPONSE=$(echo '{"command": "ping", "timestamp": 1234567890}' | python3 trader/polymarket/polymarket_bridge_fixed.py 2>/dev/null)

if echo "$RESPONSE" | grep -q "pong"; then
    echo "✅ Python桥接测试通过"
else
    echo "❌ Python桥接测试失败"
    echo "响应: $RESPONSE"
    exit 1
fi

# 5. 检查Go模块
echo "检查Go模块..."
cd "$(dirname "$0")"
if ! go list ./trader/polymarket/... 2>/dev/null; then
    echo "⚠️  Go模块可能有问题，尝试运行: go mod tidy"
    go mod tidy
fi

echo "✅ 环境检查完成"
echo ""
echo "下一步:"
echo "1. 确保你有Polygon钱包私钥和地址"
echo "2. 配置NOFX使用Polymarket交易所"
echo "3. 使用Mumbai测试网进行测试: https://rpc-mumbai.maticvigil.com"
echo "4. 测试网USDC可从Polygon水龙头获取"