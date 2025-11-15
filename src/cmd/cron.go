package cmd

import (
	"blog_api/src/repositories"
	"blog_api/src/service"
	"database/sql"
	"log"

	"github.com/robfig/cron/v3"
)

// RunFriendLinkCrawlerJob performs the crawling of friend links and discovers RSS feeds.
func RunFriendLinkCrawlerJob(db *sql.DB) {
	log.Println("[Cron] Running friend link crawler job...")
	links, err := repositories.GetAllFriendLinks(db)
	if err != nil {
		log.Printf("[Cron] 获取全部友链失败： %v", err)
		return
	}

	for _, link := range links {
		result := service.CrawlWebsite(link.Link)
		err := repositories.UpdateFriendLink(db, link, result)
		if err != nil {
			log.Printf("[Cron] Error updating friend link %s in cron job: %v", link.Name, err)
		}
		// After updating the friend link, discover and insert RSS feeds.
		if len(result.RssURLs) > 0 {
			err = repositories.InsertFriendRss(db, link.ID, result.RssURLs)
			if err != nil {
				log.Printf("[Cron] Error inserting RSS feeds for %s in cron job: %v", link.Name, err)
			}
		}
	}
}

// RunRssParserJob fetches all RSS feeds and parses them.
func RunRssParserJob(db *sql.DB) {
	log.Println("[Cron] Running RSS parser job...")
	rssFeeds, err := repositories.GetAllFriendRss(db)
	if err != nil {
		log.Printf("[Cron] Failed to get all RSS feeds: %v", err)
		return
	}

	for _, rss := range rssFeeds {
		service.ParseRssFeed(db, rss.ID, rss.RssURL)
	}
}

// StartCronJobs initializes and starts the cron jobs.
func StartCronJobs(db *sql.DB) {
	c := cron.New()

	// Schedule the friend link crawler to run every 6 hours.
	_, err := c.AddFunc("0 */6 * * *", func() {
		RunFriendLinkCrawlerJob(db)
	})
	if err != nil {
		log.Fatalf("[Cron] Could not add friend link crawler cron job: %v", err)
	}

	// Schedule the RSS parser to run every hour.
	_, err = c.AddFunc("0 * * * *", func() {
		RunRssParserJob(db)
	})
	if err != nil {
		log.Fatalf("[Cron] Could not add RSS parser cron job: %v", err)
	}

	// Run jobs once immediately on startup in separate goroutines.
	go func() {
		log.Println("[Cron] Running initial friend link crawler job...")
		RunFriendLinkCrawlerJob(db)
	}()
	go func() {
		log.Println("[Cron] Running initial RSS parser job...")
		RunRssParserJob(db)
	}()

	log.Println("[Cron] Starting cron jobs...")
	c.Start()
}
