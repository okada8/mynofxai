# /Users/tom/Desktop/nofi/polymarket_py/start.sh
#!/bin/bash
cd "$(dirname "$0")/.."

# 检查Python环境
if ! command -v python3 &> /dev/null; then
    echo "❌ Python3 not found"
    exit 1
fi

# 检查依赖
if ! pip show py-clob-client &> /dev/null; then
    echo "📦 Installing py-clob-client..."
    pip install py-clob-client web3 eth-account
fi

# 测试Python客户端
echo "🧪 Testing Python client..."
python3 -c "
from py_clob_client.client import ClobClient
from py_clob_client.constants import POLYGON
print('✅ py-clob-client import successful')
"

echo "🚀 Ready to use Polymarket integration"
