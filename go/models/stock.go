package models

import "time"

// StockData represents comprehensive stock information
type StockData struct {
	Ticker        string    `json:"ticker"`
	CompanyName   string    `json:"company_name"`
	CurrentPrice  float64   `json:"current_price"`
	FCFPerShare   float64   `json:"fcf_per_share"`
	EPS           float64   `json:"eps"`
	BookValue     float64   `json:"book_value"`
	Sector        string    `json:"sector"`
	GrowthRate    float64   `json:"growth_rate"`
	PERatio       float64   `json:"pe_ratio"`
	MarketCap     int64     `json:"market_cap"`
	FetchTime     time.Time `json:"fetch_time"`
}

// ValuationResult represents the result of stock valuation
type ValuationResult struct {
	Ticker             string  `json:"ticker"`
	FairValue          float64 `json:"fair_value"`
	CurrentPrice       float64 `json:"current_price"`
	PriceDifference    float64 `json:"price_difference"`
	BookValue          float64 `json:"book_value"`
	Status             string  `json:"status"`
	DCFValue           float64 `json:"dcf_value"`
	CompsValue         float64 `json:"comps_value"`
	UpsidePercentage   float64 `json:"upside_percentage"`
}

// IndustryPERatio represents P/E ratios by industry
type IndustryPERatio struct {
	Sector   string  `json:"sector"`
	PERatio  float64 `json:"pe_ratio"`
}

// FinancialMetrics represents key financial metrics for DCF calculation
type FinancialMetrics struct {
	FreeCashFlow      float64 `json:"free_cash_flow"`
	SharesOutstanding int64   `json:"shares_outstanding"`
	Revenue           float64 `json:"revenue"`
	NetIncome         float64 `json:"net_income"`
	TotalDebt         float64 `json:"total_debt"`
	Cash              float64 `json:"cash"`
}

// DCFParameters represents parameters for DCF calculation
type DCFParameters struct {
	DiscountRate         float64 `json:"discount_rate"`
	TerminalGrowthRate   float64 `json:"terminal_growth_rate"`
	MaxGrowthRate        float64 `json:"max_growth_rate"`
	ProjectionYears      int     `json:"projection_years"`
}

// CompsParameters represents parameters for comparable analysis
type CompsParameters struct {
	PEConservativeFactor float64 `json:"pe_conservative_factor"`
	MaxPERatio           float64 `json:"max_pe_ratio"`
	MinPERatio           float64 `json:"min_pe_ratio"`
}

// ValuationWeights represents weights for hybrid valuation
type ValuationWeights struct {
	DCFWeight   float64 `json:"dcf_weight"`
	CompsWeight float64 `json:"comps_weight"`
}

// Status constants for valuation results
const (
	StatusUnderpriced = "Underpriced"
	StatusOverpriced  = "Overpriced"
	StatusError       = "Error"
)