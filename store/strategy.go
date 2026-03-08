package store

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"gorm.io/gorm"
)

// StrategyStore strategy storage
type StrategyStore struct {
	db *gorm.DB
}

// Strategy strategy configuration
type Strategy struct {
	ID            string    `gorm:"primaryKey" json:"id"`
	UserID        string    `gorm:"column:user_id;not null;default:'';index" json:"user_id"`
	Name          string    `gorm:"not null" json:"name"`
	Description   string    `gorm:"default:''" json:"description"`
	IsActive      bool      `gorm:"column:is_active;default:false;index" json:"is_active"`
	IsDefault     bool      `gorm:"column:is_default;default:false" json:"is_default"`
	IsPublic      bool      `gorm:"column:is_public;default:false;index" json:"is_public"`    // whether visible in strategy market
	ConfigVisible bool      `gorm:"column:config_visible;default:true" json:"config_visible"` // whether config details are visible
	Config        string    `gorm:"not null;default:'{}'" json:"config"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (Strategy) TableName() string { return "strategies" }

// StrategyConfig strategy configuration details (JSON structure)
type StrategyConfig struct {
	// Strategy type: "ai_trading" (default) or "grid_trading"
	StrategyType string `json:"strategy_type,omitempty"`

	// language setting: "zh" for Chinese, "en" for English
	// This determines the language used for data formatting and prompt generation
	Language string `json:"language,omitempty"`
	// coin source configuration
	CoinSource CoinSourceConfig `json:"coin_source"`
	// quantitative data configuration
	Indicators IndicatorConfig `json:"indicators"`
	// Alpha factors configuration (new in nofx 2.0)
	AlphaFactors *AlphaFactorConfig `json:"alpha_factors,omitempty"`

	// custom prompt (appended at the end)
	CustomPrompt string `json:"custom_prompt,omitempty"`
	// risk control configuration
	RiskControl RiskControlConfig `json:"risk_control"`
	// Enhanced risk control (new in nofx 2.0)
	RiskControlEnhanced *RiskControlEnhanced `json:"risk_control_enhanced,omitempty"`

	// editable sections of System Prompt
	PromptSections PromptSectionsConfig `json:"prompt_sections,omitempty"`
	// volatility configuration
	VolatilityConfig *VolatilityConfig `json:"volatility_config,omitempty"`

	// Enable timeframe prediction
	EnableTimeframePrediction bool `json:"enable_timeframe_prediction"`

	// Trailing Stop Configuration
	TrailingStopConfig *TrailingStopConfig `json:"trailing_stop_config,omitempty"`

	// Grid trading configuration (only used when StrategyType == "grid_trading")
	GridConfig *GridStrategyConfig `json:"grid_config,omitempty"`

	// Multi-agent configuration (new in nofx 2.0)
	MultiAgent *MultiAgentConfig `json:"multi_agent,omitempty"`
}

// MultiAgentConfig multi-agent configuration
type MultiAgentConfig struct {
	Enabled         bool                   `json:"enabled"`
	Agents          map[string]AgentConfig `json:"agents"`
	VotingMechanism string                 `json:"voting_mechanism"`  // "majority", "weighted", "veto"
	MinAgreementPct float64                `json:"min_agreement_pct"` // minimum agreement percentage

	// Genetic evolution
	EnableGeneticEvolution bool            `json:"enable_genetic_evolution"`
	EvolutionConfig        EvolutionConfig `json:"evolution_config"`
}

// AgentConfig agent configuration
type AgentConfig struct {
	Enabled          bool     `json:"enabled"`
	Weight           float64  `json:"weight"`            // voting weight
	PromptTemplate   string   `json:"prompt_template"`   // agent-specific prompt
	ModelOverride    string   `json:"model_override"`    // optional model override
	DataRequirements []string `json:"data_requirements"` // required data types
}

// EvolutionConfig evolution configuration
type EvolutionConfig struct {
	PopulationSize int     `json:"population_size"`
	MaxGenerations int     `json:"max_generations"`
	MutationRate   float64 `json:"mutation_rate"`
	CrossoverRate  float64 `json:"crossover_rate"`
	EliteCount     int     `json:"elite_count"` // number of elites to keep per generation
}

// TrailingStopConfig trailing stop configuration
type TrailingStopConfig struct {
	Enabled     bool              `json:"enabled"`
	PeakDecline PeakDeclineConfig `json:"peak_decline"`
}

// PeakDeclineConfig peak decline configuration
type PeakDeclineConfig struct {
	Enabled       bool    `json:"enabled"`
	ActivationPct float64 `json:"activation_pct"`
}

// VolatilityConfig volatility configuration
type VolatilityConfig struct {
	LookbackMinutes    int      `json:"lookback_minutes"`
	HighThresholdPct   float64  `json:"high_threshold_pct"`
	MediumThresholdPct float64  `json:"medium_threshold_pct"`
	BaseSymbols        []string `json:"base_symbols"`
}

// GridStrategyConfig grid trading specific configuration
type GridStrategyConfig struct {
	// Trading pair (e.g., "BTCUSDT")
	Symbol string `json:"symbol"`
	// Number of grid levels (5-50)
	GridCount int `json:"grid_count"`
	// Total investment in USDT
	TotalInvestment float64 `json:"total_investment"`
	// Leverage (1-20)
	Leverage int `json:"leverage"`
	// Upper price boundary (0 = auto-calculate from ATR)
	UpperPrice float64 `json:"upper_price"`
	// Lower price boundary (0 = auto-calculate from ATR)
	LowerPrice float64 `json:"lower_price"`
	// Use ATR to auto-calculate bounds
	UseATRBounds bool `json:"use_atr_bounds"`
	// ATR multiplier for bound calculation (default 2.0)
	ATRMultiplier float64 `json:"atr_multiplier"`
	// Position distribution: "uniform" | "gaussian" | "pyramid"
	Distribution string `json:"distribution"`
	// Maximum drawdown percentage before emergency exit
	MaxDrawdownPct float64 `json:"max_drawdown_pct"`
	// Stop loss percentage per position
	StopLossPct float64 `json:"stop_loss_pct"`
	// Daily loss limit percentage
	DailyLossLimitPct float64 `json:"daily_loss_limit_pct"`
	// Use maker-only orders for lower fees
	UseMakerOnly bool `json:"use_maker_only"`
	// Enable automatic grid direction adjustment based on box breakouts
	EnableDirectionAdjust bool `json:"enable_direction_adjust"`
	// Direction bias ratio for long_bias/short_bias modes (default 0.7 = 70%/30%)
	DirectionBiasRatio float64 `json:"direction_bias_ratio"`
}

// PromptSectionsConfig editable sections of System Prompt
type PromptSectionsConfig struct {
	// role definition (title + description)
	RoleDefinition string `json:"role_definition,omitempty"`
	// trading frequency awareness
	TradingFrequency string `json:"trading_frequency,omitempty"`
	// entry standards
	EntryStandards string `json:"entry_standards,omitempty"`
	// decision process
	DecisionProcess string `json:"decision_process,omitempty"`
}

// CoinSourceConfig coin source configuration
type CoinSourceConfig struct {
	// source type: "static" | "ai500" | "oi_top" | "oi_low" | "mixed"
	SourceType string `json:"source_type"`
	// static coin list (used when source_type = "static")
	StaticCoins []string `json:"static_coins,omitempty"`
	// excluded coins list (filtered out from all sources)
	ExcludedCoins []string `json:"excluded_coins,omitempty"`
	// whether to use AI500 coin pool
	UseAI500 bool `json:"use_ai500"`
	// AI500 coin pool maximum count
	AI500Limit int `json:"ai500_limit,omitempty"`
	// whether to use OI Top (持仓增加榜，适合做多)
	UseOITop bool `json:"use_oi_top"`
	// OI Top maximum count
	OITopLimit int `json:"oi_top_limit,omitempty"`
	// whether to use OI Low (持仓减少榜，适合做空)
	UseOILow bool `json:"use_oi_low"`
	// OI Low maximum count
	OILowLimit int `json:"oi_low_limit,omitempty"`
	// whether to use Visual Screener (CoinAnk)
	UseScreener bool `json:"use_screener"`
	// Visual Screener maximum count
	ScreenerLimit int `json:"screener_limit,omitempty"`
	// Visual Screener duration (e.g., "1h", "4h")
	ScreenerDuration string `json:"screener_duration,omitempty"`
	// Visual Screener sort by ("oi", "price", "vol")
	ScreenerSortBy string `json:"screener_sort_by,omitempty"`
	// whether to use Hyperliquid All coins (all available perp pairs)
	UseHyperAll bool `json:"use_hyper_all"`
	// whether to use Hyperliquid Main coins (top N by 24h volume)
	UseHyperMain bool `json:"use_hyper_main"`
	// Hyperliquid Main maximum count (default 20)
	HyperMainLimit int `json:"hyper_main_limit,omitempty"`
	// Note: API URLs are now built automatically using NofxOSAPIKey from IndicatorConfig

	// Gainers/Losers configuration
	UseGainersLosers   bool `json:"use_gainers_losers"`
	GainersTop         int  `json:"gainers_top,omitempty"`
	LosersTop          int  `json:"losers_top,omitempty"`
	OnlyBinanceSymbols bool `json:"only_binance_symbols"`
}

// IndicatorConfig indicator configuration
type IndicatorConfig struct {
	// K-line configuration
	Klines KlineConfig `json:"klines"`
	// raw kline data (OHLCV) - always enabled, required for AI analysis
	EnableRawKlines bool `json:"enable_raw_klines"`
	// technical indicator switches
	EnableEMA         bool `json:"enable_ema"`
	EnableMACD        bool `json:"enable_macd"`
	EnableRSI         bool `json:"enable_rsi"`
	EnableATR         bool `json:"enable_atr"`
	EnableBOLL        bool `json:"enable_boll"` // Bollinger Bands
	EnableVolume      bool `json:"enable_volume"`
	EnableOI          bool `json:"enable_oi"`           // open interest
	EnableFundingRate bool `json:"enable_funding_rate"` // funding rate
	// Advanced indicators
	EnableADX             bool `json:"enable_adx"`
	EnableVWAP            bool `json:"enable_vwap"`
	EnableOBV             bool `json:"enable_obv"`
	EnableCMF             bool `json:"enable_cmf"`
	EnableLiquidation     bool `json:"enable_liquidation"`
	EnableCVD             bool `json:"enable_cvd"`
	EnableNetPosition     bool `json:"enable_net_position"`
	EnableSignalLabels    bool `json:"enable_signal_labels"`
	EnableIndicatorSeries bool `json:"enable_indicator_series"`

	// EMA period configuration
	EMAPeriods []int `json:"ema_periods,omitempty"` // default [20, 50]
	// RSI period configuration
	RSIPeriods []int `json:"rsi_periods,omitempty"` // default [7, 14]
	// ATR period configuration
	ATRPeriods []int `json:"atr_periods,omitempty"` // default [14]
	// BOLL period configuration (period, standard deviation multiplier is fixed at 2)
	BOLLPeriods []int `json:"boll_periods,omitempty"` // default [20] - can select multiple timeframes
	// external data sources
	ExternalDataSources []ExternalDataSource `json:"external_data_sources,omitempty"`

	// Donchian Channel
	EnableDonchianBox bool  `json:"enable_donchian_box"`
	EnableBoxPrefetch bool  `json:"enable_box_prefetch"`
	DonchianPeriods   []int `json:"donchian_periods,omitempty"`

	// Score System
	EnableHeatScore                bool                  `json:"enable_heat_score"`
	EnableVolatilityUtilScore      bool                  `json:"enable_volatility_util_score"`
	EnableVolumeSpikeScore         bool                  `json:"enable_volume_spike_score"`
	EnableFundingRateScore         bool                  `json:"enable_funding_rate_score"`
	EnableOrderbookImbalanceScore  bool                  `json:"enable_orderbook_imbalance_score"`
	EnableAdaptiveCompositeWeights bool                  `json:"enable_adaptive_composite_weights"`
	WeightHeat                     float64               `json:"weight_heat"`
	WeightVolatilityUtil           float64               `json:"weight_volatility_util"`
	WeightVolumeSpike              float64               `json:"weight_volume_spike"`
	WeightFundingRate              float64               `json:"weight_funding_rate"`
	WeightOrderbookImbalance       float64               `json:"weight_orderbook_imbalance"`
	GroupedWeights                 *GroupedWeightsConfig `json:"grouped_weights,omitempty"`

	// ========== NofxOS Unified API Configuration ==========
	// Unified API Key for all NofxOS data sources
	NofxOSAPIKey string `json:"nofxos_api_key,omitempty"`

	// quantitative data sources (capital flow, position changes, price changes)
	EnableQuantData    bool `json:"enable_quant_data"`    // whether to enable quantitative data
	EnableQuantOI      bool `json:"enable_quant_oi"`      // whether to show OI data
	EnableQuantNetflow bool `json:"enable_quant_netflow"` // whether to show Netflow data

	// OI ranking data (market-wide open interest increase/decrease rankings)
	EnableOIRanking   bool   `json:"enable_oi_ranking"`             // whether to enable OI ranking data
	OIRankingDuration string `json:"oi_ranking_duration,omitempty"` // duration: 1h, 4h, 24h
	OIRankingLimit    int    `json:"oi_ranking_limit,omitempty"`    // number of entries (default 10)

	// NetFlow ranking data (market-wide fund flow rankings - institution/personal)
	EnableNetFlowRanking   bool   `json:"enable_netflow_ranking"`             // whether to enable NetFlow ranking data
	NetFlowRankingDuration string `json:"netflow_ranking_duration,omitempty"` // duration: 1h, 4h, 24h
	NetFlowRankingLimit    int    `json:"netflow_ranking_limit,omitempty"`    // number of entries (default 10)

	// Price ranking data (market-wide gainers/losers)
	EnablePriceRanking   bool   `json:"enable_price_ranking"`             // whether to enable price ranking data
	PriceRankingDuration string `json:"price_ranking_duration,omitempty"` // durations: "1h" or "1h,4h,24h"
	PriceRankingLimit    int    `json:"price_ranking_limit,omitempty"`    // number of entries per ranking (default 10)
}

// GroupedWeightsConfig grouped weights configuration
type GroupedWeightsConfig struct {
	Scalp    WeightsConfig `json:"scalp"`
	Intraday WeightsConfig `json:"intraday"`
	Swing    WeightsConfig `json:"swing"`
	Position WeightsConfig `json:"position"`
}

// WeightsConfig weights configuration
type WeightsConfig struct {
	EnableHeat               bool    `json:"enable_heat"`
	EnableVolatilityUtil     bool    `json:"enable_volatility_util"`
	EnableVolumeSpike        bool    `json:"enable_volume_spike"`
	EnableFundingRate        bool    `json:"enable_funding_rate"`
	EnableOrderbookImbalance bool    `json:"enable_orderbook_imbalance"`
	WeightHeat               float64 `json:"weight_heat"`
	WeightVolatilityUtil     float64 `json:"weight_volatility_util"`
	WeightVolumeSpike        float64 `json:"weight_volume_spike"`
	WeightFundingRate        float64 `json:"weight_funding_rate"`
	WeightOrderbookImbalance float64 `json:"weight_orderbook_imbalance"`
}

// RiskControlEnhanced enhanced risk control configuration
type RiskControlEnhanced struct {
	// Emergency risk control fields
	DailyLossLimitPct        float64 `json:"daily_loss_limit_pct"`
	StrategyDrawdownLimitPct float64 `json:"strategy_drawdown_limit_pct"`
	MaxRiskPerTradePct       float64 `json:"max_risk_per_trade_pct"`
	AutoDisableOnLoss        bool    `json:"auto_disable_on_loss"`
	DailyLossThresholdPct    float64 `json:"daily_loss_threshold_pct"`
	RiskPerTradeMode         string  `json:"risk_per_trade_mode"` // "fixed" or "dynamic"

	// Dynamic risk adjustment
	DynamicRiskAdjustment DynamicRiskConfig `json:"dynamic_risk_adjustment"`
}

// DynamicRiskConfig dynamic risk configuration
type DynamicRiskConfig struct {
	Enabled              bool                      `json:"enabled"`
	MarketRegimeMapping  map[string]RiskAdjustment `json:"market_regime_mapping"`
	ConfidenceAdjustment ConfidenceRiskAdjustment  `json:"confidence_adjustment"`
	VolatilityAdjustment VolatilityRiskAdjustment  `json:"volatility_adjustment"`
}

// RiskAdjustment risk adjustment parameters
type RiskAdjustment struct {
	RiskMultiplier  float64 `json:"risk_multiplier"`   // 0.5-2.0
	PositionSizePct float64 `json:"position_size_pct"` // position size adjustment
	MaxLeverage     int     `json:"max_leverage"`      // leverage adjustment
}

// ConfidenceRiskAdjustment confidence-based risk adjustment
type ConfidenceRiskAdjustment struct {
	HighConfidence   RiskAdjustment `json:"high_confidence"`   // confidence > 85%
	MediumConfidence RiskAdjustment `json:"medium_confidence"` // 70-85%
	LowConfidence    RiskAdjustment `json:"low_confidence"`    // < 70%
}

// VolatilityRiskAdjustment volatility-based risk adjustment
type VolatilityRiskAdjustment struct {
	HighVolatility   RiskAdjustment `json:"high_volatility"`   // ATR > threshold
	MediumVolatility RiskAdjustment `json:"medium_volatility"` // normal range
	LowVolatility    RiskAdjustment `json:"low_volatility"`    // ATR < threshold
}

// AlphaFactorConfig alpha factor configuration
type AlphaFactorConfig struct {
	// Liquidation clusters
	EnableLiquidationClusters   bool    `json:"enable_liquidation_clusters"`
	LiquidationClusterThreshold float64 `json:"liquidation_cluster_threshold"`
	MinLiquidationUSD           float64 `json:"min_liquidation_usd"`

	// Exchange flow
	EnableExchangeFlow        bool    `json:"enable_exchange_flow"`
	ExchangeFlowLookbackHours int     `json:"exchange_flow_lookback_hours"`
	SignificantFlowThreshold  float64 `json:"significant_flow_threshold"`

	// Whale wallet tracking
	EnableWhaleWalletTracking bool     `json:"enable_whale_wallet_tracking"`
	WhaleThresholdUSD         float64  `json:"whale_threshold_usd"`
	TrackedWallets            []string `json:"tracked_wallets"`

	// Spread expansion
	EnableSpreadExpansion    bool    `json:"enable_spread_expansion"`
	SpreadExpansionThreshold float64 `json:"spread_expansion_threshold"`

	// Factor weights
	FactorWeights map[string]float64 `json:"factor_weights"`
}

// KlineConfig K-line configuration
type KlineConfig struct {
	// primary timeframe: "1m", "3m", "5m", "15m", "1h", "4h"
	PrimaryTimeframe string `json:"primary_timeframe"`
	// primary timeframe K-line count
	PrimaryCount int `json:"primary_count"`
	// longer timeframe
	LongerTimeframe string `json:"longer_timeframe,omitempty"`
	// longer timeframe K-line count
	LongerCount int `json:"longer_count,omitempty"`
	// whether to enable multi-timeframe analysis
	EnableMultiTimeframe bool `json:"enable_multi_timeframe"`
	// selected timeframe list (new: supports multi-timeframe selection)
	SelectedTimeframes []string `json:"selected_timeframes,omitempty"`
}

// ExternalDataSource external data source configuration
type ExternalDataSource struct {
	Name        string            `json:"name"`   // data source name
	Type        string            `json:"type"`   // type: "api" | "webhook"
	URL         string            `json:"url"`    // API URL
	Method      string            `json:"method"` // HTTP method
	Headers     map[string]string `json:"headers,omitempty"`
	DataPath    string            `json:"data_path,omitempty"`    // JSON data path
	RefreshSecs int               `json:"refresh_secs,omitempty"` // refresh interval (seconds)
}

// RiskControlConfig risk control configuration
type RiskControlConfig struct {
	// Max number of coins held simultaneously (CODE ENFORCED)
	MaxPositions int `json:"max_positions"`

	// BTC/ETH exchange leverage for opening positions (AI guided)
	BTCETHMaxLeverage int `json:"btc_eth_max_leverage"`
	// Altcoin exchange leverage for opening positions (AI guided)
	AltcoinMaxLeverage int `json:"altcoin_max_leverage"`

	// BTC/ETH single position max value = equity × this ratio (CODE ENFORCED, default: 5)
	BTCETHMaxPositionValueRatio float64 `json:"btc_eth_max_position_value_ratio"`
	// Altcoin single position max value = equity × this ratio (CODE ENFORCED, default: 1)
	AltcoinMaxPositionValueRatio float64 `json:"altcoin_max_position_value_ratio"`

	// Max margin utilization (e.g. 0.9 = 90%) (CODE ENFORCED)
	MaxMarginUsage float64 `json:"max_margin_usage"`
	// Min position size in USDT (CODE ENFORCED)
	MinPositionSize float64 `json:"min_position_size"`

	// Min take_profit / stop_loss ratio (AI guided)
	MinRiskRewardRatio float64 `json:"min_risk_reward_ratio"`
	// Min AI confidence to open position (AI guided)
	MinConfidence int `json:"min_confidence"`

	// Execution & Orders
	EnforceStopLossTakeProfit   bool    `json:"enforce_stop_loss_take_profit"`
	ManualStopLossPct           float64 `json:"manual_stop_loss_pct,omitempty"`
	ManualTakeProfitPct         float64 `json:"manual_take_profit_pct,omitempty"`
	LimitOrderTimeoutMin        int     `json:"limit_order_timeout_min,omitempty"`
	MinLimitOrderDistancePct    float64 `json:"min_limit_order_distance_pct,omitempty"`
	ReverseDirection            bool    `json:"reverse_direction"`
	MaxNewEntriesPerDirection   int     `json:"max_new_entries_per_direction,omitempty"`
	RequireBreakoutCloseConfirm bool    `json:"require_breakout_close_confirm"`
	BreakoutMinATRFraction      float64 `json:"breakout_min_atr_fraction,omitempty"`

	// Position Sizing
	PositionSizeMode   string  `json:"position_size_mode,omitempty"`
	FixedPositionUSD   float64 `json:"fixed_position_usd,omitempty"`
	EquityPct          float64 `json:"equity_pct,omitempty"`
	MaxRiskPerTradePct float64 `json:"max_risk_per_trade_pct,omitempty"`

	// ATR Multiples
	SLATRMult float64 `json:"sl_atr_mult,omitempty"`
	TPATRMult float64 `json:"tp_atr_mult,omitempty"`

	// Drawdown & Cooldown
	DrawdownTPActivation  float64 `json:"drawdown_tp_activation,omitempty"`
	DrawdownTPThreshold   float64 `json:"drawdown_tp_threshold,omitempty"`
	SymbolCooldownMinutes int     `json:"symbol_cooldown_minutes,omitempty"`
	TimeoutExitMinutes    int     `json:"timeout_exit_minutes,omitempty"`
	TimeoutMinProgressPct float64 `json:"timeout_min_progress_pct,omitempty"`

	// Dynamic Position Control
	MaxPositionsMin         int `json:"max_positions_min,omitempty"`
	MaxPositionsMax         int `json:"max_positions_max,omitempty"`
	ScanIntervalBaseMinutes int `json:"scan_interval_base_minutes,omitempty"`
	ScanIntervalBaseMin     int `json:"scan_interval_base_min,omitempty"`
	ScanIntervalBaseMax     int `json:"scan_interval_base_max,omitempty"`

	// Holding & Exit
	MinHoldMinutes               int     `json:"min_hold_minutes,omitempty"`
	MinCloseProfitPct            float64 `json:"min_close_profit_pct,omitempty"`
	CloseWhenProfitExceedsPct    float64 `json:"close_when_profit_exceeds_pct,omitempty"`
	CloseWhenDrawdownFromPeakPct float64 `json:"close_when_drawdown_from_peak_pct,omitempty"`

	// ATR Risk
	EnableATRRisk               bool    `json:"enable_atr_risk"`
	ATRPeriod                   int     `json:"atr_period,omitempty"`
	ATRMultiplier               float64 `json:"atr_multiplier,omitempty"`
	EnableAdaptiveATRMultiplier bool    `json:"enable_adaptive_atr_multiplier"`
	ATRMultiplierMin            float64 `json:"atr_multiplier_min,omitempty"`
	ATRMultiplierMax            float64 `json:"atr_multiplier_max,omitempty"`

	// Dynamic Stop Loss & Risk
	DefaultStopLossMinPct     float64 `json:"default_stop_loss_min_pct,omitempty"`
	EnableDynamicStopLossMin  bool    `json:"enable_dynamic_stop_loss_min"`
	StopLossMinIncreaseMaxPct float64 `json:"stop_loss_min_increase_max_pct,omitempty"`
	RiskPerTradePct           float64 `json:"risk_per_trade_pct,omitempty"`
	EnableDynamicRiskPerTrade bool    `json:"enable_dynamic_risk_per_trade"`
	RiskPerTradeMinPct        float64 `json:"risk_per_trade_min_pct,omitempty"`
	RiskPerTradeMaxPct        float64 `json:"risk_per_trade_max_pct,omitempty"`
	RiskPerTradeRecentWindow  int     `json:"risk_per_trade_recent_window,omitempty"`

	// Staged Take Profit
	EnableStagedTakeProfit bool    `json:"enable_staged_take_profit"`
	Stage1ProfitPct        float64 `json:"stage1_profit_pct,omitempty"`
	Stage1CloseRatio       float64 `json:"stage1_close_ratio,omitempty"`
	Stage2ProfitPct        float64 `json:"stage2_profit_pct,omitempty"`
	Stage2CloseRatio       float64 `json:"stage2_close_ratio,omitempty"`
	Stage3ProfitPct        float64 `json:"stage3_profit_pct,omitempty"`
	Stage3CloseRatio       float64 `json:"stage3_close_ratio,omitempty"`
	Stage4ProfitPct        float64 `json:"stage4_profit_pct,omitempty"`
	Stage4CloseRatio       float64 `json:"stage4_close_ratio,omitempty"`

	// Sideways Handling
	EnableSidewaysTimeDecayClose                 bool    `json:"enable_sideways_time_decay_close"`
	EnableSidewaysMicroGrid                      bool    `json:"enable_sideways_micro_grid"`
	SidewaysBandPct                              float64 `json:"sideways_band_pct,omitempty"`
	SidewaysMinDurationMin                       int     `json:"sideways_min_duration_min,omitempty"`
	SidewaysCloseProfitPct                       float64 `json:"sideways_close_profit_pct,omitempty"`
	SidewaysCloseRatio                           float64 `json:"sideways_close_ratio,omitempty"`
	UseSidewaysRatioThreshold                    bool    `json:"use_sideways_ratio_threshold"`
	SidewaysRatioMin                             float64 `json:"sideways_ratio_min,omitempty"`
	RequireIndicesDeteriorationForTimeDecayClose bool    `json:"require_indices_deterioration_for_time_decay_close"`
	TimeDecayDeteriorationMinSignals             int     `json:"time_decay_deterioration_min_signals,omitempty"`
	SidewaysBandLowerCoeff                       float64 `json:"sideways_band_lower_coeff,omitempty"`
	SidewaysBandUpperCoeff                       float64 `json:"sideways_band_upper_coeff,omitempty"`
	SidewaysHeatWeightedRatioThreshold           float64 `json:"sideways_heat_weighted_ratio_threshold,omitempty"`

	// Trend Stop Loss
	EnableTrendStopLoss         bool    `json:"enable_trend_stop_loss"`
	TrendStopLossMinSignals     int     `json:"trend_stop_loss_min_signals,omitempty"`
	TrendStopLossTriggerLossPct float64 `json:"trend_stop_loss_trigger_loss_pct,omitempty"`
	TrendStopLossMinHoldMinutes int     `json:"trend_stop_loss_min_hold_minutes,omitempty"`

	// Execution & Monitoring
	TakeProfitMonitorIntervalSec int `json:"take_profit_monitor_interval_sec,omitempty"`
	LossThrottleSec              int `json:"loss_throttle_sec,omitempty"`
	UnplacedTTLMs                int `json:"unplaced_ttl_ms,omitempty"`
	UnplacedTTLMsMin             int `json:"unplaced_ttl_ms_min,omitempty"`
	UnplacedTTLMsMax             int `json:"unplaced_ttl_ms_max,omitempty"`

	// Trailing Stop
	TrailingStopMode              string  `json:"trailing_stop_mode,omitempty"`
	TrailingStopCallbackRatePct   float64 `json:"trailing_stop_callback_rate_pct,omitempty"`
	TPSLOrderUpdateCooldownSec    int     `json:"tp_sl_order_update_cooldown_sec,omitempty"`
	TrailingStopAppMinIntervalSec int     `json:"trailing_stop_app_min_interval_sec,omitempty"`

	// Dynamic Take Profit
	EnableDynamicTakeProfit    bool              `json:"enable_dynamic_take_profit"`
	EnableROETPLadder          bool              `json:"enable_roe_tp_ladder"`
	StagedActivationPeakROEPct float64           `json:"staged_activation_peak_roe_pct,omitempty"`
	DynamicTPLadder            []DynamicTPConfig `json:"dynamic_tp_ladder,omitempty"`

	// Evolution & Factors
	EnableFullAutoEvolution        bool            `json:"enable_full_auto_evolution"`
	EnableAutoTuneThresholds       bool            `json:"enable_auto_tune_thresholds"`
	AutoTuneLookbackMinutes        int             `json:"auto_tune_lookback_minutes,omitempty"`
	LastAutoTuneMs                 int64           `json:"last_auto_tune_ms,omitempty"`
	EnableAIEvolution              bool            `json:"enable_ai_evolution"`
	EvolutionMode                  string          `json:"evolution_mode,omitempty"`
	AIEvolutionMinTrades           int             `json:"ai_evolution_min_trades,omitempty"`
	ReliabilityFloor               float64         `json:"reliability_floor,omitempty"`
	DirectionQuantileLevel         float64         `json:"direction_quantile_level,omitempty"`
	DirectionPoolIncludePositions  bool            `json:"direction_pool_include_positions"`
	DirectionVolumePriceDivergence bool            `json:"direction_volume_price_divergence"`
	ThresholdVolumeSpikeNormGreen  float64         `json:"threshold_volume_spike_norm_green,omitempty"`
	ThresholdFundingRatePosPct     float64         `json:"threshold_funding_rate_pos_pct,omitempty"`
	ThresholdOBImbalancePos        float64         `json:"threshold_ob_imbalance_pos,omitempty"`
	ReliabilityFloorMin            float64         `json:"reliability_floor_min,omitempty"`
	ReliabilityFloorMax            float64         `json:"reliability_floor_max,omitempty"`
	DirectionQuantileMin           float64         `json:"direction_quantile_min,omitempty"`
	DirectionQuantileMax           float64         `json:"direction_quantile_max,omitempty"`
	VolumeSpikeThresholdMin        float64         `json:"volume_spike_threshold_min,omitempty"`
	VolumeSpikeThresholdMax        float64         `json:"volume_spike_threshold_max,omitempty"`
	OIThresholdMin                 float64         `json:"oi_threshold_min,omitempty"`
	OIThresholdMax                 float64         `json:"oi_threshold_max,omitempty"`
	MTFSmallAlignThreshold         float64         `json:"mtf_small_align_threshold,omitempty"`
	MTFMedAlignThreshold           float64         `json:"mtf_med_align_threshold,omitempty"`
	MTFHighAlignThreshold          float64         `json:"mtf_high_align_threshold,omitempty"`
	SidewaysBandWeightATR          float64         `json:"sideways_band_weight_atr,omitempty"`
	SidewaysBandWeightVol          float64         `json:"sideways_band_weight_vol,omitempty"`
	SidewaysBandWeightReliability  float64         `json:"sideways_band_weight_reliability,omitempty"`
	QuantileWeightMode             string          `json:"quantile_weight_mode,omitempty"`
	DirectionMinGap                float64         `json:"direction_min_gap,omitempty"`
	ReliabilityWeightAlign         float64         `json:"reliability_weight_align,omitempty"`
	ReliabilityWeightHeatInv       float64         `json:"reliability_weight_heat_inv,omitempty"`
	ReliabilityWeightZPenInv       float64         `json:"reliability_weight_zpen_inv,omitempty"`
	ReliabilityWeightFund          float64         `json:"reliability_weight_fund,omitempty"`
	ReliabilityWeightOBInv         float64         `json:"reliability_weight_ob_inv,omitempty"`
	DirectionCoreVoteMultiplier    float64         `json:"direction_core_vote_multiplier,omitempty"`
	AdaptiveConsistencyStep        float64         `json:"adaptive_consistency_step,omitempty"`
	WinRateTightenThreshold        float64         `json:"win_rate_tighten_threshold,omitempty"`
	MaxDrawdownTightenThresholdPct float64         `json:"max_drawdown_tighten_threshold_pct,omitempty"`
	EntryConfidenceMin             float64         `json:"entry_confidence_min,omitempty"`
	EntryWindowMinMin              int             `json:"entry_window_min_min,omitempty"`
	EntryWindowMaxMin              int             `json:"entry_window_max_min,omitempty"`
	EntryUrgencyNowMaxMin          int             `json:"entry_urgency_now_max_min,omitempty"`
	EntryUrgencySoonMaxMin         int             `json:"entry_urgency_soon_max_min,omitempty"`
	EntryRSIOverboughtSoft         float64         `json:"entry_rsi_overbought_soft,omitempty"`
	EntryRSIOverboughtHard         float64         `json:"entry_rsi_overbought_hard,omitempty"`
	EntryRSIOversoldSoft           float64         `json:"entry_rsi_oversold_soft,omitempty"`
	EntryRSIOversoldHard           float64         `json:"entry_rsi_oversold_hard,omitempty"`
	MinFactorPass                  int             `json:"min_factor_pass,omitempty"`
	MinReliabilityForSubmit        float64         `json:"min_reliability_for_submit,omitempty"`
	ThresholdHeatScoreGold         float64         `json:"threshold_heat_score_gold,omitempty"`
	ThresholdATRUtilGreenPct       float64         `json:"threshold_atr_util_green_pct,omitempty"`
	FactorLibraryEnabled           map[string]bool `json:"factor_library_enabled,omitempty"`
	FirstLayerMinPassCount         int             `json:"first_layer_min_pass_count,omitempty"`
	AllowTradeWhenAIFails          bool            `json:"allow_trade_when_ai_fails"`
}

// DynamicTPConfig dynamic take profit configuration
type DynamicTPConfig struct {
	TriggerROEPct float64 `json:"trigger_roe_pct"`
	TargetTPPct   float64 `json:"target_tp_pct"`
}

// NewStrategyStore creates a new StrategyStore
func NewStrategyStore(db *gorm.DB) *StrategyStore {
	return &StrategyStore{db: db}
}

func (s *StrategyStore) initTables() error {
	// AutoMigrate will add missing columns without dropping existing data
	return s.db.AutoMigrate(&Strategy{})
}

func (s *StrategyStore) initDefaultData() error {
	// No longer pre-populate strategies - create on demand when user configures
	return nil
}

// GetDefaultStrategyConfig returns the default strategy configuration for the given language
func GetDefaultStrategyConfig(lang string) StrategyConfig {
	// Normalize language to "zh" or "en"
	normalizedLang := "en"
	if lang == "zh" {
		normalizedLang = "zh"
	}

	config := StrategyConfig{
		Language: normalizedLang,
		CoinSource: CoinSourceConfig{
			SourceType: "ai500",
			UseAI500:   true,
			AI500Limit: 10,
			UseOITop:   false,
			OITopLimit: 10,
			UseOILow:   false,
			OILowLimit: 10,
			// Screener defaults
			UseScreener:      false,
			ScreenerLimit:    10,
			ScreenerDuration: "1h",
			ScreenerSortBy:   "oi",
		},
		Indicators: IndicatorConfig{
			Klines: KlineConfig{
				PrimaryTimeframe:     "5m",
				PrimaryCount:         30,
				LongerTimeframe:      "4h",
				LongerCount:          10,
				EnableMultiTimeframe: true,
				SelectedTimeframes:   []string{"5m", "15m", "1h", "4h"},
			},
			EnableRawKlines:   true, // Required - raw OHLCV data for AI analysis
			EnableEMA:         false,
			EnableMACD:        false,
			EnableRSI:         false,
			EnableATR:         false,
			EnableBOLL:        false,
			EnableVolume:      true,
			EnableOI:          true,
			EnableFundingRate: true,
			EMAPeriods:        []int{20, 50},
			RSIPeriods:        []int{7, 14},
			ATRPeriods:        []int{14},
			BOLLPeriods:       []int{20},
			// NofxOS unified API key
			NofxOSAPIKey: getEnvOrDefault("NOFXOS_API_KEY", "cm_568c67eae410d912c54c"),
			// Quant data
			EnableQuantData:    true,
			EnableQuantOI:      true,
			EnableQuantNetflow: true,
			// OI ranking data
			EnableOIRanking:   true,
			OIRankingDuration: "1h",
			OIRankingLimit:    10,
			// NetFlow ranking data
			EnableNetFlowRanking:   true,
			NetFlowRankingDuration: "1h",
			NetFlowRankingLimit:    10,
			// Price ranking data
			EnablePriceRanking:   true,
			PriceRankingDuration: "1h,4h,24h",
			PriceRankingLimit:    10,
		},
		EnableTimeframePrediction: true,
		TrailingStopConfig: &TrailingStopConfig{
			Enabled: true,
			PeakDecline: PeakDeclineConfig{
				Enabled:       true,
				ActivationPct: 2.0,
			},
		},
		RiskControl: RiskControlConfig{
			MaxPositions:                 3,   // Max 3 coins simultaneously (CODE ENFORCED)
			BTCETHMaxLeverage:            5,   // BTC/ETH exchange leverage (AI guided)
			AltcoinMaxLeverage:           5,   // Altcoin exchange leverage (AI guided)
			BTCETHMaxPositionValueRatio:  5.0, // BTC/ETH: max position = 5x equity (CODE ENFORCED)
			AltcoinMaxPositionValueRatio: 1.0, // Altcoin: max position = 1x equity (CODE ENFORCED)
			MaxMarginUsage:               0.9, // Max 90% margin usage (CODE ENFORCED)
			MinPositionSize:              12,  // Min 12 USDT per position (CODE ENFORCED)
			MinRiskRewardRatio:           3.0, // Min 3:1 profit/loss ratio (AI guided)
			MinConfidence:                75,  // Min 75% confidence (AI guided)
		},
	}

	if lang == "zh" {
		config.PromptSections = PromptSectionsConfig{
			RoleDefinition: `# 你是一个专业的加密货币交易AI

你的任务是根据提供的市场数据做出交易决策。你是一个经验丰富的量化交易员，擅长技术分析和风险管理。`,
			TradingFrequency: `# ⏱️ 交易频率意识

- 优秀交易员：每天2-4笔 ≈ 每小时0.1-0.2笔
- 每小时超过2笔 = 过度交易
- 单笔持仓时间 ≥ 30-60分钟
如果你发现自己每个周期都在交易 → 标准太低；如果持仓不到30分钟就平仓 → 太冲动。`,
			EntryStandards: `# 🎯 入场标准（严格）

只在多个信号共振时入场。自由使用任何有效的分析方法，避免单一指标、信号矛盾、横盘震荡、或平仓后立即重新开仓等低质量行为。`,
			DecisionProcess: `# 📋 决策流程

1. 检查持仓 → 是否止盈/止损
2. 扫描候选币种 + 多时间框架 → 是否存在强信号
3. 先写思维链，再输出结构化JSON`,
		}
	} else {
		config.PromptSections = PromptSectionsConfig{
			RoleDefinition: `# You are a professional cryptocurrency trading AI

Your task is to make trading decisions based on the provided market data. You are an experienced quantitative trader skilled in technical analysis and risk management.`,
			TradingFrequency: `# ⏱️ Trading Frequency Awareness

- Excellent trader: 2-4 trades per day ≈ 0.1-0.2 trades per hour
- >2 trades per hour = overtrading
- Single position holding time ≥ 30-60 minutes
If you find yourself trading every cycle → standards are too low; if closing positions in <30 minutes → too impulsive.`,
			EntryStandards: `# 🎯 Entry Standards (Strict)

Only enter positions when multiple signals resonate. Freely use any effective analysis methods, avoid low-quality behaviors such as single indicators, contradictory signals, sideways oscillation, or immediately restarting after closing positions.`,
			DecisionProcess: `# 📋 Decision Process

1. Check positions → whether to take profit/stop loss
2. Scan candidate coins + multi-timeframe → whether strong signals exist
3. Write chain of thought first, then output structured JSON`,
		}
	}

	return config
}

// Create create a strategy
func (s *StrategyStore) Create(strategy *Strategy) error {
	return s.db.Create(strategy).Error
}

// Update update a strategy
func (s *StrategyStore) Update(strategy *Strategy) error {
	return s.db.Model(&Strategy{}).
		Where("id = ? AND user_id = ?", strategy.ID, strategy.UserID).
		Updates(map[string]interface{}{
			"name":           strategy.Name,
			"description":    strategy.Description,
			"config":         strategy.Config,
			"is_public":      strategy.IsPublic,
			"config_visible": strategy.ConfigVisible,
			"updated_at":     time.Now().UTC(),
		}).Error
}

// Delete delete a strategy
func (s *StrategyStore) Delete(userID, id string) error {
	// do not allow deleting system default strategy
	var st Strategy
	if err := s.db.Where("id = ?", id).First(&st).Error; err == nil && st.IsDefault {
		return fmt.Errorf("cannot delete system default strategy")
	}

	return s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&Strategy{}).Error
}

// List get user's strategy list
func (s *StrategyStore) List(userID string) ([]*Strategy, error) {
	var strategies []*Strategy
	err := s.db.Where("user_id = ? OR is_default = ?", userID, true).
		Order("is_default DESC, created_at DESC").
		Find(&strategies).Error
	if err != nil {
		return nil, err
	}
	return strategies, nil
}

// ListPublic get all public strategies for the strategy market
func (s *StrategyStore) ListPublic() ([]*Strategy, error) {
	var strategies []*Strategy
	err := s.db.Where("is_public = ?", true).
		Order("created_at DESC").
		Find(&strategies).Error
	if err != nil {
		return nil, err
	}
	return strategies, nil
}

// Get get a single strategy
func (s *StrategyStore) Get(userID, id string) (*Strategy, error) {
	var st Strategy
	err := s.db.Where("id = ? AND (user_id = ? OR is_default = ?)", id, userID, true).
		First(&st).Error
	if err != nil {
		return nil, err
	}
	return &st, nil
}

// GetActive get user's currently active strategy
func (s *StrategyStore) GetActive(userID string) (*Strategy, error) {
	var st Strategy
	err := s.db.Where("user_id = ? AND is_active = ?", userID, true).First(&st).Error
	if err == gorm.ErrRecordNotFound {
		// no active strategy, return system default strategy
		return s.GetDefault()
	}
	if err != nil {
		return nil, err
	}
	return &st, nil
}

// GetDefault get system default strategy
func (s *StrategyStore) GetDefault() (*Strategy, error) {
	var st Strategy
	err := s.db.Where("is_default = ?", true).First(&st).Error
	if err != nil {
		return nil, err
	}
	return &st, nil
}

// SetActive set active strategy (will first deactivate other strategies)
func (s *StrategyStore) SetActive(userID, strategyID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// first deactivate all strategies for the user
		if err := tx.Model(&Strategy{}).Where("user_id = ?", userID).
			Update("is_active", false).Error; err != nil {
			return err
		}

		// activate specified strategy
		return tx.Model(&Strategy{}).
			Where("id = ? AND (user_id = ? OR is_default = ?)", strategyID, userID, true).
			Update("is_active", true).Error
	})
}

// Duplicate duplicate a strategy (used to create custom strategy based on default strategy)
func (s *StrategyStore) Duplicate(userID, sourceID, newID, newName string) error {
	// get source strategy
	source, err := s.Get(userID, sourceID)
	if err != nil {
		return fmt.Errorf("failed to get source strategy: %w", err)
	}

	// create new strategy
	newStrategy := &Strategy{
		ID:          newID,
		UserID:      userID,
		Name:        newName,
		Description: "Created based on [" + source.Name + "]",
		IsActive:    false,
		IsDefault:   false,
		Config:      source.Config,
	}

	return s.Create(newStrategy)
}

// ParseConfig parse strategy configuration JSON
func (s *Strategy) ParseConfig() (*StrategyConfig, error) {
	var config StrategyConfig
	if err := json.Unmarshal([]byte(s.Config), &config); err != nil {
		return nil, fmt.Errorf("failed to parse strategy configuration: %w", err)
	}
	return &config, nil
}

// SetConfig set strategy configuration
func (s *Strategy) SetConfig(config *StrategyConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to serialize strategy configuration: %w", err)
	}
	s.Config = string(data)
	return nil
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
