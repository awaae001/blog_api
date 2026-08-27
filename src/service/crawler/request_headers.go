package crawlerService

import "net/http"

// CrawlerUserAgent is the formal bot signature sent with every outbound crawler request.
// It follows the conventional "Name/Version (+info URL)" format so site owners can identify the bot and look up its purpose.
const CrawlerUserAgent = "BlogApiBot/1.0 (+https://blog.awaae001.top/rss)"

// acceptLanguage advertises a preference for Chinese content with an English fallback.
const acceptLanguage = "zh-CN,zh;q=0.9,en;q=0.8"

// Accept-Encoding is deliberately left unset: net/http adds gzip automatically and
// only decompresses the response transparently when the caller has not set the
// header itself. Setting it here would force every call site to decode manually.

// setHTMLRequestHeaders prepares a request that expects an HTML document.
func setHTMLRequestHeaders(req *http.Request) {
	req.Header.Set("User-Agent", CrawlerUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", acceptLanguage)
}

// setFeedRequestHeaders prepares a request that expects an RSS or Atom feed.
func setFeedRequestHeaders(req *http.Request) {
	req.Header.Set("User-Agent", CrawlerUserAgent)
	req.Header.Set("Accept", "application/rss+xml,application/atom+xml,application/xml;q=0.9,text/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", acceptLanguage)
}

// setImageRequestHeaders prepares a request that probes an image resource.
func setImageRequestHeaders(req *http.Request) {
	req.Header.Set("User-Agent", CrawlerUserAgent)
	req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/*,*/*;q=0.8")
}
