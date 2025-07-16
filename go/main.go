package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"fair-stock-value/config"
	"fair-stock-value/models"
	"fair-stock-value/services"
	"fair-stock-value/utils"
	"fair-stock-value/valuation"
)

func main() {
	// Command line flags
	var (
		testMode     = flag.Bool("test", false, "Run in test mode with limited stocks")
		tickerFile   = flag.String("tickers", "", "Path to ticker CSV file")
		maxWorkers   = flag.Int("workers", 8, "Maximum number of parallel workers")
		showColors   = flag.Bool("colors", true, "Enable colored output")
		showProgress = flag.Bool("progress", true, "Show progress indicators")
		sortBy       = flag.String("sort", "upside", "Sort results by: upside, ticker, fair_value")
		onlyUnderpriced = flag.Bool("underpriced", false, "Show only underpriced stocks")
		maxResults   = flag.Int("limit", 0, "Maximum number of results to show (0 = no limit)")
		showExtra    = flag.Bool("extra", false, "Show additional fields (P/E, EPS, Market Cap, Sector)")
		help         = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Load configuration
	cfg := config.NewDefaultConfig()
	if *testMode {
		cfg = config.GetTestConfig()
	}

	// Override config with command line flags
	if *tickerFile != "" {
		cfg.DataSources.TickerFile = *tickerFile
	}
	if *maxWorkers > 0 {
		cfg.Processing.MaxWorkers = *maxWorkers
	}
	cfg.Output.ShowColors = *showColors
	cfg.Output.ShowProgress = *showProgress
	cfg.Output.SortBy = *sortBy
	cfg.Output.ShowOnlyUnderpriced = *onlyUnderpriced
	cfg.Output.ShowExtra = *showExtra
	if *maxResults > 0 {
		cfg.Output.MaxResults = *maxResults
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// Create application
	app := NewApplication(cfg)

	// Run the application
	if err := app.Run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}

// Application represents the main application
type Application struct {
	config      *config.Config
	dataFetcher *services.DataFetcher
	calculator  *valuation.Calculator
	tickers     []string
}

// NewApplication creates a new application instance
func NewApplication(cfg *config.Config) *Application {
	return &Application{
		config:      cfg,
		dataFetcher: services.NewDataFetcher(),
		calculator:  valuation.NewCalculator(),
	}
}

// Run runs the stock valuation analysis
func (app *Application) Run() error {
	fmt.Println("Starting stock valuation analysis...")

	// Load tickers
	if err := app.loadTickers(); err != nil {
		return fmt.Errorf("failed to load tickers: %w", err)
	}

	// Configure calculator with config parameters
	app.calculator.SetDCFParameters(app.config.DCFParams)
	app.calculator.SetCompsParameters(app.config.CompsParams)
	app.calculator.SetWeights(app.config.Weights)

	// Process stocks
	results, err := app.processStocks()
	if err != nil {
		return fmt.Errorf("failed to process stocks: %w", err)
	}

	// Display results
	utils.DisplayResults(
		results,
		app.config.Output.ShowColors,
		app.config.Output.SortBy,
		app.config.Output.ShowOnlyUnderpriced,
		app.config.Output.MaxResults,
		app.config.Output.ShowExtra,
	)

	return nil
}

// loadTickers loads ticker symbols from CSV file or uses defaults
func (app *Application) loadTickers() error {
	// Use test tickers if in test mode
	if app.config.Output.MaxResults == 10 { // Test mode indicator
		app.tickers = []string{
			"AAPL", "MSFT", "GOOGL", "AMZN", "NVDA",
			"META", "TSLA", "BRK-B", "UNH", "JNJ",
		}
		fmt.Printf("Using test tickers: %d stocks\n", len(app.tickers))
		return nil
	}

	// Try to load from CSV file
	tickers, err := app.dataFetcher.LoadTickersFromCSV(app.config.DataSources.TickerFile)
	if err != nil {
		fmt.Printf("Warning: Could not load tickers from CSV, using defaults: %v\n", err)
		// Use default tickers
		app.tickers = []string{
			"AAPL", "MSFT", "GOOGL", "AMZN", "NVDA", "META", "TSLA", "BRK-B",
			"UNH", "JNJ", "JPM", "V", "PG", "HD", "MA", "BAC", "ABBV", "PFE",
			"KO", "AVGO", "PEP", "TMO", "COST", "WMT", "MRK", "DIS", "ACN",
			"VZ", "ADBE", "NFLX", "NKE", "CRM", "DHR", "LIN", "TXN", "NEE",
			"ABT", "ORCL", "PM", "RTX", "QCOM", "HON", "WFC", "UPS", "T",
			"LOW", "SPGI", "ELV", "SCHW", "CAT",
		}
	} else {
		app.tickers = tickers
	}

	fmt.Printf("Loaded %d tickers for analysis\n", len(app.tickers))
	return nil
}

// processStocks processes all stocks and returns valuation results
func (app *Application) processStocks() ([]*models.ValuationResult, error) {
	fmt.Printf("Processing %d stocks with %d parallel workers...\n", 
		len(app.tickers), app.config.Processing.MaxWorkers)

	results := make([]*models.ValuationResult, 0, len(app.tickers))
	resultsChan := make(chan *models.ValuationResult, len(app.tickers))
	errorsChan := make(chan error, len(app.tickers))

	// Create worker pool
	workerPool := utils.NewWorkerPool(app.config.Processing.MaxWorkers)
	defer workerPool.Close()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Progress tracking
	var completed int
	var mu sync.Mutex
	
	// Process each ticker
	for i, ticker := range app.tickers {
		tickerCopy := ticker
		index := i
		
		workerPool.Submit(func() {
			if app.config.Output.ShowProgress {
				utils.ShowProgress(index+1, len(app.tickers), tickerCopy)
			}
			
			result, err := app.processStock(ctx, tickerCopy)
			if err != nil {
				errorsChan <- fmt.Errorf("failed to process %s: %w", tickerCopy, err)
				return
			}
			
			resultsChan <- result
			
			mu.Lock()
			completed++
			mu.Unlock()
		})
	}

	// Collect results
	var errors []error
	for i := 0; i < len(app.tickers); i++ {
		select {
		case result := <-resultsChan:
			results = append(results, result)
		case err := <-errorsChan:
			errors = append(errors, err)
		case <-ctx.Done():
			return nil, fmt.Errorf("processing timed out: %w", ctx.Err())
		}
	}

	// Report errors if any
	if len(errors) > 0 {
		fmt.Printf("\nWarning: %d stocks failed to process:\n", len(errors))
		for _, err := range errors {
			fmt.Printf("  - %v\n", err)
		}
	}

	if app.config.Output.ShowProgress {
		fmt.Printf("\nCompleted processing %d stocks\n", len(results))
	}

	return results, nil
}

// processStock processes a single stock and returns its valuation result
func (app *Application) processStock(ctx context.Context, ticker string) (*models.ValuationResult, error) {
	// Fetch stock data
	stockData, err := app.dataFetcher.FetchStockData(ctx, ticker)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data for %s: %w", ticker, err)
	}

	// Calculate valuation
	result := app.calculator.CalculateFairValue(stockData)
	if result == nil {
		return nil, fmt.Errorf("failed to calculate valuation for %s", ticker)
	}

	return result, nil
}

// showHelp displays help information
func showHelp() {
	fmt.Println("Stock Fair Value Estimation Tool")
	fmt.Println("=================================")
	fmt.Println()
	fmt.Println("This tool calculates fair value prices for stocks using a hybrid approach:")
	fmt.Println("- 60% Discounted Cash Flow (DCF) analysis")
	fmt.Println("- 40% Comparable Company Analysis (Comps)")
	fmt.Println("- Tangible book value as a conservative floor")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  fair-stock-value [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -test              Run in test mode with limited stocks")
	fmt.Println("  -config string     Path to configuration file")
	fmt.Println("  -tickers string    Path to ticker CSV file")
	fmt.Println("  -workers int       Maximum number of parallel workers (default 8)")
	fmt.Println("  -colors            Enable colored output (default true)")
	fmt.Println("  -progress          Show progress indicators (default true)")
	fmt.Println("  -sort string       Sort results by: upside, ticker, fair_value (default \"upside\")")
	fmt.Println("  -underpriced       Show only underpriced stocks")
	fmt.Println("  -limit int         Maximum number of results to show (0 = no limit)")
	fmt.Println("  -extra             Show additional fields (P/E, EPS, FCF/Share, Sector, Company)")
	fmt.Println("  -help              Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  fair-stock-value -test")
	fmt.Println("  fair-stock-value -workers 4 -sort ticker")
	fmt.Println("  fair-stock-value -underpriced -limit 20")
	fmt.Println("  fair-stock-value -extra -limit 10")
	fmt.Println()
}