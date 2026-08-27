package crawlerService

import (
	"blog_api/src/config"
	"blog_api/src/model"
	"context"
	"os"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// newInspectionTestDB opens a throwaway SQLite database with the friend link
// schema applied.
func newInspectionTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(t.TempDir()+"/test.db"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	content, err := os.ReadFile("../../../migrations/001_01_create_frined_link.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	if err := db.Exec(string(content)).Error; err != nil {
		t.Fatalf("apply migration: %v", err)
	}
	return db
}

// TestInspectFriendLinksFullScope checks the scope of a manual full run and the
// progress it publishes.
//
// Every target is a loopback address, which the SSRF guard rejects at dial time.
// The run therefore needs no network and every covered link fails the same way.
func TestInspectFriendLinksFullScope(t *testing.T) {
	t.Setenv("CONFIG_PATH", t.TempDir())
	if _, err := config.Load(); err != nil {
		t.Fatalf("load config: %v", err)
	}

	db := newInspectionTestDB(t)
	links := []model.FriendWebsite{
		{Name: "alive", Link: "http://127.0.0.1:1/", Info: "x", Status: "survival"},
		{Name: "died", Link: "http://127.0.0.1:2/", Info: "x", Status: "error", IsDied: true},
		{Name: "skipped", Link: "http://127.0.0.1:3/", Info: "x", Status: "survival", SkipHealthCheck: true},
	}
	for i := range links {
		if err := db.Create(&links[i]).Error; err != nil {
			t.Fatalf("insert friend link %s: %v", links[i].Name, err)
		}
	}

	progress, err := InspectFriendLinks(context.Background(), db, FullInspectionScope())
	if err != nil {
		t.Fatalf("inspect friend links: %v", err)
	}

	// Died links are included, links that opted out of health checks are not.
	if progress.Total != 2 || progress.Processed != 2 {
		t.Fatalf("progress total/processed = %d/%d, want 2/2", progress.Total, progress.Processed)
	}
	if progress.Failed != 2 || progress.Survival != 0 {
		t.Fatalf("progress failed/survival = %d/%d, want 2/0", progress.Failed, progress.Survival)
	}
	if progress.Running {
		t.Fatal("progress still reports a running run after the call returned")
	}
	if progress.DatabaseFailures != 0 {
		t.Fatalf("database failures = %d, want 0", progress.DatabaseFailures)
	}

	for _, name := range []string{"alive", "died"} {
		var stored model.FriendWebsite
		if err := db.Where("website_name = ?", name).First(&stored).Error; err != nil {
			t.Fatalf("reload %s: %v", name, err)
		}
		if stored.Times != 1 || stored.Status != StatusError {
			t.Fatalf("%s: times/status = %d/%s, want 1/%s", name, stored.Times, stored.Status, StatusError)
		}
	}

	var skipped model.FriendWebsite
	if err := db.Where("website_name = ?", "skipped").First(&skipped).Error; err != nil {
		t.Fatalf("reload skipped: %v", err)
	}
	if skipped.Times != 0 || skipped.Status != "survival" {
		t.Fatalf("skipped link was inspected: times/status = %d/%s", skipped.Times, skipped.Status)
	}

	covered := LinkInspectionSnapshot(links[0].ID)
	if !covered.InRun || !covered.Done || covered.Status != StatusError {
		t.Fatalf("covered link snapshot = %+v, want in run, done, status %s", covered, StatusError)
	}
	if excluded := LinkInspectionSnapshot(links[2].ID); excluded.InRun {
		t.Fatalf("excluded link snapshot = %+v, want in_run false", excluded)
	}
}

// TestInspectFriendLinksRejectsConcurrentRun verifies that a second inspection
// is refused while one holds the crawler.
func TestInspectFriendLinksRejectsConcurrentRun(t *testing.T) {
	friendInspectionMu.Lock()
	defer friendInspectionMu.Unlock()

	if _, err := InspectFriendLinks(context.Background(), nil, FullInspectionScope()); err != ErrInspectionBusy {
		t.Fatalf("synchronous run error = %v, want %v", err, ErrInspectionBusy)
	}
	if err := StartFriendLinksInspection(nil, FullInspectionScope()); err != ErrInspectionBusy {
		t.Fatalf("background run error = %v, want %v", err, ErrInspectionBusy)
	}
}
