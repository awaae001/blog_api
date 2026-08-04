package crawlerService

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"time"
)

// ErrBlockedCrawlTarget is returned when a crawl target fails SSRF validation.
var ErrBlockedCrawlTarget = errors.New("crawl target is not a public HTTP(S) address")

// blockedPrefixes lists non-public ranges not covered by netip helpers:
// netip.IsPrivate covers RFC1918 + fc00::/7, IsLoopback 127.0.0.0/8 + ::1,
// IsLinkLocalUnicast 169.254.0.0/16 + fe80::/10 (cloud metadata included).
var blockedPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),     // "this network" (IsUnspecified only matches 0.0.0.0)
	netip.MustParsePrefix("100.64.0.0/10"), // CGNAT shared address space
	netip.MustParsePrefix("192.0.0.0/24"),  // IETF protocol assignments
	netip.MustParsePrefix("198.18.0.0/15"), // benchmarking
	netip.MustParsePrefix("240.0.0.0/4"),   // reserved, includes broadcast
}

// ValidatePublicHTTPURL enforces the scheme whitelist and basic shape of a
// crawl target. IP-level checks happen at dial time in safeDialContext.
func ValidatePublicHTTPURL(rawURL string) (*url.URL, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid URL: %v", ErrBlockedCrawlTarget, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("%w: unsupported scheme %q", ErrBlockedCrawlTarget, u.Scheme)
	}
	if u.Hostname() == "" {
		return nil, fmt.Errorf("%w: missing host", ErrBlockedCrawlTarget)
	}
	return u, nil
}

// isBlockedIP reports whether ip is loopback, private, link-local, multicast,
// unspecified, or in another non-public range.
func isBlockedIP(ip netip.Addr) bool {
	ip = ip.Unmap()
	if !ip.IsValid() || ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsMulticast() || ip.IsUnspecified() {
		return true
	}
	for _, p := range blockedPrefixes {
		if p.Contains(ip) {
			return true
		}
	}
	return false
}

// safeDialContext resolves the host itself, rejects any non-public resolved
// address, and then dials a validated IP directly. Dialing the checked IP
// (instead of the hostname) closes the DNS-rebinding window between the
// check and the connection. TLS is unaffected: http.Transport derives the
// ServerName from the request URL, not the dialed address.
func safeDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	dialer := &net.Dialer{Timeout: 10 * time.Second}

	if ip, err := netip.ParseAddr(host); err == nil {
		if isBlockedIP(ip) {
			return nil, fmt.Errorf("%w: blocked IP %s", ErrBlockedCrawlTarget, ip)
		}
		return dialer.DialContext(ctx, network, addr)
	}

	resolved, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	if len(resolved) == 0 {
		return nil, fmt.Errorf("no addresses resolved for host %s", host)
	}

	validated := make([]net.IP, 0, len(resolved))
	for _, r := range resolved {
		addr4or6, ok := netip.AddrFromSlice(r.IP)
		if !ok || isBlockedIP(addr4or6) {
			return nil, fmt.Errorf("%w: host %s resolves to blocked IP %s", ErrBlockedCrawlTarget, host, r.IP)
		}
		validated = append(validated, r.IP)
	}

	var lastErr error
	for _, ip := range validated {
		conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

// newSafeHTTPTransport returns a transport that only connects to public IPs.
func newSafeHTTPTransport() *http.Transport {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.DialContext = safeDialContext
	return t
}
