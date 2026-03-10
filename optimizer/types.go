package optimizer

// GeneType defines the type of gene
type GeneType int

const (
	GeneTypeInt GeneType = iota
	GeneTypeFloat
)

// GeneDef defines a gene for optimization
type GeneDef struct {
	Name string   `json:"name"`
	Type GeneType `json:"type"`
	Min  float64  `json:"min"`
	Max  float64  `json:"max"`
	Step float64  `json:"step"` // 0 for continuous
}

// Chromosome represents a single solution
type Chromosome struct {
	Genes   map[string]float64 `json:"genes"`
	Fitness float64            `json:"fitness"`
}

// GAConfig configuration for Genetic Algorithm
type GAConfig struct {
	PopulationSize int       `json:"population_size"`
	Generations    int       `json:"generations"`
	MutationRate   float64   `json:"mutation_rate"`
	EliteSize      int       `json:"elite_size"`
	TournamentSize int       `json:"tournament_size"`
	Schema         []GeneDef `json:"schema"`
	RandomSeed     int64     `json:"random_seed"`

	// Optimization target
	TargetMetric string // "sharpe", "profit", "drawdown", "win_rate"

	// Complexity score for adaptive worker count (1.0 = normal, >2.0 = complex)
	ComplexityScore float64 `json:"complexity_score"`

	// Progress callback - ignore in JSON
	OnProgress func(generation int, bestFitness float64, bestChromosome Chromosome) `json:"-"`
}

// BacktestResult represents the result of a backtest run
type BacktestResult struct {
	Profit  float64
	Returns []float64 // Daily returns for Sharpe ratio
	MaxDD   float64
	Trades  int
	WinRate float64
	Error   string
}

// Backtester interface for running strategies
type Backtester interface {
	RunStrategy(c Chromosome) BacktestResult
}
