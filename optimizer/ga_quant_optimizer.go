package optimizer

import (
	"math/rand"
	"sort"
	"time"
)

// Optimizer handles the genetic algorithm execution
type Optimizer struct {
	config   GAConfig
	rnd      *rand.Rand
	stopChan chan struct{}
}

// NewOptimizer creates a new optimizer instance
func NewOptimizer(cfg GAConfig) *Optimizer {
	seed := cfg.RandomSeed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	return &Optimizer{
		config: cfg,
		rnd:    rand.New(rand.NewSource(seed)),
	}
}

// SetStopChan sets the channel to signal cancellation
func (o *Optimizer) SetStopChan(stopChan chan struct{}) {
	o.stopChan = stopChan
}

// Run executes the optimization process
func (o *Optimizer) Run(bt Backtester) Chromosome {
	// 0. Initialize population
	pop := o.initPopulation()

	// Initial evaluation
	o.evaluatePopulation(pop, bt)

	// Report initial progress
	if o.config.OnProgress != nil {
		o.config.OnProgress(0, pop[0].Fitness, pop[0])
	}

	for gen := 0; gen < o.config.Generations; gen++ {
		// Check cancellation
		select {
		case <-o.stopChan:
			return pop[0] // Return current best
		default:
		}

		// 1. Elitism: Keep the best
		newPop := make([]Chromosome, 0, o.config.PopulationSize)
		// Ensure elite size doesn't exceed population
		eliteCount := o.config.EliteSize
		if eliteCount > len(pop) {
			eliteCount = len(pop)
		}
		
		for i := 0; i < eliteCount; i++ {
			newPop = append(newPop, o.cloneChromosome(pop[i]))
		}

		// Calculate dynamic mutation rate based on progress
		// Decrease mutation rate as generations progress to fine-tune
		progress := float64(gen) / float64(o.config.Generations)
		currentMutationRate := o.config.MutationRate * (1.0 - 0.5*progress) // Reduce up to 50%

		// 2. Breed new generation
		for len(newPop) < o.config.PopulationSize {
			parent1 := o.tournament(pop)
			parent2 := o.tournament(pop)

			child := o.crossover(parent1, parent2)
			o.mutate(&child, currentMutationRate)
			
			newPop = append(newPop, child)
		}

		// 3. Evaluate new population
		o.evaluatePopulation(newPop, bt)
		pop = newPop

		// Report progress
		if o.config.OnProgress != nil {
			o.config.OnProgress(gen+1, pop[0].Fitness, pop[0])
		}
	}

	return pop[0]
}

// initPopulation creates initial random population
func (o *Optimizer) initPopulation() []Chromosome {
	pop := make([]Chromosome, o.config.PopulationSize)
	for i := range pop {
		pop[i] = o.randomChromosome()
	}
	return pop
}

// randomChromosome generates a random chromosome based on the schema
func (o *Optimizer) randomChromosome() Chromosome {
	genes := make(map[string]float64)
	for _, def := range o.config.Schema {
		val := 0.0
		if def.Type == GeneTypeInt {
			// Integer range [min, max]
			min := int(def.Min)
			max := int(def.Max)
			val = float64(min + o.rnd.Intn(max-min+1))
		} else {
			// Float range [min, max]
			val = def.Min + o.rnd.Float64()*(def.Max-def.Min)
			if def.Step > 0 {
				// Snap to step
				steps := (val - def.Min) / def.Step
				val = def.Min + float64(int(steps+0.5))*def.Step
			}
		}
		genes[def.Name] = val
	}
	return Chromosome{Genes: genes}
}

// evaluatePopulation runs backtests for all chromosomes
func (o *Optimizer) evaluatePopulation(pop []Chromosome, bt Backtester) {
	// Parallel evaluation could be implemented here
	// For now, sequential to avoid resource contention
	for i := range pop {
		// Skip if already evaluated (fitness > 0 or specific flag)
		// Here we assume 0 is not evaluated (risk: what if profit is 0?)
		// Better to use a flag, but for now re-evaluating is safer
		
		res := bt.RunStrategy(pop[i])
		pop[i].Fitness = o.calculateFitness(res)
	}
	
	// Sort by fitness descending
	sort.Slice(pop, func(i, j int) bool {
		return pop[i].Fitness > pop[j].Fitness
	})
}

// calculateFitness computes fitness score based on target metric
func (o *Optimizer) calculateFitness(res BacktestResult) float64 {
	// Handle errors
	if res.Error != "" {
		return -1000000 // Severe penalty
	}

	switch o.config.TargetMetric {
	case "sharpe":
		// Calculate Sharpe from returns
		sharpe := CalculateSharpeRatio(res.Returns)
		return sharpe
	case "drawdown":
		// Minimize drawdown (maximize negative)
		return -res.MaxDD
	case "win_rate":
		return res.WinRate
	case "profit":
		fallthrough
	default:
		return res.Profit
	}
}

// tournament selection
func (o *Optimizer) tournament(pop []Chromosome) Chromosome {
	// Optimization: pre-generate indices if tournament size is large
	// For typical small size (2-5), direct Intn calls are fine.
	
	idx := o.rnd.Intn(len(pop))
	best := pop[idx]
	
	for i := 0; i < o.config.TournamentSize; i++ {
		idx = o.rnd.Intn(len(pop))
		candidate := pop[idx]
		if candidate.Fitness > best.Fitness {
			best = candidate
		}
	}
	return o.cloneChromosome(best)
}

// crossover combines two parents to create a child
func (o *Optimizer) crossover(p1, p2 Chromosome) Chromosome {
	childGenes := make(map[string]float64)
	
	// Uniform crossover
	for name, val1 := range p1.Genes {
		if o.rnd.Float64() < 0.5 {
			childGenes[name] = val1
		} else {
			childGenes[name] = p2.Genes[name]
		}
	}
	
	return Chromosome{Genes: childGenes}
}

// mutate modifies the chromosome with dynamic rate
func (o *Optimizer) mutate(c *Chromosome, rate float64) {
	for _, def := range o.config.Schema {
		if o.rnd.Float64() < rate {
			// Mutate this gene
			// Simple mutation: re-randomize
			// Advanced: small perturbation (gaussian)
			
			// Let's use Gaussian perturbation for floats to fine-tune
			if def.Type == GeneTypeFloat {
				// 10% standard deviation range
				stdDev := (def.Max - def.Min) * 0.1
				change := o.rnd.NormFloat64() * stdDev
				newVal := c.Genes[def.Name] + change
				
				// Clamp
				if newVal < def.Min { newVal = def.Min }
				if newVal > def.Max { newVal = def.Max }
				
				if def.Step > 0 {
					steps := (newVal - def.Min) / def.Step
					newVal = def.Min + float64(int(steps+0.5))*def.Step
				}
				c.Genes[def.Name] = newVal
			} else {
				// Integer: re-roll or +/- step
				min := int(def.Min)
				max := int(def.Max)
				c.Genes[def.Name] = float64(min + o.rnd.Intn(max-min+1))
			}
		}
	}
}

func (o *Optimizer) cloneChromosome(c Chromosome) Chromosome {
	genes := make(map[string]float64)
	for k, v := range c.Genes {
		genes[k] = v
	}
	return Chromosome{
		Genes:   genes,
		Fitness: c.Fitness,
	}
}
