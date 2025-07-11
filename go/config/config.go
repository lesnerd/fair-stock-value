package config

import (
	"fmt"
	
	"fair-stock-value/models"
)

// Config holds application configuration
type Config struct {
	DCFParams     models.DCFParameters     `json:"dcf_parameters"`
	CompsParams   models.CompsParameters   `json:"comps_parameters"`
	Weights       models.ValuationWeights  `json:"valuation_weights"`
	DataSources   DataSourcesConfig        `json:"data_sources"`
	Processing    ProcessingConfig         `json:"processing"`
	Output        OutputConfig             `json:"output"`
}

// DataSourcesConfig holds configuration for data sources
type DataSourcesConfig struct {
	TickerFile          string `json:"ticker_file"`
	UseYahooFinance     bool   `json:"use_yahoo_finance"`
	UseAlphaVantage     bool   `json:"use_alpha_vantage"`
	AlphaVantageAPIKey  string `json:"alpha_vantage_api_key"`
	RequestTimeout      int    `json:"request_timeout_seconds"`
	MaxRetries          int    `json:"max_retries"`
}

// ProcessingConfig holds configuration for processing
type ProcessingConfig struct {
	MaxWorkers        int  `json:"max_workers"`
	EnableCaching     bool `json:"enable_caching"`
	CacheExpiryHours  int  `json:"cache_expiry_hours"`
	EnableParallel    bool `json:"enable_parallel"`
}

// OutputConfig holds configuration for output formatting
type OutputConfig struct {
	ShowColors        bool `json:"show_colors"`
	ShowProgress      bool `json:"show_progress"`
	SortBy            string `json:"sort_by"` // "upside", "ticker", "fair_value"
	ShowOnlyUnderpriced bool `json:"show_only_underpriced"`
	MaxResults        int  `json:"max_results"`
}

// NewDefaultConfig creates a new configuration with default values
func NewDefaultConfig() *Config {
	return &Config{
		DCFParams: models.DCFParameters{
			DiscountRate:       0.12,
			TerminalGrowthRate: 0.08,
			MaxGrowthRate:      0.08,
			ProjectionYears:    5,
		},
		CompsParams: models.CompsParameters{
			PEConservativeFactor: 0.85,
			MaxPERatio:          40.0,
			MinPERatio:          5.0,
		},
		Weights: models.ValuationWeights{
			DCFWeight:   0.6,
			CompsWeight: 0.4,
		},
		DataSources: DataSourcesConfig{
			TickerFile:         "data/fortune_500_tickers.csv",
			UseYahooFinance:    true,
			UseAlphaVantage:    false,
			AlphaVantageAPIKey: "",
			RequestTimeout:     10,
			MaxRetries:         3,
		},
		Processing: ProcessingConfig{
			MaxWorkers:       8,
			EnableCaching:    true,
			CacheExpiryHours: 24,
			EnableParallel:   true,
		},
		Output: OutputConfig{
			ShowColors:          true,
			ShowProgress:        true,
			SortBy:             "upside",
			ShowOnlyUnderpriced: false,
			MaxResults:         0, // 0 means no limit
		},
	}
}

// GetTestConfig returns a configuration optimized for testing
func GetTestConfig() *Config {
	config := NewDefaultConfig()
	
	// Modify for testing
	config.Processing.MaxWorkers = 4
	config.Processing.EnableParallel = true
	config.Output.ShowProgress = true
	config.Output.MaxResults = 10
	
	return config
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate DCF parameters
	if c.DCFParams.DiscountRate <= 0 || c.DCFParams.DiscountRate >= 1 {
		return fmt.Errorf("discount rate must be between 0 and 1")
	}
	
	if c.DCFParams.TerminalGrowthRate <= 0 || c.DCFParams.TerminalGrowthRate >= c.DCFParams.DiscountRate {
		return fmt.Errorf("terminal growth rate must be positive and less than discount rate")
	}
	
	if c.DCFParams.ProjectionYears <= 0 {
		return fmt.Errorf("projection years must be positive")
	}
	
	// Validate Comps parameters
	if c.CompsParams.PEConservativeFactor <= 0 || c.CompsParams.PEConservativeFactor > 1 {
		return fmt.Errorf("P/E conservative factor must be between 0 and 1")
	}
	
	if c.CompsParams.MinPERatio <= 0 || c.CompsParams.MinPERatio >= c.CompsParams.MaxPERatio {
		return fmt.Errorf("invalid P/E ratio bounds")
	}
	
	// Validate weights
	if c.Weights.DCFWeight < 0 || c.Weights.CompsWeight < 0 {
		return fmt.Errorf("weights cannot be negative")
	}
	
	totalWeight := c.Weights.DCFWeight + c.Weights.CompsWeight
	if totalWeight <= 0 {
		return fmt.Errorf("total weight must be positive")
	}
	
	// Normalize weights if they don't sum to 1
	if totalWeight != 1.0 {
		c.Weights.DCFWeight /= totalWeight
		c.Weights.CompsWeight /= totalWeight
	}
	
	// Validate processing parameters
	if c.Processing.MaxWorkers <= 0 {
		return fmt.Errorf("max workers must be positive")
	}
	
	if c.Processing.CacheExpiryHours < 0 {
		return fmt.Errorf("cache expiry hours cannot be negative")
	}
	
	// Validate data source parameters
	if c.DataSources.RequestTimeout <= 0 {
		return fmt.Errorf("request timeout must be positive")
	}
	
	if c.DataSources.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}
	
	return nil
}

// GetIndustryPERatios returns the default industry P/E ratios
func GetIndustryPERatios() map[string]float64 {
	return map[string]float64{
		"Technology":             22.0,
		"Healthcare":             18.0,
		"Financial Services":     10.0,
		"Consumer Cyclical":      16.0,
		"Consumer Defensive":     20.0,
		"Energy":                 12.0,
		"Industrials":            13.0,
		"Materials":              12.0,
		"Real Estate":            14.0,
		"Utilities":              16.0,
		"Communication Services": 18.0,
		"Default":                18.0,
	}
}