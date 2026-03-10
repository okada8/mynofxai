import json
import sys
import os
import traceback
from py_clob_client.client import ClobClient
from py_clob_client.headers.values import ApiCreds # Import ApiCreds explicitly

# Check if py_clob_client is installed
try:
    import py_clob_client
except ImportError:
    print(json.dumps({"error": "py_clob_client not installed"}))
    sys.exit(1)

class PolymarketBridge:
    def __init__(self):
        self.client = None
        
    def initialize(self, key, chain_id, funder=None, signature_type=1, api_key=None, api_secret=None, api_passphrase=None): # Default to Proxy (1)
        try:
            host = "https://clob.polymarket.com"
            
            creds = None
            # Use provided L2 credentials if available
            if api_key and api_secret and api_passphrase:
                creds = ApiCreds(
                    api_key=api_key,
                    api_secret=api_secret,
                    api_passphrase=api_passphrase
                )
            
            # 3. Initialize ClobClient
            self.client = ClobClient(
                host,
                key=key,
                chain_id=chain_id,
                creds=creds,
                signature_type=signature_type,
                funder=funder
            )
            
            return {
                "status": "success", 
                "api_key": creds.api_key if creds else "derived", 
                "funder": funder or "derived",
                "signature_type": signature_type
            }
        except Exception as e:
            return {"error": str(e), "traceback": traceback.format_exc()}

    def get_balance(self):
        if not self.client:
            return {"error": "Client not initialized"}
        try:
            # 这里的实现取决于 py_polymarket 提供的功能
            # 假设我们需要从链上获取 USDC 余额，或者 ClobClient 有相关方法
            # 如果 ClobClient 没有直接获取余额的方法，我们可能需要使用 web3.py
            # 为了简单起见，我们先尝试获取 API 密钥相关的余额信息（如果支持）
            # 或者返回一个占位符，因为 Go 代码中已经有了通过合约获取余额的逻辑
            # 这里我们主要关注订单操作
            return {"status": "success", "message": "Balance check handled by Go"}
        except Exception as e:
            return {"error": str(e)}

    def create_order(self, token_id, side, price, size):
        if not self.client:
            return {"error": "Client not initialized"}
        try:
            resp = self.client.create_order(
                token_id=token_id,
                price=price,
                side=side, # BUY or SELL
                size=size
            )
            return {"status": "success", "response": resp}
        except Exception as e:
            return {"error": str(e)}

    def cancel_order(self, order_id):
        if not self.client:
            return {"error": "Client not initialized"}
        try:
            resp = self.client.cancel(order_id)
            return {"status": "success", "response": resp}
        except Exception as e:
            return {"error": str(e)}

    def cancel_all(self):
        if not self.client:
            return {"error": "Client not initialized"}
        try:
            resp = self.client.cancel_all()
            return {"status": "success", "response": resp}
        except Exception as e:
            return {"error": str(e)}

    def get_market_price(self, token_id):
        # 获取市场价格（例如 mid price 或 last trade）
        if not self.client:
             return {"error": "Client not initialized"}
        try:
             # ClobClient 通常有获取 orderbook 或 ticker 的方法
             book = self.client.get_order_book(token_id)
             # 简单计算 mid price
             best_bid = 0
             best_ask = 0
             if book.bids:
                 best_bid = float(book.bids[0].price)
             if book.asks:
                 best_ask = float(book.asks[0].price)
             
             mid_price = (best_bid + best_ask) / 2 if (best_bid and best_ask) else (best_bid or best_ask)
             return {"status": "success", "price": mid_price}
        except Exception as e:
            return {"error": str(e)}

    def get_open_orders(self, token_id=None):
        if not self.client:
            return {"error": "Client not initialized"}
        try:
            # Try to fetch open orders
            # If token_id is None, fetch all (if supported), otherwise filter
            # Note: library method signature might vary, assuming standard params
            resp = self.client.get_orders(market=token_id, status="OPEN")
            return {"status": "success", "orders": resp}
        except Exception as e:
            return {"error": str(e)}
            
    def get_positions(self):
        if not self.client:
            return {"error": "Client not initialized"}
        try:
            # Attempt to get positions if supported
            # Otherwise return empty list or error
            if hasattr(self.client, 'get_positions'):
                resp = self.client.get_positions()
                return {"status": "success", "positions": resp}
            else:
                return {"status": "error", "error": "get_positions not supported by client"}
        except Exception as e:
            return {"error": str(e)}

bridge = PolymarketBridge()

def process_command(line):
    try:
        cmd = json.loads(line)
        action = cmd.get("command")
        
        if action == "ping":
            return {"status": "pong", "timestamp": cmd.get("timestamp")}
            
        elif action == "init":
            return bridge.initialize(
                cmd.get("key"), 
                cmd.get("chain_id", 137),
                funder=cmd.get("funder"),
                signature_type=cmd.get("signature_type", 1), # Default to 1 (Proxy)
                api_key=cmd.get("api_key"),
                api_secret=cmd.get("api_secret"),
                api_passphrase=cmd.get("api_passphrase")
            )
            
        elif action == "create_order":
            return bridge.create_order(
                cmd.get("token_id"), 
                cmd.get("side"), 
                cmd.get("price"), 
                cmd.get("size")
            )
            
        elif action == "cancel_order":
            return bridge.cancel_order(cmd.get("order_id"))
            
        elif action == "cancel_all":
            return bridge.cancel_all()
            
        elif action == "get_price":
            return bridge.get_market_price(cmd.get("token_id"))
            
        elif action == "get_open_orders":
            return bridge.get_open_orders(cmd.get("token_id"))
            
        elif action == "get_positions":
            return bridge.get_positions()
            
        else:
            return {"error": f"Unknown command: {action}"}
            
    except Exception as e:
        return {"error": f"Processing error: {str(e)}", "traceback": traceback.format_exc()}

def main():
    # 从标准输入读取命令
    for line in sys.stdin:
        line = line.strip()
        if not line:
            continue
            
        response = process_command(line)
        print(json.dumps(response))
        sys.stdout.flush()

if __name__ == "__main__":
    main()