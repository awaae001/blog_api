package cmd

import (
	"blog_api/src/config"
	"blog_api/src/model"
	friendsRepositories "blog_api/src/repositories/friend"
	crawlerService "blog_api/src/service/crawler"
	"context"
	"log"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

func scheduleFromNextMidnight(jobName string, interval time.Duration, job func()) {
	go func() {
		now := time.Now()
		nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		firstRunAt := nextMidnight.Add(interval)
		initialDelay := time.Until(firstRunAt)

		log.Printf("[Cron] %s 将于 %s 首次执行（下一天 0 点后 + %s），之后每 %s 执行一次", jobName, firstRunAt.Format(time.RFC3339), interval, interval)

		timer := time.NewTimer(initialDelay)
		defer timer.Stop()

		<-timer.C
		job()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			job()
		}
	}()
}

// 任务互斥锁：防止启动扫描与定时任务、以及任务超时重入导致的并发执行。
// 友链巡查自身的互斥由 crawlerService 持有，因为手动全量巡查与定时任务
// 读写同一批数据，必须共用同一把锁。
var (
	diedRssCheckMu sync.Mutex
	rssParserMu    sync.Mutex
)

// RunDiedRssCheckJob 执行失效 RSS 的探活检查（低频）
func RunDiedRssCheckJob(db *gorm.DB) {
	if !diedRssCheckMu.TryLock() {
		log.Println("[Cron] 失效 RSS 探活任务上一轮仍在执行，本次跳过")
		return
	}
	defer diedRssCheckMu.Unlock()

	log.Println("[Cron] 正在运行失效 RSS 探活任务...")
	isDied := true
	opts := model.FriendRssQueryOptions{
		IsDied: &isDied,
	}
	resp, err := friendsRepositories.QueryFriendRss(db, opts)
	if err != nil {
		log.Printf("[Cron] 获取失效 RSS 失败: %v", err)
		return
	}
	rssFeeds := resp.Feeds

	if len(rssFeeds) == 0 {
		log.Println("[Cron] 没有需要探活的失效 RSS")
		return
	}

	for _, feed := range rssFeeds {
		if feed.Status == "pause" {
			continue
		}
		crawlerService.CheckAndReviveRssFeed(db, feed.ID, feed.RssURL)
	}

	log.Println("[Cron] 失效 RSS 探活任务完成")
}

// RunRssParserJob 获取所有 RSS 订阅源并解析它们（并发模式）
func RunRssParserJob(db *gorm.DB) {
	if !rssParserMu.TryLock() {
		log.Println("[Cron] RSS 解析任务上一轮仍在执行，本次跳过")
		return
	}
	defer rssParserMu.Unlock()

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

	result := crawlerService.ParseRssFeedsConcurrently(context.Background(), db, rssFeeds)
	if result.DatabaseFailures > 0 {
		log.Printf("[Cron] RSS 解析任务有 %d 个数据库写入失败", result.DatabaseFailures)
	}
	log.Println("[Cron] RSS 解析任务完成")
}

// StartCronJobs 初始化并启动 cron 任务
func StartCronJobs(db *gorm.DB) {
	c := cron.New()

	// 安排慢检查任务从下一天 0 点开始，每 48 小时运行一次
	scheduleFromNextMidnight("失效检查（友链+RSS）", 48*time.Hour, func() {
		// 巡查自身会记录跳过与失败原因，这里无需再处理返回值。
		_, _ = crawlerService.InspectFriendLinks(context.Background(), db, crawlerService.DiedInspectionScope())
		RunDiedRssCheckJob(db)
	})
	scheduleFromNextMidnight("图片资源检查", 48*time.Hour, func() {
		crawlerService.CheckImagesHealth(db)
	})

	// 安排 RSS 解析任务每 3 小时运行一次
	c.AddFunc("0 */3 * * *", func() {
		RunRssParserJob(db)
	})
	// 安排友链爬取任务每 6 小时运行一次
	c.AddFunc("0 */6 * * *", func() {
		_, _ = crawlerService.InspectFriendLinks(context.Background(), db, crawlerService.SurvivalInspectionScope())
	})

	// 如果配置了启动时扫描，则立即运行一次任务
	if config.GetConfig().CronScanOnStartup {
		go func() {
			log.Println("[Cron] 调度启动时扫描任务")
			_, _ = crawlerService.InspectFriendLinks(context.Background(), db, crawlerService.SurvivalInspectionScope())
			RunRssParserJob(db)
		}()
	} else {
		log.Println("[Cron] 根据 CRON_SCAN_ON_STARTUP 设置跳过初始扫描")
	}

	log.Println("[Cron] 正在启动 cron 任务...")
	c.Start()
}
