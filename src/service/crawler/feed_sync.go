package crawlerService

import (
	"blog_api/src/model"
	friendsRepositories "blog_api/src/repositories/friend"
	"context"
	"errors"
	"fmt"
	"html"
	"log"
	"math/rand/v2"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/mmcdole/gofeed"
	"gorm.io/gorm"
)

// Scheduling knobs. A healthy feed is polled at roughly half its average post
// gap, clamped to [minFetchInterval, maxFetchInterval]. Failures back off
// exponentially; a dead feed is still probed every deadFetchInterval.
const (
	minFetchInterval     = time.Hour
	maxFetchInterval     = 24 * time.Hour
	defaultFetchInterval = 6 * time.Hour
	retryBaseInterval    = 30 * time.Minute
	deadFetchInterval    = 48 * time.Hour
	maxFeedFailures      = 5
	dueFeedsPerRun       = 200
	maxDescriptionRunes  = 500
)

// ErrFeedExists reports that the feed URL is already subscribed.
var ErrFeedExists = errors.New("feed already subscribed")

var stripTags = bluemonday.StrictPolicy()

// RefreshSummary summarizes one scheduler run.
type RefreshSummary struct {
	Feeds         int
	NotModified   int
	Failures      int
	InsertedItems int
}

type fetchOutcome struct {
	feed model.RssFeed
	resp feedResponse
	err  error
}

