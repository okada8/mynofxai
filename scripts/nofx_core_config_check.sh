#!/bin/bash
# NOFX核心配置检查脚本
# 精简版本：只检查交易者配置、AI提示词/策略配置、数据库配置

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_FILE="/tmp/nofx_core_config_check_$(date +%Y%m%d_%H%M%S).log"
REPORT_FILE="/tmp/nofx_core_config_report_$(date +%Y%m%d_%H%M%S).md"

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
echo "# NOFX核心配置检查报告" > "$REPORT_FILE"
echo "**生成时间**: $(date '+%Y-%m-%d %H:%M:%S %Z')" >> "$REPORT_FILE"
echo "**检查范围**: 交易者配置、AI提示词/策略配置、数据库配置" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

log "${BLUE}=== NOFX核心配置检查开始 ===${NC}"
log "检查时间: $(date '+%Y-%m-%d %H:%M:%S %Z')"
log "日志文件: $LOG_FILE"
log "报告文件: $REPORT_FILE"
log "检查范围: 交易者配置、AI提示词/策略配置、数据库配置"
log ""

# ==================== 第一部分：交易者配置检查 ====================
log "${BLUE}=== 1. 交易者配置检查 ===${NC}"
report "## 1. 交易者配置检查"

# 检查hourly_report.sh中的交易者配置
HOURLY_REPORT="$SCRIPT_DIR/hourly_report.sh"
if [ -f "$HOURLY_REPORT" ]; then
    check_pass "交易者配置脚本存在: $HOURLY_REPORT"
    
    # 提取交易者配置
    API_KEY=$(grep "API_KEY=" "$HOURLY_REPORT" | head -1 | cut -d'"' -f2 2>/dev/null || grep "API_KEY=" "$HOURLY_REPORT" | head -1 | cut -d'=' -f2 | tr -d "'\"")
    API_BASE=$(grep "API_BASE=" "$HOURLY_REPORT" | head -1 | cut -d'"' -f2 2>/dev/null || grep "API_BASE=" "$HOURLY_REPORT" | head -1 | cut -d'=' -f2 | tr -d "'\"")
    TRADER_ID=$(grep "TRADER_ID=" "$HOURLY_REPORT" | head -1 | cut -d'"' -f2 2>/dev/null || grep "TRADER_ID=" "$HOURLY_REPORT" | head -1 | cut -d'=' -f2 | tr -d "'\"")
    
    if [ -n "$API_KEY" ]; then
        check_pass "API_KEY配置存在: ${API_KEY:0:10}..."
    else
        check_fail "API_KEY配置缺失"
    fi
    
    if [ -n "$API_BASE" ]; then
        check_pass "API_BASE配置存在: $API_BASE"
    else
        check_fail "API_BASE配置缺失"
    fi
    
    if [ -n "$TRADER_ID" ]; then
        check_pass "TRADER_ID配置存在: $TRADER_ID"
    else
        check_fail "TRADER_ID配置缺失"
    fi
    
    # 检查API连通性
    log "检查API连通性..."
    if curl -s -H "X-API-KEY: $API_KEY" "${API_BASE}/health" &> /dev/null; then
        check_pass "API健康检查端点响应正常"
    elif curl -s -H "X-API-KEY: $API_KEY" "${API_BASE}/positions" &> /dev/null; then
        check_pass "API持仓端点响应正常"
    else
        check_fail "API不可访问 (端点: $API_BASE)"
    fi
    
    # 检查交易者配置访问
    log "检查交易者配置访问..."
    TRADER_CONFIG=$(curl -s -H "X-API-KEY: $API_KEY" "${API_BASE}/traders/${TRADER_ID}/config" 2>/dev/null)
    if [ -n "$TRADER_CONFIG" ] && [ "$TRADER_CONFIG" != "null" ] && [ "$TRADER_CONFIG" != "Not Found" ]; then
        check_pass "交易者配置可访问 (TRADER_ID: $TRADER_ID)"
        
        # 解析交易者配置
        TRADER_INITIAL_BALANCE=$(echo "$TRADER_CONFIG" | python3 -c "import json, sys; data=json.load(sys.stdin); print(data.get('initial_balance', 'N/A'))" 2>/dev/null || echo "N/A")
        log "  交易者初始余额: $TRADER_INITIAL_BALANCE"
    else
        check_fail "交易者配置不可访问或TRADER_ID无效"
    fi
    
