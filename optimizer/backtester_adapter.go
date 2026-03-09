package optimizer

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"strings"
	"time"

	"nofx/backtest"
	"nofx/mcp"
	"nofx/store"
)

func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// GABacktester adapts the genetic algorithm to the backtesting engine
type GABacktester struct {
	baseConfig store.StrategyConfig
	btConfig   backtest.BacktestConfig
	aiClient   mcp.AIClient
}

// NewGABacktester creates a new adapter
func NewGABacktester(base store.StrategyConfig, btCfg backtest.BacktestConfig, client mcp.AIClient) *GABacktester {
	return &GABacktester{
		baseConfig: base,
		btConfig:   btCfg,
		aiClient:   client,
	}
}

// RunStrategy executes a backtest with the given chromosome
func (b *GABacktester) RunStrategy(c Chromosome) BacktestResult {
	// 1. Clone base config
	cfg := b.baseConfig
	// Deep copy needed if nested pointers
	// For simplicity, we assume shallow copy is fine for now as we modify values
	// But in concurrent env, better to deep copy.
	// In GA sequential evaluation, it's safer.
	
	// 2. Apply genes to config
	for name, value := range c.Genes {
		if err := setFieldValue(&cfg, name, value); err != nil {
			// Log error?
			// fmt.Printf("Error setting gene %s: %v\n", name, err)
		}
	}

	// 3. Validate configuration
	// Simple validation to prevent crashes
	if cfg.RiskControl.MaxPositions <= 0 {
		cfg.RiskControl.MaxPositions = 1
	}
	if cfg.RiskControl.SLATRMult < 0 {
		cfg.RiskControl.SLATRMult = 1.0
	}
	if cfg.RiskControl.TPATRMult < 0 {
		cfg.RiskControl.TPATRMult = 1.0
	}
	// Add more validation as needed

	// 4. Run backtest
	// Use RunID to avoid collision
	btConfigCopy := b.btConfig
	if btConfigCopy.RunID == "" || strings.HasPrefix(btConfigCopy.RunID, "opt_") {
		btConfigCopy.RunID = fmt.Sprintf("opt_%d_%s", time.Now().UnixNano(), randString(6))
	}
	// Important: We must set the strategy on the COPY of the config, not the original shared one
	btConfigCopy.SetLoadedStrategy(&cfg)
	
	runner, err := backtest.NewRunner(btConfigCopy, b.aiClient)
	if err != nil {
		return BacktestResult{Error: err.Error()}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if err := runner.Start(ctx); err != nil {
		return BacktestResult{
			Profit: -1000000,
			Error:  err.Error(),
		}
	}
	defer runner.Stop()

	if err := runner.Wait(); err != nil {
		// Log error but continue
	}

	// 5. Extract metrics
	stats := runner.GetStats()

	// Calculate returns for Sharpe ratio
	// Need daily equity points
	equityPoints, _ := runner.GetEquityPoints()
	returns := CalculateReturns(equityPoints)

	// Calculate missing metrics
	initialBalance := b.btConfig.InitialBalance
	totalProfit := stats.EquityLast - initialBalance

	return BacktestResult{
		Profit:  totalProfit,
		Returns: returns,
		MaxDD:   stats.MaxDrawdownPct,
		Trades:  stats.Trades,
		WinRate: stats.WinRate,
	}
}

// Helper to set field value by path (e.g. "RiskControl.MaxPositions")
func setFieldValue(obj interface{}, path string, value float64) error {
	v := reflect.ValueOf(obj).Elem()
	parts := strings.Split(path, ".")

	for _, part := range parts {
		if v.Kind() == reflect.Struct {
			v = v.FieldByName(part)
			if !v.IsValid() {
				return fmt.Errorf("field %s not found", part)
			}
		} else {
			// Handle map or slice if needed, for now simplistic struct traversal
			return fmt.Errorf("cannot traverse %s", part)
		}
	}

	if !v.CanSet() {
		return fmt.Errorf("cannot set field %s", path)
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.SetInt(int64(value))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v.SetUint(uint64(value))
	case reflect.Float32, reflect.Float64:
		v.SetFloat(value)
	case reflect.Bool:
		v.SetBool(value > 0.5)
	default:
		return fmt.Errorf("unsupported type %s", v.Kind())
	}

	return nil
}

// CalculateReturns calculates daily returns from equity points
func CalculateReturns(points []backtest.EquityPoint) []float64 {
	if len(points) < 2 {
		return nil
	}
	
	// Group by day
	dailyEquity := make(map[string]float64)
	var dates []string
	
	for _, p := range points {
		date := time.UnixMilli(p.Timestamp).Format("2006-01-02")
		// Use the last equity of the day
		dailyEquity[date] = p.Equity
		
		// Maintain order
		if len(dates) == 0 || dates[len(dates)-1] != date {
			dates = append(dates, date)
		}
	}
	
	var returns []float64
	for i := 1; i < len(dates); i++ {
		prev := dailyEquity[dates[i-1]]
		curr := dailyEquity[dates[i]]
		if prev > 0 {
			// Calculate daily return
			ret := (curr - prev) / prev
			returns = append(returns, ret)
		}
	}
	
	return returns
}

// CalculateSharpeRatio calculates annualized Sharpe ratio
func CalculateSharpeRatio(returns []float64) float64 {
	if len(returns) == 0 {
		return 0
	}
	
	sum := 0.0
	for _, r := range returns {
		sum += r
	}
	mean := sum / float64(len(returns))
	
	variance := 0.0
	for _, r := range returns {
		variance += math.Pow(r-mean, 2)
	}
	stdDev := math.Sqrt(variance / float64(len(returns)))
	
	if stdDev == 0 {
		return 0
	}
	
	// Annualize (assuming daily returns)
	return (mean / stdDev) * math.Sqrt(365)
}
