package crawlerService

import (
	"blog_api/src/config"
	"blog_api/src/model"
	"log"
	"sync"
)

// CrawlJob 表示一个爬取任务
type CrawlJob struct {
	Link model.FriendWebsite
}

// CrawlJobResult 表示爬取任务的结果
type CrawlJobResult struct {
	Link   model.FriendWebsite
	Result model.CrawlResult
}

// CrawlWebsitesConcurrently 并发爬取多个网站
// 使用 worker pool 模式，并发数量由配置文件控制
func CrawlWebsitesConcurrently(links []model.FriendWebsite) []CrawlJobResult {
	concurrency := config.GetConfig().Crawler.Concurrency
	if concurrency <= 0 {
		concurrency = 5 // 默认并发数
	}

	// 如果链接数量少于并发数，则使用链接数量作为并发数
	if len(links) < concurrency {
		concurrency = len(links)
	}

	if len(links) == 0 {
		return []CrawlJobResult{}
	}

	log.Printf("[ConcurrentCrawler] 开始并发爬取 %d 个网站，并发数: %d", len(links), concurrency)

	// 创建任务通道和结果通道
	jobs := make(chan CrawlJob, len(links))
	results := make(chan CrawlJobResult, len(links))

	// 启动 worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go crawlWorker(i, jobs, results, &wg)
	}

	// 发送任务到任务通道
	for _, link := range links {
		jobs <- CrawlJob{Link: link}
	}
	close(jobs)

	// 等待所有 worker 完成后关闭结果通道
	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集结果
	var crawlResults []CrawlJobResult
	for result := range results {
		crawlResults = append(crawlResults, result)
	}

	log.Printf("[ConcurrentCrawler] 完成并发爬取，共处理 %d 个网站", len(crawlResults))
	return crawlResults
}

// crawlWorker 是 worker goroutine，从任务通道获取任务并执行爬取
func crawlWorker(id int, jobs <-chan CrawlJob, results chan<- CrawlJobResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		log.Printf("[ConcurrentCrawler][Worker %d] 正在爬取: %s", id, job.Link.Link)
		result := CrawlWebsite(job.Link.Link)
		results <- CrawlJobResult{
			Link:   job.Link,
			Result: result,
		}
		log.Printf("[ConcurrentCrawler][Worker %d] 完成爬取: %s, 状态: %s", id, job.Link.Link, result.Status)
	}
}

// RssParseJob 表示一个 RSS 解析任务
type RssParseJob struct {
	FriendRssID int
	RssURL      string
}

// ParseRssFeedsConcurrently 并发解析多个 RSS 订阅源
func ParseRssFeedsConcurrently(feeds []model.FriendRss, parseFunc func(friendRssID int, rssURL string)) {
	concurrency := config.GetConfig().Crawler.Concurrency
	if concurrency <= 0 {
		concurrency = 5 // 默认并发数
	}

	// 如果订阅源数量少于并发数，则使用订阅源数量作为并发数
	if len(feeds) < concurrency {
		concurrency = len(feeds)
	}

	if len(feeds) == 0 {
		return
	}

	log.Printf("[ConcurrentCrawler] 开始并发解析 %d 个 RSS 订阅源，并发数: %d", len(feeds), concurrency)

	// 创建任务通道
	jobs := make(chan RssParseJob, len(feeds))

	// 启动 worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go rssParseWorker(i, jobs, parseFunc, &wg)
	}

	// 发送任务到任务通道
	for _, feed := range feeds {
		jobs <- RssParseJob{
			FriendRssID: feed.ID,
			RssURL:      feed.RssURL,
		}
	}
	close(jobs)

	// 等待所有 worker 完成
	wg.Wait()

	log.Printf("[ConcurrentCrawler] 完成并发解析 %d 个 RSS 订阅源", len(feeds))
}

// rssParseWorker 是 RSS 解析的 worker goroutine
func rssParseWorker(id int, jobs <-chan RssParseJob, parseFunc func(friendRssID int, rssURL string), wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		log.Printf("[ConcurrentCrawler][Worker %d] 正在解析 RSS: %s", id, job.RssURL)
		parseFunc(job.FriendRssID, job.RssURL)
		log.Printf("[ConcurrentCrawler][Worker %d] 完成解析 RSS: %s", id, job.RssURL)
	}
}
