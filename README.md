# NOFX — 开源 AI 交易操作系统 (Open Source AI Trading OS)

<p align="center">
  <strong>为 AI 驱动的金融交易打造的基础设施层。</strong>
</p>

<p align="center">
  <a href="https://golang.org/"><img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go" alt="Go"></a>
  <a href="https://reactjs.org/"><img src="https://img.shields.io/badge/React-18+-61DAFB?style=flat&logo=react" alt="React"></a>
  <a href="https://www.typescriptlang.org/"><img src="https://img.shields.io/badge/TypeScript-5.0+-3178C6?style=flat&logo=typescript" alt="TypeScript"></a>
  <a href="https://github.com/NoFxAiOS/nofx/blob/main/LICENSE"><img src="https://img.shields.io/badge/License-AGPL--3.0-blue.svg?style=for-the-badge" alt="License"></a>
</p>

## 📖 项目简介

NOFX 是一个现代化的 AI 交易操作系统，旨在连接大语言模型（LLM）与加密货币交易所。它允许用户利用 DeepSeek、OpenAI、Claude 等顶尖 AI 模型进行自动化交易、策略回测、市场分析和多模型辩论。

本项目采用 **Go (Backend)** 和 **React (Frontend)** 的前后端分离架构，提供高性能的并发处理能力和现代化的用户交互体验。

## 🛠 技术栈

### 后端 (Backend)
- **语言**: Go 1.25+
- **框架**: Gin (Web Framework)
- **数据库**: SQLite (默认) / PostgreSQL
- **ORM**: GORM
- **其他库**: 
  - `go-binance`, `go-hyperliquid` (交易所 SDK)
  - `golang-jwt` (认证)
  - `gorilla/websocket` (实时数据)

### 前端 (Frontend)
- **框架**: React 18
- **构建工具**: Vite
- **语言**: TypeScript
- **样式**: Tailwind CSS
- **图表**: Lightweight Charts (TradingView), Recharts
- **状态管理**: Zustand

> **注意**: 虽然此前可能有关于 Vue 的讨论，但本项目前端完全基于 **React** 生态构建，以提供更强的组件化能力和 TypeScript 支持。

## ✨ 主要功能

- **多模型支持**: 集成 DeepSeek, Qwen, GPT-4, Claude, Gemini 等主流大模型。
- **多交易所接入**: 支持 Binance, OKX, Bybit, Bitget, Gate, KuCoin, Indodax, Hyperliquid, Aster, Lighter 等 CEX 和 DEX。
- **策略工坊 (Strategy Studio)**: 可视化构建交易策略，支持多种技术指标 (RSI, MACD, Bollinger Bands 等)。
- **AI 辩论场 (Debate Arena)**: 让多个 AI 扮演不同角色（多头、空头、风控官）进行辩论，生成综合交易决策。
- **回测实验室 (Backtest Lab)**: 基于历史数据的高性能策略回测，提供详细的盈亏分析和资金曲线。
- **智能缓存系统 (Smart Caching)**: 
  - **指标缓存**: 自动缓存技术指标计算结果，回测速度提升 50-100 倍。
  - **LLM 语义缓存**: 支持磁盘/内存二级缓存与 Prompt 智能去重，大幅降低 AI API 调用成本。
- **高性能架构**: 前端采用 Web Worker 并行计算图表指标，后端采用对象池与引用计数优化内存占用。
- **实时仪表盘**: 实时监控持仓、订单、账户权益和 AI 决策日志（包含思维链 CoT）。
- **Telegram 通知**: 实时推送交易信号、风控预警及系统状态报告到指定的 Telegram 群组或频道。

## 🚀 快速开始

### 前置要求
- **Go**: 1.25 或更高版本
- **Node.js**: 18 或更高版本
- **TA-Lib**: 技术分析库 (必须安装，否则编译会失败)

#### 安装 TA-Lib
- **macOS**: `brew install ta-lib`
- **Ubuntu/Debian**: `sudo apt-get install libta-lib0-dev`

### 本地开发运行

1. **克隆项目**
   ```bash
   git clone https://github.com/NoFxAiOS/nofx.git
   cd nofx
   ```

2. **后端配置与启动**
   ```bash
   # 进入项目根目录
   cd nofx

   # 整理依赖
   go mod tidy

   # 编译并运行
   go build -o nofx_server
   ./nofx_server
   ```
   后端服务默认运行在 `http://localhost:8080`。

3. **前端配置与启动**
   ```bash
   # 新开一个终端窗口，进入 web 目录
   cd web

   # 安装依赖
   npm install

   # 启动开发服务器
   npm run dev
   ```
   前端页面默认运行在 `http://localhost:3000`。

## ⚙️ 配置说明

项目主要通过环境变量进行配置。请参考根目录下的 `.env.example` 创建 `.env` 文件。

### 核心配置 (.env)

```ini
# --- 服务器配置 ---
PORT=8080
GIN_MODE=debug  # 生产环境请设置为 release

# --- 认证配置 ---
JWT_SECRET=your_super_secret_key_change_this

# --- 数据库配置 ---
# 支持 sqlite 或 postgres
DB_DRIVER=sqlite
DB_DSN=nofx.db

# --- 外部服务 (可选) ---
# 币安 API (也可在网页端配置)
BINANCE_API_KEY=your_binance_api_key
BINANCE_SECRET_KEY=your_binance_secret_key

# --- AI 模型配置 (可选) ---
# DeepSeek API (也可在网页端配置)
DEEPSEEK_API_KEY=your_deepseek_api_key

# --- 消息通知 (可选) ---
# Telegram Bot Token (从 @BotFather 获取)
TELEGRAM_BOT_TOKEN=your_bot_token
# 接收通知的 Chat ID (个人或群组 ID)
TELEGRAM_CHAT_ID=your_chat_id

# --- 缓存配置 (可选) ---
# 开启 LLM 响应缓存 (true/false)
ENABLE_LLM_CACHE=true
# 缓存目录 (默认 ~/.nofi/llm_cache)
LLM_CACHE_DIR=/path/to/cache
```

### 交易所配置
支持在 `.env` 中预配置，也可以在启动后的 Web 界面中动态添加和管理 API Key。建议在 Web 界面管理以支持多账户。

## 📂 项目结构

```
nofx/
├── api/            # HTTP API 处理逻辑 (Gin Handlers)
├── auth/           # 认证与授权模块
├── backtest/       # 回测引擎核心逻辑
├── cmd/            # 命令行工具入口
├── config/         # 全局配置加载
├── docs/           # 项目文档
├── kernel/         # 核心交易引擎与 AI 提示词构建
├── mcp/            # Model Context Protocol (AI 模型适配层)
├── model/          # 数据库模型定义 (GORM)
├── provider/       # 市场数据提供商 (K线, 深度等)
├── trader/         # 各大交易所的交易实现 (Binance, OKX...)
├── web/            # React 前端源代码
├── go.mod          # Go 依赖定义
└── main.go         # 程序入口
```

## 🤝 贡献指南

欢迎提交 Pull Request 或 Issue！请确保在提交代码前运行测试：

```bash
go test ./...
```

## 📄 许可证

本项目基于 **AGPL-3.0** 许可证开源。
