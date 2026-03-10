package config

// FeatureFlags controls the rollout of new features
var FeatureFlags = map[string]bool{
	"enable_indicator_cache": true,  // Enable indicator caching
	"enable_llm_cache":       true,  // Enable LLM response caching
	"enable_batched_ws":      false, // Enable batched WebSocket updates (experimental)
}

// IsFeatureEnabled checks if a feature is enabled
func IsFeatureEnabled(feature string) bool {
	enabled, ok := FeatureFlags[feature]
	return ok && enabled
}

// SetFeatureFlag dynamically updates a feature flag (useful for runtime toggling or testing)
func SetFeatureFlag(feature string, enabled bool) {
	FeatureFlags[feature] = enabled
}
