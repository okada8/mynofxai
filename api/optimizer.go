package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"nofx/backtest"
	"nofx/optimizer"
	"nofx/store"
)

// StartOptimizationRequest defines request body
type StartOptimizationRequest struct {
	StrategyID         string                 `json:"strategy_id"` // Used to fetch StrategyConfig
	StrategyConfig     *store.StrategyConfig  `json:"strategy_config,omitempty"` // Optional override
	ParameterRanges    []optimizer.GeneDef    `json:"parameter_ranges"`
	OptimizationTarget string                 `json:"optimization_target"`
	GAConfig           optimizer.GAConfig     `json:"ga_config"`
	BacktestConfig     backtest.BacktestConfig `json:"backtest_config"`
}

// handleRunOptimization starts optimization task
func (s *Server) handleRunOptimization(c *gin.Context) {
	var req StartOptimizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SafeBadRequest(c, err.Error())
		return
	}

	// 1. Resolve Strategy Config
	var strategyConfig store.StrategyConfig
	if req.StrategyConfig != nil {
		strategyConfig = *req.StrategyConfig
	} else if req.StrategyID != "" {
		// Fetch from DB
		var strategy store.Strategy
		if err := s.store.GormDB().First(&strategy, "id = ?", req.StrategyID).Error; err != nil {
			SafeBadRequest(c, "Strategy not found: "+err.Error())
			return
		}
		// Parse config string to struct
		config, err := strategy.ParseConfig()
		if err != nil {
			SafeBadRequest(c, "Failed to parse strategy config: "+err.Error())
			return
		}
		strategyConfig = *config
	} else {
		SafeBadRequest(c, "Either strategy_id or strategy_config must be provided")
		return
	}

	// 2. Default Parameter Ranges if empty
	if len(req.ParameterRanges) == 0 {
		// Auto-generate based on strategy type
		req.ParameterRanges = generateDefaultParameterRanges(strategyConfig)
		if len(req.ParameterRanges) == 0 {
			SafeBadRequest(c, "No parameters selected for optimization and auto-generation failed")
			return
		}
	}

	// Default GA Config if empty
	gaCfg := req.GAConfig
	if gaCfg.PopulationSize == 0 {
		gaCfg.PopulationSize = 20
	}
	if gaCfg.Generations == 0 {
		gaCfg.Generations = 10
	}
	if gaCfg.MutationRate == 0 {
		gaCfg.MutationRate = 0.1
	}
	if gaCfg.EliteSize == 0 {
		gaCfg.EliteSize = 2
	}
	if gaCfg.TournamentSize == 0 {
		gaCfg.TournamentSize = 3
	}
	gaCfg.Schema = req.ParameterRanges
	gaCfg.TargetMetric = req.OptimizationTarget
	if gaCfg.TargetMetric == "" {
		gaCfg.TargetMetric = "profit"
	}

	// Range validation
	if req.GAConfig.PopulationSize > 0 {
		if req.GAConfig.PopulationSize < 10 || req.GAConfig.PopulationSize > 1000 {
			SafeBadRequest(c, "Population size must be between 10 and 1000")
			return
		}
		gaCfg.PopulationSize = req.GAConfig.PopulationSize
	}
	if req.GAConfig.Generations > 0 {
		if req.GAConfig.Generations < 1 || req.GAConfig.Generations > 500 {
			SafeBadRequest(c, "Generations must be between 1 and 500")
			return
		}
		gaCfg.Generations = req.GAConfig.Generations
	}
	if req.GAConfig.MutationRate > 0 {
		if req.GAConfig.MutationRate < 0.01 || req.GAConfig.MutationRate > 0.5 {
			SafeBadRequest(c, "Mutation rate must be between 0.01 and 0.5")
			return
		}
		gaCfg.MutationRate = req.GAConfig.MutationRate
	}

	// Setup Backtest Config
	btCfg := req.BacktestConfig
	// Ensure defaults
	if btCfg.InitialBalance == 0 {
		btCfg.InitialBalance = 10000
	}
	if btCfg.StartTS == 0 {
		// Default to last 30 days
		btCfg.EndTS = time.Now().Unix()
		btCfg.StartTS = btCfg.EndTS - 30*24*3600
	}
	if len(btCfg.Symbols) == 0 {
		btCfg.Symbols = []string{"BTCUSDT"}
	}
	if len(btCfg.Timeframes) == 0 {
		btCfg.Timeframes = []string{"1h"}
	}
	
	// Ensure mandatory fields are present to prevent crashes
	if btCfg.Leverage.BTCETHLeverage == 0 {
		btCfg.Leverage.BTCETHLeverage = 10
	}
	if btCfg.Leverage.AltcoinLeverage == 0 {
		btCfg.Leverage.AltcoinLeverage = 5
	}
	if btCfg.AICfg.ModelID == "" {
		// Use a safe default or handle in backtest runner
		// Here we assume mcpClient handles it or runner defaults
	}

	// Use a temporary run ID for internal tracking, real run IDs are generated per evaluation
	btCfg.RunID = "opt-" + time.Now().Format("20060102-150405")

	// Get AI Client
	aiClient := s.mcpClient

	// Start Task
	taskID, err := s.optimizerManager.StartOptimization(strategyConfig, gaCfg, btCfg, aiClient)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Optimization started",
		"task_id": taskID,
	})
}

