package crawlerService

import (
	"context"
	"errors"
	"net"
)

// Crawl status values persisted for a friend link.
const (
	// StatusSurvival marks a reachable site.
	StatusSurvival = "survival"

	// StatusTimeout marks a transient network condition worth retrying:
	// the name resolved (or resolution itself timed out) but no usable
	// response arrived in time.
	StatusTimeout = "timeout"

	// StatusError marks a definitive failure: the host does not exist,
	// actively refuses connections, or answered with a bad response.
	StatusError = "error"
)

// classifyTransportError maps a failed HTTP round trip to a crawl status and a
// short human-readable reason for the log.
//
// The distinction that matters: a domain that no longer resolves is dead and
// should not be reported the same way as a site that was merely slow. Only
// genuine timing failures stay StatusTimeout so repeated scans can tell the
// two apart.
func classifyTransportError(err error) (status string, reason string) {
	if err == nil {
		return StatusSurvival, ""
	}

	// SSRF rejection is a configuration problem with the stored URL, never a
	// statement about the remote site's health.
	if errors.Is(err, ErrBlockedCrawlTarget) {
		return StatusError, "目标地址被安全策略拒绝"
	}

	// DNS is checked before the generic timeout branch: a *net.DNSError can
	// also report Timeout, and NXDOMAIN must not be hidden behind it.
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		switch {
		case dnsErr.IsNotFound:
			return StatusError, "域名不存在 (NXDOMAIN)"
		case dnsErr.IsTimeout:
			return StatusTimeout, "DNS 查询超时"
		case dnsErr.IsTemporary:
			return StatusTimeout, "DNS 临时故障"
		default:
			return StatusError, "DNS 解析失败: " + dnsErr.Err
		}
	}

	if errors.Is(err, context.Canceled) {
		return StatusTimeout, "请求被取消"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return StatusTimeout, "请求超时"
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return StatusTimeout, "连接超时"
	}

	// The host resolved but nothing is listening, or the peer dropped the
	// connection. Both are definitive rather than transient.
	if errors.Is(err, net.ErrClosed) {
		return StatusError, "连接已关闭"
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return StatusError, "连接失败: " + opErr.Err.Error()
	}

	return StatusError, "请求失败: " + err.Error()
}
