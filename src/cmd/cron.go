package cmd

import (
	"blog_api/src/config"
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

// rssRefreshMu 防止启动扫描与定时任务、以及上一轮超时导致的重入。
// 友链巡查自身的互斥由 crawlerService 持有。
var rssRefreshMu sync.Mutex

// RunRssRefreshJob 抓取所有到期的订阅源。每个源按自己的 next_fetch_at 调度，
// 失效的源也会以长间隔继续探活，成功后自动恢复。
func RunRssRefreshJob(db *gorm.DB) {
	if !rssRefreshMu.TryLock() {
		log.Println("[Cron] RSS 抓取任务上一轮仍在执行，本次跳过")
		return
	}
	defer rssRefreshMu.Unlock()

	if _, err := crawlerService.RefreshDueFeeds(context.Background(), db); err != nil {
		log.Printf("[Cron] RSS 抓取任务失败: %v", err)
	}
}

// StartCronJobs 初始化并启动 cron 任务
func StartCronJobs(db *gorm.DB) {
	c := cron.New()

	// 安排慢检查任务从下一天 0 点开始，每 48 小时运行一次
	scheduleFromNextMidnight("失效友链检查", 48*time.Hour, func() {
		// 巡查自身会记录跳过与失败原因，这里无需再处理返回值。
		_, _ = crawlerService.InspectFriendLinks(context.Background(), db, crawlerService.DiedInspectionScope())
	})
	scheduleFromNextMidnight("图片资源检查", 48*time.Hour, func() {
		crawlerService.CheckImagesHealth(db)
	})

	// RSS 调度器每 10 分钟检查一次到期的订阅源
	c.AddFunc("*/10 * * * *", func() {
		RunRssRefreshJob(db)
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
			RunRssRefreshJob(db)
		}()
	} else {
		log.Println("[Cron] 根据 CRON_SCAN_ON_STARTUP 设置跳过初始扫描")
	}

	log.Println("[Cron] 正在启动 cron 任务...")
	c.Start()
}