else
    check_fail "交易者配置脚本不存在: $HOURLY_REPORT"
fi

# ==================== 第二部分：AI提示词/策略配置检查 ====================
log ""
log "${BLUE}=== 2. AI提示词/策略配置检查 ===${NC}"
report ""
report "## 2. AI提示词/策略配置检查"

# 检查优化策略文件
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
    check_pass "AI策略配置文件存在: $STRATEGY_FILE"
    
    # 检查JSON格式
    if python3 -m json.tool "$STRATEGY_FILE" &> /dev/null; then
        check_pass "策略文件JSON格式正确"
        
        # 检查关键配置字段
        log "分析策略配置结构..."
        
        # 1. 检查nofxos_api_key
        if grep -q '"nofxos_api_key"' "$STRATEGY_FILE"; then
            NOFXOS_API_KEY=$(grep -o '"nofxos_api_key": *"[^"]*"' "$STRATEGY_FILE" | cut -d'"' -f4)
            check_pass "nofxos_api_key配置存在: ${NOFXOS_API_KEY:0:10}..."
        else
            check_warning "nofxos_api_key配置缺失"
        fi
        
        # 2. 检查允许交易币种
        if grep -q '"static_coins"' "$STRATEGY_FILE"; then
            STATIC_COINS_COUNT=$(grep -A 20 '"static_coins"' "$STRATEGY_FILE" | grep -o '"[A-Z0-9]*USDT"' | wc -l)
            check_pass "允许交易币种配置存在 ($STATIC_COINS_COUNT 个币种)"
            
            # 显示部分币种
            SAMPLE_COINS=$(grep -A 20 '"static_coins"' "$STRATEGY_FILE" | grep -o '"[A-Z0-9]*USDT"' | head -5 | tr '\n' ' ' | sed 's/"//g')
            log "  示例币种: $SAMPLE_COINS..."
        else
            check_fail "允许交易币种配置(static_coins)缺失"
        fi
        
        # 3. 检查AI提示词部分
        if grep -q '"prompt_sections"' "$STRATEGY_FILE"; then
            check_pass "AI提示词配置(prompt_sections)存在"
            
            # 检查关键提示词部分
            PROMPT_SECTIONS=("role_definition" "trading_frequency" "entry_standards" "decision_process")
            for section in "${PROMPT_SECTIONS[@]}"; do
                if grep -q "\"$section\"" "$STRATEGY_FILE"; then
                    check_pass "提示词部分 '$section' 存在"
                else
                    check_warning "提示词部分 '$section' 缺失"
                fi
            done
            
            # 提取决策流程部分
            DECISION_PROCESS=$(grep -A 10 '"decision_process"' "$STRATEGY_FILE" | head -5 | grep -v '"decision_process"' | tr -d '"' | sed 's/^ *//')
            if [ -n "$DECISION_PROCESS" ]; then
                log "  决策流程示例: ${DECISION_PROCESS:0:80}..."
            fi
        else
            check_fail "AI提示词配置(prompt_sections)缺失"
        fi
        
        # 4. 检查风险控制配置
        if grep -q '"risk_control"' "$STRATEGY_FILE"; then
            check_pass "风险控制配置存在"
            
            # 检查关键风险参数
            RISK_PARAMS=("max_positions" "btc_eth_max_leverage" "stop_loss_percent" "min_risk_reward_ratio")
            for param in "${RISK_PARAMS[@]}"; do
                if grep -q "\"$param\"" "$STRATEGY_FILE"; then
                    PARAM_VALUE=$(grep -o "\"$param\": *[^,}]*" "$STRATEGY_FILE" | head -1)
                    log "  $param: $PARAM_VALUE"
                else
                    check_warning "风险参数 '$param' 缺失"
                fi
            done
        else
            check_warning "风险控制配置缺失"
        fi
        
        # 5. 检查技术指标配置
        if grep -q '"indicators"' "$STRATEGY_FILE"; then
            check_pass "技术指标配置存在"
            
            # 检查启用的指标
            ENABLED_INDICATORS=$(grep '"enable_[a-z_]*": true' "$STRATEGY_FILE" | grep -o 'enable_[a-z_]*' | sed 's/enable_//g' | tr '\n' ' ' | sed 's/ $//')
            if [ -n "$ENABLED_INDICATORS" ]; then
                check_pass "启用的技术指标: $ENABLED_INDICATORS"
            else
                check_warning "未找到启用的技术指标"
            fi
            
            # 检查K线配置
            if grep -q '"primary_timeframe"' "$STRATEGY_FILE"; then
                TIMEFRAME=$(grep -o '"primary_timeframe": *"[^"]*"' "$STRATEGY_FILE" | cut -d'"' -f4)
                COUNT=$(grep -o '"primary_count": *[0-9]*' "$STRATEGY_FILE" | grep -o '[0-9]*')
                check_pass "K线配置: $TIMEFRAME ($COUNT条)"
            fi
        else
            check_fail "技术指标配置缺失"
        fi
        
    else
        check_fail "策略文件JSON格式错误"
    fi
