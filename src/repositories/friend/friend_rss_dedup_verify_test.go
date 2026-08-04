package friendsRepositories

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"blog_api/src/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// 临时验证：并发调用 CreateFriendRssFeeds 同 URL 只能落一行
func TestCreateFriendRssFeedsConcurrentDedup(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := gorm.Open(sqlite.Open(dbPath+"?_foreign_keys=on&_busy_timeout=5000&_journal_mode=WAL"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	for _, mig := range []string{
		"../../../migrations/001_02_create_friend_rss.sql",
		"../../../migrations/001_03_create_rss_post.sql",
		"../../../migrations/006_friend_rss_unique_url.sql",
	} {
		content, err := os.ReadFile(mig)
		if err != nil {
			t.Fatalf("read migration %s: %v", mig, err)
		}
		if err := db.Exec(string(content)).Error; err != nil {
			t.Fatalf("apply migration %s: %v", mig, err)
		}
	}

	const goroutines = 20
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := CreateFriendRssFeeds(db, 1, "http://x.com/feed", "A"); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent create error: %v", err)
	}

	var count int64
	if err := db.Model(&model.FriendRss{}).
		Where("friend_link_id = ? AND rss_url = ?", 1, "http://x.com/feed").
		Count(&count).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 row, got %d", count)
	}
}
