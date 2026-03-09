# 参数优化 API 文档

本文档详细描述了 `/api/strategies/optimize` 接口的请求参数。该接口用于启动遗传算法（GA）优化任务，以寻找最佳策略参数。

## 接口信息

- **路径**: `/api/strategies/optimize`
- **方法**: `POST`
- **Content-Type**: `application/json`

## 请求体结构 (JSON)

请求体是一个 JSON 对象，包含以下字段。必须提供 `strategy_id` 或 `strategy_config` 之一来指定要优化的策略。

### 1. 核心标识 (必选其一)

| 字段名 | 类型 | 描述 |
| :--- | :--- | :--- |
| `strategy_id` | `string` | 现有策略的唯一 ID（UUID）。如果提供，后端将从数据库加载策略配置。 |
| `strategy_config` | `object` | 完整的策略配置对象。如果提供，将直接使用此配置，覆盖数据库中的配置。适用于未保存的临时策略。 |

### 2. 优化参数范围 (`parameter_ranges`)

定义遗传算法需要搜索的参数及其取值范围。

- **类型**: `array` of objects
- **必填**: 否（若为空，系统尝试自动生成，但建议显式提供）

| 字段名 | 类型 | 描述 | 示例 |
| :--- | :--- | :--- | :--- |
| `name` | `string` | 参数在配置对象中的路径（点号分隔） | `"risk_control.min_confidence"` |
| `type` | `int` | 参数类型：`0` = 整数 (Int), `1` = 浮点数 (Float) | `0` |
| `min` | `number` | 搜索范围下限 | `60` |
| `max` | `number` | 搜索范围上限 | `90` |
| `step` | `number` | 搜索步长（0 表示连续浮点数） | `5` |

**示例:**
```json
"parameter_ranges": [
  {
    "name": "indicators.rsi_periods.0",
    "type": 0,
    "min": 7,
    "max": 21,
    "step": 1
  },
  {
    "name": "risk_control.min_risk_reward_ratio",
    "type": 1,
    "min": 1.5,
    "max": 5.0,
    "step": 0.1
  }
]
```

### 3. 优化目标 (`optimization_target`)

指定遗传算法优化的目标函数。

- **类型**: `string`
- **默认值**: `"profit"`
- **可选值**:
    - `"profit"`: 最大化总收益（Total Profit）
    - `"sharpe"`: 最大化夏普比率（Sharpe Ratio）
    - `"drawdown"`: 最小化最大回撤（Max Drawdown，内部处理为最大化负回撤）
    - `"win_rate"`: 最大化胜率（Win Rate）

### 4. 遗传算法配置 (`ga_config`)

控制遗传算法运行参数。

- **类型**: `object`
- **必填**: 是（可以传空对象 `{}` 使用默认值）

| 字段名 | 类型 | 默认值 | 范围限制 | 描述 |
| :--- | :--- | :--- | :--- | :--- |
| `population_size` | `int` | 20 | 10 - 1000 | 每一代包含的个体（策略组合）数量。 |
| `generations` | `int` | 10 | 1 - 500 | 进化的总代数。 |
| `mutation_rate` | `float` | 0.1 | 0.01 - 0.5 | 变异概率（0.1 = 10%）。 |
| `elite_size` | `int` | 2 | < 种群大小 | 每一代保留的最佳个体数量（不参与交叉变异）。 |
| `tournament_size` | `int` | 3 | > 1 | 锦标赛选择算法中每次参与竞争的个体数量。 |

### 5. 回测环境配置 (`backtest_config`)

定义用于评估每个参数组合性能的回测环境。

- **类型**: `object`
- **必填**: 是

| 字段名 | 类型 | 默认值 | 描述 |
| :--- | :--- | :--- | :--- |
| `symbols` | `string[]` | `["BTCUSDT"]` | 用于回测的交易对列表。 |
| `timeframes` | `string[]` | `["1h"]` | K 线时间周期（如 `"15m"`, `"1h"`, `"4h"`）。 |
| `start_time` | `int64` | 30天前 | 回测开始时间戳（秒）。 |
| `end_time` | `int64` | 当前时间 | 回测结束时间戳（秒）。 |
| `initial_balance` | `number` | 10000 | 初始资金（USDT）。 |
| `leverage` | `object` | - | 杠杆配置对象。 |

**Leverage 对象结构:**
```json
"leverage": {
  "btc_eth_leverage": 10,
  "altcoin_leverage": 5
}
```

## 完整请求示例

```json
{
  "strategy_id": "550e8400-e29b-41d4-a716-446655440000",
  "parameter_ranges": [
    {
      "name": "risk_control.min_confidence",
      "type": 0,
      "min": 60,
      "max": 90,
      "step": 5
    }
  ],
  "optimization_target": "sharpe",
  "ga_config": {
    "population_size": 50,
    "generations": 20,
    "mutation_rate": 0.15,
    "elite_size": 2,
    "tournament_size": 3
  },
  "backtest_config": {
    "symbols": ["ETHUSDT"],
    "timeframes": ["15m"],
    "start_time": 1704067200,
    "end_time": 1706659200,
    "initial_balance": 5000,
    "leverage": {
      "btc_eth_leverage": 5,
      "altcoin_leverage": 2
    }
  }
}
```

## 常见错误 (400 Bad Request)

1.  **缺少策略标识**: 未提供 `strategy_id` 也未提供 `strategy_config`。
2.  **无效的 GA 参数**:
    - `population_size` < 10 或 > 1000
    - `generations` > 500
    - `mutation_rate` < 0.01 或 > 0.5
3.  **缺少优化参数**: `parameter_ranges` 为空且系统无法自动推断（例如策略未启用相关指标）。
