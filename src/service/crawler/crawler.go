package crawlerService

import (
	"blog_api/src/model"
	"context"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
)

const maxHTMLResponseBytes = int64(4 << 20)

var crawlerHTTPClient = &http.Client{
	Timeout:   10 * time.Second,
	Transport: newSafeHTTPTransport(),
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

// CrawlWebsite inspects one website and returns its health and discovered metadata.
// Transport and parsing failures are represented by the result status.
func CrawlWebsite(ctx context.Context, rawURL string) model.CrawlResult {
	if _, err := ValidatePublicHTTPURL(rawURL); err != nil {
		log.Printf("[crawler]拒绝爬取非法目标 %s: %v", rawURL, err)
		return model.CrawlResult{Status: StatusError}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		log.Printf("[crawler]创建获取 %s 的请求时出错: %v", rawURL, err)
		return model.CrawlResult{Status: StatusError}
	}
	setHTMLRequestHeaders(req)
	resp, err := crawlerHTTPClient.Do(req)
	if err != nil {
		status, reason := classifyTransportError(err)
		log.Printf("[crawler]获取 URL %s 失败，判定为 %s (%s): %v", rawURL, status, reason, err)
		return model.CrawlResult{Status: status}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		redirectLocation := resp.Header.Get("Location")
		absoluteRedirectURL := toAbsoluteURL(resp.Request.URL, redirectLocation)
		log.Printf("[crawler]检测到 %s 重定向到 %s (resolved to %s)", rawURL, redirectLocation, absoluteRedirectURL)
		return model.CrawlResult{Status: StatusSurvival, RedirectURL: absoluteRedirectURL}
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("[crawler]错误: %s 的状态码非 200: %d", rawURL, resp.StatusCode)
		return model.CrawlResult{Status: StatusError}
	}
	limitedBody := &limitedReader{reader: resp.Body, remaining: maxHTMLResponseBytes}
	utf8Reader, err := charset.NewReader(limitedBody, resp.Header.Get("Content-Type"))
	if err != nil {
		log.Printf("[crawler]创建 %s 的字符集解码器时出错: %v", rawURL, err)
		return model.CrawlResult{Status: StatusError}
	}
	doc, err := goquery.NewDocumentFromReader(utf8Reader)
	if err != nil {
		log.Printf("[crawler]解析 %s 的 HTML 时出错: %v", rawURL, err)
		return model.CrawlResult{Status: StatusError}
	}
	description := doc.Find("meta[name='description']").AttrOr("content", "")
	iconURL, exists := doc.Find("link[rel='icon']").Attr("href")
	if !exists {
		iconURL, exists = doc.Find("link[rel='apple-touch-icon']").Attr("href")
		if !exists {
			iconURL = doc.Find("link[rel='shortcut icon']").AttrOr("href", "")
		}
	}

	return model.CrawlResult{
		Description: description,
		IconURL:     iconURL,
		Status:      StatusSurvival,
		Feeds:       extractFeedLinks(doc, resp.Request.URL),
	}
}

// toAbsoluteURL 根据基础 URL 将相对 URL 转换为绝对 URL
func toAbsoluteURL(base *url.URL, href string) string {
	relativeURL, err := url.Parse(href)
	if err != nil {
		return ""
	}
	return base.ResolveReference(relativeURL).String()
}
