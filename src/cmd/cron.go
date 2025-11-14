package main

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

	// Schedule the crawler to run at 1:00 AM every day.
	_, err := c.AddFunc("0 1 * * *", func() {
		log.Println("Running friend link crawler job...")
		links, err := repositories.GetAllFriendLinks(db)
		if err != nil {
			log.Printf("Error getting friend links for cron job: %v", err)
			return
		}

		for _, link := range links {
			result := service.CrawlWebsite(link.Link)
			err := repositories.UpdateFriendLink(db, link.ID, result)
			if err != nil {
				log.Printf("Error updating friend link %s in cron job: %v", link.Name, err)
			}
		}
	})
	if err != nil {
		log.Fatalf("Could not add cron job: %v", err)
	}

	log.Println("Starting cron jobs...")
	c.Start()
}
