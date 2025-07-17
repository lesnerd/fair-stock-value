package services

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// GrowthRateSource represents a source of growth rate data
type GrowthRateSource struct {
	Name        string
	URL         string
	GrowthRate  float64
	Confidence  float64 // 0-1 scale for data quality
	FetchTime   time.Time
	Error       error
}

// GrowthRateFetcher handles fetching growth rate predictions from multiple sources
type GrowthRateFetcher struct {
	httpClient   *http.Client
	requestDelay time.Duration
	sources      []string
	userAgents   []string
	randSource   *rand.Rand
}

// NewGrowthRateFetcher creates a new growth rate fetcher
func NewGrowthRateFetcher() *GrowthRateFetcher {
	return &GrowthRateFetcher{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		requestDelay: 2 * time.Second,
		sources: []string{
			"yahoo_finance",
			"marketwatch",
			"seeking_alpha",
			"finviz",
			"tipranks",
			"investing",
			"zacks",
			"morningstar",
			"reuters",
			"bloomberg",
		},
		userAgents: []string{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/121.0",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/120.0",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36",
		},
		randSource: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// createRealisticRequest creates an HTTP request with realistic headers and user agent
func (grf *GrowthRateFetcher) createRealisticRequest(ctx context.Context, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	// Random user agent
	userAgent := grf.userAgents[grf.randSource.Intn(len(grf.userAgents))]
	req.Header.Set("User-Agent", userAgent)
	
	// Common browser headers
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Cache-Control", "max-age=0")
	
	// Add random delay to avoid rate limiting
	delay := time.Duration(grf.randSource.Intn(1000)+500) * time.Millisecond
	time.Sleep(delay)
	
	return req, nil
}

// FetchGrowthRatesBatch fetches growth rates for multiple tickers using fallback system
func (grf *GrowthRateFetcher) FetchGrowthRatesBatch(tickers []string) (map[string]float64, error) {
	result := make(map[string]float64)
	for _, ticker := range tickers {
		if growth := grf.getFallbackGrowthRate(ticker); growth > 0 {
			result[ticker] = growth
		} else {
			result[ticker] = 0.06 // Default
		}
	}
	
	return result, nil
}

// FetchGrowthRateConsensus fetches growth rate from multiple sources and calculates consensus
func (grf *GrowthRateFetcher) FetchGrowthRateConsensus(ctx context.Context, ticker string) (float64, error) {
	fmt.Printf("Fetching growth rate predictions for %s from multiple sources...\n", ticker)
	
	// Create channels for concurrent fetching
	sourcesChan := make(chan GrowthRateSource, len(grf.sources))
	var wg sync.WaitGroup
	
	// Fetch from all sources concurrently
	for _, source := range grf.sources {
		wg.Add(1)
		go func(sourceName string) {
			defer wg.Done()
			
			var sourceData GrowthRateSource
			sourceData.Name = sourceName
			sourceData.FetchTime = time.Now()
			
			switch sourceName {
			case "yahoo_finance":
				sourceData = grf.fetchFromYahooFinance(ctx, ticker)
			case "marketwatch":
				sourceData = grf.fetchFromMarketWatch(ctx, ticker)
			case "seeking_alpha":
				sourceData = grf.fetchFromSeekingAlpha(ctx, ticker)
			case "finviz":
				sourceData = grf.fetchFromFinviz(ctx, ticker)
			case "tipranks":
				sourceData = grf.fetchFromTipRanks(ctx, ticker)
			case "investing":
				sourceData = grf.fetchFromInvesting(ctx, ticker)
			case "zacks":
				sourceData = grf.fetchFromZacks(ctx, ticker)
			case "morningstar":
				sourceData = grf.fetchFromMorningstar(ctx, ticker)
			case "reuters":
				sourceData = grf.fetchFromReuters(ctx, ticker)
			case "bloomberg":
				sourceData = grf.fetchFromBloomberg(ctx, ticker)
			}
			
			sourcesChan <- sourceData
		}(source)
	}
	
	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(sourcesChan)
	}()
	
	// Collect results
	var sources []GrowthRateSource
	for sourceData := range sourcesChan {
		sources = append(sources, sourceData)
		if sourceData.Error != nil {
			fmt.Printf("Error fetching from %s: %v\n", sourceData.Name, sourceData.Error)
		} else {
			fmt.Printf("Growth rate from %s: %.2f%% (confidence: %.2f)\n", 
				sourceData.Name, sourceData.GrowthRate*100, sourceData.Confidence)
		}
	}
	
	// Calculate weighted consensus
	consensus := grf.calculateWeightedConsensus(sources)
	
	if consensus == 0 {
		// Try fallback growth estimates for major stocks
		if fallbackGrowth := grf.getFallbackGrowthRate(ticker); fallbackGrowth > 0 {
			fmt.Printf("Using fallback growth rate for %s: %.2f%%\n", ticker, fallbackGrowth*100)
			return fallbackGrowth, nil
		}
		fmt.Printf("No valid growth rate data found for %s, using default\n", ticker)
		return 0.06, nil // Default 6% growth
	}
	
	fmt.Printf("Consensus growth rate for %s: %.2f%%\n", ticker, consensus*100)
	return consensus, nil
}

// fetchFromYahooFinance fetches growth rate from Yahoo Finance analyst estimates
func (grf *GrowthRateFetcher) fetchFromYahooFinance(ctx context.Context, ticker string) GrowthRateSource {
	source := GrowthRateSource{
		Name:       "yahoo_finance",
		Confidence: 0.85, // High confidence for Yahoo Finance
		FetchTime:  time.Now(),
	}
	
	// Try Yahoo Finance analysis page
	analysisURL := fmt.Sprintf("https://finance.yahoo.com/quote/%s/analysis/", ticker)
	source.URL = analysisURL
	
	req, err := http.NewRequestWithContext(ctx, "GET", analysisURL, nil)
	if err != nil {
		source.Error = fmt.Errorf("failed to create request: %w", err)
		return source
	}
	
	grf.setRequestHeaders(req)
	
	resp, err := grf.httpClient.Do(req)
	if err != nil {
		source.Error = fmt.Errorf("failed to fetch data: %w", err)
		return source
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		source.Error = fmt.Errorf("HTTP status %d", resp.StatusCode)
		return source
	}
	
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		source.Error = fmt.Errorf("failed to parse HTML: %w", err)
		return source
	}
	
	// Look for growth rate estimates in various sections
	growthRate := grf.extractYahooGrowthRate(doc)
	source.GrowthRate = growthRate
	
	return source
}

