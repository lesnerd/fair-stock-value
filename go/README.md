# Fair Stock Value - Go Implementation

A well-organized Go application for calculating fair value prices for Fortune 500 stocks using a hybrid valuation approach.

## Overview

This application calculates fair value prices using:
- **60% Discounted Cash Flow (DCF) analysis** - Projects future cash flows and discounts them to present value
- **40% Comparable Company Analysis (Comps)** - Uses P/E ratios and earnings multiples
- **Conservative floor** - Uses tangible book value as a minimum valuation

## Features

- **Clean Architecture**: Well-organized code with separate packages for models, services, valuation, config, and utilities
- **Parallel Processing**: Multi-threaded stock analysis for improved performance
- **Configurable Parameters**: Customizable DCF parameters, P/E ratios, and weights
- **Multiple Data Sources**: Support for various financial data providers
- **Rich CLI Interface**: Command-line flags for different modes and options
- **Colored Output**: Terminal-friendly display with color-coded results
- **Progress Tracking**: Real-time progress indicators during processing

## Project Structure

```
go/
├── main.go                 # CLI interface and main application
├── models/                 # Data structures and models
│   └── stock.go           # Stock data models
├── services/              # External data fetching services
│   └── data_fetcher.go    # Stock data fetching logic
├── valuation/             # Valuation calculation logic
│   └── calculator.go      # DCF and Comps calculations
├── config/                # Configuration management
│   └── config.go          # Application configuration
├── utils/                 # Common utilities
│   ├── display.go         # Terminal display utilities
│   └── parallel.go        # Parallel processing utilities
├── data/                  # Data files
│   └── fortune_500_tickers.csv # Stock ticker symbols
├── go.mod                 # Go module file
└── README.md              # This file
```

## Installation

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd fair-stock-value/go
   ```

2. **Install dependencies**:
   ```bash
   go mod tidy
   ```

3. **Build the application**:
   ```bash
   go build -o fair-stock-value
   ```

## Usage

### Basic Usage

```bash
# Run with default settings
./fair-stock-value

# Run in test mode with limited stocks
./fair-stock-value -test

# Show help
./fair-stock-value -help
```

### Command Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `-test` | Run in test mode with limited stocks | false |
| `-tickers` | Path to ticker CSV file | `data/fortune_500_tickers.csv` |
| `-workers` | Maximum number of parallel workers | 8 |
| `-colors` | Enable colored output | true |
| `-progress` | Show progress indicators | true |
| `-sort` | Sort results by: upside, ticker, fair_value | upside |
| `-underpriced` | Show only underpriced stocks | false |
| `-limit` | Maximum number of results to show (0 = no limit) | 0 |
| `-help` | Show help message | false |

### Examples

```bash
# Test mode with 4 workers
./fair-stock-value -test -workers 4

# Show only underpriced stocks, sorted by ticker
./fair-stock-value -underpriced -sort ticker

# Show top 20 results sorted by fair value
./fair-stock-value -sort fair_value -limit 20

# Run without colors or progress (for scripting)
./fair-stock-value -colors=false -progress=false
```

## Configuration

The application uses default configuration values that can be customized:

### DCF Parameters
- **Discount Rate**: 12% (cost of capital)
- **Terminal Growth Rate**: 8% (long-term growth)
- **Max Growth Rate**: 8% (cap on growth projections)
- **Projection Years**: 5 years

### Comps Parameters
- **P/E Conservative Factor**: 85% (15% discount for conservatism)
- **Max P/E Ratio**: 40x (cap on extreme valuations)
- **Min P/E Ratio**: 5x (minimum valuation floor)

### Valuation Weights
- **DCF Weight**: 60%
- **Comps Weight**: 40%

## Output

The application displays results in a formatted table with:
- **Ticker**: Stock symbol
- **Fair Value**: Calculated fair value price
- **Current Price**: Current market price
- **Difference**: Price difference (fair value - current price)
- **Book Value**: Tangible book value per share
- **Status**: Underpriced (green) or Overpriced (red)

### Sample Output

```
================================================================================
Stock Fair Value Analysis - 2024-01-15 10:30:45
================================================================================
Ticker   Fair Value   Current Price Difference   Book Value   Status      
--------------------------------------------------------------------------------
AAPL     $195.50      $180.00      $15.50       $4.00        Underpriced
MSFT     $385.20      $350.00      $35.20       $16.00       Underpriced
GOOGL    $155.80      $140.00      $15.80       $22.00       Underpriced
AMZN     $165.40      $140.00      $25.40       $18.00       Underpriced
NVDA     $450.20      $480.00      -$29.80      $26.00       Overpriced
================================================================================
Summary:
Total stocks analyzed: 50
Underpriced: 32
Overpriced: 18
Average upside for underpriced stocks: $18.75
================================================================================
```

## Architecture

### Models Package
- Defines data structures for stocks, valuation results, and configuration
- Provides type safety and clear interfaces

### Services Package
- Handles external data fetching from financial APIs
- Implements caching and fallback mechanisms
- Supports multiple data sources

### Valuation Package
- Implements DCF and Comps valuation models
- Provides configurable parameters
- Ensures conservative valuation approach

### Config Package
- Manages application configuration
- Provides validation and defaults
- Supports different operational modes

### Utils Package
- Display utilities for terminal output
- Parallel processing utilities
- Common helper functions

## Data Sources

The application currently uses fallback data for major stocks. In a production environment, you would integrate with:
- Yahoo Finance API
- Alpha Vantage API
- Financial Modeling Prep API
- IEX Cloud API

## Performance

- **Parallel Processing**: Uses configurable worker pools for concurrent stock analysis
- **Caching**: Implements P/E ratio caching to avoid repeated API calls
- **Timeout Management**: Includes request timeouts and context cancellation
- **Memory Efficient**: Processes stocks in batches to manage memory usage

## Error Handling

- Graceful handling of API failures with fallback data
- Comprehensive error reporting
- Timeout management for long-running operations
- Validation of input parameters

## Testing

```bash
# Run in test mode
./fair-stock-value -test

# Run with specific test parameters
./fair-stock-value -test -workers 2 -progress=true
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License.

## Comparison with Python Version

The Go version offers several advantages over the Python implementation:

1. **Better Organization**: Clear separation of concerns with dedicated packages
2. **Performance**: Compiled binary with better performance characteristics
3. **Concurrency**: Native goroutines for efficient parallel processing
4. **Type Safety**: Static typing prevents runtime errors
5. **Deployment**: Single binary deployment without dependency management
6. **Memory Management**: Better memory usage patterns
7. **Error Handling**: Explicit error handling with proper error types

## Future Enhancements

- Configuration file support (JSON/YAML)
- Database integration for historical data
- Web API interface
- Real-time data streaming
- Advanced charting and visualization
- Portfolio-level analysis
- Risk metrics calculation
- Backtesting capabilities