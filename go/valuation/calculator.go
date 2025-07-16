package valuation

import (
	"math"

	"fair-stock-value/models"
)

// Calculator handles stock valuation calculations
type Calculator struct {
	dcfParams     models.DCFParameters
	compsParams   models.CompsParameters
	weights       models.ValuationWeights
}

// NewCalculator creates a new valuation calculator with default parameters
func NewCalculator() *Calculator {
	return &Calculator{
		dcfParams: models.DCFParameters{
			DiscountRate:       0.12, // 12% discount rate
			TerminalGrowthRate: 0.08, // 8% terminal growth rate
			MaxGrowthRate:      0.08, // 8% max growth rate cap
			ProjectionYears:    5,    // 5 year projection
		},
		compsParams: models.CompsParameters{
			PEConservativeFactor: 0.85, // 15% discount for conservatism
			MaxPERatio:          40.0,  // Cap P/E at 40x
			MinPERatio:          5.0,   // Minimum P/E of 5x
		},
		weights: models.ValuationWeights{
			DCFWeight:   0.6, // 60% weight for DCF
			CompsWeight: 0.4, // 40% weight for Comps
		},
	}
}

// CalculateFairValue calculates the hybrid fair value using DCF and Comps
func (c *Calculator) CalculateFairValue(stockData *models.StockData) *models.ValuationResult {
	dcfValue := c.calculateDCFValue(stockData)
	compsValue := c.calculateCompsValue(stockData)
	
	// Weighted average: 60% DCF + 40% Comps
	fairValue := (dcfValue * c.weights.DCFWeight) + (compsValue * c.weights.CompsWeight)
	
	// Ensure fair value is not below book value (conservative floor)
	fairValue = math.Max(fairValue, stockData.BookValue)
	
	// Calculate metrics
	priceDifference := fairValue - stockData.CurrentPrice
	upsidePercentage := (priceDifference / stockData.CurrentPrice) * 100
	
	status := models.StatusOverpriced
	if stockData.CurrentPrice < fairValue {
		status = models.StatusUnderpriced
	}
	
	return &models.ValuationResult{
		Ticker:           stockData.Ticker,
		FairValue:        fairValue,
		CurrentPrice:     stockData.CurrentPrice,
		PriceDifference:  priceDifference,
		BookValue:        stockData.BookValue,
		Status:           status,
		DCFValue:         dcfValue,
		CompsValue:       compsValue,
		UpsidePercentage: upsidePercentage,
		
		// Additional optional fields
		PERatio:          stockData.PERatio,
		EPS:              stockData.EPS,
		FCFPerShare:      stockData.FCFPerShare,
		MarketCap:        stockData.MarketCap,
		Sector:           stockData.Sector,
		GrowthRate:       stockData.GrowthRate,
		CompanyName:      stockData.CompanyName,
	}
}

// calculateDCFValue calculates fair value using Discounted Cash Flow model
func (c *Calculator) calculateDCFValue(stockData *models.StockData) float64 {
	fcfPerShare := stockData.FCFPerShare
	growthRate := math.Min(stockData.GrowthRate, c.dcfParams.MaxGrowthRate)
	
	// If FCF is negative or zero, use a conservative estimate
	if fcfPerShare <= 0 {
		fcfPerShare = 2.0 // Conservative fallback
	}
	
	// Project FCF for the specified number of years
	var projectedFCF []float64
	for year := 1; year <= c.dcfParams.ProjectionYears; year++ {
		fcf := fcfPerShare * math.Pow(1+growthRate, float64(year))
		projectedFCF = append(projectedFCF, fcf)
	}
	
	// Calculate present value of projected FCF
	var pvFCF float64
	for i, fcf := range projectedFCF {
		pvFCF += fcf / math.Pow(1+c.dcfParams.DiscountRate, float64(i+1))
	}
	
	// Calculate terminal value using Gordon Growth Model
	terminalFCF := projectedFCF[len(projectedFCF)-1] * (1 + c.dcfParams.TerminalGrowthRate)
	terminalValue := terminalFCF / (c.dcfParams.DiscountRate - c.dcfParams.TerminalGrowthRate)
	pvTerminalValue := terminalValue / math.Pow(1+c.dcfParams.DiscountRate, float64(c.dcfParams.ProjectionYears))
	
	// Total DCF value
	dcfValue := pvFCF + pvTerminalValue
	
	// Use book value as floor
	return math.Max(dcfValue, stockData.BookValue)
}

// calculateCompsValue calculates fair value using Comparable Company Analysis
func (c *Calculator) calculateCompsValue(stockData *models.StockData) float64 {
	eps := stockData.EPS
	peRatio := stockData.PERatio
	
	// Apply conservative adjustments to P/E ratio
	conservativePE := peRatio * c.compsParams.PEConservativeFactor
	conservativePE = math.Max(c.compsParams.MinPERatio, math.Min(conservativePE, c.compsParams.MaxPERatio))
	
	// If EPS is negative, use a conservative approach
	if eps <= 0 {
		eps = 1.0 // Conservative fallback
	}
	
	// Calculate value using P/E multiple
	compsValue := eps * conservativePE
	
	// Use book value as floor
	return math.Max(compsValue, stockData.BookValue)
}

// SetDCFParameters allows customization of DCF parameters
func (c *Calculator) SetDCFParameters(params models.DCFParameters) {
	c.dcfParams = params
}

// SetCompsParameters allows customization of Comps parameters
func (c *Calculator) SetCompsParameters(params models.CompsParameters) {
	c.compsParams = params
}

// SetWeights allows customization of valuation weights
func (c *Calculator) SetWeights(weights models.ValuationWeights) {
	c.weights = weights
}

// GetDCFParameters returns current DCF parameters
func (c *Calculator) GetDCFParameters() models.DCFParameters {
	return c.dcfParams
}

// GetCompsParameters returns current Comps parameters
func (c *Calculator) GetCompsParameters() models.CompsParameters {
	return c.compsParams
}

// GetWeights returns current valuation weights
func (c *Calculator) GetWeights() models.ValuationWeights {
	return c.weights
}