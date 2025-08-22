package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

func parseURLsFromHTML(baseURL, url string) ([]string, error) {
	html, err := getHTML(url)
	if err != nil {
		fmt.Println("could not get HTML:", err)
		return nil, err
	}

	return getURLsFromHTML(html, baseURL)
}

func getHTML(rawURL string) (string, error) {
	resp, err := http.Get(rawURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return "", errors.New("client error: " + resp.Status)
	}
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		return "", errors.New("must be content-type text/html")
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func getURLsFromHTML(htmlBody, rawBaseURL string) ([]string, error) {
	baseURL := strings.TrimSuffix(rawBaseURL, "/")
	doc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		return nil, err
	}
	var urls []string
	traverse(doc, baseURL, &urls)
	return urls, nil
}

func traverse(n *html.Node, baseURL string, urls *[]string) {
	if n == nil {
		return
	}
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {
			if attr.Key == "href" {
				url := attr.Val
				if strings.HasPrefix(url, "/") {
					url, err := normalizeURL(url)
					if err != nil {
						continue
					}
					*urls = append(*urls, baseURL+url)
				} else {
					*urls = append(*urls, strings.TrimPrefix(url, "/"))
				}
				break
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		traverse(c, baseURL, urls)
	}
}
