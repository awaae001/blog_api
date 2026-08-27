package crawlerService

import (
	"blog_api/src/model"
	friendsRepositories "blog_api/src/repositories/friend"
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

// ErrInspectionBusy reports that another friend link inspection still holds the
// crawler. Scheduled jobs and manual full scans read and write the same rows,
// so exactly one run is allowed at a time.
var ErrInspectionBusy = errors.New("friend link inspection is already running")

// friendInspectionMu serializes every friend link inspection: the scheduled
// survival scan, the scheduled died-link scan and the manual full scan.
var friendInspectionMu sync.Mutex

// InspectionScope selects which friend links a run covers and what it does
// with the crawl results.
type InspectionScope struct {
	// IsDied filters by the stored died flag. A nil value covers every link.
	IsDied *bool

	// SkipHealthCheck filters by the "do not inspect" flag. A nil value
	// ignores the flag entirely.
	SkipHealthCheck *bool

	// DiscoverRss enables RSS feed discovery for links that opted in.
	DiscoverRss bool

	// Label names the run in logs and in the reported progress.
	Label string
}

// InspectionProgress is an observable snapshot of the current or last run. It
// doubles as the result of a finished run, so counters are never tracked twice.
type InspectionProgress struct {
	Running          bool   `json:"running"`
	Label            string `json:"label,omitempty"`
	StartedAt        int64  `json:"started_at,omitempty"`
	FinishedAt       int64  `json:"finished_at,omitempty"`
	Total            int    `json:"total"`
	Processed        int    `json:"processed"`
	Survival         int    `json:"survival"`
	Failed           int    `json:"failed"`
	DatabaseFailures int    `json:"database_failures"`
	RssDiscovered    int    `json:"rss_discovered"`
	Error            string `json:"error,omitempty"`
}

// LinkInspectionProgress describes one link's position inside the current or
// last run.
type LinkInspectionProgress struct {
	// InRun reports whether the run covered this link at all.
	InRun bool `json:"in_run"`

	// Done reports whether the run already produced a result for this link.
	Done bool `json:"done"`

	// Status is the crawl status recorded by this run, empty until Done.
	Status string `json:"status,omitempty"`

	// Run is the overall progress of the run this link belongs to.
	Run InspectionProgress `json:"run"`
}

// inspectionState holds the in-memory progress of the single active run. It is
// intentionally not persisted: it describes one process lifetime only.
var inspectionState = struct {
	mu       sync.Mutex
	progress InspectionProgress
	links    map[int]string // link id -> crawl status ("" while pending)
	done     map[int]bool
}{}

// InspectFriendLinks runs one inspection synchronously and returns its final
// progress. It returns ErrInspectionBusy when another run is in flight, which
// callers that only schedule work can ignore: the skip is already logged.
func InspectFriendLinks(ctx context.Context, db *gorm.DB, scope InspectionScope) (InspectionProgress, error) {
	label := scopeLabel(scope)
	if !friendInspectionMu.TryLock() {
		log.Printf("[Inspection][%s] 已有巡查任务在执行，本次跳过", label)
		return InspectionProgressSnapshot(), ErrInspectionBusy
	}
	defer friendInspectionMu.Unlock()

	beginRun(label)
	return runInspection(ctx, db, scope)
}

// StartFriendLinksInspection claims the crawler and runs one inspection in the
// background. Progress becomes observable before this call returns, so a client
// that polls immediately after a successful start always sees a running run.
func StartFriendLinksInspection(db *gorm.DB, scope InspectionScope) error {
	label := scopeLabel(scope)
	if !friendInspectionMu.TryLock() {
		log.Printf("[Inspection][%s] 已有巡查任务在执行，拒绝启动", label)
		return ErrInspectionBusy
	}

	beginRun(label)
	go func() {
		defer friendInspectionMu.Unlock()
		// The run outlives the HTTP request that started it, so it cannot
		// inherit that request's context.
		runInspection(context.Background(), db, scope)
	}()
	return nil
}

// FullInspectionScope is the scope of a manual "inspect everything" run. Links
// that opted out of health checks stay excluded: that flag is an explicit
// instruction not to crawl the site on a schedule or in bulk.
func FullInspectionScope() InspectionScope {
	skipHealthCheck := false
	return InspectionScope{
		SkipHealthCheck: &skipHealthCheck,
		DiscoverRss:     true,
		Label:           "全量巡查",
	}
}

// SurvivalInspectionScope covers the links that are currently alive.
func SurvivalInspectionScope() InspectionScope {
	isDied := false
	skipHealthCheck := false
	return InspectionScope{
		IsDied:          &isDied,
		SkipHealthCheck: &skipHealthCheck,
		DiscoverRss:     true,
		Label:           "友链爬取",
	}
}

// DiedInspectionScope covers the links already marked as died.
func DiedInspectionScope() InspectionScope {
	isDied := true
	skipHealthCheck := false
	return InspectionScope{
		IsDied:          &isDied,
		SkipHealthCheck: &skipHealthCheck,
		DiscoverRss:     false,
		Label:           "失效友链检查",
	}
}

// InspectionProgressSnapshot returns the current or last run progress.
func InspectionProgressSnapshot() InspectionProgress {
	inspectionState.mu.Lock()
	defer inspectionState.mu.Unlock()
	return inspectionState.progress
}

// LinkInspectionSnapshot returns one link's position inside the current or last
// run.
func LinkInspectionSnapshot(linkID int) LinkInspectionProgress {
	inspectionState.mu.Lock()
	defer inspectionState.mu.Unlock()

	status, inRun := inspectionState.links[linkID]
	return LinkInspectionProgress{
		InRun:  inRun,
		Done:   inspectionState.done[linkID],
		Status: status,
		Run:    inspectionState.progress,
	}
}

// runInspection performs one inspection. The caller must already hold
// friendInspectionMu and must have started a progress run.
func runInspection(ctx context.Context, db *gorm.DB, scope InspectionScope) (InspectionProgress, error) {
	label := scopeLabel(scope)
	log.Printf("[Inspection][%s] 开始巡查", label)

	opts := model.FriendLinkQueryOptions{
		IsDied:          scope.IsDied,
		SkipHealthCheck: scope.SkipHealthCheck,
	}
	resp, err := friendsRepositories.QueryFriendLinks(db, opts)
	if err != nil {
		queryErr := fmt.Errorf("查询友链失败: %w", err)
		log.Printf("[Inspection][%s] %v", label, queryErr)
		return finishRun(queryErr), queryErr
	}

	links := resp.Links
	inspectionState.mu.Lock()
	inspectionState.links = make(map[int]string, len(links))
	inspectionState.done = make(map[int]bool, len(links))
	for _, link := range links {
		inspectionState.links[link.ID] = ""
	}
	inspectionState.progress.Total = len(links)
	inspectionState.mu.Unlock()

	if len(links) == 0 {
		log.Printf("[Inspection][%s] 没有需要巡查的友链", label)
		return finishRun(nil), nil
	}

	crawlErr := CrawlWebsitesConcurrently(ctx, links, func(jobResult CrawlJobResult) error {
		link := jobResult.Link
		result := jobResult.Result

		databaseFailed := false
		if err := friendsRepositories.UpdateFriendLink(db, link, result); err != nil {
			databaseFailed = true
			log.Printf("[Inspection][%s] 更新友链 %s 失败: %v", label, link.Name, err)
		}

		discovered := 0
		if scope.DiscoverRss && link.EnableRss && len(result.RssURLs) > 0 {
			discovered = discoverRssFeeds(db, label, link, result.RssURLs)
		}

		// consume 由单一 goroutine 顺序调用，加锁只是为了让 HTTP 读到一致快照。
		inspectionState.mu.Lock()
		inspectionState.links[link.ID] = result.Status
		inspectionState.done[link.ID] = true
		inspectionState.progress.Processed++
		if result.Status == StatusSurvival {
			inspectionState.progress.Survival++
		} else {
			inspectionState.progress.Failed++
		}
		if databaseFailed {
			inspectionState.progress.DatabaseFailures++
		}
		inspectionState.progress.RssDiscovered += discovered
		inspectionState.mu.Unlock()

		// One failed row must not abort the batch: later links still deserve
		// to be inspected, so the failure is counted instead of returned.
		return nil
	})

	progress := finishRun(crawlErr)
	if crawlErr != nil {
		log.Printf("[Inspection][%s] 巡查中断: %v", label, crawlErr)
		return progress, crawlErr
	}

	log.Printf(
		"[Inspection][%s] 巡查完成：共 %d 个，处理 %d 个，存活 %d 个，失败 %d 个，数据库失败 %d 个，新增 RSS %d 个",
		label,
		progress.Total,
		progress.Processed,
		progress.Survival,
		progress.Failed,
		progress.DatabaseFailures,
		progress.RssDiscovered,
	)
	return progress, nil
}

// discoverRssFeeds registers the feeds found for one link and returns how many
// were stored.
func discoverRssFeeds(db *gorm.DB, label string, link model.FriendWebsite, rssURLs []string) int {
	inserted := 0
	for _, rssURL := range rssURLs {
		name, err := GetRssTitle(rssURL)
		if err != nil {
			log.Printf("[Inspection][%s] 获取 RSS 标题失败 %s: %v", label, rssURL, err)
			continue
		}
		if _, err := friendsRepositories.CreateFriendRssFeeds(db, link.ID, rssURL, name); err != nil {
			log.Printf("[Inspection][%s] 为 %s 插入 RSS 订阅源失败: %v", label, link.Name, err)
			continue
		}
		inserted++
	}
	return inserted
}

func scopeLabel(scope InspectionScope) string {
	if scope.Label == "" {
		return "友链巡查"
	}
	return scope.Label
}

// beginRun publishes a fresh running state before any crawling starts.
func beginRun(label string) {
	inspectionState.mu.Lock()
	defer inspectionState.mu.Unlock()

	inspectionState.progress = InspectionProgress{
		Running:   true,
		Label:     label,
		StartedAt: time.Now().Unix(),
	}
	inspectionState.links = make(map[int]string)
	inspectionState.done = make(map[int]bool)
}

// finishRun closes the run and returns its final progress.
func finishRun(err error) InspectionProgress {
	inspectionState.mu.Lock()
	defer inspectionState.mu.Unlock()

	inspectionState.progress.Running = false
	inspectionState.progress.FinishedAt = time.Now().Unix()
	if err != nil {
		inspectionState.progress.Error = err.Error()
	}
	return inspectionState.progress
}
