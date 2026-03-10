#!/usr/bin/env python3
import sys
import os
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

# 测试Python桥接是否可以导入
try:
    from trader.polymarket.polymarket_bridge_fixed import main
    print("✅ polymarket_bridge_fixed.py 导入成功")
except ImportError as e:
    print(f"❌ 导入失败: {e}")
    sys.exit(1)

# 测试py-clob-client是否可以导入
try:
    from py_clob_client.client import ClobClient
    print("✅ py_clob_client 导入成功")
except ImportError as e:
    print(f"❌ py_clob_client 导入失败: {e}")
    print("运行: pip install py-clob-client")
    sys.exit(1)

# 测试web3和eth-account
try:
    from web3 import Web3
    from eth_account import Account
    print("✅ web3 和 eth_account 导入成功")
except ImportError as e:
    print(f"❌ web3 或 eth_account 导入失败: {e}")
    print("运行: pip install web3 eth-account")
    sys.exit(1)

print("\n✅ 所有Python依赖检查通过")
print("\n测试Python桥接通信...")

# 创建一个简单的测试
import subprocess
import json

# 启动Python桥接进程
script_path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "trader/polymarket/polymarket_bridge_fixed.py")
if not os.path.exists(script_path):
    print(f"❌ 脚本不存在: {script_path}")
    sys.exit(1)

print(f"使用脚本: {script_path}")

# 测试进程通信
try:
    proc = subprocess.Popen(
        ['python3', script_path],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True
    )
    
    # 发送ping命令
    ping_cmd = json.dumps({"command": "ping", "timestamp": 1234567890})
    print(f"发送: {ping_cmd}")
    
    stdout, stderr = proc.communicate(input=ping_cmd + "\n", timeout=5)
    
    if stderr:
        print(f"⚠️  stderr: {stderr}")
    
    print(f"接收: {stdout.strip()}")
    
    response = json.loads(stdout.strip())
    if response.get("status") == "pong":
        print("✅ Python桥接通信测试通过")
    else:
        print(f"❌ 响应异常: {response}")
        
except Exception as e:
    print(f"❌ 测试失败: {e}")
    sys.exit(1)

print("\n🎯 Python桥接测试完成")
print("\n注意: 实际使用时需要:")
print("1. 真实的Polygon钱包私钥")
print("2. 配置正确的RPC URL")
print("3. 钱包中有USDC用于交易")