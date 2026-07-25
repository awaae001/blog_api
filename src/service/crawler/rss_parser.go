package crawlerService

import (
	"blog_api/src/config"
	"blog_api/src/model"
	friendsRepositories "blog_api/src/repositories/friend"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/mmcdole/gofeed"
	"gorm.io/gorm"
)

const (
	maxRssParseFailures = 4
	maxRSSResponseBytes = int64(8 << 20)
)

var rssHTTPClient = &http.Client{}

// ErrRssSource identifies failures caused by fetching or parsing a remote feed.
var ErrRssSource = errors.New("RSS source failure")

func parseFeedURL(ctx context.Context, rawURL string) (*gofeed.Feed, error) {
	timeoutSeconds := config.GetConfig().Crawler.RssTimeoutSeconds
	if timeoutSeconds <= 0 {
		timeoutSeconds = 15
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create RSS request: %w", err)
	}
	req.Header.Set("User-Agent", "blog_api RSS crawler")
	resp, err := rssHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch RSS: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch RSS: unexpected HTTP status %d", resp.StatusCode)
	}

	feed, err := gofeed.NewParser().Parse(&limitedReader{reader: resp.Body, remaining: maxRSSResponseBytes})
	if err != nil {
		return nil, fmt.Errorf("parse RSS: %w", err)
	}
	return feed, nil
}

// ParseRssFeed fetches one RSS feed, stores new articles, and updates its health state.
// A successful call always restores the feed to a healthy state. Remote failures
// wrap ErrRssSource; database failures do not.
func ParseRssFeed(ctx context.Context, db *gorm.DB, friendRssID int, rssURL string) (model.RssFetchResult, error) {
	var result model.RssFetchResult
	feed, err := parseFeedURL(ctx, rssURL)
	if err != nil {
		log.Printf("解析 RSS feed %s 时出错: %v", rssURL, err)
		if ctx.Err() != nil {
			return result, ctx.Err()
		}
		if stateErr := updateRssParseState(db, friendRssID, false); stateErr != nil {
			return result, fmt.Errorf("record RSS fetch failure: %w", stateErr)
		}
		return result, fmt.Errorf("%w: %v", ErrRssSource, err)
	}
	result.CheckedItems = len(feed.Items)
	err = db.Transaction(func(tx *gorm.DB) error {
		friendRssName := ""
		var friendRss model.FriendRss
		if err := tx.Select("name").Where("id = ?", friendRssID).First(&friendRss).Error; err != nil {
			return fmt.Errorf("load RSS feed name: %w", err)
		}
		friendRssName = friendRss.Name

		p := bluemonday.StripTagsPolicy()
		for _, item := range feed.Items {
			publishedTime := item.PublishedParsed
			if publishedTime == nil {
				publishedTime = item.UpdatedParsed
				if publishedTime == nil {
					continue
				}
			}

			publishedUnix := publishedTime.Unix()
			if publishedUnix < 0 {
				publishedUnix = 0
			}

			author := rssItemAuthor(item, friendRssName)
			post := &model.RssPost{
				RssID:       friendRssID,
				Title:       item.Title,
				Link:        item.Link,
				Description: p.Sanitize(item.Description),
				Author:      author,
				Time:        publishedUnix,
			}

			inserted, err := friendsRepositories.InsertRssPost(tx, post)
			if err != nil {
				return fmt.Errorf("insert RSS article %q: %w", item.Title, err)
			}
			if inserted {
				result.InsertedItems++
			}
		}

		return updateRssParseState(tx, friendRssID, true)
	})
	if err != nil {
		return model.RssFetchResult{}, err
	}

	log.Printf("RSS %s 共检查 %d 篇文章，新增 %d 篇", rssURL, result.CheckedItems, result.InsertedItems)
	return result, nil
}

func rssItemAuthor(item *gofeed.Item, fallback string) string {
	if item.Author != nil {
		if item.Author.Name != "" {
			return item.Author.Name
		}
		if item.Author.Email != "" {
			return item.Author.Email
		}
	}
	for _, candidate := range item.Authors {
		if candidate == nil {
			continue
		}
		if candidate.Name != "" {
			return candidate.Name
		}
		if candidate.Email != "" {
			return candidate.Email
		}
	}
	return fallback
}

func updateRssParseState(db *gorm.DB, friendRssID int, success bool) error {
	var rss model.FriendRss
	if err := db.Select("id, times, status, is_died").Where("id = ?", friendRssID).First(&rss).Error; err != nil {
		log.Printf("更新 RSS 解析状态前查询失败 (id=%d): %v", friendRssID, err)
		return err
	}

	newTimes, newStatus, reachedThreshold := model.ComputeFailureState(
		rss.Times,
		success,
		maxRssParseFailures,
		"survival",
		"timeout",
		"error",
	)
	newIsDied := false
	if !success {
		newIsDied = rss.IsDied
	}
	if reachedThreshold {
		newIsDied = true
	}

	if rss.Times == newTimes && rss.Status == newStatus && rss.IsDied == newIsDied {
		return nil
	}

	if err := db.Model(&model.FriendRss{}).
		Where("id = ?", friendRssID).
		Updates(map[string]interface{}{
			"times":   newTimes,
			"status":  newStatus,
			"is_died": newIsDied,
		}).Error; err != nil {
		log.Printf("更新 RSS 解析状态失败 (id=%d): %v", friendRssID, err)
		return err
	}

	log.Printf("RSS 解析状态更新 (id=%d, success=%t, times=%d, status=%s, is_died=%t)", friendRssID, success, newTimes, newStatus, newIsDied)
	return nil
}

// GetRssTitle fetches and returns the title of an RSS feed.
func GetRssTitle(rssURL string) (string, error) {
	feed, err := parseFeedURL(context.Background(), rssURL)
	if err != nil {
		log.Printf("解析 RSS feed %s 时出错: %v", rssURL, err)
		return "", err
	}
	return feed.Title, nil
}

// CheckAndReviveRssFeed probes a died RSS feed and revives it on success.
func CheckAndReviveRssFeed(db *gorm.DB, friendRssID int, rssURL string) {
	if _, err := parseFeedURL(context.Background(), rssURL); err != nil {
		log.Printf("失效 RSS 探活失败 %s (id=%d): %v", rssURL, friendRssID, err)
		return
	}

	if err := db.Model(&model.FriendRss{}).
		Where("id = ?", friendRssID).
		Updates(map[string]interface{}{
			"times":   0,
			"status":  "survival",
			"is_died": false,
		}).Error; err != nil {
		log.Printf("失效 RSS 复活状态写入失败 (id=%d): %v", friendRssID, err)
		return
	}

	log.Printf("失效 RSS 已复活 (id=%d, url=%s)", friendRssID, rssURL)
}
