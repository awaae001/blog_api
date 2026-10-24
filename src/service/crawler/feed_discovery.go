package crawlerService

import (
	"blog_api/src/model"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
)

// feedMIMETypes maps the <link type> values that advertise a feed to a short
// feed type. Plain application/json is left out on purpose: WordPress uses it
// for REST API links, not feeds.
var feedMIMETypes = map[string]string{
	"application/rss+xml":   "rss",
	"application/atom+xml":  "atom",
	"application/feed+json": "json",
	"application/rdf+xml":   "rdf",
}

// DiscoverFeeds loads a web page and returns the feeds it advertises, best first.
func DiscoverFeeds(ctx context.Context, pageURL string) ([]model.FeedCandidate, error) {
	if _, err := ValidatePublicHTTPURL(pageURL); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return nil, err
	}
	setHTMLRequestHeaders(req)
	resp, err := feedHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := charset.NewReader(&limitedReader{reader: resp.Body, remaining: maxHTMLResponseBytes}, resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, err
	}
	return extractFeedLinks(doc, resp.Request.URL), nil
}

// extractFeedLinks collects every <link rel="alternate"> that points at a feed.
// Document order is kept, except that comment feeds sink to the end so the
// first candidate is the main article feed.
func extractFeedLinks(doc *goquery.Document, pageURL *url.URL) []model.FeedCandidate {
	base := pageURL
	if href, ok := doc.Find("base[href]").First().Attr("href"); ok {
		if parsed, err := pageURL.Parse(strings.TrimSpace(href)); err == nil {
			base = parsed
		}
	}

	var candidates []model.FeedCandidate
	seen := make(map[string]bool)
	doc.Find("link[href]").Each(func(_ int, s *goquery.Selection) {
		if !slices.Contains(strings.Fields(strings.ToLower(s.AttrOr("rel", ""))), "alternate") {
			return
		}
		mime, _, _ := strings.Cut(strings.ToLower(s.AttrOr("type", "")), ";")
		feedType, ok := feedMIMETypes[strings.TrimSpace(mime)]
		if !ok {
			return
		}
		ref, err := base.Parse(strings.TrimSpace(s.AttrOr("href", "")))
		if err != nil || (ref.Scheme != "http" && ref.Scheme != "https") {
			return
		}
		ref.Fragment = ""
		abs := ref.String()
		if seen[abs] {
			return
		}
		seen[abs] = true
		candidates = append(candidates, model.FeedCandidate{
			URL:   abs,
			Title: strings.TrimSpace(s.AttrOr("title", "")),
			Type:  feedType,
		})
	})

	slices.SortStableFunc(candidates, func(a, b model.FeedCandidate) int {
		return boolRank(isCommentFeed(a)) - boolRank(isCommentFeed(b))
	})
	return candidates
}

func isCommentFeed(c model.FeedCandidate) bool {
	text := strings.ToLower(c.URL + " " + c.Title)
	return strings.Contains(text, "comment") || strings.Contains(text, "评论")
}

func boolRank(b bool) int {
	if b {
		return 1
	}
	return 0
}
