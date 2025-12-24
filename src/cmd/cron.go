package cmd

import (
	"blog_api/src/config"
	"blog_api/src/model"
	friendsRepositories "blog_api/src/repositories/friend"
	"blog_api/src/service"
	crawlerService "blog_api/src/service/crawler"
	"log"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// RunFriendLinkCrawlerJob 执行友链爬取并发现 RSS 订阅源（并发模式）
func RunFriendLinkCrawlerJob(db *gorm.DB) {
	log.Println("[Cron] 正在运行友链爬取任务（并发模式）...")
	isDied := false
	opts := model.FriendLinkQueryOptions{
		Statuses: []string{"ignored"},
		NotIn:    true,
		IsDied:   &isDied,
	}
	resp, err := friendsRepositories.QueryFriendLinks(db, opts)
	if err != nil {
		log.Printf("[Cron] 获取全部友链失败： %v", err)
		return
	}
	links := resp.Links

	if len(links) == 0 {
		log.Println("[Cron] 没有需要爬取的友链")
		return
	}

	// 使用并发爬虫
	results := crawlerService.CrawlWebsitesConcurrently(links)

	// 处理爬取结果
	for _, crawlResult := range results {
		link := crawlResult.Link
		result := crawlResult.Result

		err := friendsRepositories.UpdateFriendLink(db, link, result)
		if err != nil {
			log.Printf("[Cron] 在 cron 任务中更新友链 %s 失败: %v", link.Name, err)
		}
		// 更新友链后，发现并插入 RSS 订阅源
		if link.EnableRss && len(result.RssURLs) > 0 {
			for _, rssURL := range result.RssURLs {
				name, err := crawlerService.GetRssTitle(rssURL)
				if err != nil {
					log.Printf("[Cron] 获取 RSS 标题失败 %s: %v", rssURL, err)
					continue
				}
				_, err = friendsRepositories.CreateFriendRssFeeds(db, link.ID, rssURL, name)
				if err != nil {
					log.Printf("[Cron] 在 cron 任务中为 %s 插入 RSS 订阅源失败: %v", link.Name, err)
				}
			}
		}
	}
	log.Println("[Cron] 友链爬取任务完成")
}

// RunDiedFriendLinkCheckJob 执行失效友链的检查（并发模式）
func RunDiedFriendLinkCheckJob(db *gorm.DB) {
	log.Println("[Cron] 正在运行失效友链检查任务（并发模式）...")
	isDied := true
	opts := model.FriendLinkQueryOptions{
		IsDied: &isDied,
	}
	resp, err := friendsRepositories.QueryFriendLinks(db, opts)
	if err != nil {
		log.Printf("[Cron] 获取全部 died 友链失败： %v", err)
		return
	}
	links := resp.Links

	if len(links) == 0 {
		log.Println("[Cron] 没有需要检查的失效友链")
		return
	}

	// 使用并发爬虫
	results := crawlerService.CrawlWebsitesConcurrently(links)

	// 处理爬取结果
	for _, crawlResult := range results {
		link := crawlResult.Link
		result := crawlResult.Result
		// 如果链接仍然有效，状态将更新为"存活"并重置计数
		err := friendsRepositories.UpdateFriendLink(db, link, result)
		if err != nil {
			log.Printf("[Cron] 在 cron 任务中更新失效友链 %s 失败: %v", link.Name, err)
		}
	}
	log.Println("[Cron] 失效友链检查任务完成")
}

// RunRssParserJob 获取所有 RSS 订阅源并解析它们（并发模式）
func RunRssParserJob(db *gorm.DB) {
	log.Println("[Cron] 正在运行 RSS 解析任务（并发模式）...")
	opts := model.FriendRssQueryOptions{Status: "valid"}
	resp, err := friendsRepositories.QueryFriendRss(db, opts)
	if err != nil {
		log.Printf("[Cron] 获取所有 RSS 订阅源失败: %v", err)
		return
	}
	rssFeeds := resp.Feeds

	if len(rssFeeds) == 0 {
		log.Println("[Cron] 没有需要解析的 RSS 订阅源")
		return
	}

	// 使用并发解析
	crawlerService.ParseRssFeedsConcurrently(rssFeeds, func(friendRssID int, rssURL string) {
		crawlerService.ParseRssFeed(db, friendRssID, rssURL)
	})
	log.Println("[Cron] RSS 解析任务完成")
}

// RunImageCheckJob 执行图片资源检查任务
func RunImageCheckJob(db *gorm.DB) {
	log.Println("[Cron] 正在运行图片资源检查任务...")
	crawlerService.CheckImagesHealth(db)
	log.Println("[Cron] 图片资源检查任务完成")
}

// StartCronJobs 初始化并启动 cron 任务
func StartCronJobs(db *gorm.DB) {
	c := cron.New()

	// 安排友链爬取任务每 6 小时运行一次
	c.AddFunc("0 */6 * * *", func() {
		RunFriendLinkCrawlerJob(db)
	})

	// 安排失效友链检查任务每 24 小时运行一次
	c.AddFunc("0 0 * * *", func() {
		RunDiedFriendLinkCheckJob(db)
	})

	// 安排 RSS 解析任务每 3 小时运行一次
	c.AddFunc("0 */3 * * *", func() {
		RunRssParserJob(db)
	})

	// 安排图片资源检查任务每 24 小时运行一次
	c.AddFunc("30 0 * * *", func() {
		RunImageCheckJob(db)
	})

	// 如果启用了状态日志，则安排任务
	if config.GetConfig().EnableStatusLog {
		// 安排系统状态日志记录任务每 5 分钟运行一次
		c.AddFunc("*/5 * * * *", func() {
			service.LogSystemStatus(db)
		})
		log.Println("[Cron] 已启用系统状态日志记录任务")
	}

	// 如果配置了启动时扫描，则立即运行一次任务
	if config.GetConfig().CronScanOnStartup {
		go func() {
			log.Println("[Cron] 调度启动时扫描任务")
			if config.GetConfig().EnableStatusLog {
				service.LogSystemStatus(db)
			}
			RunFriendLinkCrawlerJob(db)
			RunRssParserJob(db)
		}()
	} else {
		log.Println("[Cron] 根据 CRON_SCAN_ON_STARTUP 设置跳过初始扫描")
	}

	log.Println("[Cron] 正在启动 cron 任务...")
	c.Start()
}
