#!/usr/bin/env python3
"""
Stock Fair Value Estimation Application - Demo Version
=====================================================

This demo version uses sample realistic data to demonstrate the valuation methodology.
The full version would fetch real-time data from yfinance.

Features:
- 60% DCF + 40% Comps hybrid valuation
- Realistic sample data for major Fortune 500 stocks
- Terminal-based output with clear formatting
- Scheduling capability for daily 9:00 AM ET runs

Requirements:
- pip install schedule pytz tabulate

Usage:
- python stock_fair_value_demo.py (runs once)
- python stock_fair_value_demo.py --schedule (runs daily at 9:00 AM ET)
"""

import schedule
import time
import pytz
from datetime import datetime
from tabulate import tabulate
import sys
import argparse
import csv
import os
import statistics
import random

class StockValuatorDemo:
    """
    Demo version of stock valuation using hybrid DCF + Comps methodology
    """
    
    def __init__(self):
        self.discount_rate = 0.12  # 12% discount rate for DCF
        self.max_growth_rate = 0.08  # 8% max growth rate cap
        self.terminal_growth_rate = 0.08  # 8% terminal growth rate
        self.dcf_weight = 0.6  # 60% weight for DCF
        self.comps_weight = 0.4  # 40% weight for Comps
        
        # Conservative industry-specific P/E ratios (more conservative than market)
        self.default_industry_pe_ratios = {
            'Technology': 22,          # Reduced from 25
            'Healthcare': 18,          # Reduced from 20
            'Financial Services': 10,  # Reduced from 12
            'Consumer Cyclical': 16,   # Reduced from 18
            'Consumer Defensive': 20,  # Reduced from 22
            'Energy': 12,             # Reduced from 15
            'Industrials': 13,        # Reduced from 15
            'Materials': 12,          # Reduced from 14
            'Real Estate': 14,        # Reduced from 16
            'Utilities': 16,          # Reduced from 18
            'Communication Services': 18,  # Reduced from 20
            'Default': 18             # Reduced from 20
        }
        
        # Cache for P/E ratios to avoid repeated API calls
        self.pe_ratio_cache = {}
        
        # Load Fortune 500 stock tickers from CSV file
        self.available_tickers = self.load_tickers_from_csv()
        
        # Sample realistic data for major Fortune 500 stocks
        self.sample_stock_data = {
            'AAPL': {
                'ticker': 'AAPL',
                'current_price': 170.10,
                'fcf_per_share': 6.23,
                'eps': 6.13,
                'book_value': 4.40,
                'sector': 'Technology',
                'growth_rate': 0.07,
                'company_name': 'Apple Inc.'
            },
            'MSFT': {
                'ticker': 'MSFT',
                'current_price': 415.26,
                'fcf_per_share': 11.05,
                'eps': 12.05,
                'book_value': 30.59,
                'sector': 'Technology',
                'growth_rate': 0.08,
                'company_name': 'Microsoft Corporation'
            },
            'GOOGL': {
                'ticker': 'GOOGL',
                'current_price': 179.58,
                'fcf_per_share': 5.44,
                'eps': 6.31,
                'book_value': 73.97,
                'sector': 'Communication Services',
                'growth_rate': 0.06,
                'company_name': 'Alphabet Inc.'
            },
            'AMZN': {
                'ticker': 'AMZN',
                'current_price': 223.22,
                'fcf_per_share': 3.17,
                'eps': 3.01,
                'book_value': 20.28,
                'sector': 'Consumer Cyclical',
                'growth_rate': 0.08,
                'company_name': 'Amazon.com Inc.'
            },
            'NVDA': {
                'ticker': 'NVDA',
                'current_price': 892.54,
                'fcf_per_share': 20.83,
                'eps': 24.69,
                'book_value': 34.58,
                'sector': 'Technology',
                'growth_rate': 0.08,
                'company_name': 'NVIDIA Corporation'
            },
            'META': {
                'ticker': 'META',
                'current_price': 563.34,
                'fcf_per_share': 18.69,
                'eps': 19.29,
                'book_value': 54.28,
                'sector': 'Communication Services',
                'growth_rate': 0.07,
                'company_name': 'Meta Platforms Inc.'
            },
            'TSLA': {
                'ticker': 'TSLA',
                'current_price': 248.50,
                'fcf_per_share': 2.53,
                'eps': 3.12,
                'book_value': 29.76,
                'sector': 'Consumer Cyclical',
                'growth_rate': 0.08,
                'company_name': 'Tesla Inc.'
            },
            'BRK-B': {
                'ticker': 'BRK-B',
                'current_price': 462.67,
                'fcf_per_share': 28.92,
                'eps': 22.85,
                'book_value': 330.44,
                'sector': 'Financial Services',
                'growth_rate': 0.05,
                'company_name': 'Berkshire Hathaway Inc.'
            },
            'UNH': {
                'ticker': 'UNH',
                'current_price': 590.14,
                'fcf_per_share': 25.73,
                'eps': 27.45,
                'book_value': 128.56,
                'sector': 'Healthcare',
                'growth_rate': 0.07,
                'company_name': 'UnitedHealth Group Inc.'
            },
            'JNJ': {
                'ticker': 'JNJ',
                'current_price': 155.39,
                'fcf_per_share': 6.34,
                'eps': 7.75,
                'book_value': 29.85,
                'sector': 'Healthcare',
                'growth_rate': 0.04,
                'company_name': 'Johnson & Johnson'
            }
        }
        
    def load_tickers_from_csv(self, filename='fortune_500_tickers.csv'):
        """
        Load Fortune 500 stock tickers from CSV file
        
        Args:
            filename (str): Name of the CSV file containing tickers
            
        Returns:
            list: List of stock ticker symbols
        """
        tickers = []
        csv_path = os.path.join(os.path.dirname(__file__), filename)
        
        try:
            with open(csv_path, 'r', newline='') as csvfile:
                reader = csv.DictReader(csvfile)
                for row in reader:
                    tickers.append(row['ticker'])
            
            print(f"Loaded {len(tickers)} tickers from {filename}")
            return tickers
            
        except FileNotFoundError:
            print(f"Warning: {filename} not found. Using sample data only.")
            return list(self.sample_stock_data.keys())
        except Exception as e:
            print(f"Error loading {filename}: {str(e)}. Using sample data only.")
            return list(self.sample_stock_data.keys())
    
    def get_sample_pe_ratio(self, ticker):
        """
        Get sample P/E ratio for demo stocks - simulates real-time fetching
        
        Args:
            ticker (str): Stock ticker symbol
            
        Returns:
            float: Sample P/E ratio
        """
        # Simulate multi-source P/E fetching for demonstration
        return self.simulate_multi_source_pe_fetch(ticker)
    
    def simulate_multi_source_pe_fetch(self, ticker):
        """
        Simulate fetching P/E ratios from multiple sources for demonstration
        
        Args:
            ticker (str): Stock ticker symbol
            
        Returns:
            float: Aggregated P/E ratio from simulated sources
        """
        # Base P/E ratios (realistic market values)
        base_pe_ratios = {
            'AAPL': 28.5, 'MSFT': 32.1, 'GOOGL': 22.8, 'AMZN': 45.2, 'NVDA': 65.3,
            'META': 24.1, 'TSLA': 58.7, 'BRK-B': 9.8, 'UNH': 18.5, 'JNJ': 15.2
        }
        
        base_pe = base_pe_ratios.get(ticker, 20.0)
        
        # Simulate different sources with slight variations (realistic scenario)
        simulated_sources = {
            'YFinance': base_pe * random.uniform(0.95, 1.05),     # ±5% variation
            'Alpha Vantage': base_pe * random.uniform(0.92, 1.08), # ±8% variation
            'FMP': base_pe * random.uniform(0.96, 1.04),          # ±4% variation
            'Finviz': base_pe * random.uniform(0.94, 1.06),       # ±6% variation
        }
        
        print(f"Multi-source P/E fetch for {ticker}:")
        pe_ratios = []
        
        for source, pe_value in simulated_sources.items():
            # Simulate some sources failing (realistic)
            if random.random() > 0.2:  # 80% success rate per source
                pe_ratios.append(pe_value)
                print(f"  {source}: {pe_value:.2f}")
            else:
                print(f"  {source}: Failed")
        
        if not pe_ratios:
            print(f"  All sources failed for {ticker}, using fallback")
            stock_data = self.sample_stock_data.get(ticker, {})
            sector = stock_data.get('sector', 'Default')
            return self.default_industry_pe_ratios.get(sector, self.default_industry_pe_ratios['Default'])
        
        # Apply wisdom of the crowd aggregation
        aggregated_pe = self.aggregate_pe_ratios(pe_ratios)
        
        # Apply conservative factor (15% discount)
        conservative_pe = aggregated_pe * 0.85
        conservative_pe = max(5, min(conservative_pe, 40))  # Apply bounds
        
        print(f"  Aggregated: {aggregated_pe:.2f} -> Conservative: {conservative_pe:.2f}")
        return conservative_pe
    
    def aggregate_pe_ratios(self, pe_ratios):
        """
        Aggregate multiple P/E ratios using wisdom of the crowd
        
        Args:
            pe_ratios (list): List of P/E ratios from different sources
            
        Returns:
            float: Aggregated P/E ratio
        """
        if not pe_ratios:
            return None
        
        # Remove outliers using IQR method for 3+ values
        if len(pe_ratios) >= 3:
            q1 = statistics.quantiles(pe_ratios, n=4)[0]
            q3 = statistics.quantiles(pe_ratios, n=4)[2]
            iqr = q3 - q1
            lower_bound = q1 - 1.5 * iqr
            upper_bound = q3 + 1.5 * iqr
            
            filtered_ratios = [pe for pe in pe_ratios if lower_bound <= pe <= upper_bound]
            if filtered_ratios:
                pe_ratios = filtered_ratios
                print(f"    Filtered outliers: {len(pe_ratios)} values remaining")
        
        # Use weighted average (60% median + 40% mean for robustness)
        if len(pe_ratios) == 1:
            return pe_ratios[0]
        elif len(pe_ratios) == 2:
            return statistics.mean(pe_ratios)
        else:
            median_pe = statistics.median(pe_ratios)
            mean_pe = statistics.mean(pe_ratios)
            weighted_pe = 0.6 * median_pe + 0.4 * mean_pe
            print(f"    Wisdom of crowd: Median={median_pe:.2f}, Mean={mean_pe:.2f}, Weighted={weighted_pe:.2f}")
            return weighted_pe
    
    def calculate_dcf_value(self, stock_data):
        """
        Calculate fair value using Discounted Cash Flow (DCF) model
        
        Args:
            stock_data (dict): Stock financial data
            
        Returns:
            float: DCF-based fair value per share
        """
        try:
            fcf_per_share = stock_data['fcf_per_share']
            growth_rate = min(stock_data['growth_rate'], self.max_growth_rate)
            
            # If FCF is negative or zero, use a conservative estimate
            if fcf_per_share <= 0:
                fcf_per_share = 2.0  # Conservative fallback
            
            # Project FCF for 5 years
            projected_fcf = []
            for year in range(1, 6):
                fcf = fcf_per_share * ((1 + growth_rate) ** year)
                projected_fcf.append(fcf)
            
            # Calculate present value of projected FCF
            pv_fcf = sum([fcf / ((1 + self.discount_rate) ** (i + 1)) 
                         for i, fcf in enumerate(projected_fcf)])
            
            # Calculate terminal value using Gordon Growth Model
            terminal_fcf = projected_fcf[-1] * (1 + self.terminal_growth_rate)
            terminal_value = terminal_fcf / (self.discount_rate - self.terminal_growth_rate)
            pv_terminal_value = terminal_value / ((1 + self.discount_rate) ** 5)
            
            # Total DCF value
            dcf_value = pv_fcf + pv_terminal_value
            
            return max(dcf_value, stock_data['book_value'])  # Use book value as floor
            
        except Exception as e:
            print(f"Error calculating DCF for {stock_data['ticker']}: {str(e)}")
            return stock_data['book_value']
    
    def calculate_comps_value(self, stock_data):
        """
        Calculate fair value using Comparable Company Analysis with dynamic P/E ratios
        
        Args:
            stock_data (dict): Stock financial data
            
        Returns:
            float: Comps-based fair value per share
        """
        try:
            eps = stock_data['eps']
            ticker = stock_data['ticker']
            
            # Get sample P/E ratio (simulates real-time fetching)
            pe_ratio = self.get_sample_pe_ratio(ticker)
            
            # If EPS is negative, use a conservative approach
            if eps <= 0:
                eps = 1.0  # Conservative fallback
            
            # Calculate value using P/E multiple
            comps_value = eps * pe_ratio
            
            return max(comps_value, stock_data['book_value'])  # Use book value as floor
            
        except Exception as e:
            print(f"Error calculating Comps for {stock_data['ticker']}: {str(e)}")
            return stock_data['book_value']
    
    def calculate_fair_value(self, stock_data):
        """
        Calculate hybrid fair value using weighted DCF and Comps
        
        Args:
            stock_data (dict): Stock financial data
            
        Returns:
            float: Hybrid fair value per share
        """
        dcf_value = self.calculate_dcf_value(stock_data)
        comps_value = self.calculate_comps_value(stock_data)
        
        # Weighted average: 60% DCF + 40% Comps
        fair_value = (dcf_value * self.dcf_weight) + (comps_value * self.comps_weight)
        
        # Ensure fair value is not below book value (conservative floor)
        return max(fair_value, stock_data['book_value'])
    
    def determine_status(self, fair_value, current_price):
        """
        Determine if stock is overpriced or underpriced
        
        Args:
            fair_value (float): Calculated fair value
            current_price (float): Current market price
            
        Returns:
            str: 'Underpriced' or 'Overpriced'
        """
        return 'Underpriced' if current_price < fair_value else 'Overpriced'
    
    def run_valuation(self):
        """
        Run valuation for sample stocks and display results
        """
        print("Starting stock valuation analysis...")
        print("Using sample Fortune 500 stock data...")
        print(f"Total tickers available in CSV: {len(self.available_tickers)}")
        print(f"Sample data available for: {len(self.sample_stock_data)} stocks")
        
        results = []
        
        for ticker, stock_data in self.sample_stock_data.items():
            print(f"Processing {ticker} - {stock_data['company_name']}")
            
            # Calculate fair value
            fair_value = self.calculate_fair_value(stock_data)
            
            # Determine status
            status = self.determine_status(fair_value, stock_data['current_price'])
            
            # Calculate percentage difference
            price_diff = ((fair_value - stock_data['current_price']) / stock_data['current_price']) * 100
            
            results.append({
                'Ticker': ticker,
                'Company': stock_data['company_name'][:25] + '...' if len(stock_data['company_name']) > 25 else stock_data['company_name'],
                'Fair Value': f"${fair_value:.2f}",
                'Current Price': f"${stock_data['current_price']:.2f}",
                'Book Value': f"${stock_data['book_value']:.2f}",
                'Status': status,
                'Diff %': f"{price_diff:+.1f}%"
            })
        
        # Display results
        self.display_results(results)
        
        return results
    
    def display_results(self, results):
        """
        Display valuation results in a formatted table
        
        Args:
            results (list): List of valuation results
        """
        # Get current time in Eastern Time
        et_tz = pytz.timezone('US/Eastern')
        current_time = datetime.now(et_tz)
        
        print("\n" + "="*100)
        print(f"Stock Fair Value Analysis - {current_time.strftime('%Y-%m-%d %H:%M:%S ET')}")
        print("="*100)
        print("Methodology: 60% DCF + 40% Comparable Company Analysis")
        print("="*100)
        
        # Create table
        table = tabulate(results, headers="keys", tablefmt="grid")
        print(table)
        
        # Summary statistics
        underpriced_count = sum(1 for r in results if r['Status'] == 'Underpriced')
        overpriced_count = len(results) - underpriced_count
        
        print(f"\nSummary:")
        print(f"Total stocks analyzed: {len(results)}")
        print(f"Underpriced: {underpriced_count}")
        print(f"Overpriced: {overpriced_count}")
        print("\nNote: This demo uses sample data. The full version fetches real-time data from yfinance.")
        print("="*100)

def run_scheduled_analysis():
    """
    Run the stock valuation analysis (used for scheduling)
    """
    print(f"Running scheduled stock valuation at {datetime.now()}")
    valuator = StockValuatorDemo()
    valuator.run_valuation()

def main():
    """
    Main function to handle command line arguments and execution
    """
    parser = argparse.ArgumentParser(description='Stock Fair Value Estimation - Demo Version')
    parser.add_argument('--schedule', action='store_true', 
                       help='Run scheduled analysis daily at 9:00 AM ET')
    
    args = parser.parse_args()
    
    if args.schedule:
        # Schedule daily run at 9:00 AM ET
        et_tz = pytz.timezone('US/Eastern')
        schedule.every().day.at("09:00").do(run_scheduled_analysis)
        
        print("Stock valuation scheduled to run daily at 9:00 AM ET")
        print("Press Ctrl+C to stop")
        
        try:
            while True:
                schedule.run_pending()
                time.sleep(60)  # Check every minute
        except KeyboardInterrupt:
            print("\nScheduled analysis stopped.")
    else:
        # Run once immediately
        valuator = StockValuatorDemo()
        valuator.run_valuation()

if __name__ == "__main__":
    main()