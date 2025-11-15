package service

import (
	"blog_api/src/model"
	"blog_api/src/repositories"
	"database/sql"
	"log"

	"github.com/mmcdole/gofeed"
)

// ParseRssFeed parses an RSS feed and saves the articles to the database.
func ParseRssFeed(db *sql.DB, friendRssID int, rssURL string) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(rssURL)
	if err != nil {
		log.Printf("Error parsing RSS feed %s: %v", rssURL, err)
		return
	}

	for _, item := range feed.Items {
		publishedTime := item.PublishedParsed
		if publishedTime == nil {
			// If PublishedParsed is nil, use UpdatedParsed
			publishedTime = item.UpdatedParsed
			if publishedTime == nil {
				log.Printf("Skipping post with no publish or update time: %s", item.Title)
				continue
			}
		}

		post := &model.RssPost{
			FriendRssID: friendRssID,
			Title:       item.Title,
			Link:        item.Link,
			Description: item.Description,
			Time:        *publishedTime,
		}

		err := repositories.InsertRssPost(db, post)
		if err != nil {
			log.Printf("Error inserting post '%s': %v", item.Title, err)
		}
	}
}