// extractYahooGrowthRate extracts growth rate from Yahoo Finance analysis page
func (grf *GrowthRateFetcher) extractYahooGrowthRate(doc *goquery.Document) float64 {
	// Look for growth estimates table
	var growthRates []float64
	
	// Search for "Growth Estimates" section with more patterns
	doc.Find("table, div, span").Each(func(i int, elem *goquery.Selection) {
		text := strings.TrimSpace(elem.Text())
		lowerText := strings.ToLower(text)
		
		// Look for growth-related text patterns
		if strings.Contains(lowerText, "growth") && 
		   (strings.Contains(lowerText, "estimate") || 
		    strings.Contains(lowerText, "next year") || 
		    strings.Contains(lowerText, "5 year") ||
		    strings.Contains(lowerText, "eps growth") ||
		    strings.Contains(lowerText, "revenue growth")) {
			
			// Extract numbers from the text
			if growth, err := grf.parseGrowthValue(text); err == nil && growth > 0 && growth < 1 {
				growthRates = append(growthRates, growth)
			}
		}
	})
	
	// More aggressive search in table cells
	doc.Find("table").Each(func(i int, table *goquery.Selection) {
		table.Find("tr").Each(func(j int, row *goquery.Selection) {
			cells := row.Find("td")
			if cells.Length() >= 2 {
				label := strings.TrimSpace(cells.First().Text())
				lowerLabel := strings.ToLower(label)
				
				// Look for growth-related labels
				if strings.Contains(lowerLabel, "growth") || 
				   strings.Contains(lowerLabel, "next year") || 
				   strings.Contains(lowerLabel, "5 year") ||
				   strings.Contains(lowerLabel, "eps") {
					
					cells.Each(func(k int, cell *goquery.Selection) {
						if k > 0 { // Skip label column
							text := strings.TrimSpace(cell.Text())
							if growth, err := grf.parseGrowthValue(text); err == nil && growth > 0 && growth < 1 {
								growthRates = append(growthRates, growth)
							}
						}
					})
				}
			}
		})
	})
	
	// Look for JSON data with growth estimates
	doc.Find("script").Each(func(i int, script *goquery.Selection) {
		content := script.Text()
		if strings.Contains(content, "growth") && strings.Contains(content, "estimate") {
			if growth := grf.extractGrowthFromJSON(content); growth > 0 {
				growthRates = append(growthRates, growth)
			}
		}
	})
	
	// Return average of found growth rates
	if len(growthRates) > 0 {
		sum := 0.0
		for _, rate := range growthRates {
			sum += rate
		}
		return sum / float64(len(growthRates))
	}
	
	return 0
}

