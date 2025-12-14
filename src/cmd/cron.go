package cmd

import (
	"blog_api/src/config"
	"blog_api/src/model"
	"blog_api/src/repositories"
	"blog_api/src/service"
	"log"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// RunFriendLinkCrawlerJob 执行友链爬取并发现 RSS 订阅源（并发模式）
func RunFriendLinkCrawlerJob(db *gorm.DB) {
	log.Println("[Cron] 正在运行友链爬取任务（并发模式）...")
	opts := model.FriendLinkQueryOptions{
		Statuses: []string{"died", "ignored"},
		NotIn:    true,
	}
	resp, err := repositories.QueryFriendLinks(db, opts)
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
	results := service.CrawlWebsitesConcurrently(links)

	// 处理爬取结果
	for _, crawlResult := range results {
		link := crawlResult.Link
		result := crawlResult.Result

		err := repositories.UpdateFriendLink(db, link, result)
		if err != nil {
			log.Printf("[Cron] 在 cron 任务中更新友链 %s 失败: %v", link.Name, err)
		}
		// 更新友链后，发现并插入 RSS 订阅源
		if link.EnableRss && len(result.RssURLs) > 0 {
			for _, rssURL := range result.RssURLs {
				name, err := service.GetRssTitle(rssURL)
				if err != nil {
					log.Printf("[Cron] 获取 RSS 标题失败 %s: %v", rssURL, err)
					continue
				}
				_, err = repositories.CreateFriendRssFeeds(db, link.ID, rssURL, name)
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
	opts := model.FriendLinkQueryOptions{
		Status: "died",
	}
	resp, err := repositories.QueryFriendLinks(db, opts)
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
	results := service.CrawlWebsitesConcurrently(links)

	// 处理爬取结果
	for _, crawlResult := range results {
		link := crawlResult.Link
		result := crawlResult.Result
		// 如果链接仍然有效，状态将更新为"存活"并重置计数
		err := repositories.UpdateFriendLink(db, link, result)
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
	resp, err := repositories.QueryFriendRss(db, opts)
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
	service.ParseRssFeedsConcurrently(rssFeeds, func(friendRssID int, rssURL string) {
		service.ParseRssFeed(db, friendRssID, rssURL)
	})
	log.Println("[Cron] RSS 解析任务完成")
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

	// 安排 RSS 解析任务每小时运行一次
	c.AddFunc("0 * * * *", func() {
		RunRssParserJob(db)
	})

	// 如果配置了启动时扫描，则立即运行一次任务
	if config.GetConfig().CronScanOnStartup {
		go func() {
			log.Println("[Cron] 正在运行初始友链爬取任务...")
			RunFriendLinkCrawlerJob(db)
			log.Println("[Cron] 正在运行初始 RSS 解析任务...")
			RunRssParserJob(db)
		}()
	} else {
		log.Println("[Cron] 根据 CRON_SCAN_ON_STARTUP 设置跳过初始扫描")
	}

	log.Println("[Cron] 正在启动 cron 任务...")
	c.Start()
}