// Helper to generate default parameter ranges
func generateDefaultParameterRanges(cfg store.StrategyConfig) []optimizer.GeneDef {
	var ranges []optimizer.GeneDef

	// Common Indicators
	if cfg.Indicators.EnableRSI {
		ranges = append(ranges, optimizer.GeneDef{
			Name: "Indicators.RSIPeriods.0", // Assuming single period for now or need smarter mapping
			Type: optimizer.GeneTypeInt,
			Min:  7,
			Max:  21,
			Step: 1,
		})
	}
	if cfg.Indicators.EnableEMA {
		ranges = append(ranges, optimizer.GeneDef{
			Name: "Indicators.EMAPeriods.0",
			Type: optimizer.GeneTypeInt,
			Min:  5,
			Max:  50,
			Step: 1,
		})
	}
	
	// Risk Control
	ranges = append(ranges, optimizer.GeneDef{
		Name: "RiskControl.MinConfidence",
		Type: optimizer.GeneTypeInt, // It's float but treated as int 0-100 often, or float 0-1
		Min:  50,
		Max:  90,
		Step: 5,
	})
	
	// Grid Specific
	if cfg.StrategyType == "grid_trading" && cfg.GridConfig != nil {
		ranges = append(ranges, optimizer.GeneDef{
			Name: "GridConfig.GridCount",
			Type: optimizer.GeneTypeInt,
			Min:  5,
			Max:  50,
			Step: 5,
		})
		ranges = append(ranges, optimizer.GeneDef{
			Name: "GridConfig.Leverage",
			Type: optimizer.GeneTypeInt,
			Min:  1,
			Max:  10, // Safe range
			Step: 1,
		})
	}

	return ranges
}

// handleGetOptimizationStatus gets task status
func (s *Server) handleGetOptimizationStatus(c *gin.Context) {
	taskID := c.Param("id")
	task, ok := s.optimizerManager.GetTask(taskID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// handleCancelOptimization cancels optimization
func (s *Server) handleCancelOptimization(c *gin.Context) {
	taskID := c.Param("id")
	// Check if task exists
	_, ok := s.optimizerManager.GetTask(taskID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if err := s.optimizerManager.CancelTask(taskID); err != nil {
		SafeBadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task cancelled"})
}

// handleApplyOptimizationResult applies result to strategy
func (s *Server) handleApplyOptimizationResult(c *gin.Context) {
	taskID := c.Param("id")
	task, ok := s.optimizerManager.GetTask(taskID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if task.Status != "completed" || task.Result == nil {
		SafeBadRequest(c, "Optimization not completed or no result")
		return
	}

	// Create new strategy with best config
	newConfig := task.Result.BestConfig
	
	// Serialize config
	configBytes, err := json.Marshal(newConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize config"})
		return
	}

	// Create Strategy in DB
	strategy := &store.Strategy{
		ID:          "opt-" + time.Now().Format("20060102-150405"), // Generate ID
		UserID:      "system", // Or get from context if available
		Name:        "Optimized " + time.Now().Format("0102-1504"),
		Description: "Generated by GA Optimizer",
		Config:      string(configBytes),
		IsActive:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	if err := s.store.Strategy().Create(strategy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save strategy"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Strategy saved",
		"strategy_id": strategy.ID,
	})
}
