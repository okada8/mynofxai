# NOFX 策略文件配置指南

本文档列出了 NOFX 系统支持的策略 JSON 文件字段、数据类型及其含义。上传策略文件时，请确保文件格式符合以下规范。

## 1. 根对象结构 (Root)

| 字段名 | 类型 | 必填 | 含义 |
| :--- | :--- | :--- | :--- |
| `strategy_type` | String | 否 | 策略类型，默认为 `"ai_trading"` (AI 交易)，也可选 `"grid_trading"` (网格交易) |
| `language` | String | 否 | 语言设置，`"zh"` (中文) 或 `"en"` (英文)，影响 AI 提示词语言 |
| `coin_source` | Object | **是** | [币种来源配置](#2-coin_source-币种来源配置) |
| `indicators` | Object | **是** | [技术指标配置](#3-indicators-技术指标配置) |
| `risk_control` | Object | **是** | [风险控制配置](#4-risk_control-风险控制配置) |
| `risk_control_enhanced` | Object | 否 | [增强风控配置](#41-risk_control_enhanced-增强风控配置) (v2.0+) |
| `alpha_factors` | Object | 否 | [Alpha 因子配置](#9-alpha_factors-alpha-因子配置) (v2.0+) |
| `multi_agent` | Object | 否 | [多智能体配置](#10-multi_agent-多智能体配置) (v2.0+) |
| `macro_config` | Object | 否 | [宏观配置](#11-macro_config-宏观配置) (v3.0+) |
| `prompt_sections` | Object | 否 | [自定义 AI 提示词段落](#5-prompt_sections-ai-提示词段落) |
| `custom_prompt` | String | 否 | 追加的自定义提示词文本 |
| `volatility_config` | Object | 否 | [波动率配置](#7-volatility_config-波动率配置) |
| `trailing_stop_config` | Object | 否 | [移动止损配置](#8-trailing_stop_config-移动止损配置) |
| `enable_timeframe_prediction` | Boolean | 否 | 是否启用时间框架预测 (默认 true) |
| `grid_config` | Object | 否 | [网格交易配置](#6-grid_config-网格交易配置) (仅当 `strategy_type` 为 `"grid_trading"` 时有效) |

---

## 2. `coin_source` (币种来源配置)

控制系统如何选择交易标的。

| 字段名 | 类型 | 含义 |
| :--- | :--- | :--- |
| `source_type` | String | 来源类型：`"static"` (静态列表), `"ai500"` (AI500池), `"oi_top"` (持仓增加榜), `"oi_low"` (持仓减少榜), `"mixed"` (混合) |
| `static_coins` | Array&lt;String&gt; | 静态币种列表 (例如 `["BTCUSDT", "ETHUSDT"]`)，仅当 `source_type="static"` 时生效 |
| `excluded_coins` | Array&lt;String&gt; | 排除币种列表 (黑名单) |
| `use_ai500` | Boolean | 是否使用 AI500 优选币池 |
| `ai500_limit` | Integer | AI500 池最大数量限制 |
| `use_oi_top` | Boolean | 是否使用持仓量增长榜 (适合做多) |
| `oi_top_limit` | Integer | 持仓增长榜最大数量 |
| `use_oi_low` | Boolean | 是否使用持仓量减少榜 (适合做空) |
| `oi_low_limit` | Integer | 持仓减少榜最大数量 |
| `use_hyper_all` | Boolean | 是否使用 Hyperliquid 所有可用币种 |
| `use_hyper_main` | Boolean | 是否使用 Hyperliquid 主要币种 (高交易量) |
| `hyper_main_limit` | Integer | Hyperliquid 主要币种数量限制 |
| `use_gainers_losers` | Boolean | 是否使用涨跌幅榜 |
| `gainers_top` | Integer | 涨幅榜取样数量 (如 5) |
| `losers_top` | Integer | 跌幅榜取样数量 (如 5) |
| `only_binance_symbols` | Boolean | 是否仅限币安交易对 |
| `macro_screening` | Object | [宏观筛选配置](#111-macro_screening-宏观筛选配置) |

---

## 3. `indicators` (技术指标配置)

控制系统采集哪些市场数据喂给 AI 模型。

| 字段名 | 类型 | 含义 |
| :--- | :--- | :--- |
| `klines` | Object | [K线配置](#31-klines-k线配置) |
| `enable_raw_klines` | Boolean | 是否启用原始 K 线数据 (OHLCV)，**建议开启** |
| `enable_ema` | Boolean | 是否计算并提供 EMA 指标 |
| `enable_macd` | Boolean | 是否计算并提供 MACD 指标 |
| `enable_rsi` | Boolean | 是否计算并提供 RSI 指标 |
| `enable_atr` | Boolean | 是否计算并提供 ATR 指标 |
| `enable_boll` | Boolean | 是否计算并提供布林带 (BOLL) 指标 |
| `enable_volume` | Boolean | 是否提供成交量数据 |
| `enable_oi` | Boolean | 是否提供持仓量 (Open Interest) 数据 |
| `enable_funding_rate` | Boolean | 是否提供资金费率数据 |
| `enable_adx` | Boolean | 是否启用 ADX 趋势强度指标 |
| `enable_vwap` | Boolean | 是否启用成交量加权平均价 (VWAP) |
| `enable_obv` | Boolean | 是否启用能量潮 (OBV) 指标 |
| `enable_cmf` | Boolean | 是否启用蔡金资金流量 (CMF) 指标 |
| `enable_liquidation` | Boolean | 是否启用爆仓数据 |
| `enable_cvd` | Boolean | 是否启用累计成交量偏差 (CVD) |
| `enable_net_position` | Boolean | 是否启用净持仓数据 |
| `enable_signal_labels` | Boolean | 是否显示信号标签 |
| `enable_indicator_series` | Boolean | 是否显示指标序列 |
| `external_data_sources` | Array&lt;Object&gt; | [外部数据源配置](#33-external_data_sources-外部数据源配置) |
| `oi_ranking_limit` | Integer | OI 排行榜条目数量限制 (默认 10) |
| `netflow_ranking_limit` | Integer | 净流入排行榜条目数量限制 (默认 10) |
| `price_ranking_limit` | Integer | 涨跌幅排行榜条目数量限制 (默认 10) |
| `ema_periods` | Array&lt;Integer&gt; | EMA 周期列表 (例如 `[20, 50]`) |
| `rsi_periods` | Array&lt;Integer&gt; | RSI 周期列表 (例如 `[7, 14]`) |
| `atr_periods` | Array&lt;Integer&gt; | ATR 周期列表 (例如 `[14]`) |
| `boll_periods` | Array&lt;Integer&gt; | 布林带周期列表 (例如 `[20]`) |
| `nofxos_api_key` | String | NofxOS 统一数据 API 密钥 |
| `enable_quant_data` | Boolean | 是否启用量化数据 (资金流向等) |
| `enable_quant_oi` | Boolean | 是否启用量化 OI 数据 |
| `enable_quant_netflow` | Boolean | 是否启用量化净流入数据 |
| `enable_oi_ranking` | Boolean | 是否启用全市场 OI 涨跌幅排行 |
| `oi_ranking_duration` | String | OI 排行时间周期 (`"1h"`, `"4h"`, `"24h"`) |
| `enable_netflow_ranking` | Boolean | 是否启用全市场资金流向排行 |
| `netflow_ranking_duration` | String | 资金流排行时间周期 (`"1h"`, `"4h"`, `"24h"`) |
| `enable_price_ranking` | Boolean | 是否启用全市场涨跌幅排行 |
| `price_ranking_duration` | String | 涨跌幅排行周期 (`"1h"`, `"1h,4h,24h"`) |
| `enable_donchian_box` | Boolean | 是否启用唐奇安通道箱体 |
| `enable_box_prefetch` | Boolean | 是否预取箱体数据 |
| `donchian_periods` | Array&lt;Integer&gt; | 唐奇安通道周期列表 (如 `[72, 240, 500]`) |
| `enable_heat_score` | Boolean | 是否启用热度评分 |
| `enable_volatility_util_score` | Boolean | 是否启用波动率利用评分 |
| `enable_volume_spike_score` | Boolean | 是否启用成交量突增评分 |
| `enable_funding_rate_score` | Boolean | 是否启用资金费率评分 |
| `enable_orderbook_imbalance_score` | Boolean | 是否启用订单簿不平衡评分 |
| `enable_adaptive_composite_weights` | Boolean | 是否启用自适应组合权重 |
| `weight_heat` | Float | 热度评分基础权重 (0.0-1.0) |
| `weight_volatility_util` | Float | 波动率利用评分基础权重 |
| `weight_volume_spike` | Float | 成交量突增评分基础权重 |
| `weight_funding_rate` | Float | 资金费率评分基础权重 |
| `weight_orderbook_imbalance` | Float | 订单簿不平衡评分基础权重 |
| `grouped_weights` | Object | [分组权重配置](#32-grouped_weights-分组权重配置) (Scalp/Intraday/Swing/Position) |

### 3.1 `klines` (K线配置)

| 字段名 | 类型 | 含义 |
| :--- | :--- | :--- |
| `primary_timeframe` | String | 主时间周期 (例如 `"15m"`, `"1h"`) |
| `primary_count` | Integer | 主周期 K 线数量 |
| `longer_timeframe` | String | 更长周期 (辅助趋势判断) |
| `longer_count` | Integer | 更长周期 K 线数量 |
| `enable_multi_timeframe` | Boolean | 是否启用多周期分析 |
| `selected_timeframes` | Array&lt;String&gt; | 启用的时间周期列表 (例如 `["5m", "15m", "1h", "4h"]`) |

### 3.2 `grouped_weights` (分组权重配置)

包含 `scalp` (头皮), `intraday` (日内), `swing` (波段), `position` (趋势) 四组独立配置。每组包含以下字段：

| 字段名 | 类型 | 含义 |
| :--- | :--- | :--- |
| `enable_heat` | Boolean | 是否启用热度评分 |
| `enable_volatility_util` | Boolean | 是否启用波动率评分 |
| `enable_volume_spike` | Boolean | 是否启用成交量突增评分 |
| `enable_funding_rate` | Boolean | 是否启用资金费率评分 |
| `enable_orderbook_imbalance` | Boolean | 是否启用订单簿不平衡评分 |
| `weight_heat` | Float | 热度权重 (0.0-1.0) |
| `weight_volatility_util` | Float | 波动率权重 |
| `weight_volume_spike` | Float | 成交量突增权重 |
| `weight_funding_rate` | Float | 资金费率权重 |
| `weight_orderbook_imbalance` | Float | 订单簿权重 |

### 3.3 `external_data_sources` (外部数据源配置)

| 字段名 | 类型 | 含义 |
| :--- | :--- | :--- |
| `name` | String | 数据源名称 |
| `type` | String | 数据源类型 |
| `url` | String | 数据源 URL |
| `method` | String | 请求方法 (如 "GET") |
| `headers` | Map&lt;String, String&gt; | 请求头 |
| `data_path` | String | 数据路径 (JSON Path) |
| `refresh_secs` | Integer | 刷新间隔 (秒) |

---

## 4. `risk_control` (风险控制配置)

系统硬性风控规则 + AI 辅助风控参数。

| 字段名 | 类型 | 含义 | 备注 |
| :--- | :--- | :--- | :--- |
| `max_positions` | Integer | 最大同时持仓数量 | **代码强制执行** |
| `btc_eth_max_leverage` | Integer | BTC/ETH 最大开仓杠杆 | AI 参考 |
| `altcoin_max_leverage` | Integer | 山寨币最大开仓杠杆 | AI 参考 |
| `btc_eth_max_position_value_ratio` | Float | BTC/ETH 单笔最大持仓价值系数 (权益的倍数) | **代码强制执行** (默认 5.0) |
| `altcoin_max_position_value_ratio` | Float | 山寨币单笔最大持仓价值系数 (权益的倍数) | **代码强制执行** (默认 1.0) |
| `max_margin_usage` | Float | 最大保证金使用率 (0.0-1.0) | **代码强制执行** (默认 0.9) |
| `min_position_size` | Float | 最小开仓金额 (USDT) | **代码强制执行** (默认 12.0) |
| `min_risk_reward_ratio` | Float | 最小盈亏比 | AI 参考 (默认 3.0) |
| `min_confidence` | Integer | 最小开仓置信度 (0-100) | AI 参考 (默认 75) |
| `macro_risk_adjustments` | Map&lt;String, Object&gt; | 宏观风险调整配置 | 见下文 |
| `enforce_stop_loss_take_profit` | Boolean | 是否强制执行止盈止损 | |
| `manual_stop_loss_pct` | Float | 手动止损比例 (如 0.05) | |
| `manual_take_profit_pct` | Float | 手动止盈比例 (如 0.1) | |
| `limit_order_timeout_min` | Integer | 限价单超时时间 (分钟) | |
| `min_limit_order_distance_pct` | Float | 最小限价单距离百分比 | |
| `position_size_mode` | String | 仓位计算模式 (`"equity_pct"`, `"fixed_usd"`) | |
| `fixed_position_usd` | Float | 固定仓位金额 (USDT) | |
| `equity_pct` | Float | 权益百分比 (如 0.1) | |
| `max_risk_per_trade_pct` | Float | 单笔最大风险比例 | |
| `sl_atr_mult` | Float | 止损 ATR 倍数 | |
| `tp_atr_mult` | Float | 止盈 ATR 倍数 | |
| `drawdown_tp_activation` | Float | 回撤止盈激活阈值 | |
| `drawdown_tp_threshold` | Float | 回撤止盈触发阈值 | |
| `symbol_cooldown_minutes` | Integer | 币种冷却时间 (分钟) | |
| `reverse_direction` | Boolean | 是否反向交易 | |
| `max_new_entries_per_direction` | Integer | 同方向最大开仓数 | |
| `require_breakout_close_confirm` | Boolean | 是否要求突破收盘确认 | |
| `breakout_min_atr_fraction` | Float | 突破最小 ATR 分数 | |
| `timeout_exit_minutes` | Integer | 超时离场时间 (分钟) | |
| `timeout_min_progress_pct` | Float | 超时最小进度百分比 | |
| `max_positions_min` | Integer | 动态最大持仓数量下限 | |
| `max_positions_max` | Integer | 动态最大持仓数量上限 | |
| `scan_interval_base_minutes` | Integer | 基础扫描间隔 (分钟) | |
| `scan_interval_base_min` | Integer | 扫描间隔下限 (分钟) | |
| `scan_interval_base_max` | Integer | 扫描间隔上限 (分钟) | |
| `min_hold_minutes` | Integer | 最小持仓时间 (分钟) | |
| `min_close_profit_pct` | Float | 最小平仓利润百分比 | |
| `close_when_profit_exceeds_pct` | Float | 利润超过此百分比时触发平仓 | |
| `close_when_drawdown_from_peak_pct` | Float | 利润回撤超过此百分比时触发平仓 | |
| `enable_atr_risk` | Boolean | 是否启用 ATR 动态风控 | |
| `atr_period` | Integer | ATR 周期 (如 14) | |
| `atr_multiplier` | Float | ATR 倍数 (如 2.0) | |
| `enable_adaptive_atr_multiplier` | Boolean | 是否启用自适应 ATR 倍数 | |
| `atr_multiplier_min` | Float | ATR 倍数下限 | |
| `atr_multiplier_max` | Float | ATR 倍数上限 | |
| `enable_staged_take_profit` | Boolean | 是否启用阶梯止盈 | |
| `stage1_profit_pct` | Float | 第一阶段止盈比例 | |
| `stage1_close_ratio` | Float | 第一阶段平仓比例 (0.0-1.0) | |
| `stage2_profit_pct` | Float | 第二阶段止盈比例 | |
| `stage2_close_ratio` | Float | 第二阶段平仓比例 (0.0-1.0) | |
| `stage3_profit_pct` | Float | 第三阶段止盈比例 | |
| `stage3_close_ratio` | Float | 第三阶段平仓比例 (0.0-1.0) | |
| `stage4_profit_pct` | Float | 第四阶段止盈比例 | |
| `stage4_close_ratio` | Float | 第四阶段平仓比例 (0.0-1.0) | |
| `enable_sideways_time_decay_close` | Boolean | 是否启用横盘时间衰减平仓 | |
| `sideways_band_pct` | Float | 横盘判定区间比例 | |
| `enable_sideways_micro_grid` | Boolean | 是否启用横盘微网格 | |
| `sideways_min_duration_min` | Integer | 横盘最小持续时间 (分钟) | |
| `sideways_close_profit_pct` | Float | 横盘平仓利润百分比 | |
| `sideways_close_ratio` | Float | 横盘平仓比例 | |
| `use_sideways_ratio_threshold` | Boolean | 是否使用横盘比例阈值 | |
| `sideways_ratio_min` | Float | 横盘比例下限 | |
| `require_indices_deterioration_for_time_decay_close` | Boolean | 时间衰减平仓是否需要指标恶化 | |
| `time_decay_deterioration_min_signals` | Integer | 时间衰减平仓所需最少恶化信号数 | |
| `sideways_band_lower_coeff` | Float | 横盘区间下限系数 | |
| `sideways_band_upper_coeff` | Float | 横盘区间上限系数 | |
| `sideways_heat_weighted_ratio_threshold` | Float | 横盘热度加权比例阈值 | |
| `trailing_stop_mode` | String | 移动止损模式 (如 `"app"`) | |
| `trailing_stop_callback_rate_pct` | Float | 移动止损回调比例 | |
| `trailing_stop_app_min_interval_sec` | Integer | 移动止损最小间隔 (秒) | |
| `enable_dynamic_take_profit` | Boolean | 是否启用动态止盈 | |
| `enable_roe_tp_ladder` | Boolean | 是否启用 ROE 阶梯止盈 | |
| `staged_activation_peak_roe_pct` | Float | 阶梯止盈激活峰值 ROE 百分比 | |
| `dynamic_tp_ladder` | Array&lt;Object&gt; | [动态止盈阶梯配置](#82-dynamic_tp_ladder-动态止盈阶梯配置) | |
| `enable_full_auto_evolution` | Boolean | 是否启用全自动进化 | |
| `enable_auto_tune_thresholds` | Boolean | 是否启用阈值自动调整 | |
| `auto_tune_lookback_minutes` | Integer | 自动调整回溯时间 (分钟) | |
| `enable_ai_evolution` | Boolean | 是否启用 AI 进化 | |
| `ai_evolution_min_trades` | Integer | AI 进化最小交易数 | |
| `evolution_mode` | String | 进化模式 (如 `"ai_primary"`) | |
| `factor_library_enabled` | Object | 因子库启用配置 (Key-Value) | |
| `default_stop_loss_min_pct` | Float | 默认最小止损比例 | |
| `enable_dynamic_stop_loss_min` | Boolean | 是否启用动态最小止损 | |
| `stop_loss_min_increase_max_pct` | Float | 止损最小增加上限比例 | |
| `risk_per_trade_pct` | Float | 单笔风险比例 | |
| `enable_dynamic_risk_per_trade` | Boolean | 是否启用动态单笔风险 | |
| `risk_per_trade_min_pct` | Float | 单笔风险比例下限 | |
| `risk_per_trade_max_pct` | Float | 单笔风险比例上限 | |
| `risk_per_trade_recent_window` | Integer | 动态风险计算窗口 | |
| `enable_trend_stop_loss` | Boolean | 是否启用趋势止损 | |
| `trend_stop_loss_min_signals` | Integer | 趋势止损最小信号数 | |
| `trend_stop_loss_trigger_loss_pct` | Float | 趋势止损触发亏损比例 | |
| `trend_stop_loss_min_hold_minutes` | Integer | 趋势止损最小持仓时间 (分钟) | |
| `take_profit_monitor_interval_sec` | Integer | 止盈监控间隔 (秒) | |
| `loss_throttle_sec` | Integer | 亏损节流时间 (秒) | |
| `unplaced_ttl_ms` | Integer | 未成交订单 TTL (毫秒) | |
| `unplaced_ttl_ms_min` | Integer | 未成交订单 TTL 下限 (毫秒) | |
| `unplaced_ttl_ms_max` | Integer | 未成交订单 TTL 上限 (毫秒) | |
| `tp_sl_order_update_cooldown_sec` | Integer | 止盈止损订单更新冷却 (秒) | |
| `entry_confidence_min` | Float | 最小入场置信度 | |
| `entry_window_min_min` | Integer | 入场窗口最小时间 (分钟) | |
| `entry_window_max_min` | Integer | 入场窗口最大时间 (分钟) | |
| `entry_urgency_now_max_min` | Integer | 紧急入场最大时间 (分钟) | |
| `entry_urgency_soon_max_min` | Integer | 快速入场最大时间 (分钟) | |
| `entry_rsi_overbought_soft` | Float | RSI 超买软阈值 | |
| `entry_rsi_overbought_hard` | Float | RSI 超买硬阈值 | |
| `entry_rsi_oversold_soft` | Float | RSI 超卖软阈值 | |
| `entry_rsi_oversold_hard` | Float | RSI 超卖硬阈值 | |
| `min_factor_pass` | Integer | 最小通过因子数 | |
| `min_reliability_for_submit` | Float | 提交所需最小可靠性 | |
| `threshold_heat_score_gold` | Float | 黄金热度评分阈值 | |
| `threshold_atr_util_green_pct` | Float | 绿色 ATR 利用率阈值 | |
| `first_layer_min_pass_count` | Integer | 第一层过滤最小通过数 | |
| `allow_trade_when_ai_fails` | Boolean | AI 失败时是否允许交易 | |

### 4.1 `risk_control_enhanced` (增强风控配置)

| 字段名 | 类型 | 含义 |
| :--- | :--- | :--- |
| `daily_loss_limit_pct` | Float | 每日亏损限制百分比 |
| `strategy_drawdown_limit_pct` | Float | 策略回撤限制百分比 |
| `max_risk_per_trade_pct` | Float | 单笔最大风险比例 |
| `auto_disable_on_loss` | Boolean | 触发亏损限制是否自动停用策略 |
| `daily_loss_threshold_pct` | Float | 每日亏损阈值百分比 |
| `risk_per_trade_mode` | String | 单笔风险模式 (`"fixed"`, `"dynamic"`) |
| `dynamic_risk_adjustment` | Object | 动态风险调整配置 (含 `market_regime_mapping`, `confidence_adjustment`, `volatility_adjustment`) |

---

## 5. `prompt_sections` (AI 提示词段落)

允许覆盖系统内置提示词的特定段落。

| 字段名 | 类型 | 含义 |
| :--- | :--- | :--- |
| `role_definition` | String | AI 角色定义 (Role Definition) |
| `trading_frequency` | String | 交易频率意识 (Trading Frequency Awareness) |
| `entry_standards` | String | 入场标准 (Entry Standards) |
| `decision_process` | String | 决策流程 (Decision Process) |

---

## 6. `grid_config` (网格交易配置)

仅当 `strategy_type` 为 `"grid_trading"` 时有效。

| 字段名 | 类型 | 含义 |
| :--- | :--- | :--- |
| `symbol` | String | 交易对 (例如 `"BTCUSDT"`) |
| `grid_count` | Integer | 网格格数 (5-50) |
| `total_investment` | Float | 总投入金额 (USDT) |
| `leverage` | Integer | 杠杆倍数 (1-20) |
| `upper_price` | Float | 价格区间上限 (0 = 自动计算) |
| `lower_price` | Float | 价格区间下限 (0 = 自动计算) |
| `use_atr_bounds` | Boolean | 是否使用 ATR 自动计算区间 |
| `atr_multiplier` | Float | ATR 倍数 (默认 2.0) |
| `distribution` | String | 网格分布：`"uniform"` (等差), `"gaussian"` (正态), `"pyramid"` (金字塔) |
| `max_drawdown_pct` | Float | 最大回撤止损百分比 |
| `stop_loss_pct` | Float | 单笔止损百分比 |
| `use_maker_only` | Boolean | 是否仅使用 Maker 单 (挂单) |
| `direction_bias_ratio` | Float | 方向偏好比例 (默认 0.7) |
| `daily_loss_limit_pct` | Float | 每日亏损限制百分比 |
| `enable_direction_adjust` | Boolean | 是否启用基于箱体突破的方向自动调整 |

---

## 7. `volatility_config` (波动率配置)

| 字段名 | 类型 | 含义 |
| :--- | :--- | :--- |
| `lookback_minutes` | Integer | 波动率计算回溯时间 (分钟) |
| `high_threshold_pct` | Float | 高波动率阈值百分比 |
| `medium_threshold_pct` | Float | 中波动率阈值百分比 |
| `base_symbols` | Array&lt;String&gt; | 基础参考币种列表 (如 `["BTCUSDT", "ETHUSDT"]`) |

---

## 8. `trailing_stop_config` (移动止损配置)

| 字段名 | 类型 | 含义 |
| :--- | :--- | :--- |
| `enabled` | Boolean | 是否启用移动止损 |
| `peak_decline` | Object | [峰值回撤配置](#81-peak_decline-峰值回撤配置) |

### 8.1 `peak_decline` (峰值回撤配置)

| 字段名 | 类型 | 含义 |
| :--- | :--- | :--- |
| `enabled` | Boolean | 是否启用峰值回撤检测 |
| `activation_pct` | Float | 激活回撤阈值百分比 |

### 8.2 `dynamic_tp_ladder` (动态止盈阶梯配置)

| 字段名 | 类型 | 含义 |
| :--- | :--- | :--- |
| `trigger_roe_pct` | Float | 触发 ROE 百分比 |
| `target_tp_pct` | Float | 目标止盈百分比 |

---

## 9. `alpha_factors` (Alpha 因子配置)

| 字段名 | 类型 | 含义 |
| :--- | :--- | :--- |
| `enable_liquidation_clusters` | Boolean | 是否启用清算簇分析 |
| `liquidation_cluster_threshold` | Float | 清算簇阈值 |
| `min_liquidation_usd` | Float | 最小清算金额 (USD) |
| `enable_exchange_flow` | Boolean | 是否启用交易所资金流分析 |
| `exchange_flow_lookback_hours` | Integer | 资金流回溯时间 (小时) |
| `significant_flow_threshold` | Float | 显著资金流阈值 |
| `enable_whale_wallet_tracking` | Boolean | 是否启用鲸鱼钱包追踪 |
| `whale_threshold_usd` | Float | 鲸鱼阈值 (USD) |
| `tracked_wallets` | Array&lt;String&gt; | 追踪的钱包地址列表 |
| `enable_spread_expansion` | Boolean | 是否启用价差扩张分析 |
| `spread_expansion_threshold` | Float | 价差扩张阈值 |
| `factor_weights` | Map&lt;String, Float&gt; | 各因子权重配置 |

---

## 10. `multi_agent` (多智能体配置)

| 字段名 | 类型 | 含义 |
| :--- | :--- | :--- |
| `enabled` | Boolean | 是否启用多智能体模式 |
| `agents` | Map&lt;String, AgentConfig&gt; | 智能体列表 (如 "bull", "bear", "risk_manager") |
| `voting_mechanism` | String | 投票机制 (`"majority"`, `"weighted"`, `"veto"`) |
| `min_agreement_pct` | Float | 最小一致性百分比 |
| `enable_genetic_evolution` | Boolean | 是否启用遗传进化 |
| `evolution_config` | Object | 进化配置 (种群大小, 变异率等) |

---

## 11. `macro_config` (宏观配置)

| 字段名 | 类型 | 含义 |
| :--- | :--- | :--- |
| `current_regime` | String | 当前宏观体制 (如 "inflationary_growth") |
| `assumed_conditions` | Map&lt;String, String&gt; | 假设条件 (如 {"interest_rate": "rising"}) |
| `regime_detection_frequency` | String | 体制检测频率 |
| `data_sources` | Array&lt;String&gt; | 数据源列表 |

### 11.1 `macro_screening` (宏观筛选配置)

| 字段名 | 类型 | 含义 |
| :--- | :--- | :--- |
| `enable_macro_filter` | Boolean | 是否启用宏观筛选 |
| `sector_allocation` | Map&lt;String, Float&gt; | 板块配置比例 |
| `max_sector_exposure` | Float | 最大板块敞口 |
