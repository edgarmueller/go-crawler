package main

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
	"sync"
)

func normalizeURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return u.Host + strings.TrimSuffix(u.Path, "/"), nil
}

func (cfg *crawler) crawlPage(rawCurrentURL string) {
	cfg.concurrencyControl <- struct{}{}
	defer cfg.stopCrawl()
	if cfg.hasMaxPages() {
		return
	}

	parsedURL, err := url.Parse(rawCurrentURL)
	if err != nil {
		fmt.Println("could not parse URL:", err)
		return
	}

	if parsedURL.Hostname() != cfg.baseURL.Hostname() {
		fmt.Println("skipping URL with different host or scheme:", rawCurrentURL)
		return
	}

	normalizedURL, err := normalizeURL(rawCurrentURL)
	if err != nil {
		fmt.Println("could not normalize URL:", err)
		return
	}

	isFirst := cfg.addPageVisit(normalizedURL)
	if !isFirst {
		return
	}

	fmt.Println("crawling:", normalizedURL)
	nextURLs, err := parseURLsFromHTML(cfg.baseURL.String(), rawCurrentURL)
	if err != nil {
		fmt.Println("could not get URLs from HTML:", err)
		return
	}

	for _, nextURL := range nextURLs {
		cfg.wg.Add(1)
		go cfg.crawlPage(nextURL)
	}
}

func (cfg *crawler) stopCrawl() {
	<-cfg.concurrencyControl
	cfg.wg.Done()
}

func (cfg *crawler) addPageVisit(normalizedURL string) (isFirst bool) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	if _, exists := cfg.pages[normalizedURL]; exists {
		cfg.pages[normalizedURL] += 1
		return false
	} else {
		cfg.pages[normalizedURL] = 1
		return true
	}
}

func (cfg *crawler) hasMaxPages() bool {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	return len(cfg.pages) >= cfg.maxPages
}

type crawler struct {
	pages              map[string]int
	baseURL            *url.URL
	mu                 *sync.Mutex
	concurrencyControl chan struct{}
	wg                 *sync.WaitGroup
	maxPages           int
}

func (cfg *crawler) printReport() {
	fmt.Println("=============================")
	fmt.Println("REPORT for", cfg.baseURL.String())
	fmt.Println("=============================")
	sortedPages := sortPagesByLinkCount(cfg.pages)
	for _, entry := range sortedPages {
		fmt.Printf("Found %d internal links to %s\n", entry.count, entry.url)
	}
}

type page struct {
	url   string
	count int
}

func sortPagesByLinkCount(pages map[string]int) []page {
	sortedPages := []page{}
	for url, count := range pages {
		sortedPages = append(sortedPages, page{url: url, count: count})
	}
	sort.Slice(sortedPages, func(i, j int) bool {
		if sortedPages[i].count == sortedPages[j].count {
			return sortedPages[i].url < sortedPages[j].url
		}
		return sortedPages[i].count > sortedPages[j].count
	})
	return sortedPages
}