// RefreshDueFeeds fetches every feed whose schedule is due. Fetching runs on a
// worker pool; results are written by this goroutine alone.
func RefreshDueFeeds(ctx context.Context, db *gorm.DB) (RefreshSummary, error) {
	var summary RefreshSummary
	feeds, err := friendsRepositories.DueRssFeeds(db, time.Now().Unix(), dueFeedsPerRun)
	if err != nil || len(feeds) == 0 {
		return summary, err
	}

	jobs := make(chan model.RssFeed)
	outcomes := make(chan fetchOutcome)
	var wg sync.WaitGroup
	for range effectiveConcurrency(len(feeds)) {
		wg.Go(func() {
			for feed := range jobs {
				resp, err := fetchFeed(ctx, feed.FeedURL, feed.ETag, feed.LastModified)
				outcomes <- fetchOutcome{feed: feed, resp: resp, err: err}
			}
		})
	}
	go func() {
		defer close(jobs)
		for _, feed := range feeds {
			select {
			case jobs <- feed:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		wg.Wait()
		close(outcomes)
	}()

	for out := range outcomes {
		if ctx.Err() != nil {
			continue
		}
		summary.Feeds++
		result, err := applyFetchOutcome(db, out)
		switch {
		case err != nil:
			log.Printf("[RSS] 写入订阅源 %d 失败: %v", out.feed.ID, err)
			summary.Failures++
		case out.err != nil:
			summary.Failures++
		case result.NotModified:
			summary.NotModified++
		}
		summary.InsertedItems += result.InsertedItems
	}
	log.Printf("[RSS] 本轮抓取 %d 个订阅源：未变化 %d，失败 %d，新增文章 %d", summary.Feeds, summary.NotModified, summary.Failures, summary.InsertedItems)
	return summary, ctx.Err()
}

// RefreshFeed fetches one feed right away, ignoring its schedule and cached
// validators. Remote failures wrap ErrRssSource.
func RefreshFeed(ctx context.Context, db *gorm.DB, feed model.RssFeed) (model.RssFetchResult, error) {
	resp, err := fetchFeed(ctx, feed.FeedURL, "", "")
	if ctx.Err() != nil {
		return model.RssFetchResult{}, ctx.Err()
	}
	result, dbErr := applyFetchOutcome(db, fetchOutcome{feed: feed, resp: resp, err: err})
	if dbErr != nil {
		return result, dbErr
	}
	if err != nil {
		return result, fmt.Errorf("%w: %s", ErrRssSource, describeFetchError(err))
	}
	return result, nil
}

// Subscribe validates a feed URL by fetching it, stores the feed and its
// current articles. A zero friendLinkID means no owner; an empty name falls
// back to the feed title.
func Subscribe(ctx context.Context, db *gorm.DB, feedURL string, friendLinkID int, name string) (model.RssFeed, model.RssFetchResult, error) {
	feedURL = strings.TrimSpace(feedURL)
	if taken, err := friendsRepositories.RssFeedURLTaken(db, feedURL, 0); err != nil || taken {
		return model.RssFeed{}, model.RssFetchResult{}, cmpErr(err, ErrFeedExists)
	}

	resp, err := fetchFeed(ctx, feedURL, "", "")
	if err != nil {
		return model.RssFeed{}, model.RssFetchResult{}, fmt.Errorf("%w: %s", ErrRssSource, describeFetchError(err))
	}
	if resp.movedTo != "" {
		feedURL = resp.movedTo
		resp.movedTo = ""
		if taken, err := friendsRepositories.RssFeedURLTaken(db, feedURL, 0); err != nil || taken {
			return model.RssFeed{}, model.RssFetchResult{}, cmpErr(err, ErrFeedExists)
		}
	}

	feed := model.RssFeed{
		Name:    strings.TrimSpace(name),
		FeedURL: feedURL,
		Status:  model.FeedStatusActive,
	}
	if feed.Name == "" {
		feed.Name = strings.TrimSpace(resp.feed.Title)
	}
	if feed.Name == "" {
		u, _ := url.Parse(feedURL)
		feed.Name = u.Host
	}
	if friendLinkID > 0 {
		feed.FriendLinkID = &friendLinkID
	}
	if err := friendsRepositories.CreateRssFeed(db, &feed); err != nil {
		return model.RssFeed{}, model.RssFetchResult{}, fmt.Errorf("create feed: %w", err)
	}

	result, err := applyFetchOutcome(db, fetchOutcome{feed: feed, resp: resp})
	if err != nil {
		return feed, result, err
	}
	feed, err = friendsRepositories.GetRssFeedByID(db, feed.ID)
	return feed, result, err
}

// SubscribeDiscovered subscribes a friend link to its best discovered feed
// when it has none yet, and reports whether a feed was added. Links that
// already own a feed are left alone, so a feed removed by the admin stays
// removed and comment feeds never pile up.
func SubscribeDiscovered(ctx context.Context, db *gorm.DB, link model.FriendWebsite, candidates []model.FeedCandidate) bool {
	if !link.EnableRss || len(candidates) == 0 {
		return false
	}
	has, err := friendsRepositories.FriendLinkHasRssFeeds(db, link.ID)
	if err != nil || has {
		return false
	}
	for _, candidate := range candidates {
		feed, _, err := Subscribe(ctx, db, candidate.URL, link.ID, "")
		switch {
		case err == nil:
			log.Printf("[RSS] 为友链 %s 订阅 %s", link.Name, feed.FeedURL)
			return true
		case errors.Is(err, ErrFeedExists):
			return false
		default:
			log.Printf("[RSS] 友链 %s 的候选订阅源 %s 不可用: %v", link.Name, candidate.URL, err)
		}
	}
	return false
}

// applyFetchOutcome persists one fetch: new articles plus the feed's health
// and next schedule, in a single transaction.
func applyFetchOutcome(db *gorm.DB, out fetchOutcome) (model.RssFetchResult, error) {
	now := time.Now()
	feed := out.feed
	updates := map[string]any{"last_fetched_at": now.Unix()}

	if out.err != nil {
		updates["error_count"] = feed.ErrorCount + 1
		updates["last_error"] = describeFetchError(out.err)
		status, delay := failureSchedule(feed, out.err)
		updates["status"] = status
		updates["next_fetch_at"] = now.Add(delay).Unix()
		log.Printf("[RSS] 抓取 %s 失败 (%d 次，%s): %s", feed.FeedURL, feed.ErrorCount+1, status, updates["last_error"])
		_, err := friendsRepositories.UpdateRssFeed(db, feed.ID, updates)
		return model.RssFetchResult{}, err
	}

	resp := out.resp
	result := model.RssFetchResult{NotModified: resp.notModified}
	interval := time.Duration(feed.FetchInterval) * time.Second
	var posts []model.RssPost
	if resp.feed != nil {
		posts = buildPosts(feed, resp.feed, now)
		result.CheckedItems = len(resp.feed.Items)
		interval = pollInterval(posts)
		updates["feed_type"] = resp.feed.FeedType
		updates["site_url"] = resolveRef(feed.FeedURL, resp.feed.Link)
	}
	if interval <= 0 {
		interval = defaultFetchInterval
	}
	if feed.Status != model.FeedStatusPaused {
		updates["status"] = model.FeedStatusActive
	}
	updates["error_count"] = 0
	updates["last_error"] = ""
	updates["etag"] = resp.etag
	updates["last_modified"] = resp.lastModified
	updates["last_success_at"] = now.Unix()
	updates["fetch_interval"] = int64(interval / time.Second)
	updates["next_fetch_at"] = now.Add(jitter(interval)).Unix()

	err := db.Transaction(func(tx *gorm.DB) error {
		if resp.movedTo != "" {
			taken, err := friendsRepositories.RssFeedURLTaken(tx, resp.movedTo, feed.ID)
			if err != nil {
				return err
			}
			if !taken {
				log.Printf("[RSS] 订阅源 %d 已永久迁移到 %s", feed.ID, resp.movedTo)
				updates["feed_url"] = resp.movedTo
			}
		}
		inserted, err := friendsRepositories.InsertRssPosts(tx, posts)
		if err != nil {
			return err
		}
		result.InsertedItems = inserted
		_, err = friendsRepositories.UpdateRssFeed(tx, feed.ID, updates)
		return err
	})
	if err != nil {
		return model.RssFetchResult{}, err
	}
	if result.InsertedItems > 0 {
		log.Printf("[RSS] %s 新增 %d 篇文章", feed.FeedURL, result.InsertedItems)
	}
	return result, nil
}

// failureSchedule decides the status and retry delay after a failed fetch.
func failureSchedule(feed model.RssFeed, err error) (string, time.Duration) {
	status := feed.Status
	var statusErr *httpStatusError
	gone := errors.As(err, &statusErr) && statusErr.code == http.StatusGone
	if status != model.FeedStatusPaused && (gone || feed.ErrorCount+1 >= maxFeedFailures) {
		status = model.FeedStatusDead
	}
	if status == model.FeedStatusDead {
		return status, deadFetchInterval
	}

	delay := min(retryBaseInterval<<feed.ErrorCount, maxFetchInterval)
	if statusErr != nil && statusErr.retryAfter > delay {
		delay = min(statusErr.retryAfter, maxFetchInterval)
	}
	return status, delay
}

// pollInterval estimates how often to poll from the gaps between the newest
// posts: half the average gap, clamped.
func pollInterval(posts []model.RssPost) time.Duration {
	times := make([]int64, 0, len(posts))
	for _, p := range posts {
		times = append(times, p.PublishedAt)
	}
	slices.Sort(times)
	slices.Reverse(times)
	times = times[:min(len(times), 10)]
	if len(times) < 2 {
		return defaultFetchInterval
	}
	avgGap := time.Duration(times[0]-times[len(times)-1]) * time.Second / time.Duration(len(times)-1)
	return min(max(avgGap/2, minFetchInterval), maxFetchInterval)
}

// jitter spreads feeds over +-10% so they do not stay in lockstep.
func jitter(d time.Duration) time.Duration {
	return d + time.Duration((rand.Float64()*0.2-0.1)*float64(d))
}

// buildPosts converts feed items into rows. Items without any link are
// dropped; a missing date falls back to the fetch time.
func buildPosts(feed model.RssFeed, parsed *gofeed.Feed, now time.Time) []model.RssPost {
	base := resolveRef(feed.FeedURL, parsed.Link)
	posts := make([]model.RssPost, 0, len(parsed.Items))
	for _, item := range parsed.Items {
		link := resolveRef(base, strings.TrimSpace(item.Link))
		guid := strings.TrimSpace(item.GUID)
		if link == "" && strings.HasPrefix(guid, "http") {
			link = guid
		}
		if link == "" {
			continue
		}
		if guid == "" {
			guid = link
		}

		published := now
		if item.PublishedParsed != nil {
			published = *item.PublishedParsed
		} else if item.UpdatedParsed != nil {
			published = *item.UpdatedParsed
		}
		if published.After(now) || published.Unix() < 0 {
			published = now
		}

		title := strings.TrimSpace(item.Title)
		if title == "" {
			title = link
		}
		description := item.Description
		if description == "" {
			description = item.Content
		}

		posts = append(posts, model.RssPost{
			FeedID:      feed.ID,
			GUID:        guid,
			Title:       title,
			Link:        link,
			Description: summarize(description),
			Author:      itemAuthor(item, feed.Name),
			PublishedAt: published.Unix(),
			FetchedAt:   now.Unix(),
		})
	}
	return posts
}

func itemAuthor(item *gofeed.Item, fallback string) string {
	people := item.Authors
	if item.Author != nil {
		people = append([]*gofeed.Person{item.Author}, people...)
	}
	for _, p := range people {
		if p == nil {
			continue
		}
		if p.Name != "" {
			return p.Name
		}
		if p.Email != "" {
			return p.Email
		}
	}
	return fallback
}

// resolveRef resolves ref against base; it returns ref unchanged when either
// side does not parse and "" for an empty ref.
func resolveRef(base, ref string) string {
	if ref == "" {
		return ""
	}
	b, err := url.Parse(base)
	if err != nil {
		return ref
	}
	r, err := b.Parse(ref)
	if err != nil {
		return ref
	}
	return r.String()
}

// summarize strips markup and cuts the text to maxDescriptionRunes. The
// result stays HTML-escaped, so consumers may render it as HTML safely.
func summarize(s string) string {
	text := strings.Join(strings.Fields(html.UnescapeString(stripTags.Sanitize(s))), " ")
	if runes := []rune(text); len(runes) > maxDescriptionRunes {
		text = string(runes[:maxDescriptionRunes]) + "…"
	}
	return html.EscapeString(text)
}

// cmpErr returns err when set, otherwise fallback.
func cmpErr(err, fallback error) error {
	if err != nil {
		return err
	}
	return fallback
}