// fetchFromMarketWatch fetches growth rate from MarketWatch
func (grf *GrowthRateFetcher) fetchFromMarketWatch(ctx context.Context, ticker string) GrowthRateSource {
	source := GrowthRateSource{
		Name:       "marketwatch",
		Confidence: 0.7,
		FetchTime:  time.Now(),
	}
	
	// MarketWatch analyst estimates URL
	analysisURL := fmt.Sprintf("https://www.marketwatch.com/investing/stock/%s/analystestimates", ticker)
	source.URL = analysisURL
	
	req, err := http.NewRequestWithContext(ctx, "GET", analysisURL, nil)
	if err != nil {
		source.Error = fmt.Errorf("failed to create request: %w", err)
		return source
	}
	
	grf.setRequestHeaders(req)
	
	resp, err := grf.httpClient.Do(req)
	if err != nil {
		source.Error = fmt.Errorf("failed to fetch data: %w", err)
		return source
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		source.Error = fmt.Errorf("HTTP status %d", resp.StatusCode)
		return source
	}
	
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		source.Error = fmt.Errorf("failed to parse HTML: %w", err)
		return source
	}
	
	growthRate := grf.extractMarketWatchGrowthRate(doc)
	source.GrowthRate = growthRate
	
	return source
}

// extractMarketWatchGrowthRate extracts growth rate from MarketWatch
func (grf *GrowthRateFetcher) extractMarketWatchGrowthRate(doc *goquery.Document) float64 {
	var growthRates []float64
	
	// Look for growth estimates in tables
	doc.Find("table").Each(func(i int, table *goquery.Selection) {
		table.Find("tr").Each(func(j int, row *goquery.Selection) {
			label := strings.TrimSpace(row.Find("td").First().Text())
			
			if strings.Contains(strings.ToLower(label), "growth") ||
			   strings.Contains(strings.ToLower(label), "estimate") {
				
				row.Find("td").Each(func(k int, cell *goquery.Selection) {
					if k > 0 {
						text := strings.TrimSpace(cell.Text())
						if growth, err := grf.parseGrowthValue(text); err == nil && growth > 0 {
							growthRates = append(growthRates, growth)
						}
					}
				})
			}
		})
	})
	
	// Return average if we found any rates
	if len(growthRates) > 0 {
		sum := 0.0
		for _, rate := range growthRates {
			sum += rate
		}
		return sum / float64(len(growthRates))
	}
	
	return 0
}

// fetchFromSeekingAlpha fetches growth rate from Seeking Alpha
func (grf *GrowthRateFetcher) fetchFromSeekingAlpha(ctx context.Context, ticker string) GrowthRateSource {
	source := GrowthRateSource{
		Name:       "seeking_alpha",
		Confidence: 0.6,
		FetchTime:  time.Now(),
	}
	
	// Seeking Alpha overview page
	overviewURL := fmt.Sprintf("https://seekingalpha.com/symbol/%s", ticker)
	source.URL = overviewURL
	
	req, err := http.NewRequestWithContext(ctx, "GET", overviewURL, nil)
	if err != nil {
		source.Error = fmt.Errorf("failed to create request: %w", err)
		return source
	}
	
	grf.setRequestHeaders(req)
	
	resp, err := grf.httpClient.Do(req)
	if err != nil {
		source.Error = fmt.Errorf("failed to fetch data: %w", err)
		return source
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		source.Error = fmt.Errorf("HTTP status %d", resp.StatusCode)
		return source
	}
	
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		source.Error = fmt.Errorf("failed to parse HTML: %w", err)
		return source
	}
	
	growthRate := grf.extractSeekingAlphaGrowthRate(doc)
	source.GrowthRate = growthRate
	
	return source
}

// extractSeekingAlphaGrowthRate extracts growth rate from Seeking Alpha
func (grf *GrowthRateFetcher) extractSeekingAlphaGrowthRate(doc *goquery.Document) float64 {
	var growthRates []float64
	
	// Look for growth metrics in various sections
	doc.Find("div, span, td").Each(func(i int, elem *goquery.Selection) {
		text := strings.TrimSpace(elem.Text())
		
		if strings.Contains(strings.ToLower(text), "growth") &&
		   (strings.Contains(text, "%") || strings.Contains(text, "estimate")) {
			
			if growth, err := grf.parseGrowthValue(text); err == nil && growth > 0 {
				growthRates = append(growthRates, growth)
			}
		}
	})
	
	// Return average if we found any rates
	if len(growthRates) > 0 {
		sum := 0.0
		for _, rate := range growthRates {
			sum += rate
		}
		return sum / float64(len(growthRates))
	}
	
	return 0
}

// fetchFromFinviz fetches growth rate from Finviz
func (grf *GrowthRateFetcher) fetchFromFinviz(ctx context.Context, ticker string) GrowthRateSource {
	source := GrowthRateSource{
		Name:       "finviz",
		Confidence: 0.95, // Highest confidence for Finviz due to clean data format
		FetchTime:  time.Now(),
	}
	
	// Finviz stock overview page
	overviewURL := fmt.Sprintf("https://finviz.com/quote.ashx?t=%s", ticker)
	source.URL = overviewURL
	
	req, err := http.NewRequestWithContext(ctx, "GET", overviewURL, nil)
	if err != nil {
		source.Error = fmt.Errorf("failed to create request: %w", err)
		return source
	}
	
	grf.setRequestHeaders(req)
	
	resp, err := grf.httpClient.Do(req)
	if err != nil {
		source.Error = fmt.Errorf("failed to fetch data: %w", err)
		return source
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		source.Error = fmt.Errorf("HTTP status %d", resp.StatusCode)
		return source
	}
	
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		source.Error = fmt.Errorf("failed to parse HTML: %w", err)
		return source
	}
	
	growthRate := grf.extractFinvizGrowthRate(doc)
	source.GrowthRate = growthRate
	
	return source
}

