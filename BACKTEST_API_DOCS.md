# 回测功能 API 文档

本文档描述了策略回测功能（Strategy Backtest / Test Run）以及历史数据回测（Historical Backtest）的 API 接口参数、类型及其含义。

## 1. 策略试运行 (Strategy Test Run)

- **接口路径**: `POST /api/strategies/test-run`
- **功能描述**: 执行单次策略试运行。该接口不会执行真实交易，而是通过模拟环境或仅调用 AI 进行分析，返回 AI 的决策结果、生成的提示词（Prompt）以及候选币种等信息。通常用于策略调试和 Prompt 预览。

### 请求参数 (Request Body)

请求体为一个 JSON 对象，包含以下字段：

| 参数名 | 类型 | 必填 | 默认值 | 含义与说明 |
| :--- | :--- | :--- | :--- | :--- |
| `config` | `Object` | **是** | - | **策略配置对象**。包含了回测所需的完整策略参数，详情见 [策略配置文档](STRATEGY_CONFIG_DOCS.md)。 |
| `prompt_variant` | `String` | 否 | `"balanced"` | **提示词变体**。用于控制生成 System Prompt 的风格。可选值：<br>• `"balanced"`: 平衡型（默认）<br>• `"aggressive"`: 激进型<br>• `"conservative"`: 保守型 |
| `ai_model_id` | `String` | 否 | - | **AI 模型 ID**。指定用于回测的 AI 模型 ID（对应 `ai_models` 表中的 ID）。<br>• 如果为空，回测将仅生成 Prompt 但不调用 AI。<br>• 如果不为空且 `run_real_ai` 为 `true`，则会调用该模型进行分析。 |
| `run_real_ai` | `Boolean` | 否 | `false` | **是否调用真实 AI**。<br>• `true`: 调用 `ai_model_id` 指定的模型进行真实推理，返回 AI 的分析结果。<br>• `false`: 仅执行数据获取和 Prompt 构建流程，不产生 AI 费用。 |

### 响应数据 (Response)

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

---

## 2. 历史数据回测 (Historical Backtest)

- **接口路径**: `POST /api/backtest/start`
- **功能描述**: 启动一个完整的历史数据回测任务。系统将在指定的时间范围内，针对选定的币种和策略配置，模拟执行交易并生成详细的性能报告。

### 请求参数 (Request Body)

请求体包含一个 `config` 对象（对应 `BacktestConfig` 结构）：

```json
{
  "config": {
    "run_id": "bt_20231027_001",
    "strategy_id": "uuid-of-strategy",
    "start_ts": 1672531200,
    "end_ts": 1675123200,
    ...
  }
}
```

#### BacktestConfig 字段详解

| 字段名 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `run_id` | `String` | 否 | 回测任务 ID。如未提供，系统将自动生成 (例如 `bt_YYYYMMDD_HHMMSS`)。 |
| `strategy_id` | `String` | 否 | **策略 ID**。如果提供，系统将加载该策略的配置作为基础配置。 |
| `symbols` | `Array<String>` | 否 | **回测币种列表** (例如 `["BTCUSDT", "ETHUSDT"]`)。如果不提供且指定了 `strategy_id`，将使用策略中定义的选币逻辑自动解析币种。 |
| `start_ts` | `Integer` | **是** | **开始时间戳** (秒)。 |
| `end_ts` | `Integer` | **是** | **结束时间戳** (秒)。 |
| `initial_balance` | `Float` | 否 | 初始资金 (默认 10000.0)。 |
| `decision_timeframe` | `String` | 否 | 决策时间周期 (例如 `"15m"`, `"1h"`)。默认为策略中配置的主周期。 |
| `cache_ai` | `Boolean` | 否 | **启用 AI 缓存**。如果为 `true`，系统将尝试读取和保存 AI 响应到本地缓存，以加速回测并节省成本。 |
| `ai_cache_path` | `String` | 否 | 自定义 AI 缓存路径。默认使用系统全局缓存。 |
| `replay_only` | `Boolean` | 否 | **仅重放模式**。如果为 `true`，回测引擎将只使用缓存中的 AI 决策，遇到缓存未命中时不会调用真实 AI，而是跳过或使用默认操作。适合复盘分析。 |
| `custom_prompt` | `String` | 否 | 自定义 System Prompt 后缀。 |
| `ai_model_id` | `String` | 否 | 指定使用的 AI 模型 ID。如果不提供，将使用用户的默认模型。 |
| `ai_cfg` | `Object` | 否 | **临时 AI 配置**。允许在回测时临时覆盖 AI 模型参数 (如 `temperature`, `max_tokens`)。 |

### 响应数据 (Response)

请求成功后立即返回回测任务的初始元数据：

```json
{
  "run_id": "bt_20231027_001",
  "state": "running",
  "user_id": "user_123",
  "strategy_id": "...",
  "created_at": "2023-10-27T10:00:00Z",
  "summary": {
    "progress_pct": 0,
    "equity_last": 10000.0
    ...
  }
}
```

### 任务管理接口

- **查询状态**: `GET /api/backtest/status?run_id=...`
- **暂停任务**: `POST /api/backtest/pause` (`{"run_id": "..."}`)
- **恢复任务**: `POST /api/backtest/resume` (`{"run_id": "..."}`)
- **停止任务**: `POST /api/backtest/stop` (`{"run_id": "..."}`)
- **删除任务**: `POST /api/backtest/delete` (`{"run_id": "..."}`)
- **获取权益曲线**: `GET /api/backtest/equity?run_id=...`
- **获取交易记录**: `GET /api/backtest/trades?run_id=...`
- **导出结果**: `GET /api/backtest/export?run_id=...` (下载 ZIP 包)
