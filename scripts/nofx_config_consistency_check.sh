#!/bin/bash
# NOFX配置一致性检查脚本
# 检查交易者配置、AI提示词/策略配置、数据库配置之间的一致性

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_FILE="/tmp/nofx_consistency_check_$(date +%Y%m%d_%H%M%S).log"
REPORT_FILE="/tmp/nofx_consistency_report_$(date +%Y%m%d_%H%M%S).md"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 检查计数器
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0
WARNING_CHECKS=0

# 日志函数
log() {
    echo -e "$1" | tee -a "$LOG_FILE"
}

report() {
    echo -e "$1" | tee -a "$REPORT_FILE"
}

check_pass() {
    ((TOTAL_CHECKS++))
    ((PASSED_CHECKS++))
    log "${GREEN}✅ PASS${NC}: $1"
    report "✅ **PASS**: $1"
}

check_fail() {
    ((TOTAL_CHECKS++))
    ((FAILED_CHECKS++))
    log "${RED}❌ FAIL${NC}: $1"
    report "❌ **FAIL**: $1"
}

check_warning() {
    ((TOTAL_CHECKS++))
    ((WARNING_CHECKS++))
    log "${YELLOW}⚠️ WARNING${NC}: $1"
    report "⚠️ **WARNING**: $1"
}

# 初始化报告文件
echo "# NOFX配置一致性检查报告" > "$REPORT_FILE"
echo "**生成时间**: $(date '+%Y-%m-%d %H:%M:%S %Z')" >> "$REPORT_FILE"
echo "**检查类型**: 配置一致性检查（跨模块验证）" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

log "${BLUE}=== NOFX配置一致性检查开始 ===${NC}"
log "检查时间: $(date '+%Y-%m-%d %H:%M:%S %Z')"
log "日志文件: $LOG_FILE"
log "报告文件: $REPORT_FILE"
log "检查目标: 验证交易者配置、AI策略配置、数据库配置之间的一致性"
log ""

# ==================== 第一步：提取配置信息 ====================
log "${BLUE}=== 1. 提取配置信息 ===${NC}"
report "## 1. 配置信息提取"

# 1.1 从hourly_report.sh提取交易者配置
HOURLY_REPORT="$SCRIPT_DIR/hourly_report.sh"
if [ -f "$HOURLY_REPORT" ]; then
    log "从 $HOURLY_REPORT 提取交易者配置..."
    
    TRADER_API_KEY=$(grep "API_KEY=" "$HOURLY_REPORT" | head -1 | cut -d'"' -f2 2>/dev/null || grep "API_KEY=" "$HOURLY_REPORT" | head -1 | cut -d'=' -f2 | tr -d "'\"")
    TRADER_API_BASE=$(grep "API_BASE=" "$HOURLY_REPORT" | head -1 | cut -d'"' -f2 2>/dev/null || grep "API_BASE=" "$HOURLY_REPORT" | head -1 | cut -d'=' -f2 | tr -d "'\"")
    TRADER_ID=$(grep "TRADER_ID=" "$HOURLY_REPORT" | head -1 | cut -d'"' -f2 2>/dev/null || grep "TRADER_ID=" "$HOURLY_REPORT" | head -1 | cut -d'=' -f2 | tr -d "'\"")
    
    log "  交易者API_KEY: ${TRADER_API_KEY:0:10}..."
    log "  交易者API_BASE: $TRADER_API_BASE"
    log "  交易者ID: $TRADER_ID"
    
    # 尝试从API获取交易者配置详情
    log "  查询交易者配置详情..."
    TRADER_CONFIG_JSON=$(curl -s -H "X-API-KEY: $TRADER_API_KEY" "${TRADER_API_BASE}/traders/${TRADER_ID}/config" 2>/dev/null)
    if [ -n "$TRADER_CONFIG_JSON" ] && [ "$TRADER_CONFIG_JSON" != "null" ] && [ "$TRADER_CONFIG_JSON" != "Not Found" ]; then
        TRADER_INITIAL_BALANCE=$(echo "$TRADER_CONFIG_JSON" | python3 -c "import json, sys; data=json.load(sys.stdin); print(data.get('initial_balance', 'N/A'))" 2>/dev/null || echo "N/A")
        TRADER_ALLOWED_SYMBOLS=$(echo "$TRADER_CONFIG_JSON" | python3 -c "import json, sys; data=json.load(sys.stdin); print(','.join(data.get('trading_symbols', [])))" 2>/dev/null || echo "N/A")
        log "  交易者初始余额: $TRADER_INITIAL_BALANCE"
        log "  交易者允许币种: ${TRADER_ALLOWED_SYMBOLS:0:50}..."
    else
        TRADER_INITIAL_BALANCE="N/A"
        TRADER_ALLOWED_SYMBOLS="N/A"
        log "  无法获取交易者配置详情"
    fi
    
    check_pass "交易者配置提取完成"