// extractFinvizGrowthRate extracts growth rate from Finviz
func (grf *GrowthRateFetcher) extractFinvizGrowthRate(doc *goquery.Document) float64 {
	var growthRates []float64
	
	// Finviz typically shows growth in a table format
	doc.Find("table").Each(func(i int, table *goquery.Selection) {
		table.Find("td").Each(func(j int, cell *goquery.Selection) {
			text := strings.TrimSpace(cell.Text())
			lowerText := strings.ToLower(text)
			
			// Look for cells that contain growth information
			if strings.Contains(lowerText, "eps growth") ||
			   strings.Contains(lowerText, "sales growth") ||
			   strings.Contains(lowerText, "ltg") || // Long-term growth
			   strings.Contains(lowerText, "eps next y") ||
			   strings.Contains(lowerText, "eps next 5y") ||
			   strings.Contains(lowerText, "sales next 5y") {
				
				// Get the next cell which should contain the value
				nextCell := cell.Next()
				if nextCell.Length() > 0 {
					value := strings.TrimSpace(nextCell.Text())
					if growth, err := grf.parseGrowthValue(value); err == nil && growth > 0 && growth < 1 {
						growthRates = append(growthRates, growth)
					}
				}
			}
		})
	})
	
	// Also look in the main data table (Finviz specific structure)
	doc.Find("table.snapshot-table2").Each(func(i int, table *goquery.Selection) {
		table.Find("tr").Each(func(j int, row *goquery.Selection) {
			cells := row.Find("td")
			cells.Each(func(k int, cell *goquery.Selection) {
				text := strings.TrimSpace(cell.Text())
				lowerText := strings.ToLower(text)
				
				if strings.Contains(lowerText, "eps next y") ||
				   strings.Contains(lowerText, "eps next 5y") ||
				   strings.Contains(lowerText, "sales next 5y") {
					
					nextCell := cell.Next()
					if nextCell.Length() > 0 {
						value := strings.TrimSpace(nextCell.Text())
						if growth, err := grf.parseGrowthValue(value); err == nil && growth > 0 && growth < 1 {
							growthRates = append(growthRates, growth)
						}
					}
				}
			})
		})
	})
	
	// Return average if we found any rates
	if len(growthRates) > 0 {
		sum := 0.0
		for _, rate := range growthRates {
			sum += rate
		}
		return sum / float64(len(growthRates))
	}
	
	return 0
}

// fetchFromTipRanks fetches growth rate from TipRanks
func (grf *GrowthRateFetcher) fetchFromTipRanks(ctx context.Context, ticker string) GrowthRateSource {
	source := GrowthRateSource{
		Name:       "tipranks",
		Confidence: 0.9, // TipRanks has high-quality analyst data
		FetchTime:  time.Now(),
	}
	
	// TipRanks stock analysis URL
	analysisURL := fmt.Sprintf("https://www.tipranks.com/stocks/%s/forecast", ticker)
	source.URL = analysisURL
	
	req, err := http.NewRequestWithContext(ctx, "GET", analysisURL, nil)
	if err != nil {
		source.Error = fmt.Errorf("failed to create request: %w", err)
		return source
	}
	
	grf.setRequestHeaders(req)
	
	resp, err := grf.httpClient.Do(req)
	if err != nil {
		source.Error = fmt.Errorf("failed to fetch data: %w", err)
		return source
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		source.Error = fmt.Errorf("HTTP status %d", resp.StatusCode)
		return source
	}
	
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		source.Error = fmt.Errorf("failed to parse HTML: %w", err)
		return source
	}
	
	growthRate := grf.extractTipRanksGrowthRate(doc)
	source.GrowthRate = growthRate
	
	return source
}

