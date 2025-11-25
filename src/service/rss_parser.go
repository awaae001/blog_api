package service

import (
	"blog_api/src/model"
	"blog_api/src/repositories"
	"log"

	"github.com/microcosm-cc/bluemonday"
	"github.com/mmcdole/gofeed"
	"gorm.io/gorm"
)

// ParseRssFeed parses an RSS feed and saves the articles to the database.
func ParseRssFeed(db *gorm.DB, friendRssID int, rssURL string) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(rssURL)
	if err != nil {
		log.Printf("解析 RSS feed %s 时出错: %v", rssURL, err)
		return
	}

	p := bluemonday.StripTagsPolicy()
	for _, item := range feed.Items {
		publishedTime := item.PublishedParsed
		if publishedTime == nil {
			// If PublishedParsed is nil, use UpdatedParsed
			publishedTime = item.UpdatedParsed
			if publishedTime == nil {
				log.Printf("跳过没有发布或更新时间的文章: %s", item.Title)
				continue
			}
		}

		post := &model.RssPost{
			FriendRssID: friendRssID,
			Title:       item.Title,
			Link:        item.Link,
			Description: p.Sanitize(item.Description),
			Time:        *publishedTime,
		}

		err := repositories.InsertRssPost(db, post)
		if err != nil {
			log.Printf("插入文章 '%s' 时出错: %v", item.Title, err)
		}
	}
}