else
    check_fail "交易者配置文件不存在: $HOURLY_REPORT"
    TRADER_API_KEY=""
    TRADER_API_BASE=""
    TRADER_ID=""
    TRADER_INITIAL_BALANCE="N/A"
    TRADER_ALLOWED_SYMBOLS="N/A"
fi

# 1.2 从优化策略文件提取AI策略配置
PROMPT_DIR="$SCRIPT_DIR/prompt"
STRATEGY_FILE=""

# 遍历目录查找包含 risk_control 的 json 文件
if [ -d "$PROMPT_DIR" ]; then
    log "在 $PROMPT_DIR 中查找策略文件..."
    while IFS= read -r file; do
        if grep -q "risk_control" "$file"; then
            STRATEGY_FILE="$file"
            log "找到策略文件: $(basename "$file")"
            break
        fi
    done < <(find "$PROMPT_DIR" -name "*.json" -maxdepth 1)
fi

if [ -n "$STRATEGY_FILE" ] && [ -f "$STRATEGY_FILE" ]; then
    log "从 $STRATEGY_FILE 提取AI策略配置..."
    
    # 检查JSON格式
    if python3 -m json.tool "$STRATEGY_FILE" &> /dev/null; then
        # 提取nofxos_api_key
        STRATEGY_NOFXOS_API_KEY=$(grep -o '"nofxos_api_key": *"[^"]*"' "$STRATEGY_FILE" | cut -d'"' -f4)
        
        # 提取允许交易币种
        STRATEGY_ALLOWED_COINS=$(grep -A 30 '"static_coins"' "$STRATEGY_FILE" | grep -o '"[A-Z0-9]*USDT"' | tr '\n' ',' | sed 's/,$//' | sed 's/"//g')
        STRATEGY_COINS_COUNT=$(echo "$STRATEGY_ALLOWED_COINS" | tr ',' '\n' | wc -l | tr -d ' ')
        
        # 提取风险控制参数
        STRATEGY_MAX_POSITIONS=$(grep -o '"max_positions": *[0-9]*' "$STRATEGY_FILE" | grep -o '[0-9]*' | head -1)
        STRATEGY_BTC_LEVERAGE=$(grep -o '"btc_eth_max_leverage": *[0-9]*' "$STRATEGY_FILE" | grep -o '[0-9]*' | head -1)
        STRATEGY_ALTCOIN_LEVERAGE=$(grep -o '"altcoin_max_leverage": *[0-9]*' "$STRATEGY_FILE" | grep -o '[0-9]*' | head -1)
        
        # 提取止损参数
        STRATEGY_BTC_STOP_LOSS=$(grep -A 5 '"stop_loss_percent"' "$STRATEGY_FILE" | grep -o '"btc": *[0-9.]*' | grep -o '[0-9.]*' | head -1)
        STRATEGY_MAIN_COINS_STOP_LOSS=$(grep -A 5 '"stop_loss_percent"' "$STRATEGY_FILE" | grep -o '"main_coins": *[0-9.]*' | grep -o '[0-9.]*' | head -1)
        STRATEGY_ALTCOINS_STOP_LOSS=$(grep -A 5 '"stop_loss_percent"' "$STRATEGY_FILE" | grep -o '"altcoins": *[0-9.]*' | grep -o '[0-9.]*' | head -1)
        
        log "  策略NOFXOS_API_KEY: ${STRATEGY_NOFXOS_API_KEY:0:10}..."
        log "  策略允许币种数量: $STRATEGY_COINS_COUNT"
        log "  策略最大持仓: $STRATEGY_MAX_POSITIONS"
        log "  BTC杠杆限制: ${STRATEGY_BTC_LEVERAGE}x"
        log "  BTC止损: ${STRATEGY_BTC_STOP_LOSS}%"
        
        check_pass "AI策略配置提取完成"
    else
        check_fail "策略文件JSON格式错误"
        STRATEGY_NOFXOS_API_KEY=""
        STRATEGY_ALLOWED_COINS=""
        STRATEGY_COINS_COUNT=0
        STRATEGY_MAX_POSITIONS=""
        STRATEGY_BTC_LEVERAGE=""
        STRATEGY_ALTCOIN_LEVERAGE=""
        STRATEGY_BTC_STOP_LOSS=""
        STRATEGY_MAIN_COINS_STOP_LOSS=""
        STRATEGY_ALTCOINS_STOP_LOSS=""
    fi
