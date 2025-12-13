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

// RunFriendLinkCrawlerJob 执行友链爬取并发现 RSS 订阅源
func RunFriendLinkCrawlerJob(db *gorm.DB) {
	log.Println("[Cron] 正在运行友链爬取任务...")
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

	for _, link := range links {
		result := service.CrawlWebsite(link.Link)
		err := repositories.UpdateFriendLink(db, link, result)
		if err != nil {
			log.Printf("[Cron] 在 cron 任务中更新友链 %s 失败: %v", link.Name, err)
		}
		// 更新友链后，发现并插入 RSS 订阅源
		if len(result.RssURLs) > 0 {
			_, err = repositories.CreateFriendRssFeeds(db, link.ID, result.RssURLs)
			if err != nil {
				log.Printf("[Cron] 在 cron 任务中为 %s 插入 RSS 订阅源失败: %v", link.Name, err)
			}
		}
	}
}

// RunDiedFriendLinkCheckJob 执行失效友链的检查
func RunDiedFriendLinkCheckJob(db *gorm.DB) {
	log.Println("[Cron] 正在运行失效友链检查任务...")
	opts := model.FriendLinkQueryOptions{
		Status: "died",
	}
	resp, err := repositories.QueryFriendLinks(db, opts)
	if err != nil {
		log.Printf("[Cron] 获取全部 died 友链失败： %v", err)
		return
	}
	links := resp.Links

	for _, link := range links {
		result := service.CrawlWebsite(link.Link)
		// 如果链接仍然有效，状态将更新为“存活”并重置计数
		err := repositories.UpdateFriendLink(db, link, result)
		if err != nil {
			log.Printf("[Cron] 在 cron 任务中更新失效友链 %s 失败: %v", link.Name, err)
		}
	}
}

// RunRssParserJob 获取所有 RSS 订阅源并解析它们
func RunRssParserJob(db *gorm.DB) {
	log.Println("[Cron] 正在运行 RSS 解析任务...")
	opts := model.FriendRssQueryOptions{Status: "valid"}
	resp, err := repositories.QueryFriendRss(db, opts)
	if err != nil {
		log.Printf("[Cron] 获取所有 RSS 订阅源失败: %v", err)
		return
	}
	rssFeeds := resp.Feeds

	for _, rss := range rssFeeds {
		service.ParseRssFeed(db, rss.ID, rss.RssURL)
	}
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
