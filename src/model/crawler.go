package model

// CrawlResult holds the data extracted from a website.
type CrawlResult struct {
	Description string
	IconURL     string
	Status      string          // e.g., "survival", "timeout", "error"
	RedirectURL string          // The new URL if a redirect occurs
	Feeds       []FeedCandidate // Feeds advertised by the page, best first
}

// FeedCandidate is one feed advertised by a page via <link rel="alternate">.
type FeedCandidate struct {
	URL   string `json:"url"`
	Title string `json:"title"`
	Type  string `json:"type"` // rss / atom / json / rdf
}
