package optimizer

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"

	"nofx/backtest"
	"nofx/mcp"
	"nofx/store"
)

// OptimizationResult represents final result
type OptimizationResult struct {
	BestChromosome Chromosome
	BestConfig     store.StrategyConfig
}

// GenerationStats tracks history
type GenerationStats struct {
	Generation  int
	BestFitness float64
}

// OptimizationProgress tracks progress
type OptimizationProgress struct {
	Generation  int               `json:"generation"`
	BestFitness float64           `json:"best_fitness"`
	BestGenes   map[string]float64 `json:"best_genes"`
	History     []GenerationStats `json:"history"`
}

// OptimizationTask represents an optimization task
type OptimizationTask struct {
	ID        string               `json:"id"`
	Status    string               `json:"status"` // "running", "completed", "failed", "cancelled"
	Config    GAConfig             `json:"config"`
	Progress  OptimizationProgress `json:"progress"`
	Result    *OptimizationResult  `json:"result,omitempty"`
	CreatedAt time.Time            `json:"created_at"`
	Error     string               `json:"error,omitempty"`
	stopChan  chan struct{}        // Channel to signal stop
	mu        sync.RWMutex         // Protects concurrent access
}

// Manager manages optimization tasks
type Manager struct {
	tasks         sync.Map
	strategyStore *store.StrategyStore
	mcpClient     mcp.AIClient
}

// NewManager creates a new manager
func NewManager(strategyStore *store.StrategyStore, mcpClient mcp.AIClient) *Manager {
	m := &Manager{
		strategyStore: strategyStore,
		mcpClient:     mcpClient,
	}
	// Start cleanup routine
	go m.cleanupRoutine()
	return m
}

func (m *Manager) cleanupRoutine() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		m.cleanupOldTasks(24 * time.Hour)
	}
}

func (m *Manager) cleanupOldTasks(maxAge time.Duration) {
	now := time.Now()
	m.tasks.Range(func(key, value interface{}) bool {
		task, ok := value.(*OptimizationTask)
		if !ok {
			return true
		}
		if now.Sub(task.CreatedAt) > maxAge {
			m.tasks.Delete(key)
		}
		return true
	})
}

// StartOptimization starts a new optimization task
func (m *Manager) StartOptimization(baseConfig store.StrategyConfig, gaCfg GAConfig, btCfg backtest.BacktestConfig, aiClient mcp.AIClient) (string, error) {
	id := uuid.New().String()

	task := &OptimizationTask{
		ID:        id,
		Status:    "running",
		Config:    gaCfg,
		CreatedAt: time.Now(),
		Progress: OptimizationProgress{
			History: make([]GenerationStats, 0),
		},
		stopChan: make(chan struct{}),
	}

	// Set up progress callback
	// Note: We need to be careful with concurrency here as OnProgress is called from GA
	gaCfg.OnProgress = func(gen int, bestFitness float64, bestChrom Chromosome) {
		task.mu.Lock()
		defer task.mu.Unlock()
		task.Progress.Generation = gen
		task.Progress.BestFitness = bestFitness
		task.Progress.BestGenes = bestChrom.Genes
		task.Progress.History = append(task.Progress.History, GenerationStats{
			Generation:  gen,
			BestFitness: bestFitness,
		})
	}

	m.tasks.Store(id, task)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				task.mu.Lock()
				task.Status = "failed"
				task.Error = "Panic during optimization"
				task.mu.Unlock()
			}
		}()

		adapter := NewGABacktester(baseConfig, btCfg, aiClient)
		optimizer := NewOptimizer(gaCfg)
		// Set stop channel for optimizer
		optimizer.SetStopChan(task.stopChan)

		// Run optimization
		bestChrom := optimizer.Run(adapter)
		
		task.mu.Lock()
		defer task.mu.Unlock()
		
		select {
		case <-task.stopChan:
			task.Status = "cancelled"
			return
		default:
		}

		// Apply genes to get best config
		bestConfig := baseConfig // Shallow copy
		// Deep copy needed
		configBytes, _ := json.Marshal(baseConfig)
		json.Unmarshal(configBytes, &bestConfig)
		
		for name, val := range bestChrom.Genes {
			if err := setFieldValue(&bestConfig, name, val); err != nil {
				// Log error?
			}
		}

		task.Result = &OptimizationResult{
			BestChromosome: bestChrom,
			BestConfig:     bestConfig,
		}
		task.Status = "completed"
	}()

	return id, nil
}

// CancelTask cancels an optimization task
func (m *Manager) CancelTask(id string) error {
	val, ok := m.tasks.Load(id)
	if !ok {
		return nil
	}
	task := val.(*OptimizationTask)
	
	// Non-blocking send to avoid deadlock if task already finished
	select {
	case task.stopChan <- struct{}{}:
	default:
	}
	
	task.mu.Lock()
	task.Status = "cancelled"
	task.mu.Unlock()
	return nil
}

// GetTask retrieves a task copy
func (m *Manager) GetTask(id string) (*OptimizationTask, bool) {
	val, ok := m.tasks.Load(id)
	if !ok {
		return nil, false
	}
	task := val.(*OptimizationTask)
	
	task.mu.RLock()
	defer task.mu.RUnlock()
	
	// Return a copy to avoid race conditions during JSON marshalling
	copyTask := *task
	// Deep copy history slice
	if task.Progress.History != nil {
		historyCopy := make([]GenerationStats, len(task.Progress.History))
		copy(historyCopy, task.Progress.History)
		copyTask.Progress.History = historyCopy
	}
	// Deep copy genes map
	if task.Progress.BestGenes != nil {
		genesCopy := make(map[string]float64, len(task.Progress.BestGenes))
		for k, v := range task.Progress.BestGenes {
			genesCopy[k] = v
		}
		copyTask.Progress.BestGenes = genesCopy
	}
	
	return &copyTask, true
}

