package crawlerService

import (
	"blog_api/src/config"
	"blog_api/src/model"
	"context"
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

// ImageCheckJob 表示一个图片检查任务
type ImageCheckJob struct {
	Image model.Image
}

// CrawlWebsitesConcurrently crawls links with a bounded worker pool.
// Results are consumed synchronously so database writes retain one owner.
func CrawlWebsitesConcurrently(ctx context.Context, links []model.FriendWebsite, consume func(CrawlJobResult) error) error {
	if len(links) == 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	concurrency := effectiveConcurrency(len(links))
	log.Printf("[ConcurrentCrawler] 开始并发爬取 %d 个网站，并发数: %d", len(links), concurrency)

	jobs := make(chan CrawlJob, concurrency)
	results := make(chan CrawlJobResult, concurrency)

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go crawlWorker(ctx, i, jobs, results, &wg)
	}

	go func() {
		defer close(jobs)
		for _, link := range links {
			select {
			case jobs <- CrawlJob{Link: link}:
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var consumeErr error
	processed := 0
	for result := range results {
		if consumeErr != nil {
			continue
		}
		if err := consume(result); err != nil {
			consumeErr = err
			cancel()
			continue
		}
		processed++
	}

	log.Printf("[ConcurrentCrawler] 完成并发爬取，共处理 %d 个网站", processed)
	return consumeErr
}

// crawlWorker 是 worker goroutine，从任务通道获取任务并执行爬取
func crawlWorker(ctx context.Context, id int, jobs <-chan CrawlJob, results chan<- CrawlJobResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobs:
			if !ok {
				return
			}
			log.Printf("[ConcurrentCrawler][Worker %d] 正在爬取: %s", id, job.Link.Link)
			result := CrawlWebsite(ctx, job.Link.Link)
			select {
			case results <- CrawlJobResult{Link: job.Link, Result: result}:
			case <-ctx.Done():
				return
			}
			log.Printf("[ConcurrentCrawler][Worker %d] 完成爬取: %s, 状态: %s", id, job.Link.Link, result.Status)
		}
	}
}

// CheckImagesConcurrently 并发检查图片
func CheckImagesConcurrently(images []model.Image, checkFunc func(image model.Image)) {
	if len(images) == 0 {
		return
	}

	concurrency := effectiveConcurrency(len(images))
	log.Printf("[ConcurrentCrawler] 开始并发检查 %d 张图片，并发数: %d", len(images), concurrency)

	jobs := make(chan ImageCheckJob, concurrency)
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go imageCheckWorker(i, jobs, checkFunc, &wg)
	}

	for _, img := range images {
		jobs <- ImageCheckJob{Image: img}
	}
	close(jobs)

	wg.Wait()
	log.Printf("[ConcurrentCrawler] 完成并发检查 %d 张图片", len(images))
}

func imageCheckWorker(id int, jobs <-chan ImageCheckJob, checkFunc func(image model.Image), wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		log.Printf("[ConcurrentCrawler][Worker %d] 正在检查图片: %s", id, job.Image.URL)
		checkFunc(job.Image)
		log.Printf("[ConcurrentCrawler][Worker %d] 完成检查图片: %s", id, job.Image.URL)
	}
}

func effectiveConcurrency(total int) int {
	return min(total, config.GetConfig().Crawler.Concurrency)
}
