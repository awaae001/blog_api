package model

// CrawlResult holds the data extracted from a website.
type CrawlResult struct {
	Description string
	IconURL     string
	Status      string // e.g., "survival", "timeout", "error"
}
