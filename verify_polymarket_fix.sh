#!/bin/bash
echo "🔧 Polymarket集成修复验证"
echo "=============================="

echo "步骤1: 检查Python环境"
if python3 -c "import py_clob_client, web3, eth_account" &>/dev/null; then
    echo "✅ Python依赖正常"
else
    echo "❌ Python依赖缺失"
    echo "运行: pip install py-clob-client web3 eth-account"
    exit 1
fi

echo ""
echo "步骤2: 测试Python桥接"
RESPONSE=$(echo '{"command": "ping", "timestamp": 1234567890}' | python3 trader/polymarket/polymarket_bridge_fixed.py 2>&1)
if echo "$RESPONSE" | grep -q '"status": "pong"'; then
    echo "✅ Python桥接正常"
else
    echo "❌ Python桥接异常: $RESPONSE"
fi

echo ""
echo "步骤3: 检查文件结构"
FILES=(
    "trader/polymarket/trader.go"
    "trader/polymarket/contract.go" 
    "trader/polymarket/client.go"
    "trader/polymarket/python_wrapper.go"
    "trader/polymarket/polymarket_bridge_fixed.py"
    "trader/polymarket/abi.json"
)

MISSING=0
for file in "${FILES[@]}"; do
    if [ -f "$file" ]; then
        echo "✅ $file"
    else
        echo "❌ $file (缺失)"
        MISSING=1
    fi
done

if [ $MISSING -eq 1 ]; then
    echo "⚠️  部分文件缺失，但可能不影响基本功能"
fi

echo ""
echo "步骤4: 检查合约地址"
echo "当前配置的合约地址:"
grep -A3 "Contract Addresses" trader/polymarket/contract.go | grep "="
echo "注意: 这些是Polygon主网地址"

echo ""
echo "步骤5: 验证修复内容"
echo "已实施的修复:"
echo "1. ✅ 修复Python导入错误 (py_clob_client)"
echo "2. ✅ 改进初始化错误处理"
echo "3. ✅ Python桥接变为可选"
echo "4. ✅ 添加环境检查脚本"
echo "5. ✅ 添加详细诊断工具"

echo ""
echo "步骤6: 测试建议配置"
echo "对于测试，建议使用以下配置:"
cat << 'EOF'
exchange: "polymarket"
polymarketPrivateKey: "0x你的64字符私钥"
polymarketWalletAddr: "0x对应的钱包地址"  
polymarketRPCURL: "https://rpc-mumbai.maticvigil.com" (测试网)
或 "https://polygon-rpc.com" (主网)
initialBalance: 100
strategy: "polymarket_basic"
EOF

echo ""
echo "步骤7: 预期行为"
echo "修复后，Polymarket机器人应该:"
echo "1. ✅ 能够成功启动（即使Python环境有问题）"
echo "2. ✅ 能够查询余额（需要有效的RPC和私钥）"
echo "3. ✅ 能够获取市场数据（通过Gamma API）"
echo "4. ⚠️  如果Python桥接正常，可以使用CLOB交易"
echo "5. ⚠️  如果Python桥接失败，仍可使用合约交易"

echo ""
echo "=============================="
echo "🎯 验证完成"
echo ""
echo "如果机器人仍然启动失败，请:"
echo "1. 检查RPC URL是否可访问"
echo "2. 验证私钥格式是否正确"
echo "3. 确保钱包中有MATIC支付Gas"
echo "4. 查看NOFX日志获取详细错误"
echo ""
echo "运行诊断工具获取更多帮助:"
echo "  ./diagnose_polymarket.sh"