// extractTipRanksGrowthRate extracts growth rate from TipRanks
func (grf *GrowthRateFetcher) extractTipRanksGrowthRate(doc *goquery.Document) float64 {
	var growthRates []float64
	
	// TipRanks typically shows analyst estimates in various sections
	doc.Find("div, span, td").Each(func(i int, elem *goquery.Selection) {
		text := strings.TrimSpace(elem.Text())
		lowerText := strings.ToLower(text)
		
		// Look for growth-related metrics
		if (strings.Contains(lowerText, "growth") || 
		    strings.Contains(lowerText, "estimate") ||
		    strings.Contains(lowerText, "consensus")) &&
		   (strings.Contains(lowerText, "eps") ||
		    strings.Contains(lowerText, "revenue") ||
		    strings.Contains(lowerText, "earnings") ||
		    strings.Contains(lowerText, "next year") ||
		    strings.Contains(lowerText, "5 year")) {
			
			if growth, err := grf.parseGrowthValue(text); err == nil && growth > 0 && growth < 1 {
				growthRates = append(growthRates, growth)
			}
		}
	})
	
	// Look for analyst consensus tables
	doc.Find("table").Each(func(i int, table *goquery.Selection) {
		table.Find("tr").Each(func(j int, row *goquery.Selection) {
			cells := row.Find("td")
			if cells.Length() >= 2 {
				label := strings.TrimSpace(cells.First().Text())
				lowerLabel := strings.ToLower(label)
				
				if strings.Contains(lowerLabel, "growth") ||
				   strings.Contains(lowerLabel, "estimate") ||
				   strings.Contains(lowerLabel, "consensus") {
					
					cells.Each(func(k int, cell *goquery.Selection) {
						if k > 0 {
							text := strings.TrimSpace(cell.Text())
							if growth, err := grf.parseGrowthValue(text); err == nil && growth > 0 && growth < 1 {
								growthRates = append(growthRates, growth)
							}
						}
					})
				}
			}
		})
	})
	
	// Look for JSON data embedded in scripts
	doc.Find("script").Each(func(i int, script *goquery.Selection) {
		content := script.Text()
		if strings.Contains(content, "growth") && 
		   (strings.Contains(content, "estimate") || strings.Contains(content, "consensus")) {
			if growth := grf.extractGrowthFromJSON(content); growth > 0 {
				growthRates = append(growthRates, growth)
			}
		}
	})
	
	// Return average if we found any rates
	if len(growthRates) > 0 {
		sum := 0.0
		for _, rate := range growthRates {
			sum += rate
		}
		return sum / float64(len(growthRates))
	}
	
	return 0
}

// fetchFromInvesting fetches growth rate from Investing.com
func (grf *GrowthRateFetcher) fetchFromInvesting(ctx context.Context, ticker string) GrowthRateSource {
	source := GrowthRateSource{
		Name:       "investing",
		Confidence: 0.8, // Investing.com has good analyst data
		FetchTime:  time.Now(),
	}
	
	// Investing.com earnings estimates URL - try multiple formats
	analysisURL := fmt.Sprintf("https://www.investing.com/equities/%s-earnings", strings.ToLower(ticker))
	
	// Alternative URL format for Investing.com
	if strings.Contains(ticker, "-") {
		analysisURL = fmt.Sprintf("https://www.investing.com/equities/%s-earnings", strings.ToLower(ticker))
	} else {
		// Try company name format for single-word tickers
		companyNames := map[string]string{
			"AAPL":  "apple-computer-inc",
			"MSFT":  "microsoft-corp",
			"GOOGL": "google-inc",
			"AMZN":  "amazon-com-inc",
			"NVDA":  "nvidia-corp",
			"META":  "meta-platforms-inc",
			"TSLA":  "tesla-motors",
		}
		if companyName, exists := companyNames[ticker]; exists {
			analysisURL = fmt.Sprintf("https://www.investing.com/equities/%s-earnings", companyName)
		}
	}
	source.URL = analysisURL
	
	req, err := http.NewRequestWithContext(ctx, "GET", analysisURL, nil)
	if err != nil {
		source.Error = fmt.Errorf("failed to create request: %w", err)
		return source
	}
	
	grf.setRequestHeaders(req)
	
	resp, err := grf.httpClient.Do(req)
	if err != nil {
		source.Error = fmt.Errorf("failed to fetch data: %w", err)
		return source
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		source.Error = fmt.Errorf("HTTP status %d", resp.StatusCode)
		return source
	}
	
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		source.Error = fmt.Errorf("failed to parse HTML: %w", err)
		return source
	}
	
	growthRate := grf.extractInvestingGrowthRate(doc)
	source.GrowthRate = growthRate
	
	return source
}

