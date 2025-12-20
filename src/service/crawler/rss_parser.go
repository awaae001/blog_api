package crawlerService

import (
	"blog_api/src/model"
	friendsRepositories "blog_api/src/repositories/friend"
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

		time := publishedTime.Unix()
		if time < 0 {
			time = 0
		}

		post := &model.RssPost{
			RssID:       friendRssID,
			Title:       item.Title,
			Link:        item.Link,
			Description: p.Sanitize(item.Description),
			Time:        time,
		}

		err := friendsRepositories.InsertRssPost(db, post)
		if err != nil {
			log.Printf("插入文章 '%s' 时出错: %v", item.Title, err)
		}
	}
}

// GetRssTitle fetches and returns the title of an RSS feed.
func GetRssTitle(rssURL string) (string, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(rssURL)
	if err != nil {
		log.Printf("解析 RSS feed %s 时出错: %v", rssURL, err)
		return "", err
	}
	return feed.Title, nil
}
