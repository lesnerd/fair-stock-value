package utils

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"fair-stock-value/models"
)

// Colors for terminal output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
)

// DisplayResults displays the valuation results in a formatted table
func DisplayResults(results []*models.ValuationResult, showColors bool, sortBy string, showOnlyUnderpriced bool, maxResults int, showExtra bool) {
	if len(results) == 0 {
		fmt.Println("No results to display!")
		return
	}

	// Filter results if needed
	filteredResults := results
	if showOnlyUnderpriced {
		filteredResults = filterUnderpriced(results)
	}

	// Sort results
	sortResults(filteredResults, sortBy)

	// Limit results if specified
	if maxResults > 0 && len(filteredResults) > maxResults {
		filteredResults = filteredResults[:maxResults]
	}

	// Display header
	displayHeader(showColors)

	// Display table
	displayTable(filteredResults, showColors, showExtra)

	// Display summary
	displaySummary(results, showColors)
}

// filterUnderpriced filters results to show only underpriced stocks
func filterUnderpriced(results []*models.ValuationResult) []*models.ValuationResult {
	var filtered []*models.ValuationResult
	for _, result := range results {
		if result.Status == models.StatusUnderpriced {
			filtered = append(filtered, result)
		}
	}
	return filtered
}

// sortResults sorts results based on the specified criteria
func sortResults(results []*models.ValuationResult, sortBy string) {
	switch sortBy {
	case "upside":
		sort.Slice(results, func(i, j int) bool {
			if results[i].Status == models.StatusUnderpriced && results[j].Status == models.StatusOverpriced {
				return true
			}
			if results[i].Status == models.StatusOverpriced && results[j].Status == models.StatusUnderpriced {
				return false
			}
			if results[i].Status == models.StatusUnderpriced && results[j].Status == models.StatusUnderpriced {
				return results[i].PriceDifference > results[j].PriceDifference
			}
			if results[i].Status == models.StatusOverpriced && results[j].Status == models.StatusOverpriced {
				return results[i].PriceDifference > results[j].PriceDifference
			}
			return results[i].Ticker < results[j].Ticker
		})
	case "ticker":
		sort.Slice(results, func(i, j int) bool {
			return results[i].Ticker < results[j].Ticker
		})
	case "fair_value":
		sort.Slice(results, func(i, j int) bool {
			return results[i].FairValue > results[j].FairValue
		})
	default:
		// Default to upside sorting
		sortResults(results, "upside")
	}
}

// displayHeader displays the table header
func displayHeader(showColors bool) {
	currentTime := time.Now()
	
	separator := strings.Repeat("=", 98)
	title := fmt.Sprintf("Stock Fair Value Analysis - %s", currentTime.Format("2006-01-02 15:04:05"))
	
	if showColors {
		fmt.Printf("%s%s%s%s\n", ColorBold, ColorCyan, separator, ColorReset)
		fmt.Printf("%s%s%s%s\n", ColorBold, ColorCyan, title, ColorReset)
		fmt.Printf("%s%s%s%s\n", ColorBold, ColorCyan, separator, ColorReset)
	} else {
		fmt.Println(separator)
		fmt.Println(title)
		fmt.Println(separator)
	}
}

// displayTable displays the results in a formatted table
func displayTable(results []*models.ValuationResult, showColors bool, showExtra bool) {
	// Table header
	if showExtra {
		if showColors {
			fmt.Printf("%s%-8s %-12s %-12s %-12s %-8s %-12s %-12s %-8s %-6s %-8s %-12s %-20s %-12s%s\n", 
				ColorBold, "Ticker", "Fair Value", "Current Price", "Difference", "Pct", "Book Value", "Status", "Growth", "P/E", "EPS", "FCF/Share", "Sector", "Company", ColorReset)
		} else {
			fmt.Printf("%-8s %-12s %-12s %-12s %-8s %-12s %-12s %-8s %-6s %-8s %-12s %-20s %-12s\n", 
				"Ticker", "Fair Value", "Current Price", "Difference", "Pct", "Book Value", "Status", "Growth", "P/E", "EPS", "FCF/Share", "Sector", "Company")
		}
	} else {
		if showColors {
			fmt.Printf("%s%-8s %-12s %-12s %-12s %-8s %-12s %-12s %-8s%s\n", 
				ColorBold, "Ticker", "Fair Value", "Current Price", "Difference", "Pct", "Book Value", "Status", "Growth", ColorReset)
		} else {
			fmt.Printf("%-8s %-12s %-12s %-12s %-8s %-12s %-12s %-8s\n", 
				"Ticker", "Fair Value", "Current Price", "Difference", "Pct", "Book Value", "Status", "Growth")
		}
	}
	
	// Separator line
	separatorLength := 98
	if showExtra {
		separatorLength = 168
	}
	fmt.Println(strings.Repeat("-", separatorLength))
	
	// Table rows
	for _, result := range results {
		displayRow(result, showColors, showExtra)
	}
}