else
    check_fail "策略配置文件不存在或未找到包含 risk_control 的 JSON 文件"
    log "搜索路径: $PROMPT_DIR"
    STRATEGY_NOFXOS_API_KEY=""
    STRATEGY_ALLOWED_COINS=""
    STRATEGY_COINS_COUNT=0
    STRATEGY_MAX_POSITIONS=""
    STRATEGY_BTC_LEVERAGE=""
    STRATEGY_ALTCOIN_LEVERAGE=""
    STRATEGY_BTC_STOP_LOSS=""
    STRATEGY_MAIN_COINS_STOP_LOSS=""
    STRATEGY_ALTCOINS_STOP_LOSS=""
fi

# 1.3 从数据库提取实际交易数据
log "从数据库提取实际交易数据..."
if command -v docker &> /dev/null && docker ps --format '{{.Names}}' | grep -q "postgres_db"; then
    # 检查交易者权益数据
    DB_TRADER_DATA=$(docker exec postgres_db psql -U postgres -d nofx -t -c "
    SELECT 
        COUNT(*) as record_count,
        TO_CHAR(MAX(timestamp) AT TIME ZONE 'Asia/Bangkok', 'YYYY-MM-DD HH24:MI:SS') as latest_record,
        MIN(total_equity) as min_equity,
        MAX(total_equity) as max_equity
    FROM trader_equity_snapshots 
    WHERE trader_id = '$TRADER_ID'" 2>/dev/null | tr -d '|' | sed 's/^ *//;s/ *$//')
    
    if [ -n "$DB_TRADER_DATA" ] && [ "$DB_TRADER_DATA" != "" ]; then
        DB_RECORD_COUNT=$(echo "$DB_TRADER_DATA" | awk '{print $1}')
        DB_LATEST_RECORD=$(echo "$DB_TRADER_DATA" | awk '{$1=""; print $0}' | sed 's/^ //' | cut -d' ' -f1-2)
        DB_MIN_EQUITY=$(echo "$DB_TRADER_DATA" | awk '{print $(NF-1)}')
        DB_MAX_EQUITY=$(echo "$DB_TRADER_DATA" | awk '{print $NF}')
        
        log "  数据库记录数: $DB_RECORD_COUNT"
        log "  最新记录时间: $DB_LATEST_RECORD"
        log "  最小权益: $DB_MIN_EQUITY"
        log "  最大权益: $DB_MAX_EQUITY"
        
        # 获取实际交易过的币种
        DB_TRADED_SYMBOLS=$(docker exec postgres_db psql -U postgres -d nofx -t -c "
        SELECT DISTINCT symbol FROM (
            SELECT symbol FROM trades WHERE trader_id = '$TRADER_ID'
            UNION
            SELECT SUBSTRING(symbol FROM '^[A-Z0-9]+') as symbol FROM positions WHERE trader_id = '$TRADER_ID'
        ) t ORDER BY symbol" 2>/dev/null | tr '\n' ',' | sed 's/,$//' | sed 's/^,//' | sed 's/,,/,/g')
        
        if [ -n "$DB_TRADED_SYMBOLS" ] && [ "$DB_TRADED_SYMBOLS" != "" ]; then
            DB_TRADED_COUNT=$(echo "$DB_TRADED_SYMBOLS" | tr ',' '\n' | wc -l | tr -d ' ')
            log "  实际交易过 $DB_TRADED_COUNT 个币种: ${DB_TRADED_SYMBOLS:0:50}..."
        else
            DB_TRADED_SYMBOLS="N/A"
            DB_TRADED_COUNT=0
            log "  未找到实际交易记录"
        fi
        
        # 获取当前持仓
        DB_CURRENT_POSITIONS=$(docker exec postgres_db psql -U postgres -d nofx -t -c "
        SELECT 
            symbol,
            side,
            entry_price,
            mark_price,
            quantity,
            leverage,
            unrealized_pnl
        FROM positions 
        WHERE trader_id = '$TRADER_ID' AND quantity > 0" 2>/dev/null | head -10)
        
        if [ -n "$DB_CURRENT_POSITIONS" ] && [ "$DB_CURRENT_POSITIONS" != "" ]; then
            DB_POSITION_COUNT=$(echo "$DB_CURRENT_POSITIONS" | wc -l)
            log "  当前有 $DB_POSITION_COUNT 个持仓"
        else
            DB_CURRENT_POSITIONS="N/A"
            DB_POSITION_COUNT=0
            log "  当前无持仓"
        fi
        
        check_pass "数据库配置提取完成"
    else
        check_warning "数据库中未找到交易者记录 (TRADER_ID: $TRADER_ID)"
        DB_RECORD_COUNT=0
        DB_LATEST_RECORD="N/A"
        DB_MIN_EQUITY="N/A"
        DB_MAX_EQUITY="N/A"
        DB_TRADED_SYMBOLS="N/A"
        DB_TRADED_COUNT=0
        DB_CURRENT_POSITIONS="N/A"
        DB_POSITION_COUNT=0
    fi
else
    check_fail "数据库不可访问"
    DB_RECORD_COUNT=0
    DB_LATEST_RECORD="N/A"
    DB_MIN_EQUITY="N/A"
    DB_MAX_EQUITY="N/A"
    DB_TRADED_SYMBOLS="N/A"
    DB_TRADED_COUNT=0
    DB_CURRENT_POSITIONS="N/A"
    DB_POSITION_COUNT=0
fi

# ==================== 第二步：配置一致性检查 ====================
log ""
log "${BLUE}=== 2. 配置一致性检查 ===${NC}"
report ""
report "## 2. 配置一致性检查"

# 2.1 API密钥一致性检查
log "检查API密钥一致性..."
if [ -n "$TRADER_API_KEY" ] && [ -n "$STRATEGY_NOFXOS_API_KEY" ]; then
    if [ "$TRADER_API_KEY" = "$STRATEGY_NOFXOS_API_KEY" ]; then
        check_pass "API密钥一致: 交易者API_KEY与策略NOFXOS_API_KEY相同"
    else
        check_warning "API密钥不同: 交易者API_KEY与策略NOFXOS_API_KEY不同"
        log "  交易者API_KEY: ${TRADER_API_KEY:0:15}..."
        log "  策略NOFXOS_API_KEY: ${STRATEGY_NOFXOS_API_KEY:0:15}..."
        log "  说明: 这可能是正常的，因为NOFX API和NOFX OS API可能使用不同的密钥"
    fi
else
    check_fail "无法检查API密钥一致性: 缺少一个或两个API密钥"
fi

# 2.2 交易者ID存在性检查
log "检查交易者ID存在性..."
if [ -n "$TRADER_ID" ]; then
    if [ "$DB_RECORD_COUNT" -gt 0 ]; then
        check_pass "交易者ID在数据库中存在记录 ($DB_RECORD_COUNT 条记录)"
    else
        check_fail "交易者ID在数据库中无记录: $TRADER_ID"
    fi
else
    check_fail "交易者ID为空，无法检查存在性"
fi

# 2.3 允许交易币种一致性检查
log "检查允许交易币种一致性..."
if [ -n "$STRATEGY_ALLOWED_COINS" ] && [ "$STRATEGY_COINS_COUNT" -gt 0 ]; then
    # 将策略允许币种转换为数组
    IFS=',' read -ra STRATEGY_COINS_ARRAY <<< "$STRATEGY_ALLOWED_COINS"
    
    if [ "$DB_TRADED_COUNT" -gt 0 ] && [ "$DB_TRADED_SYMBOLS" != "N/A" ]; then
        # 检查每个实际交易过的币种是否在允许列表中
        IFS=',' read -ra TRADED_COINS_ARRAY <<< "$DB_TRADED_SYMBOLS"
        VIOLATIONS=0
        
        for traded_coin in "${TRADED_COINS_ARRAY[@]}"; do
            # 添加USDT后缀（如果还没有）
            if [[ ! "$traded_coin" =~ USDT$ ]]; then
                traded_coin="${traded_coin}USDT"
            fi
            
            FOUND=0
            for strategy_coin in "${STRATEGY_COINS_ARRAY[@]}"; do
                if [ "$traded_coin" = "$strategy_coin" ]; then
                    FOUND=1
                    break
                fi
            done
            
            if [ "$FOUND" -eq 0 ]; then
                ((VIOLATIONS++))
                log "  ❌ 违规交易: $traded_coin 不在允许列表中"
            fi
        done
        
        if [ "$VIOLATIONS" -eq 0 ]; then
            check_pass "所有实际交易币种都在策略允许列表中 ($DB_TRADED_COUNT/$STRATEGY_COINS_COUNT)"
        else
            check_fail "发现 $VIOLATIONS 个违规交易: 有币种不在策略允许列表中"
        fi
    else
        check_warning "无实际交易记录，无法验证币种一致性"
    fi
else
    check_fail "策略允许币种列表为空，无法检查一致性"
fi

# 2.4 持仓数量与策略限制一致性检查
log "检查持仓数量与策略限制一致性..."
if [ -n "$STRATEGY_MAX_POSITIONS" ] && [ "$STRATEGY_MAX_POSITIONS" != "" ]; then
    if [ "$DB_POSITION_COUNT" -le "$STRATEGY_MAX_POSITIONS" ]; then
        check_pass "当前持仓数量 ($DB_POSITION_COUNT) 符合策略限制 ($STRATEGY_MAX_POSITIONS)"
    else
        check_fail "当前持仓数量 ($DB_POSITION_COUNT) 超出策略限制 ($STRATEGY_MAX_POSITIONS)"
    fi
else
    check_warning "策略未设置最大持仓限制，无法检查"
fi

# 2.5 当前持仓杠杆一致性检查（如果有持仓）
log "检查当前持仓杠杆一致性..."
if [ "$DB_POSITION_COUNT" -gt 0 ] && [ "$DB_CURRENT_POSITIONS" != "N/A" ]; then
    if [ -n "$STRATEGY_BTC_LEVERAGE" ] && [ -n "$STRATEGY_ALTCOIN_LEVERAGE" ]; then
        # 解析持仓数据检查杠杆
        LEVERAGE_VIOLATIONS=0
        echo "$DB_CURRENT_POSITIONS" | while read -r position_line; do
            if [ -n "$position_line" ]; then
                # 解析持仓行：symbol | side | entry_price | mark_price | quantity | leverage | unrealized_pnl
                SYMBOL=$(echo "$position_line" | awk -F '|' '{print $1}' | tr -d ' ')
                LEVERAGE=$(echo "$position_line" | awk -F '|' '{print $6}' | tr -d ' ')
                
                if [[ "$SYMBOL" == *"BTC"* ]]; then
                    MAX_LEVERAGE="$STRATEGY_BTC_LEVERAGE"
                    COIN_TYPE="BTC/ETH"
                else
                    MAX_LEVERAGE="$STRATEGY_ALTCOIN_LEVERAGE"
                    COIN_TYPE="其他币种"
                fi
                
                if [ -n "$LEVERAGE" ] && [ "$LEVERAGE" != "" ]; then
                    if [ "$LEVERAGE" -le "$MAX_LEVERAGE" ]; then
                        log "  ✅ $SYMBOL 杠杆: ${LEVERAGE}x (限制: ${MAX_LEVERAGE}x, $COIN_TYPE)"
                    else
                        ((LEVERAGE_VIOLATIONS++))
                        log "  ❌ $SYMBOL 杠杆违规: ${LEVERAGE}x > 限制 ${MAX_LEVERAGE}x ($COIN_TYPE)"
                    fi
                fi
            fi
        done
        
        if [ "$LEVERAGE_VIOLATIONS" -eq 0 ]; then
            check_pass "所有持仓杠杆符合策略限制"
        else
            check_fail "发现 $LEVERAGE_VIOLATIONS 个持仓杠杆超出策略限制"
        fi
    else
        check_warning "策略杠杆限制未设置，无法检查持仓杠杆"
    fi
else
    check_pass "无持仓，杠杆检查跳过"
fi

# 2.6 初始余额一致性检查（如果数据可用）
log "检查初始余额一致性..."
if [ "$TRADER_INITIAL_BALANCE" != "N/A" ] && [ "$DB_MIN_EQUITY" != "N/A" ]; then
    # 将余额转换为数值进行比较
    INITIAL_BALANCE_NUM=$(echo "$TRADER_INITIAL_BALANCE" | grep -o '[0-9.]*' | head -1)
    MIN_EQUITY_NUM=$(echo "$DB_MIN_EQUITY" | grep -o '[0-9.]*' | head -1)
    
    if [ -n "$INITIAL_BALANCE_NUM" ] && [ -n "$MIN_EQUITY_NUM" ]; then
        # 允许10%的差异（因为可能有手续费等）
        DIFF_PERCENT=$(echo "scale=2; ($MIN_EQUITY_NUM - $INITIAL_BALANCE_NUM) / $INITIAL_BALANCE_NUM * 100" | bc 2>/dev/null || echo "0")
        
        if [ -n "$DIFF_PERCENT" ] && (($(echo "${DIFF_PERCENT#-} < 10" | bc -l 2>/dev/null || echo "1"))); then
            check_pass "初始余额一致性检查通过 (配置: $INITIAL_BALANCE_NUM, 数据库最小: $MIN_EQUITY_NUM, 差异: ${DIFF_PERCENT}%)"
        else
            check_warning "初始余额差异较大 (配置: $INITIAL_BALANCE_NUM, 数据库最小: $MIN_EQUITY_NUM, 差异: ${DIFF_PERCENT}%)"
        fi
    else
        check_warning "无法解析余额数值进行一致性检查"
    fi
else
    check_warning "初始余额或数据库权益数据缺失，无法检查一致性"
fi

# ==================== 第三步：生成一致性报告 ====================
log ""
log "${BLUE}=== 3. 一致性检查总结 ===${NC}"

if [ "$TOTAL_CHECKS" -gt 0 ]; then
    CONSISTENCY_SCORE=$((PASSED_CHECKS * 100 / TOTAL_CHECKS))
else
    CONSISTENCY_SCORE=0
fi

log ""
log "${BLUE}一致性检查结果:${NC}"
log "  总检查项: $TOTAL_CHECKS"
log "  一致通过: $PASSED_CHECKS"
log "  不一致失败: $FAILED_CHECKS"
log "  警告/差异: $WARNING_CHECKS"
log "  一致性得分: $CONSISTENCY_SCORE%"

# 添加到报告
echo "" >> "$REPORT_FILE"
echo "## 3. 一致性检查总结" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"
echo "| 检查类别 | 结果 |" >> "$REPORT_FILE"
echo "|----------|------|" >> "$REPORT_FILE"
echo "| 总检查项 | $TOTAL_CHECKS |" >> "$REPORT_FILE"
echo "| 一致通过 | $PASSED_CHECKS |" >> "$REPORT_FILE"
echo "| 不一致失败 | $FAILED_CHECKS |" >> "$REPORT_FILE"
echo "| 警告/差异 | $WARNING_CHECKS |" >> "$REPORT_FILE"
echo "| 一致性得分 | $CONSISTENCY_SCORE% |" >> "$REPORT_FILE"

echo "" >> "$REPORT_FILE"
echo "## 4. 配置摘要" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"
echo "### 交易者配置" >> "$REPORT_FILE"
echo "- **交易者ID**: $TRADER_ID" >> "$REPORT_FILE"
echo "- **API基础地址**: $TRADER_API_BASE" >> "$REPORT_FILE"
echo "- **初始余额**: $TRADER_INITIAL_BALANCE" >> "$REPORT_FILE"
echo "- **允许币种**: ${TRADER_ALLOWED_SYMBOLS:0:100}..." >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

echo "### AI策略配置" >> "$REPORT_FILE"
echo "- **策略文件**: $(basename "$STRATEGY_FILE")" >> "$REPORT_FILE"
echo "- **允许币种数量**: $STRATEGY_COINS_COUNT" >> "$REPORT_FILE"
echo "- **最大持仓数**: $STRATEGY_MAX_POSITIONS" >> "$REPORT_FILE"
echo "- **BTC/ETH杠杆限制**: ${STRATEGY_BTC_LEVERAGE}x" >> "$REPORT_FILE"
echo "- **其他币杠杆限制**: ${STRATEGY_ALTCOIN_LEVERAGE}x" >> "$REPORT_FILE"
echo "- **BTC止损**: ${STRATEGY_BTC_STOP_LOSS}%" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

echo "### 数据库配置" >> "$REPORT_FILE"
echo "- **记录数量**: $DB_RECORD_COUNT" >> "$REPORT_FILE"
echo "- **最新记录**: $DB_LATEST_RECORD" >> "$REPORT_FILE"
echo "- **最小权益**: $DB_MIN_EQUITY" >> "$REPORT_FILE"
echo "- **最大权益**: $DB_MAX_EQUITY" >> "$REPORT_FILE"
echo "- **实际交易币种数**: $DB_TRADED_COUNT" >> "$REPORT_FILE"
echo "- **当前持仓数**: $DB_POSITION_COUNT" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

echo "## 5. 不一致性分析" >> "$REPORT_FILE"
if [ "$FAILED_CHECKS" -eq 0 ] && [ "$WARNING_CHECKS" -eq 0 ]; then
    echo "✅ **所有配置完全一致** - 交易者配置、AI策略配置、数据库配置之间无差异。" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    echo "**系统状态**: 配置一致性优秀，系统运行在预期配置范围内。" >> "$REPORT_FILE"
elif [ "$FAILED_CHECKS" -eq 0 ]; then
    echo "⚠️ **配置基本一致，但有 $WARNING_CHECKS 个差异点**" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    echo "**差异点分析**: " >> "$REPORT_FILE"
    grep "⚠️ WARNING" "$LOG_FILE" | sed 's/.*⚠️ WARNING: //' | while read -r warn_item; do
        echo "1. $warn_item" >> "$REPORT_FILE"
    done
    echo "" >> "$REPORT_FILE"
    echo "**建议**: 这些差异可能需要关注，但不一定需要立即修复。" >> "$REPORT_FILE"
else
    echo "❌ **发现 $FAILED_CHECKS 个不一致性问题**" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    echo "**不一致性问题**: " >> "$REPORT_FILE"
    grep "❌ FAIL" "$LOG_FILE" | sed 's/.*❌ FAIL: //' | while read -r fail_item; do
        echo "1. $fail_item" >> "$REPORT_FILE"
    done
    echo "" >> "$REPORT_FILE"
    echo "**修复优先级**: " >> "$REPORT_FILE"
    echo "1. 交易币种违规问题 (高风险)" >> "$REPORT_FILE"
    echo "2. 持仓数量/杠杆超限问题 (中风险)" >> "$REPORT_FILE"
    echo "3. 交易者ID/API密钥问题 (基础问题)" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    echo "**建议**: 立即修复不一致性问题，避免系统运行在非预期状态。" >> "$REPORT_FILE"
fi

if [ "$WARNING_CHECKS" -gt 0 ]; then
    echo "" >> "$REPORT_FILE"
    echo "## 6. 警告项目详情" >> "$REPORT_FILE"
    grep "⚠️ WARNING" "$LOG_FILE" | sed 's/.*⚠️ WARNING: //' | while read -r warn_item; do
        echo "- $warn_item" >> "$REPORT_FILE"
    done
fi

echo "" >> "$REPORT_FILE"
echo "## 7. 检查详情" >> "$REPORT_FILE"
echo "完整检查日志请查看: $LOG_FILE" >> "$REPORT_FILE"
echo "报告生成时间: $(date '+%Y-%m-%d %H:%M:%S %Z')" >> "$REPORT_FILE"

# 输出报告位置
log ""
log "${GREEN}一致性检查报告已生成: $REPORT_FILE${NC}"
log "${GREEN}详细日志已保存: $LOG_FILE${NC}"
log ""
log "=== NOFX配置一致性检查完成 ==="

# 显示报告摘要
echo ""
echo "📋 NOFX配置一致性检查报告摘要:"
echo "   报告文件: $REPORT_FILE"
echo "   日志文件: $LOG_FILE"
echo "   检查结果: $PASSED_CHECKS/$TOTAL_CHECKS 项一致 ($CONSISTENCY_SCORE%)"
echo "   检查范围: 交易者配置 ↔ AI策略配置 ↔ 数据库配置"
echo ""

# 发送Telegram通知
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

if [ -n "$TELEGRAM_BOT_TOKEN" ] && [ -n "$TELEGRAM_CHAT_ID" ]; then
    log "正在发送Telegram报告..."
    
    # 读取报告内容，截取前4000个字符（Telegram消息限制）
    # 使用Python安全转义JSON特殊字符
    if [ -f "$REPORT_FILE" ]; then
        REPORT_CONTENT=$(cat "$REPORT_FILE")
        # 使用Python处理：读取文件内容 -> 截取前4000字符 -> JSON转义 -> 去掉首尾双引号
        SAFE_MESSAGE=$(python3 -c "import json, sys; print(json.dumps(sys.stdin.read().strip()[:4000])[1:-1])" <<< "$REPORT_CONTENT")
        
        # 构建JSON Payload
        JSON_PAYLOAD=$(cat <<EOF
{
  "chat_id": "$TELEGRAM_CHAT_ID",
  "text": "$SAFE_MESSAGE"
}
EOF
)
        
        # 发送请求
        RESPONSE=$(curl -s -X POST "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/sendMessage" \
            -H "Content-Type: application/json" \
            -d "$JSON_PAYLOAD")
            
        if echo "$RESPONSE" | grep -q '"ok":true'; then
            log "${GREEN}✅ Telegram报告发送成功${NC}"
            # 只有发送成功后才删除临时文件
            rm "$REPORT_FILE" "$LOG_FILE" 2>/dev/null
        else
            log "${RED}❌ Telegram报告发送失败: $RESPONSE${NC}"
            
            # 重试不带parse_mode（纯文本模式）
            log "尝试以纯文本重发..."
            JSON_PAYLOAD_PLAIN=$(cat <<EOF
{
  "chat_id": "$TELEGRAM_CHAT_ID",
  "text": "$SAFE_MESSAGE"
}
EOF
)
            curl -s -X POST "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/sendMessage" \
                -H "Content-Type: application/json" \
                -d "$JSON_PAYLOAD_PLAIN" > /dev/null
            
            # 即使重试也尝试清理，或者选择保留以便调试
            rm "$REPORT_FILE" "$LOG_FILE" 2>/dev/null
        fi
    else
        log "${RED}❌ 报告文件不存在: $REPORT_FILE${NC}"
    fi
else
    log "${YELLOW}⚠️ Telegram配置缺失，跳过发送报告${NC}"
fi

# 如果不一致性问题严重，以非零退出码退出
if [ "$FAILED_CHECKS" -gt 0 ]; then
    exit 2  # 2表示有不一致性问题
elif [ "$WARNING_CHECKS" -gt 3 ]; then
    exit 1  # 1表示有较多警告
else
    exit 0  # 0表示基本一致
fi