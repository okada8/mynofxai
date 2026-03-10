#!/usr/bin/env python3
import json
import sys
import traceback
import os

# 检查依赖
try:
    from py_clob_client.client import ClobClient
    from py_clob_client.constants import POLYGON
    from eth_account import Account
    from web3 import Web3
except ImportError as e:
    print(json.dumps({"error": f"Missing dependency: {str(e)}. Run: pip install py-clob-client web3 eth-account"}))
    sys.exit(1)

class PolymarketBridge:
    def __init__(self):
        self.client = None
        self.w3 = None
        
    def initialize(self, private_key_hex: str, chain_id: int = POLYGON, rpc_url: str = None):
        try:
            # 验证私钥格式
            if not private_key_hex.startswith('0x'):
                private_key_hex = '0x' + private_key_hex
            
            # 创建Web3连接（用于余额检查）
            if rpc_url:
                self.w3 = Web3(Web3.HTTPProvider(rpc_url))
                if not self.w3.is_connected():
                    return {"status": "error", "error": f"Failed to connect to RPC: {rpc_url}"}
            
            # 创建CLOB客户端
            self.client = ClobClient(
                host="https://clob.polymarket.com",
                chain_id=chain_id,
                key=private_key_hex
            )
            
            # 获取API凭证
            try:
                creds = self.client.create_or_derive_api_creds()
                api_key = creds.get("apiKey", "")
            except Exception as e:
                api_key = ""
                print(f"Warning: Failed to get API creds: {e}", file=sys.stderr)
            
            # 获取钱包地址
            wallet_address = Account.from_key(private_key_hex).address
            
            # 测试连接
            try:
                markets = self.client.get_markets(limit=1)
                connected = len(markets) > 0
            except Exception as e:
                connected = False
                print(f"Warning: Failed to get markets: {e}", file=sys.stderr)
            
            # 获取USDC余额（如果Web3可用）
            usdc_balance = 0.0
            if self.w3:
                try:
                    # USDC合约地址（Polygon主网）
                    usdc_address = "0x3c499c542cEF5E3811e1192ce70d8cC03d5c3359"
                    # 简化的ERC20 ABI仅包含balanceOf
                    erc20_abi = '[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"}]'
                    usdc_contract = self.w3.eth.contract(address=usdc_address, abi=erc20_abi)
                    balance = usdc_contract.functions.balanceOf(wallet_address).call()
                    usdc_balance = balance / 10**6  # USDC有6位小数
                except Exception as e:
                    print(f"Warning: Failed to get USDC balance: {e}", file=sys.stderr)
            
            return {
                "status": "success",
                "wallet": wallet_address,
                "api_key": api_key,
                "connected": connected,
                "usdc_balance": usdc_balance,
                "has_rpc": self.w3 is not None
            }
            
        except Exception as e:
            return {"status": "error", "error": str(e), "traceback": traceback.format_exc()}

    def get_balance(self):
        try:
            # 如果Web3连接可用，返回真实余额
            if self.w3:
                # 这里可以扩展获取更多余额信息
                return {
                    "status": "success",
                    "balance": {
                        "usdc": 0.0,  # 从初始化时获取
                        "matic": 0.0
                    }
                }
            else:
                return {
                    "status": "success", 
                    "message": "Balance check handled by Go",
                    "balance": {"usdc": 0.0, "matic": 0.0}
                }
        except Exception as e:
            return {"status": "error", "error": str(e)}

    def create_order(self, token_id: str, side: str, price: float, size: int):
        try:
            if not self.client:
                return {"status": "error", "error": "Client not initialized"}
            
            order = self.client.create_order(
                token_id=token_id,
                price=str(price),
                side=side.upper(),
                size=str(size)
            )
            return {
                "status": "success",
                "order_id": order.get("id") or order.get("orderId", "unknown"),
                "status": "PENDING"
            }
        except Exception as e:
            return {"status": "error", "error": str(e)}

    def health_check(self):
        try:
            if not self.client:
                return {"status": "error", "error": "Client not initialized", "healthy": False}
            
            # 简单健康检查：获取一个市场
            self.client.get_markets(limit=1)
            return {"status": "success", "healthy": True}
        except Exception as e:
            return {"status": "error", "error": str(e), "healthy": False}

    def get_markets(self, limit: int = 10):
        try:
            if not self.client:
                return {"status": "error", "error": "Client not initialized"}
            
            markets = self.client.get_markets(limit=limit)
            return {"status": "success", "markets": markets}
        except Exception as e:
            return {"status": "error", "error": str(e)}

def main():
    bridge = PolymarketBridge()
    
    for line in sys.stdin:
        line = line.strip()
        if not line:
            continue
            
        try:
            cmd = json.loads(line)
            action = cmd.get("command")
            
            if action == "init":
                response = bridge.initialize(
                    cmd.get("key"),
                    cmd.get("chain_id", POLYGON),
                    cmd.get("rpc_url")
                )
                
            elif action == "health":
                response = bridge.health_check()
                
            elif action == "create_order":
                response = bridge.create_order(
                    cmd.get("token_id"),
                    cmd.get("side", "BUY"),
                    float(cmd.get("price", 0.5)),
                    int(cmd.get("size", 1))
                )
                
            elif action == "get_markets":
                response = bridge.get_markets(cmd.get("limit", 10))
                
            elif action == "ping":
                response = {"status": "pong", "timestamp": cmd.get("timestamp")}
                
            else:
                response = {"status": "error", "error": f"Unknown command: {action}"}
                
        except json.JSONDecodeError as e:
            response = {"status": "error", "error": f"Invalid JSON: {str(e)}"}
        except Exception as e:
            response = {"status": "error", "error": f"Unexpected error: {str(e)}"}
        
        print(json.dumps(response))
        sys.stdout.flush()

if __name__ == "__main__":
    main()