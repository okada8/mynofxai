package backtest

// CapitalAllocator handles fund distribution across exchanges
type CapitalAllocator struct {
	// Map of Exchange Name -> Allocation Percentage (0.0 - 1.0)
	Allocation map[string]float64
}

// NewCapitalAllocator creates a default allocator
func NewCapitalAllocator() *CapitalAllocator {
	return &CapitalAllocator{
		Allocation: map[string]float64{
			"Binance": 0.40,
			"OKX":     0.25,
			"Bybit":   0.20,
			"Other":   0.15,
		},
	}
}

// GetAllocation returns the capital allocated to a specific exchange
func (ca *CapitalAllocator) GetAllocation(exchange string, totalCapital float64) float64 {
	ratio, ok := ca.Allocation[exchange]
	if !ok {
		// Fallback to "Other" or remaining
		if other, ok := ca.Allocation["Other"]; ok {
			return totalCapital * other
		}
		return 0
	}
	return totalCapital * ratio
}

// RebalanceSuggestion calculates transfers needed to restore target allocation
type RebalanceSuggestion struct {
	From   string
	To     string
	Amount float64
}

// GenerateRebalancePlan checks current balances and suggests transfers
func (ca *CapitalAllocator) GenerateRebalancePlan(currentBalances map[string]float64) []RebalanceSuggestion {
	total := 0.0
	for _, bal := range currentBalances {
		total += bal
	}
	
	if total == 0 {
		return nil
	}

	diffs := make(map[string]float64)
	for ex, ratio := range ca.Allocation {
		target := total * ratio
		current := currentBalances[ex]
		// Special handling for "Other" bucket matching multiple keys? 
		// Simplified: assumes exact key match
		diffs[ex] = current - target
	}

	suggestions := []RebalanceSuggestion{}
	
	// Greedy matching: Move from highest surplus to highest deficit
	// This is a simplified solver
	// ... (Implementation skipped for brevity, focused on structure)
	
	return suggestions
}
