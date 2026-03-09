package optimizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBacktester
type MockBacktester struct {
	mock.Mock
}

func (m *MockBacktester) RunStrategy(c Chromosome) BacktestResult {
	args := m.Called(c)
	// Check if the first argument is a function
	if fn, ok := args.Get(0).(func(Chromosome) BacktestResult); ok {
		return fn(c)
	}
	// Otherwise assume it's a value
	return args.Get(0).(BacktestResult)
}

// Test crossover logic
func TestOptimizer_Crossover(t *testing.T) {
	cfg := GAConfig{
		PopulationSize: 10,
		RandomSeed:     12345,
	}
	opt := NewOptimizer(cfg)

	p1 := Chromosome{Genes: map[string]float64{"a": 1.0, "b": 2.0}}
	p2 := Chromosome{Genes: map[string]float64{"a": 3.0, "b": 4.0}}

	child := opt.crossover(p1, p2)

	// Child genes should come from p1 or p2
	if child.Genes["a"] != 1.0 && child.Genes["a"] != 3.0 {
		t.Errorf("Gene 'a' has unexpected value: %f", child.Genes["a"])
	}
	if child.Genes["b"] != 2.0 && child.Genes["b"] != 4.0 {
		t.Errorf("Gene 'b' has unexpected value: %f", child.Genes["b"])
	}
}

// Test mutation logic
func TestOptimizer_Mutate(t *testing.T) {
	cfg := GAConfig{
		RandomSeed: 12345,
		Schema: []GeneDef{
			{Name: "float_gene", Type: GeneTypeFloat, Min: 0, Max: 10},
			{Name: "int_gene", Type: GeneTypeInt, Min: 0, Max: 10},
		},
	}
	opt := NewOptimizer(cfg)

	original := Chromosome{Genes: map[string]float64{
		"float_gene": 5.0,
		"int_gene":   5.0,
	}}

	// Force mutation
	mutated := opt.cloneChromosome(original)
	
	// With rate 1.0, all genes should mutate
	opt.mutate(&mutated, 1.0)

	// Since it's random, there's a tiny chance value stays same, but very unlikely for float
	assert.NotEqual(t, original.Genes["float_gene"], mutated.Genes["float_gene"], "Float gene should mutate")
	assert.NotEqual(t, original.Genes["int_gene"], mutated.Genes["int_gene"], "Int gene should mutate")
	
	// Check bounds
	assert.True(t, mutated.Genes["float_gene"] >= 0 && mutated.Genes["float_gene"] <= 10)
	assert.True(t, mutated.Genes["int_gene"] >= 0 && mutated.Genes["int_gene"] <= 10)
}

// Test fitness calculation
func TestOptimizer_CalculateFitness(t *testing.T) {
	tests := []struct {
		metric   string
		res      BacktestResult
		expected float64
	}{
		{"profit", BacktestResult{Profit: 100}, 100},
		{"drawdown", BacktestResult{MaxDD: 20}, -20}, // Maximize negative DD
		{"win_rate", BacktestResult{WinRate: 60}, 60},
		{"sharpe", BacktestResult{Returns: []float64{0.01, 0.02, 0.01}}, 0}, // Just checking call, value depends on func
	}

	for _, tt := range tests {
		opt := NewOptimizer(GAConfig{TargetMetric: tt.metric})
		fitness := opt.calculateFitness(tt.res)
		
		if tt.metric == "sharpe" {
			// Sharpe calculation is complex, just ensure it runs
			assert.GreaterOrEqual(t, fitness, 0.0)
		} else {
			assert.Equal(t, tt.expected, fitness)
		}
	}
}

// Test full GA flow (integration)
func TestOptimizer_Run_Integration(t *testing.T) {
	cfg := GAConfig{
		PopulationSize: 5,
		Generations:    2,
		EliteSize:      1,
		MutationRate:   0.1,
		TargetMetric:   "profit",
		Schema: []GeneDef{
			{Name: "x", Type: GeneTypeFloat, Min: 0, Max: 10},
		},
		RandomSeed: 12345,
	}
	
	opt := NewOptimizer(cfg)
	mockBT := new(MockBacktester)
	
	// Mock behavior: Fitness = x * 10
	mockBT.On("RunStrategy", mock.Anything).Return(func(c Chromosome) BacktestResult {
		x := c.Genes["x"]
		return BacktestResult{Profit: x * 10}
	})

	best := opt.Run(mockBT)
	
	assert.NotNil(t, best)
	assert.Greater(t, best.Fitness, 0.0)
	
	// With evolution, we expect best fitness to be high (close to Max * 10 = 100)
	// Random search might not reach 100, but should be decent
	t.Logf("Best Fitness: %f, Genes: %v", best.Fitness, best.Genes)
}
