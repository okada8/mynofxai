#!/bin/bash
# NOFX每小时持仓报告脚本

API_KEY="nk_rpLQBBEueKmPe4tAmd550WzZoyeN4r1r"
API_BASE="http://localhost:8080/api"
TRADER_ID="421265fc_562011f9-7d1d-4b2e-8c38-41693f8eec20_deepseek_1772543540"
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S %Z')

echo "=== NOFX持仓报告 $TIMESTAMP ==="

# 1. 获取持仓
echo "📊 获取持仓数据..."
POSITIONS_JSON=$(curl -s -H "X-API-KEY: $API_KEY" "${API_BASE}/positions" 2>/dev/null)

# 2. 获取最新权益数据
echo "💰 获取账户权益..."
EQUITY_DATA=$(docker exec postgres_db psql -U postgres -d nofx -t -c "
SELECT 
  TO_CHAR(timestamp AT TIME ZONE 'Asia/Bangkok', 'YYYY-MM-DD HH24:MI:SS') as local_time,
  total_equity,
  balance,
  unrealized_pnl,
  position_count
FROM trader_equity_snapshots 
WHERE trader_id = '$TRADER_ID' 
ORDER BY timestamp DESC 
LIMIT 1;" 2>/dev/null | head -1)

# 3. 获取交易者配置中的初始余额（备用）
INITIAL_BALANCE=$(curl -s -H "X-API-KEY: $API_KEY" "${API_BASE}/traders/${TRADER_ID}/config" 2>/dev/null | \
  python3 -c "import json, sys; data=json.load(sys.stdin); print(data.get('initial_balance', 'N/A'))" 2>/dev/null || echo "N/A")

# 4. 解析持仓数据
POSITION_COUNT=0
POSITION_DETAILS=""
TOTAL_UNREALIZED_PNL=0

if [ -n "$POSITIONS_JSON" ] && [ "$POSITIONS_JSON" != "null" ] && [ "$POSITIONS_JSON" != "[]" ]; then
  POSITION_COUNT=$(echo "$POSITIONS_JSON" | python3 -c "
import json, sys
try:
    data = json.load(sys.stdin)
    print(len(data))
    total_pnl = 0
    for pos in data:
        total_pnl += float(pos['unrealized_pnl'])
    print(f'{total_pnl:.2f}')
except:
    print('0')
    print('0.00')
" 2>/dev/null)
  
  POSITION_COUNT=$(echo "$POSITION_COUNT" | head -1)
  TOTAL_UNREALIZED_PNL=$(echo "$POSITION_COUNT" | tail -1)
  
  # 提取持仓详情
  POSITION_DETAILS=$(echo "$POSITIONS_JSON" | python3 -c "
import json, sys
try:
    data = json.load(sys.stdin)
    details = []
    for pos in data:
        symbol = pos['symbol']
        side = pos['side']
        entry = float(pos['entry_price'])
        mark = float(pos['mark_price'])
        qty = float(pos['quantity'])
        leverage = int(pos['leverage'])
        margin = float(pos['margin_used'])
        pnl_usd = float(pos['unrealized_pnl'])
        pnl_pct = float(pos['unrealized_pnl_pct'])  # API返回的是百分比，不需要乘以100
        
        # 计算基于价格的百分比变化（额外信息）
        price_pct = ((mark - entry) / entry * 100) if entry > 0 else 0
        
        # 计算风险距离
        if side.upper() == 'LONG':
            stop_loss = entry * 0.985  # -1.5%
            take_profit = entry * 1.0225  # +2.25%
            distance_to_sl = (mark - stop_loss) / mark * 100
        else:  # SHORT
            stop_loss = entry * 1.015  # +1.5%
            take_profit = entry * 0.9775  # -2.25%
            distance_to_sl = (stop_loss - mark) / mark * 100
        
        details.append(f'• {symbol} {side}: {qty:.4f} @ {entry:.6f}')
        details.append(f'  当前: {mark:.6f}, 浮盈: {pnl_pct:.2f}% (${pnl_usd:.2f})')
        details.append(f'  价格变化: {price_pct:.2f}%, 杠杆: {leverage}x, 保证金: ${margin:.2f}')
        details.append(f'  止损距离: {distance_to_sl:.2f}%')
    
    print('\\n'.join(details))
except Exception as e:
    print('持仓解析失败')
" 2>/dev/null)
fi

# 5. 解析权益数据
if [ -n "$EQUITY_DATA" ]; then
  IFS='|' read -r LOCAL_TIME TOTAL_EQUITY BALANCE UNREALIZED_PNL POS_COUNT <<< "$(echo "$EQUITY_DATA" | sed 's/ *| */|/g')"
  LOCAL_TIME=$(echo "$LOCAL_TIME" | sed 's/^ *//;s/ *$//')
  TOTAL_EQUITY=$(echo "$TOTAL_EQUITY" | sed 's/^ *//;s/ *$//')
  BALANCE=$(echo "$BALANCE" | sed 's/^ *//;s/ *$//')
  UNREALIZED_PNL=$(echo "$UNREALIZED_PNL" | sed 's/^ *//;s/ *$//')
  POS_COUNT=$(echo "$POS_COUNT" | sed 's/^ *//;s/ *$//')
else
  LOCAL_TIME="N/A"
  TOTAL_EQUITY="N/A"
  BALANCE="N/A"
  UNREALIZED_PNL="0.00"
  POS_COUNT="0"
fi

# 6. 计算总盈亏（如果数据库未实现盈亏为0但API有持仓数据）
if [ "$POSITION_COUNT" -gt 0 ] && [ "$UNREALIZED_PNL" = "0" ] || [ "$UNREALIZED_PNL" = "0.00" ]; then
  UNREALIZED_PNL="$TOTAL_UNREALIZED_PNL"
fi

# 7. 格式化消息
REPORT_MESSAGE="🕐 *NOFX持仓报告* ($TIMESTAMP)

📊 *持仓概览*:
持仓数量: $POSITION_COUNT
总权益: \$$TOTAL_EQUITY
可用余额: \$$BALANCE
未实现盈亏: \$$UNREALIZED_PNL"

# 添加持仓详情
if [ "$POSITION_COUNT" -gt 0 ] && [ -n "$POSITION_DETAILS" ]; then
  REPORT_MESSAGE="$REPORT_MESSAGE

🔔 *持仓详情*:
$POSITION_DETAILS"
elif [ "$POSITION_COUNT" -eq 0 ]; then
  REPORT_MESSAGE="$REPORT_MESSAGE

✅ *无持仓* - 资金100%可用，等待AI交易信号"
else
  REPORT_MESSAGE="$REPORT_MESSAGE

⚠️ *持仓状态未知* - 请手动检查"
fi

# 添加数据时间戳
REPORT_MESSAGE="$REPORT_MESSAGE

📈 *数据时间*: $LOCAL_TIME (GMT+7)
💰 *初始余额*: \$$INITIAL_BALANCE"

echo "$REPORT_MESSAGE"
echo ""
echo "=== 报告生成完成 ==="

# 8. 保存报告到日志文件
LOG_FILE="/tmp/nofx_hourly_report.log"
echo "=== NOFX持仓报告 $TIMESTAMP ===" >> "$LOG_FILE"
echo "$REPORT_MESSAGE" >> "$LOG_FILE"
echo "" >> "$LOG_FILE"

# 9. 发送 Telegram 消息
# Load .env for credentials
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
ENV_FILE="$ROOT_DIR/.env"

if [ -f "$ENV_FILE" ]; then
    set -a
    set +e
    source "$ENV_FILE" 2>/dev/null || true
    set -e
    set +a
fi

if [ ! -z "$TELEGRAM_BOT_TOKEN" ] && [ ! -z "$TELEGRAM_CHAT_ID" ]; then
    # Construct JSON payload for HTML parsing
    # Escape special characters for JSON using Python for reliability
    SAFE_MESSAGE=$(echo "$REPORT_MESSAGE" | python3 -c "import json, sys; print(json.dumps(sys.stdin.read().strip())[1:-1])")
    
    # Send message using Markdown mode for better formatting (*bold*)
    
    JSON_PAYLOAD=$(cat <<EOF
{
  "chat_id": "$TELEGRAM_CHAT_ID",
  "text": "$SAFE_MESSAGE",
  "parse_mode": "Markdown"
}
EOF
)
    
    echo "Sending Telegram notification..."
    RESPONSE=$(curl -s -X POST "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/sendMessage" \
        -H "Content-Type: application/json" \
        -d "$JSON_PAYLOAD")
        
    if echo "$RESPONSE" | grep -q '"ok":true'; then
        echo "✅ Telegram notification sent."
    else
        echo "❌ Failed to send Telegram notification. Response: $RESPONSE"
        
        # Fallback: Try sending without parse_mode if Markdown fails
        echo "Retrying without parse_mode..."
        
        # Need to re-escape for JSON but without parse_mode
        # The SAFE_MESSAGE is already escaped for JSON structure
        
        JSON_PAYLOAD_PLAIN=$(cat <<EOF
{
  "chat_id": "$TELEGRAM_CHAT_ID",
  "text": "$SAFE_MESSAGE"
}
EOF
)
        RESPONSE_RETRY=$(curl -s -X POST "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/sendMessage" \
            -H "Content-Type: application/json" \
            -d "$JSON_PAYLOAD_PLAIN")

        if echo "$RESPONSE_RETRY" | grep -q '"ok":true'; then
             echo "✅ Telegram notification sent - plain text."
        else
             echo "❌ Failed to send Telegram notification - retry. Response: $RESPONSE_RETRY"
             exit 1
        fi
    fi

else
    echo "⚠️ Telegram credentials not found in .env, skipping notification."
fi

# 输出报告内容（供cron任务读取）
echo "REPORT_CONTENT_START"
echo "$REPORT_MESSAGE"
echo "REPORT_CONTENT_END"