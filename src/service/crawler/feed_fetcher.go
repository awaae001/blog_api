package crawlerService

import (
	"blog_api/src/config"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/mmcdole/gofeed"
)

const maxFeedResponseBytes = int64(8 << 20)

// feedHTTPClient follows redirects; fetchFeed wraps it per request to tell
// permanent moves apart from temporary ones.
var feedHTTPClient = &http.Client{
	Transport: newSafeHTTPTransport(),
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return errors.New("too many redirects")
		}
		return nil
	},
}

// ErrRssSource identifies failures caused by fetching or parsing a remote feed.
var ErrRssSource = errors.New("RSS source failure")

// httpStatusError is a non-2xx feed response.
type httpStatusError struct {
	code       int
	retryAfter time.Duration
}

func (e *httpStatusError) Error() string {
	return fmt.Sprintf("HTTP %d %s", e.code, http.StatusText(e.code))
}

// feedResponse is the outcome of one successful round trip.
type feedResponse struct {
	feed         *gofeed.Feed // nil when NotModified
	notModified  bool
	movedTo      string // final URL when every redirect hop was permanent
	etag         string
	lastModified string
}

// fetchFeed downloads and parses one feed. etag/lastModified turn it into a
// conditional request; pass empty strings to force a full download.
func fetchFeed(ctx context.Context, feedURL, etag, lastModified string) (feedResponse, error) {
	if _, err := ValidatePublicHTTPURL(feedURL); err != nil {
		return feedResponse{}, err
	}
	timeout := time.Duration(config.GetConfig().Crawler.RssTimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return feedResponse{}, err
	}
	setFeedRequestHeaders(req)
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}
	if lastModified != "" {
		req.Header.Set("If-Modified-Since", lastModified)
	}

	permanent := true
	client := *feedHTTPClient
	client.CheckRedirect = func(r *http.Request, via []*http.Request) error {
		if code := r.Response.StatusCode; code != http.StatusMovedPermanently && code != http.StatusPermanentRedirect {
			permanent = false
		}
		return feedHTTPClient.CheckRedirect(r, via)
	}

	resp, err := client.Do(req)
	if err != nil {
		return feedResponse{}, err
	}
	defer resp.Body.Close()

	out := feedResponse{
		etag:         resp.Header.Get("ETag"),
		lastModified: resp.Header.Get("Last-Modified"),
	}
	if final := resp.Request.URL.String(); permanent && final != feedURL {
		out.movedTo = final
	}

	switch {
	case resp.StatusCode == http.StatusNotModified:
		out.notModified = true
		// Some servers omit validators on 304; keep the ones we sent.
		if out.etag == "" {
			out.etag = etag
		}
		if out.lastModified == "" {
			out.lastModified = lastModified
		}
		return out, nil
	case resp.StatusCode < 200 || resp.StatusCode >= 300:
		return feedResponse{}, &httpStatusError{code: resp.StatusCode, retryAfter: parseRetryAfter(resp.Header.Get("Retry-After"))}
	}

	feed, err := gofeed.NewParser().Parse(&limitedReader{reader: resp.Body, remaining: maxFeedResponseBytes})
	if err != nil {
		return feedResponse{}, fmt.Errorf("parse feed: %w", err)
	}
	out.feed = feed
	return out, nil
}

// parseRetryAfter accepts both forms of the Retry-After header.
func parseRetryAfter(value string) time.Duration {
	if value == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(value); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	if at, err := http.ParseTime(value); err == nil {
		return max(time.Until(at), 0)
	}
	return 0
}

// describeFetchError turns a fetch failure into the short text stored in
// friend_rss.last_error.
func describeFetchError(err error) string {
	var statusErr *httpStatusError
	if errors.As(err, &statusErr) {
		return statusErr.Error()
	}
	if _, reason := classifyTransportError(err); reason != "" {
		return reason
	}
	return err.Error()
}
