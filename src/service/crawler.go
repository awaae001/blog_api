package service

import (
	"blog_api/src/model"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// CrawlWebsite fetches and parses a website to extract SEO information.
func CrawlWebsite(url string) model.CrawlResult {
	client := &http.Client{
		Timeout: 10 * time.Second, // Set a timeout to prevent hanging
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Do not follow redirects
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		log.Printf("[Crawler]Error fetching URL %s: %v", url, err)
		return model.CrawlResult{Status: "timeout"}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		redirectURL := resp.Header.Get("Location")
		log.Printf("[Crawler]Redirect detected for %s to %s", url, redirectURL)
		return model.CrawlResult{Status: "survival", RedirectURL: redirectURL}
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("[Crawler]Error: Non-200 status code for %s: %d", url, resp.StatusCode)
		return model.CrawlResult{Status: "error"}
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("[Crawler]Error parsing HTML for %s: %v", url, err)
		return model.CrawlResult{Status: "error"}
	}

	// Find the description
	description := doc.Find("meta[name='description']").AttrOr("content", "")

	// Find the web icon
	iconURL, exists := doc.Find("link[rel='icon']").Attr("href")
	if !exists {
		// Fallback for apple-touch-icon or shortcut icon
		iconURL, exists = doc.Find("link[rel='apple-touch-icon']").Attr("href")
		if !exists {
			iconURL = doc.Find("link[rel='shortcut icon']").AttrOr("href", "")
		}
	}

	// Find RSS feeds
	var rssURLs []string
	doc.Find("link[rel='alternate']").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			// Check if the type is related to RSS or Atom
			linkType, _ := s.Attr("type")
			if linkType == "application/rss+xml" || linkType == "application/atom+xml" {
				// Resolve relative URL to absolute
				absoluteURL := toAbsoluteURL(resp.Request.URL, href)
				if absoluteURL != "" {
					rssURLs = append(rssURLs, absoluteURL)
				}
			}
		}
	})

	return model.CrawlResult{
		Description: description,
		IconURL:     iconURL,
		Status:      "survival",
		RssURLs:     rssURLs,
	}
}

// toAbsoluteURL converts a relative URL to an absolute URL based on the base URL.
func toAbsoluteURL(base *url.URL, href string) string {
	relativeURL, err := url.Parse(href)
	if err != nil {
		return ""
	}
	return base.ResolveReference(relativeURL).String()
}
