# Stock Fair Value Estimation Application

A comprehensive Python application that calculates fair value prices for Fortune 500 stocks using advanced financial modeling techniques. The application combines Discounted Cash Flow (DCF) analysis with Comparable Company Analysis to provide realistic stock valuations.

## Features

- **Hybrid Valuation Methodology**: 60% DCF + 40% Comparable Company Analysis
- **Real-time Data Fetching**: Uses yfinance API to fetch current stock data
- **Dynamic P/E Ratios**: Fetches real-time P/E ratios for each stock with industry fallbacks
- **Automated Scheduling**: Runs daily at 9:00 AM ET (30 minutes before market open)
- **Terminal-based Interface**: Clean, tabulated output in the terminal
- **Conservative Approach**: Uses book value as a valuation floor
- **Error Handling**: Robust fallback mechanisms for missing data
- **Caching System**: Avoids repeated API calls for P/E ratios
- **Fortune 500 Coverage**: Analyzes major Fortune 500 stocks

## Methodology

### Discounted Cash Flow (DCF) - 60% Weight
- Fetches real-time free cash flow per share (FCF)
- Projects FCF for 5 years using analyst-derived growth rates (capped at 8%)
- Applies 12% discount rate
- Calculates terminal value using Gordon Growth Model

### Comparable Company Analysis - 40% Weight
- Uses real-time earnings per share (EPS)
- **Dynamic P/E Ratios**: Fetches actual P/E ratios from yfinance for each stock
- **Intelligent Fallbacks**: Uses industry-specific P/E ratios when real-time data unavailable
- Sector-based valuation multiples

### Conservative Floor
- Uses tangible book value per share as minimum valuation
- Ensures realistic estimates aligned with analyst expectations

## Installation

### Prerequisites
- Python 3.7 or higher
- pip package manager

### Install Dependencies
```bash
pip install -r requirements.txt
```

### Required Packages
- `yfinance>=0.2.18` - Stock data fetching
- `schedule>=1.2.0` - Task scheduling
- `pytz>=2023.3` - Timezone handling
- `tabulate>=0.9.0` - Table formatting

## Usage

### Run Once (Immediate Analysis)
```bash
python stock_fair_value.py
```

### Run Demo Version (Works Offline)
```bash
python stock_fair_value_demo.py
```

### Run Test with Sample Stocks
```bash
python stock_fair_value.py --test
```

### Schedule Daily Analysis at 9:00 AM ET
```bash
python stock_fair_value.py --schedule
```

## Sample Output

```
====================================================================================================
Stock Fair Value Analysis - 2025-07-03 09:00:00 ET
====================================================================================================
Methodology: 60% DCF + 40% Comparable Company Analysis
====================================================================================================
+----------+-------------------------+--------------+-----------------+--------------+-------------+----------+
| Ticker   | Company                 | Fair Value   | Current Price   | Book Value   | Status      | Diff %   |
+==========+=========================+==============+=================+==============+=============+==========+
| AAPL     | Apple Inc.              | $157.95      | $170.10         | $4.40        | Overpriced  | -7.1%    |
+----------+-------------------------+--------------+-----------------+--------------+-------------+----------+
| MSFT     | Microsoft Corporation   | $299.51      | $415.26         | $30.59       | Overpriced  | -27.9%   |
+----------+-------------------------+--------------+-----------------+--------------+-------------+----------+
| GOOGL    | Alphabet Inc.           | $131.28      | $179.58         | $73.97       | Overpriced  | -26.9%   |
+----------+-------------------------+--------------+-----------------+--------------+-------------+----------+
| AMZN     | Amazon.com Inc.         | $73.03       | $223.22         | $20.28       | Overpriced  | -67.3%   |
+----------+-------------------------+--------------+-----------------+--------------+-------------+----------+
| BRK-B    | Berkshire Hathaway Inc. | $543.25      | $462.67         | $330.44      | Underpriced | +17.4%   |
+----------+-------------------------+--------------+-----------------+--------------+-------------+----------+

Summary:
Total stocks analyzed: 10
Underpriced: 2
Overpriced: 8
====================================================================================================
```

## Key Parameters

- **Discount Rate**: 12% (reflects market risk)
- **Maximum Growth Rate**: 8% (conservative cap)
- **Terminal Growth Rate**: 8% (for DCF terminal value)
- **DCF Weight**: 60% (primary valuation method)
- **Comps Weight**: 40% (secondary validation)

## P/E Ratio System

### Dynamic P/E Ratios
The application first attempts to fetch real-time P/E ratios for each individual stock using yfinance. This provides the most accurate and current valuation multiples.

### Fallback Industry P/E Ratios
When real-time data is unavailable, the system uses these industry-specific defaults:

- Technology: 25x
- Healthcare: 20x
- Financial Services: 12x
- Consumer Cyclical: 18x
- Consumer Defensive: 22x
- Energy: 15x
- Industrials: 15x
- Materials: 14x
- Communication Services: 20x

### Caching
P/E ratios are cached during execution to minimize API calls and improve performance.

## Files Structure

```
fair-stock-value/
├── stock_fair_value.py          # Main application (requires internet)
├── stock_fair_value_demo.py     # Demo version (works offline)
├── fortune_500_tickers.csv      # Fortune 500 stock tickers (easily editable)
├── requirements.txt             # Python dependencies
└── README.md                   # This file
```

## Customizing Stock List

You can easily customize which stocks to analyze by editing the `fortune_500_tickers.csv` file:

```csv
ticker,company_name,sector
AAPL,Apple Inc.,Technology
MSFT,Microsoft Corporation,Technology
GOOGL,Alphabet Inc.,Communication Services
AMZN,Amazon.com Inc.,Consumer Cyclical
# Add or remove stocks as needed
```

### CSV File Format
- **ticker**: Stock symbol (e.g., AAPL, MSFT)
- **company_name**: Full company name
- **sector**: Industry sector for P/E ratio calculation

The application will automatically load all tickers from this file. If the file is missing, it falls back to a default list.

## Troubleshooting

### SSL Certificate Issues
If you encounter SSL certificate errors with yfinance:
1. Use the demo version: `python stock_fair_value_demo.py`
2. Or update your system's SSL certificates

### Missing Data Handling
The application includes robust fallback mechanisms:
- Default FCF values based on stock price
- Conservative EPS estimates
- Realistic book value assumptions
- Industry-standard growth rates

## Limitations

- Requires internet connection for real-time data
- yfinance API rate limits may apply
- Some data points may have delays
- Conservative approach may undervalue growth stocks

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

[MIT](https://choosealicense.com/licenses/mit/)

## Disclaimer

This application is for educational and research purposes only. The valuations provided should not be used as the sole basis for investment decisions. Always conduct your own research and consult with financial professionals before making investment decisions.