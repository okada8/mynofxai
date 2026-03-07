# 回测功能 API 文档

本文档描述了策略回测功能（Strategy Backtest / Test Run）的 API 接口参数、类型及其含义。

## 接口信息

- **接口路径**: `POST /api/strategies/test-run`
- **功能描述**: 执行策略回测。该接口不会执行真实交易，而是通过模拟环境或仅调用 AI 进行分析，返回 AI 的决策结果、生成的提示词（Prompt）以及候选币种等信息。

## 请求参数 (Request Body)

请求体为一个 JSON 对象，包含以下字段：

| 参数名 | 类型 | 必填 | 默认值 | 含义与说明 |
| :--- | :--- | :--- | :--- | :--- |
| `config` | `Object` | **是** | - | **策略配置对象**。包含了回测所需的完整策略参数，详情见下文 [StrategyConfig 结构](#strategyconfig-结构)。 |
| `prompt_variant` | `String` | 否 | `"balanced"` | **提示词变体**。用于控制生成 System Prompt 的风格。可选值：<br>• `"balanced"`: 平衡型（默认）<br>• `"aggressive"`: 激进型<br>• `"conservative"`: 保守型 |
| `ai_model_id` | `String` | 否 | - | **AI 模型 ID**。指定用于回测的 AI 模型 ID（对应 `ai_models` 表中的 ID）。<br>• 如果为空，回测将仅生成 Prompt 但不调用 AI。<br>• 如果不为空且 `run_real_ai` 为 `true`，则会调用该模型进行分析。 |
| `run_real_ai` | `Boolean` | 否 | `false` | **是否调用真实 AI**。<br>• `true`: 调用 `ai_model_id` 指定的模型进行真实推理，返回 AI 的分析结果。<br>• `false`: 仅执行数据获取和 Prompt 构建流程，不产生 AI 费用。 |

---

### StrategyConfig 结构

`config` 字段是一个复杂的对象，定义了策略的运行规则。

| 字段名 | 类型 | 说明 |
| :--- | :--- | :--- |
| `strategy_type` | `String` | 策略类型，默认为 `"ai_trading"`，也可为 `"grid_trading"`。 |
| `language` | `String` | 语言设置 (`"zh"` 或 `"en"`)，影响 Prompt 的生成语言。 |
| `coin_source` | `Object` | **选币源配置**，决定回测时的候选币种来源。详见 [CoinSourceConfig](#coinsourceconfig)。 |
| `indicators` | `Object` | **指标与数据配置**，决定回测时使用哪些技术指标和数据源。详见 [IndicatorConfig](#indicatorconfig)。 |
| `risk_control` | `Object` | **风控配置**，定义仓位管理和杠杆规则。详见 [RiskControlConfig](#riskcontrolconfig)。 |
| `custom_prompt` | `String` | 自定义 Prompt 后缀，会追加在 System Prompt 末尾。 |
| `prompt_sections`| `Object` | Prompt 的各部分模板（角色定义、交易频率等）。 |

#### CoinSourceConfig

| 字段名 | 类型 | 说明 |
| :--- | :--- | :--- |
| `source_type` | `String` | 选币源类型。<br>• `"ai500"`: AI500 优选池<br>• `"static"`: 静态币表<br>• `"oi_top"`: 持仓增量榜<br>• `"oi_low"`: 持仓减量榜<br>• `"mixed"`: 混合模式 |
| `static_coins` | `Array<String>` | 静态币种列表（当 `source_type="static"` 时使用），如 `["BTCUSDT", "ETHUSDT"]`。 |
| `use_ai500` | `Boolean` | 是否启用 AI500 选币。 |
| `ai500_limit` | `Integer` | AI500 选币数量限制。 |
| `use_oi_top` | `Boolean` | 是否启用 OI 增量榜。 |
| `use_oi_low` | `Boolean` | 是否启用 OI 减量榜。 |

#### IndicatorConfig

| 字段名 | 类型 | 说明 |
| :--- | :--- | :--- |
| `klines` | `Object` | **K线配置**。<br>• `primary_timeframe`: 主时间周期 (如 `"5m"`，**重要**，决定回测数据粒度)<br>• `primary_count`: 主周期 K 线数量 (如 `30`)<br>• `selected_timeframes`: 多周期列表 (如 `["5m", "1h", "4h"]`) |
| `enable_raw_klines`| `Boolean` | 是否启用原始 K 线数据（通常为 `true`）。 |
| `enable_ema` | `Boolean` | 是否启用 EMA 指标。 |
| `enable_rsi` | `Boolean` | 是否启用 RSI 指标。 |
| `enable_volume` | `Boolean` | 是否启用成交量分析。 |
| `enable_oi` | `Boolean` | 是否启用持仓量 (Open Interest) 数据。 |
| `enable_funding_rate`| `Boolean` | 是否启用资金费率数据。 |
| `enable_quant_data`| `Boolean` | 是否启用量化数据（资金流向等）。 |
| `nofxos_api_key` | `String` | NofxOS 数据服务的 API Key。 |

#### RiskControlConfig

| 字段名 | 类型 | 说明 |
| :--- | :--- | :--- |
| `max_positions` | `Integer` | 最大同时持仓数量。 |
| `btc_eth_max_leverage` | `Integer` | BTC/ETH 最大杠杆倍数。 |
| `altcoin_max_leverage` | `Integer` | 山寨币最大杠杆倍数。 |
| `min_confidence` | `Integer` | AI 开仓的最低置信度阈值 (0-100)。 |

## 响应数据 (Response)

成功请求将返回以下 JSON 数据：

```json
{
  "system_prompt": "...",       // 生成的系统提示词
  "user_prompt": "...",         // 生成的用户提示词（包含市场数据）
  "candidate_count": 10,        // 筛选出的候选币种数量
  "candidates": [...],          // 候选币种详情列表
  "prompt_variant": "balanced", // 使用的提示词变体
  "ai_response": "...",         // AI 的分析结果（如果 run_real_ai=true）
  "note": "..."                 // 状态备注
}
```