else
    # 检查原始策略文件作为备用
    check_fail "策略配置文件不存在或未找到包含 risk_control 的 JSON 文件"
    log "搜索路径: $PROMPT_DIR"
    
    ORIGINAL_STRATEGY="$SCRIPT_DIR/optimized_strategy.json"
    if [ -f "$ORIGINAL_STRATEGY" ]; then
        check_warning "修复版策略文件不存在，使用原始策略文件: $ORIGINAL_STRATEGY"
        STRATEGY_FILE="$ORIGINAL_STRATEGY"
        
        if python3 -m json.tool "$STRATEGY_FILE" &> /dev/null; then
            check_pass "原始策略文件JSON格式正确"
        else
            check_fail "原始策略文件JSON格式错误"
        fi
    else
        check_fail "未找到任何策略配置文件"
    fi
fi

# ==================== 第三部分：数据库配置检查 ====================
log ""
log "${BLUE}=== 3. 数据库配置检查 ===${NC}"
report ""
report "## 3. 数据库配置检查"

# 检查Docker是否安装
if command -v docker &> /dev/null; then
    check_pass "Docker已安装"
    
    # 检查PostgreSQL容器状态
    if docker ps --format '{{.Names}}' | grep -q "postgres_db"; then
        check_pass "PostgreSQL容器 (postgres_db) 正在运行"
        
        # 测试数据库连接
        log "测试数据库连接..."
        # Remove -t option to allow headers, which makes parsing slightly different but output more reliable
        # Or better: check if output contains "1" or is not empty and exit code is 0
        if docker exec postgres_db psql -U postgres -d nofx -c "SELECT 1" &>/dev/null; then
             check_pass "PostgreSQL数据库连接成功 (数据库: nofx, 用户: postgres)"
            
            # 检查关键表结构
            log "检查数据库表结构..."
            
            # 1. trader_equity_snapshots表
            EQUITY_TABLE_CHECK=$(docker exec postgres_db psql -U postgres -d nofx -t -c "
            SELECT COUNT(*) FROM information_schema.tables 
            WHERE table_name = 'trader_equity_snapshots'" 2>/dev/null | tr -d ' ')
            
            if [ "$EQUITY_TABLE_CHECK" = "1" ]; then
                check_pass "关键表 'trader_equity_snapshots' 存在"
                
                # 检查表结构
                EQUITY_COLUMNS=$(docker exec postgres_db psql -U postgres -d nofx -t -c "
                SELECT column_name FROM information_schema.columns 
                WHERE table_name = 'trader_equity_snapshots'
                ORDER BY ordinal_position" 2>/dev/null | tr '\n' ' ' | sed 's/ $//')
                
                REQUIRED_EQUITY_COLUMNS=("trader_id" "timestamp" "total_equity" "balance" "unrealized_pnl")
                for col in "${REQUIRED_EQUITY_COLUMNS[@]}"; do
                    if echo "$EQUITY_COLUMNS" | grep -q "$col"; then
                        check_pass "  权益快照表包含 '$col' 列"
                    else
                        check_warning "  权益快照表缺少 '$col' 列"
                    fi
                done
            else
                check_fail "关键表 'trader_equity_snapshots' 不存在"
            fi
            
            # 2. 检查positions表或类似表
            POSITIONS_TABLE_CHECK=$(docker exec postgres_db psql -U postgres -d nofx -t -c "
            SELECT COUNT(*) FROM information_schema.tables 
            WHERE table_name IN ('positions', 'trades', 'orders')" 2>/dev/null | tr -d ' ')
            
            if [ "$POSITIONS_TABLE_CHECK" -ge 1 ]; then
                check_pass "持仓/交易相关表存在 ($POSITIONS_TABLE_CHECK 个表)"
            else
                check_warning "未找到持仓/交易相关表"
            fi
            
            # 3. 检查最近数据
            log "检查最近数据..."
            RECENT_DATA=$(docker exec postgres_db psql -U postgres -d nofx -t -c "
            SELECT 
                TO_CHAR(MAX(timestamp) AT TIME ZONE 'Asia/Bangkok', 'YYYY-MM-DD HH24:MI:SS') as latest_snapshot,
                COUNT(*) as total_records
            FROM trader_equity_snapshots 
            WHERE trader_id = '$TRADER_ID'" 2>/dev/null | sed 's/|/ - /g' | tr -d ' ')
            
            if [ -n "$RECENT_DATA" ] && [ "$RECENT_DATA" != "-" ]; then
                check_pass "交易者权益数据存在: $RECENT_DATA"
            else
                check_warning "未找到交易者权益数据 (TRADER_ID: $TRADER_ID)"
            fi
            
        else
            check_fail "PostgreSQL数据库连接失败"
        fi
    else
        check_fail "PostgreSQL容器 (postgres_db) 未运行"
        
        # 检查容器是否存在但未运行
        if docker ps -a --format '{{.Names}}' | grep -q "postgres_db"; then
            check_warning "PostgreSQL容器存在但未运行"
        else
            check_fail "PostgreSQL容器不存在"
        fi
    fi
else
    check_fail "Docker未安装 - 数据库检查需要Docker"
fi

# ==================== 生成总结报告 ====================
log ""
log "${BLUE}=== 检查完成 ===${NC}"

if [ "$TOTAL_CHECKS" -gt 0 ]; then
    SUMMARY_PASS_RATE=$((PASSED_CHECKS * 100 / TOTAL_CHECKS))
else
    SUMMARY_PASS_RATE=0
fi

log ""
log "${BLUE}核心配置检查总结:${NC}"
log "  总检查项: $TOTAL_CHECKS"
log "  通过: $PASSED_CHECKS"
log "  失败: $FAILED_CHECKS"
log "  警告: $WARNING_CHECKS"
log "  通过率: $SUMMARY_PASS_RATE%"

# 添加到报告
echo "" >> "$REPORT_FILE"
echo "## 检查总结" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"
echo "| 检查类别 | 结果 |" >> "$REPORT_FILE"
echo "|----------|------|" >> "$REPORT_FILE"
echo "| 总检查项 | $TOTAL_CHECKS |" >> "$REPORT_FILE"
echo "| 通过 | $PASSED_CHECKS |" >> "$REPORT_FILE"
echo "| 失败 | $FAILED_CHECKS |" >> "$REPORT_FILE"
echo "| 警告 | $WARNING_CHECKS |" >> "$REPORT_FILE"
echo "| 通过率 | $SUMMARY_PASS_RATE% |" >> "$REPORT_FILE"

echo "" >> "$REPORT_FILE"
echo "## 核心配置状态" >> "$REPORT_FILE"

if [ "$FAILED_CHECKS" -eq 0 ]; then
    echo "✅ **所有核心配置检查通过** - NOFX系统核心配置完整。" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    echo "**配置摘要**: " >> "$REPORT_FILE"
    echo "1. **交易者配置**: API连接正常，交易者ID: $TRADER_ID" >> "$REPORT_FILE"
    echo "2. **AI策略配置**: 策略文件完整，包含$STATIC_COINS_COUNT个允许币种" >> "$REPORT_FILE"
    echo "3. **数据库配置**: PostgreSQL运行正常，表结构完整" >> "$REPORT_FILE"
else
    echo "⚠️ **发现 $FAILED_CHECKS 个失败项** - 需要修复以下核心配置问题:" >> "$REPORT_FILE"
    
    # 提取失败项
    grep "❌ FAIL" "$LOG_FILE" | sed 's/.*❌ FAIL: //' | while read -r fail_item; do
        echo "1. $fail_item" >> "$REPORT_FILE"
    done
    
    echo "" >> "$REPORT_FILE"
    echo "**修复优先级**: " >> "$REPORT_FILE"
    echo "1. 交易者配置问题 (API连接、TRADER_ID)" >> "$REPORT_FILE"
    echo "2. 数据库配置问题 (PostgreSQL连接)" >> "$REPORT_FILE"
    echo "3. AI策略配置问题 (策略文件完整性)" >> "$REPORT_FILE"
fi

if [ "$WARNING_CHECKS" -gt 0 ]; then
    echo "" >> "$REPORT_FILE"
    echo "## 警告项目 (建议修复)" >> "$REPORT_FILE"
    grep "⚠️ WARNING" "$LOG_FILE" | sed 's/.*⚠️ WARNING: //' | while read -r warn_item; do
        echo "1. $warn_item" >> "$REPORT_FILE"
    done
fi

echo "" >> "$REPORT_FILE"
echo "## 检查详情" >> "$REPORT_FILE"
echo "完整检查日志请查看: $LOG_FILE" >> "$REPORT_FILE"
echo "报告生成时间: $(date '+%Y-%m-%d %H:%M:%S %Z')" >> "$REPORT_FILE"

# 输出报告位置
log ""
log "${GREEN}核心配置检查报告已生成: $REPORT_FILE${NC}"
log "${GREEN}详细日志已保存: $LOG_FILE${NC}"
log ""
log "=== NOFX核心配置检查完成 ==="

# 显示报告摘要
echo ""
echo "📋 NOFX核心配置检查报告摘要:"
echo "   报告文件: $REPORT_FILE"
echo "   日志文件: $LOG_FILE"
echo "   检查结果: $PASSED_CHECKS/$TOTAL_CHECKS 项通过 ($SUMMARY_PASS_RATE%)"
echo "   检查范围: 交易者配置、AI提示词/策略配置、数据库配置"
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
        
        # 构建JSON Payload (纯文本模式，避免Bad Request)
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
            # 发送成功后清理文件
            rm "$REPORT_FILE" "$LOG_FILE" 2>/dev/null
        else
            log "${RED}❌ Telegram报告发送失败: $RESPONSE${NC}"
        fi
    else
        log "${RED}❌ 报告文件不存在: $REPORT_FILE${NC}"
    fi
else
    log "${YELLOW}⚠️ Telegram配置缺失，跳过发送报告${NC}"
fi

# 如果有失败项，以非零退出码退出
if [ "$FAILED_CHECKS" -gt 0 ]; then
    exit 1
else
    exit 0
fi