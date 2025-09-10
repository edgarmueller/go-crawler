package main

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func getImagesFromHTML(htmlBody string, baseURL *url.URL) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlBody))
	if err != nil {
		return nil, fmt.Errorf("couldn't parse HTML: %w", err)
	}

	var imageURLs []string
	doc.Find("img").Each(func(_ int, s *goquery.Selection) {
		src, ok := s.Attr("src")
		if !ok || strings.TrimSpace(src) == "" {
			return
		}

		u, err := url.Parse(src)
		if err != nil {
			fmt.Printf("couldn't parse src %q: %v\n", src, err)
			return
		}

		absolute := baseURL.ResolveReference(u)
		imageURLs = append(imageURLs, absolute.String())
	})

	return imageURLs, nil
}

func getURLsFromHTML(htmlBody string, baseURL *url.URL) ([]string, error) {
	if baseURL == nil {
		return nil, fmt.Errorf("invalid base URL")
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlBody))
	if err != nil {
		return nil, fmt.Errorf("couldn't parse HTML: %w", err)
	}

	var urls []string
	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if !ok {
			return
		}
		href = strings.TrimSpace(href)
		if href == "" {
			return
		}

		u, err := url.Parse(href)
		if err != nil {
			fmt.Printf("couldn't parse href %q: %v\n", href, err)
			return
		}

		resolved := baseURL.ResolveReference(u)
		urls = append(urls, resolved.String())
	})

	return urls, nil
}

func getH1FromHTML(html string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return ""
	}
	h1 := doc.Find("h1").First().Text()
	return strings.TrimSpace(h1)
}

func getFirstParagraphFromHTML(html string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return ""
	}

	main := doc.Find("main")
	var p string
	if main.Length() > 0 {
		p = main.Find("p").First().Text()
	} else {
		p = doc.Find("p").First().Text()
	}

	return strings.TrimSpace(p)
}

type PageData struct {
	URL            string
	H1             string
	FirstParagraph string
	OutgoingLinks  []string
	ImageURLs      []string
}

func extractPageData(html, pageURL string) PageData {
	h1 := getH1FromHTML(html)
	firstParagraph := getFirstParagraphFromHTML(html)

	// Parse the page URL once
	parsedURL, err := url.Parse(pageURL)
	if err != nil {
		// If it's invalid, bail gracefully with minimal data
		return PageData{
			URL:            pageURL,
			H1:             h1,
			FirstParagraph: firstParagraph,
			OutgoingLinks:  nil,
			ImageURLs:      nil,
		}
	}

	outgoingLinks, err := getURLsFromHTML(html, parsedURL)
	if err != nil {
		outgoingLinks = nil
	}

	imageURLs, err := getImagesFromHTML(html, parsedURL)
	if err != nil {
		imageURLs = nil
	}

	return PageData{
		URL:            pageURL,
		H1:             h1,
		FirstParagraph: firstParagraph,
		OutgoingLinks:  outgoingLinks,
		ImageURLs:      imageURLs,
	}
}
