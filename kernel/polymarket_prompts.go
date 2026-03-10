package kernel

const PolymarketAnalysisTemplate = `
你是一个预测市场专家。请分析以下事件：

事件: {{.Question}}
当前 YES 价格: ${{.YesPrice}} ({{.YesPercent}}%)
当前 NO 价格: ${{.NoPrice}} ({{.NoPercent}}%)
流动性: ${{.Liquidity | formatCurrency}}
24小时交易量: ${{.Volume24h | formatCurrency}}
结束时间: {{.EndTime | formatTime}}
剩余时间: {{.TimeLeft}}

历史数据:
{{range .History}}
- {{.Time | formatTime}}: YES=${{.YesPrice}}, NO=${{.NoPrice}}, 成交量=${{.Volume}}
{{end}}

请从以下角度分析：
1. **市场效率分析**: 当前价格是否反映了所有公开信息？
2. **风险评估**: 流动性风险、时间衰减、信息更新风险
3. **机会识别**: 是否存在定价错误或套利机会？
4. **交易建议**:
   - 建议仓位: {{if .Liquidity > 500000}}正常{{else}}轻仓{{end}}
   - 建议方向: {{if .YesPercent > 60}}YES{{else if .YesPercent < 40}}NO{{else}}观望{{end}}
   - 止损建议: 如果价格反向移动 {{if .Liquidity > 1000000}}15{{else}}10{{end}}%

请给出具体、可执行的交易建议。
`

const PolymarketDebateRoles = `
辩论角色分配:
1. **多头分析师**: 相信事件会发生，寻找支持 YES 的证据
2. **空头分析师**: 相信事件不会发生，寻找支持 NO 的证据 
3. **风险控制官**: 评估流动性风险、时间价值和极端情况
4. **市场结构专家**: 分析市场微观结构、流动性和交易成本

每个角色请提供:
- 核心论点 (1-2个关键理由)
- 风险评估 (主要风险点)
- 具体交易建议 (仓位、价格、时机)
`