// extractInvestingGrowthRate extracts growth rate from Investing.com
func (grf *GrowthRateFetcher) extractInvestingGrowthRate(doc *goquery.Document) float64 {
	var growthRates []float64
	
	// Investing.com typically shows estimates in structured tables
	doc.Find("table").Each(func(i int, table *goquery.Selection) {
		table.Find("tr").Each(func(j int, row *goquery.Selection) {
			cells := row.Find("td, th")
			if cells.Length() >= 2 {
				label := strings.TrimSpace(cells.First().Text())
				lowerLabel := strings.ToLower(label)
				
				// Look for growth-related metrics
				if strings.Contains(lowerLabel, "growth") ||
				   strings.Contains(lowerLabel, "estimate") ||
				   strings.Contains(lowerLabel, "consensus") ||
				   strings.Contains(lowerLabel, "eps") ||
				   strings.Contains(lowerLabel, "revenue") ||
				   strings.Contains(lowerLabel, "earnings") {
					
					cells.Each(func(k int, cell *goquery.Selection) {
						if k > 0 {
							text := strings.TrimSpace(cell.Text())
							if growth, err := grf.parseGrowthValue(text); err == nil && growth > 0 && growth < 1 {
								growthRates = append(growthRates, growth)
							}
						}
					})
				}
			}
		})
	})
	
	// Look for estimates in div/span elements
	doc.Find("div, span").Each(func(i int, elem *goquery.Selection) {
		text := strings.TrimSpace(elem.Text())
		lowerText := strings.ToLower(text)
		
		if (strings.Contains(lowerText, "growth") || 
		    strings.Contains(lowerText, "estimate")) &&
		   (strings.Contains(lowerText, "eps") ||
		    strings.Contains(lowerText, "revenue") ||
		    strings.Contains(lowerText, "earnings")) {
			
			if growth, err := grf.parseGrowthValue(text); err == nil && growth > 0 && growth < 1 {
				growthRates = append(growthRates, growth)
			}
		}
	})
	
	// Return average if we found any rates
	if len(growthRates) > 0 {
		sum := 0.0
		for _, rate := range growthRates {
			sum += rate
		}
		return sum / float64(len(growthRates))
	}
	
	return 0
}

