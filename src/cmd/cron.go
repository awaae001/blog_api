package cmd

import (
	"blog_api/src/repositories"
	"blog_api/src/service"
	"database/sql"
	"log"

	"github.com/robfig/cron/v3"
)

// StartCronJobs initializes and starts the cron jobs.
func StartCronJobs(db *sql.DB) {
	c := cron.New()

	// Schedule the crawler to run every 3 hours.
	_, err := c.AddFunc("0 */3 * * *", func() {
		log.Println("Running friend link crawler job...")
		links, err := repositories.GetAllFriendLinks(db)
		if err != nil {
			log.Printf("Error getting friend links for cron job: %v", err)
			return
		}

		for _, link := range links {
			result := service.CrawlWebsite(link.Link)
			err := repositories.UpdateFriendLink(db, link, result)
			if err != nil {
				log.Printf("Error updating friend link %s in cron job: %v", link.Name, err)
			}
			// After updating the friend link, discover and insert RSS feeds.
			if len(result.RssURLs) > 0 {
				err = repositories.InsertFriendRss(db, link.ID, result.RssURLs)
				if err != nil {
					log.Printf("Error inserting RSS feeds for %s in cron job: %v", link.Name, err)
				}
			}
		}
	})
	if err != nil {
		log.Fatalf("Could not add cron job: %v", err)
	}

	log.Println("Starting cron jobs...")
	c.Start()
}
