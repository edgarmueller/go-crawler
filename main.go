package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"sync"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("no website provided")
		os.Exit(1)
	}
	if len(os.Args) > 4 {
		fmt.Println("too many arguments provided")
	}
	baseURL := os.Args[1]
	// Default values
	maxCurrencyStr := "3" // Default max concurrency
	maxPagesStr := "10"   // Default max pages

	if len(os.Args) > 2 {
		maxCurrencyStr = os.Args[2]
	}
	if len(os.Args) > 3 {
		maxPagesStr = os.Args[3]
	}

	maxCurrency, err := strconv.Atoi(maxCurrencyStr)
	if err != nil {
		fmt.Println("invalid maxCurrency value:", err)
		os.Exit(1)
	}
	maxPages, err := strconv.Atoi(maxPagesStr)
	if err != nil {
		fmt.Println("invalid maxCurrency value:", err)
		os.Exit(1)
	}

	fmt.Println("running with max concurrency:", maxCurrency, "and max pages:", maxPages)
	fmt.Println("starting crawl of:", baseURL)
	pages := make(map[string]int)
	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil {
		fmt.Println("invalid base URL:", err)
		os.Exit(1)
	}
	crawler := &crawler{
		pages:              pages,
		baseURL:            parsedBaseURL,
		mu:                 &sync.Mutex{},
		concurrencyControl: make(chan struct{}, maxCurrency),
		wg:                 &sync.WaitGroup{},
		maxPages:           maxPages,
	}

	crawler.wg.Add(1)
	crawler.crawlPage(baseURL)
	crawler.wg.Wait()

	crawler.printReport()
}