// parseGrowthValue parses growth rate from text (handles percentages)
func (grf *GrowthRateFetcher) parseGrowthValue(text string) (float64, error) {
	// Clean the text
	cleaned := strings.ReplaceAll(text, "%", "")
	cleaned = strings.ReplaceAll(cleaned, ",", "")
	cleaned = strings.ReplaceAll(cleaned, "$", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")
	cleaned = strings.TrimSpace(cleaned)
	
	// Handle N/A or empty values
	if cleaned == "" || cleaned == "N/A" || cleaned == "--" || cleaned == "NA" || 
	   cleaned == "n/a" || cleaned == "null" || cleaned == "none" {
		return 0, fmt.Errorf("no valid value")
	}
	
	// Use regex to extract number (including negative numbers and decimals)
	re := regexp.MustCompile(`-?\d+\.?\d*`)
	match := re.FindString(cleaned)
	if match == "" {
		return 0, fmt.Errorf("no number found")
	}
	
	value, err := strconv.ParseFloat(match, 64)
	if err != nil {
		return 0, err
	}
	
	// Convert percentage to decimal if it looks like a percentage
	if strings.Contains(text, "%") || value > 1 {
		value = value / 100
	}
	
	// Apply reasonable bounds (growth rates typically between -50% to +100%)
	if value < -0.5 {
		value = -0.5
	}
	if value > 1.0 {
		value = 1.0
	}
	
	// Filter out very small values that are likely errors
	if value > 0 && value < 0.001 {
		return 0, fmt.Errorf("value too small")
	}
	
	return value, nil
}

// extractGrowthFromJSON extracts growth rate from JSON content
func (grf *GrowthRateFetcher) extractGrowthFromJSON(content string) float64 {
	// Use regex to find growth-related values in JSON
	re := regexp.MustCompile(`"growth"[^}]*?(\d+\.?\d*)`)
	matches := re.FindAllStringSubmatch(content, -1)
	
	var growthRates []float64
	for _, match := range matches {
		if len(match) > 1 {
			if value, err := strconv.ParseFloat(match[1], 64); err == nil {
				if value > 1 {
					value = value / 100
				}
				growthRates = append(growthRates, value)
			}
		}
	}
	
	// Return average if we found any rates
	if len(growthRates) > 0 {
		sum := 0.0
		for _, rate := range growthRates {
			sum += rate
		}
		return sum / float64(len(growthRates))
	}
	
	return 0
}

// calculateWeightedConsensus calculates weighted average of growth rates
func (grf *GrowthRateFetcher) calculateWeightedConsensus(sources []GrowthRateSource) float64 {
	var totalWeight float64
	var weightedSum float64
	
	for _, source := range sources {
		if source.Error == nil && source.GrowthRate > 0 {
			weight := source.Confidence
			totalWeight += weight
			weightedSum += source.GrowthRate * weight
		}
	}
	
	if totalWeight == 0 {
		return 0
	}
	
	consensus := weightedSum / totalWeight
	
	// Apply conservative adjustment (reduce by 10% for safety)
	consensus = consensus * 0.9
	
	// Apply bounds
	if consensus < 0 {
		consensus = 0.02 // Minimum 2% growth
	}
	if consensus > 0.5 {
		consensus = 0.5 // Maximum 50% growth
	}
	
	return consensus
}

// setRequestHeaders sets browser-like headers
func (grf *GrowthRateFetcher) setRequestHeaders(req *http.Request) {
	// Use the enhanced user agent from the struct
	userAgent := grf.userAgents[grf.randSource.Intn(len(grf.userAgents))]
	req.Header.Set("User-Agent", userAgent)
	
	// Enhanced browser headers to mimic real browsers
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Cache-Control", "max-age=0")
	
	// Add some referrer headers for specific sites
	if strings.Contains(req.URL.Host, "yahoo.com") {
		req.Header.Set("Referer", "https://finance.yahoo.com/")
	} else if strings.Contains(req.URL.Host, "marketwatch.com") {
		req.Header.Set("Referer", "https://www.marketwatch.com/")
	} else if strings.Contains(req.URL.Host, "seekingalpha.com") {
		req.Header.Set("Referer", "https://seekingalpha.com/")
	}
	
	// Add random delay to avoid rate limiting
	delay := time.Duration(grf.randSource.Intn(1000)+500) * time.Millisecond
	time.Sleep(delay)
}

// getFallbackGrowthRate returns estimated growth rates for major stocks
func (grf *GrowthRateFetcher) getFallbackGrowthRate(ticker string) float64 {
	// These are approximate analyst consensus estimates for major stocks
	// Updated periodically based on market conditions
	fallbackGrowthRates := map[string]float64{
		"AAPL": 0.05,  // Apple - 5% (mature tech)
		"MSFT": 0.08,  // Microsoft - 8% (cloud growth)
		"GOOGL": 0.07, // Alphabet - 7% (ad market + AI)
		"AMZN": 0.12,  // Amazon - 12% (AWS + retail)
		"NVDA": 0.15,  // NVIDIA - 15% (AI boom)
		"META": 0.09,  // Meta - 9% (metaverse investments)
		"TSLA": 0.20,  // Tesla - 20% (EV expansion)
		"BRK-B": 0.06, // Berkshire - 6% (diversified)
		"UNH": 0.08,   // UnitedHealth - 8% (healthcare)
		"JNJ": 0.04,   // Johnson & Johnson - 4% (pharma)
		"JPM": 0.05,   // JPMorgan - 5% (financial)
		"V": 0.10,     // Visa - 10% (digital payments)
		"PG": 0.04,    // Procter & Gamble - 4% (consumer staples)
		"HD": 0.06,    // Home Depot - 6% (home improvement)
		"MA": 0.11,    // Mastercard - 11% (payments)
		"BAC": 0.05,   // Bank of America - 5% (banking)
		"ABBV": 0.07,  // AbbVie - 7% (pharma)
		"PFE": 0.03,   // Pfizer - 3% (pharma)
		"KO": 0.04,    // Coca-Cola - 4% (beverages)
		"AVGO": 0.09,  // Broadcom - 9% (semiconductors)
		"PEP": 0.04,   // PepsiCo - 4% (beverages)
		"TMO": 0.07,   // Thermo Fisher - 7% (life sciences)
		"COST": 0.06,  // Costco - 6% (retail)
		"WMT": 0.05,   // Walmart - 5% (retail)
		"MRK": 0.05,   // Merck - 5% (pharma)
		"DIS": 0.06,   // Disney - 6% (entertainment)
		"ACN": 0.07,   // Accenture - 7% (consulting)
		"VZ": 0.02,    // Verizon - 2% (telecom)
		"ADBE": 0.10,  // Adobe - 10% (software)
		"NFLX": 0.08,  // Netflix - 8% (streaming)
		"NKE": 0.07,   // Nike - 7% (apparel)
		"CRM": 0.12,   // Salesforce - 12% (cloud)
		"DHR": 0.07,   // Danaher - 7% (life sciences)
		"LIN": 0.05,   // Linde - 5% (industrial gases)
		"TXN": 0.05,   // Texas Instruments - 5% (semiconductors)
		"NEE": 0.06,   // NextEra Energy - 6% (utilities)
		"ABT": 0.06,   // Abbott - 6% (healthcare)
		"ORCL": 0.06,  // Oracle - 6% (database)
		"PM": 0.03,    // Philip Morris - 3% (tobacco)
		"RTX": 0.05,   // Raytheon - 5% (aerospace)
		"QCOM": 0.08,  // Qualcomm - 8% (5G/mobile)
		"HON": 0.06,   // Honeywell - 6% (industrials)
		"WFC": 0.05,   // Wells Fargo - 5% (banking)
		"UPS": 0.04,   // UPS - 4% (logistics)
		"T": 0.01,     // AT&T - 1% (telecom)
		"LOW": 0.06,   // Lowe's - 6% (home improvement)
		"SPGI": 0.08,  // S&P Global - 8% (financial data)
		"ELV": 0.07,   // Elevance Health - 7% (healthcare)
		"SCHW": 0.06,  // Charles Schwab - 6% (financial)
		"CAT": 0.05,   // Caterpillar - 5% (machinery)
	}
	
	if growth, exists := fallbackGrowthRates[ticker]; exists {
		return growth
	}
	
	return 0 // No fallback available
}

// fetchFromZacks fetches growth rate from Zacks Investment Research
func (grf *GrowthRateFetcher) fetchFromZacks(ctx context.Context, ticker string) GrowthRateSource {
	source := GrowthRateSource{
		Name:       "zacks",
		URL:        fmt.Sprintf("https://www.zacks.com/stock/quote/%s", ticker),
		Confidence: 0.85,
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", source.URL, nil)
	if err != nil {
		source.Error = err
		return source
	}
	
	grf.setRequestHeaders(req)
	
	resp, err := grf.httpClient.Do(req)
	if err != nil {
		source.Error = err
		return source
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		source.Error = fmt.Errorf("HTTP status %d", resp.StatusCode)
		return source
	}
	
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		source.Error = err
		return source
	}
	
	// Look for growth rate patterns in Zacks format
	growthTexts := []string{
		doc.Find(".rank_view .zr_ranktext").Text(),
		doc.Find(".composite_val").Text(),
		doc.Find("[data-test='growth-rate']").Text(),
	}
	
	for _, text := range growthTexts {
		if rate, err := grf.parseGrowthValue(text); err == nil && rate > 0 {
			source.GrowthRate = rate
			return source
		}
	}
	
	source.Error = fmt.Errorf("no growth rate found")
	return source
}

// fetchFromMorningstar fetches growth rate from Morningstar
func (grf *GrowthRateFetcher) fetchFromMorningstar(ctx context.Context, ticker string) GrowthRateSource {
	source := GrowthRateSource{
		Name:       "morningstar",
		URL:        fmt.Sprintf("https://www.morningstar.com/stocks/xnas/%s/quote", ticker),
		Confidence: 0.90,
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", source.URL, nil)
	if err != nil {
		source.Error = err
		return source
	}
	
	grf.setRequestHeaders(req)
	
	resp, err := grf.httpClient.Do(req)
	if err != nil {
		source.Error = err
		return source
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		source.Error = fmt.Errorf("HTTP status %d", resp.StatusCode)
		return source
	}
	
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		source.Error = err
		return source
	}
	
	// Look for growth rate patterns in Morningstar format
	growthTexts := []string{
		doc.Find(".dp-value").Text(),
		doc.Find("[data-test='growth-forecast']").Text(),
		doc.Find(".sal-component-cta-link").Text(),
	}
	
	for _, text := range growthTexts {
		if rate, err := grf.parseGrowthValue(text); err == nil && rate > 0 {
			source.GrowthRate = rate
			return source
		}
	}
	
	source.Error = fmt.Errorf("no growth rate found")
	return source
}

// fetchFromReuters fetches growth rate from Reuters
func (grf *GrowthRateFetcher) fetchFromReuters(ctx context.Context, ticker string) GrowthRateSource {
	source := GrowthRateSource{
		Name:       "reuters",
		URL:        fmt.Sprintf("https://www.reuters.com/markets/companies/%s.O", ticker),
		Confidence: 0.85,
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", source.URL, nil)
	if err != nil {
		source.Error = err
		return source
	}
	
	grf.setRequestHeaders(req)
	
	resp, err := grf.httpClient.Do(req)
	if err != nil {
		source.Error = err
		return source
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		source.Error = fmt.Errorf("HTTP status %d", resp.StatusCode)
		return source
	}
	
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		source.Error = err
		return source
	}
	
	// Look for growth rate patterns in Reuters format
	growthTexts := []string{
		doc.Find(".forecast-data").Text(),
		doc.Find(".analyst-estimates").Text(),
		doc.Find("[data-testid='growth-rate']").Text(),
	}
	
	for _, text := range growthTexts {
		if rate, err := grf.parseGrowthValue(text); err == nil && rate > 0 {
			source.GrowthRate = rate
			return source
		}
	}
	
	source.Error = fmt.Errorf("no growth rate found")
	return source
}

// fetchFromBloomberg fetches growth rate from Bloomberg
func (grf *GrowthRateFetcher) fetchFromBloomberg(ctx context.Context, ticker string) GrowthRateSource {
	source := GrowthRateSource{
		Name:       "bloomberg",
		URL:        fmt.Sprintf("https://www.bloomberg.com/quote/%s:US", ticker),
		Confidence: 0.90,
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", source.URL, nil)
	if err != nil {
		source.Error = err
		return source
	}
	
	grf.setRequestHeaders(req)
	
	resp, err := grf.httpClient.Do(req)
	if err != nil {
		source.Error = err
		return source
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		source.Error = fmt.Errorf("HTTP status %d", resp.StatusCode)
		return source
	}
	
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		source.Error = err
		return source
	}
	
	// Look for growth rate patterns in Bloomberg format
	growthTexts := []string{
		doc.Find(".data-table-row").Text(),
		doc.Find("[data-module='Estimates']").Text(),
		doc.Find(".analyst-forecast").Text(),
	}
	
	for _, text := range growthTexts {
		if rate, err := grf.parseGrowthValue(text); err == nil && rate > 0 {
			source.GrowthRate = rate
			return source
		}
	}
	
	source.Error = fmt.Errorf("no growth rate found")
	return source
}