// displayRow displays a single result row
func displayRow(result *models.ValuationResult, showColors bool, showExtra bool) {
	var color string
	if showColors {
		if result.Status == models.StatusUnderpriced {
			color = ColorGreen
		} else {
			color = ColorRed
		}
	}
	
	if showExtra {
		// Truncate company name if too long
		companyName := result.CompanyName
		if len(companyName) > 20 {
			companyName = companyName[:17] + "..."
		}
		
		// Truncate sector if too long
		sector := result.Sector
		if len(sector) > 18 {
			sector = sector[:15] + "..."
		}
		
		fmt.Printf("%s%-8s $%-11.2f $%-11.2f $%-11.2f %6.1f%% $%-11.2f %-12s %5.1f%% %5.1f $%-7.2f $%-11.2f %-20s %-12s%s\n",
			color,
			result.Ticker,
			result.FairValue,
			result.CurrentPrice,
			result.PriceDifference,
			result.UpsidePercentage,
			result.BookValue,
			result.Status,
			result.GrowthRate*100,
			result.PERatio,
			result.EPS,
			result.FCFPerShare,
			sector,
			companyName,
			ColorReset)
	} else {
		fmt.Printf("%s%-8s $%-11.2f $%-11.2f $%-11.2f %6.1f%% $%-11.2f %-12s %5.1f%%%s\n",
			color,
			result.Ticker,
			result.FairValue,
			result.CurrentPrice,
			result.PriceDifference,
			result.UpsidePercentage,
			result.BookValue,
			result.Status,
			result.GrowthRate*100,
			ColorReset)
	}
}

// formatMarketCap formats market cap in human-readable format
func formatMarketCap(marketCap int64) string {
	if marketCap == 0 {
		return "N/A"
	}
	
	if marketCap >= 1000000000000 {
		return fmt.Sprintf("%.1fT", float64(marketCap)/1000000000000)
	} else if marketCap >= 1000000000 {
		return fmt.Sprintf("%.1fB", float64(marketCap)/1000000000)
	} else if marketCap >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(marketCap)/1000000)
	} else {
		return fmt.Sprintf("%.1fK", float64(marketCap)/1000)
	}
}

// displaySummary displays summary statistics
func displaySummary(results []*models.ValuationResult, showColors bool) {
	underpriced := 0
	overpriced := 0
	totalUpside := 0.0
	
	for _, result := range results {
		if result.Status == models.StatusUnderpriced {
			underpriced++
			totalUpside += result.PriceDifference
		} else {
			overpriced++
		}
	}
	
	avgUpside := 0.0
	if underpriced > 0 {
		avgUpside = totalUpside / float64(underpriced)
	}
	
	separator := strings.Repeat("=", 98)
	
	if showColors {
		fmt.Printf("\n%s%s%s%s\n", ColorBold, ColorCyan, separator, ColorReset)
		fmt.Printf("%sSummary:%s\n", ColorBold, ColorReset)
		fmt.Printf("Total stocks analyzed: %d\n", len(results))
		fmt.Printf("%sUnderpriced: %d%s\n", ColorGreen, underpriced, ColorReset)
		fmt.Printf("%sOverpriced: %d%s\n", ColorRed, overpriced, ColorReset)
		if underpriced > 0 {
			fmt.Printf("%sAverage upside for underpriced stocks: $%.2f%s\n", ColorGreen, avgUpside, ColorReset)
		}
		fmt.Printf("%s%s%s%s\n", ColorBold, ColorCyan, separator, ColorReset)
	} else {
		fmt.Printf("\n%s\n", separator)
		fmt.Println("Summary:")
		fmt.Printf("Total stocks analyzed: %d\n", len(results))
		fmt.Printf("Underpriced: %d\n", underpriced)
		fmt.Printf("Overpriced: %d\n", overpriced)
		if underpriced > 0 {
			fmt.Printf("Average upside for underpriced stocks: $%.2f\n", avgUpside)
		}
		fmt.Printf("%s\n", separator)
	}
}

// ShowProgress displays a progress indicator
func ShowProgress(current, total int, ticker string) {
	percentage := float64(current) / float64(total) * 100
	fmt.Printf("\rProcessing %s (%d/%d - %.1f%%)", ticker, current, total, percentage)
	
	if current == total {
		fmt.Println() // New line when complete
	}
}

// ClearLine clears the current line in the terminal
func ClearLine() {
	fmt.Print("\r" + strings.Repeat(" ", 80) + "\r")
}

// IsTerminal checks if stdout is a terminal
func IsTerminal() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}