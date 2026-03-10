# /Users/tom/Desktop/nofi/polymarket_py/client.py
import json
import os
import sys
from typing import Dict, Any, Optional
from py_clob_client.client import ClobClient
from py_clob_client.constants import POLYGON
from eth_account import Account
import time

class PolymarketPythonClient:
    """Polymarket Python客户端，通过stdin/stdout与Go通信"""
    
    def __init__(self):
        self.client = None
        self.credentials = None
        
    def initialize(self, private_key: str, chain_id: int = POLYGON):
        """初始化客户端"""
        try:
            # 创建客户端
            self.client = ClobClient(
                host="https://clob.polymarket.com",
                chain_id=chain_id,
                key=private_key
            )
            
            # 获取或创建API凭证
            self.credentials = self.client.create_or_derive_api_creds()
            
            return {
                "status": "success",
                "wallet": Account.from_key(private_key).address,
                "api_key": self.credentials["apiKey"],
                "chain_id": chain_id
            }
        except Exception as e:
            return {
                "status": "error",
                "error": str(e)
            }
    
    def get_balance(self):
        """获取余额"""
        try:
            # 注意：Polymarket余额在链上，不是通过API
            # 这里返回模拟数据或通过web3获取
            return {
                "status": "success",
                "balance": {
                    "usdc": 0.0,  # 需要实现USDC余额查询
                    "matic": 0.0   # 需要实现MATIC余额查询
                }
            }
        except Exception as e:
            return {"status": "error", "error": str(e)}
    
    def get_markets(self, limit: int = 20):
        """获取市场列表"""
        try:
            markets = self.client.get_markets(limit=limit)
            return {
                "status": "success",
                "markets": markets
            }
        except Exception as e:
            return {"status": "error", "error": str(e)}
    
    def buy_shares(self, token_id: str, price: float, size: int, side: str = "BUY"):
        """购买份额"""
        try:
            order = self.client.create_and_post_order({
                "token_id": token_id,
                "price": price,
                "size": size,
                "side": side
            })
            
            return {
                "status": "success",
                "order_id": order.get("order_id"),
                "status": order.get("status")
            }
        except Exception as e:          
            return {"status": "error", "error": str(e)}
    
    def run_command(self, command: Dict[str, Any]) -> Dict[str, Any]:
        """执行命令"""
        cmd = command.get("command")
        
        if cmd == "initialize":return self.initialize(
                command["private_key"],
                command.get("chain_id", POLYGON)
            )
        elif cmd == "get_balance":
            return self.get_balance()
        elif cmd == "get_markets":
            return self.get_markets(command.get("limit", 20))
        elif cmd == "buy_shares":
            return self.buy_shares(
                command["token_id"],
                command["price"],
                command["size"],
                command.get("side", "BUY")
            )
        elif cmd == "sell_shares":
            return self.buy_shares(
                command["token_id"],
                command["price"],
                command["size"],
                "SELL"
            )
        else:
            return {"status": "error", "error": f"Unknown command: {cmd}"}

def main():
    """主函数：通过stdin/stdout与Go通信"""
    client = PolymarketPythonClient()
    
    # 从stdin读取JSON命令
    for line in sys.stdin:
        if not line.strip():
            continue
            
        try:
            command = json.loads(line)
            result = client.run_command(command)
            
            # 写入结果到stdout
            print(json.dumps(result))
            sys.stdout.flush()
            
        except json.JSONDecodeError as e:
            error_result = {
                "status": "error",
                "error": f"JSON decode error: {str(e)}"
            }
            print(json.dumps(error_result))
            sys.stdout.flush()
        except Exception as e:
            error_result = {
                "status": "error", 
                "error": f"Unexpected error: {str(e)}"
            }
            print(json.dumps(error_result))
            sys.stdout.flush()

if __name__ == "__main__":
    main()



