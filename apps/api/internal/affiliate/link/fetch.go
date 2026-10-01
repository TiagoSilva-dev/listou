package link

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"syscall"
	"time"
)

const (
	maxBodyBytes = 1 << 20
	maxRedirects = 5
	fetchTimeout = 8 * time.Second
	maxURLLength = 2048
	userAgent    = "ListouLinkPreview/1.0 (+https://listou.app)"
)

var (
	errInvalidURL = errors.New("link: invalid url")
	errBlockedIP  = errors.New("link: destination not allowed")
)

// Non-global ranges that netip does not already classify as private/loopback.
var blockedPrefixes = []netip.Prefix{
	netip.MustParsePrefix("100.64.0.0/10"), // CGNAT
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("64:ff9b::/96"),
}

func ipAllowed(a netip.Addr) bool {
	a = a.Unmap()
	if !a.IsValid() || a.IsPrivate() || a.IsLoopback() || a.IsLinkLocalUnicast() || a.IsLinkLocalMulticast() ||
		a.IsMulticast() || a.IsUnspecified() {
		return false
	}
	for _, p := range blockedPrefixes {
		if p.Contains(a) {
			return false
		}
	}
	return true
}

// checkURL validates the shape of a user-supplied URL. The resolved IP is
// checked again at connection time (see guardedControl).
func checkURL(u *url.URL) error {
	if u.Scheme != "http" && u.Scheme != "https" {
		return errInvalidURL
	}
	if u.Hostname() == "" || u.User != nil {
		return errInvalidURL
	}
	if p := u.Port(); p != "" && p != "80" && p != "443" {
		return errInvalidURL
	}
	if len(u.String()) > maxURLLength {
		return errInvalidURL
	}
	return nil
}

// guardedControl rejects the connection after DNS resolution, which defeats
// DNS rebinding and hostnames that resolve to internal addresses.
func guardedControl(_, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return errBlockedIP
	}
	addr, err := netip.ParseAddr(host)
	if err != nil || !ipAllowed(addr) {
		return errBlockedIP
	}
	return nil
}

type fetcher struct {
	client *http.Client
}

// newFetcher builds an HTTP client that can only reach public hosts.
// allowPrivate exists for tests against httptest servers on loopback.
func newFetcher(allowPrivate bool) *fetcher {
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	if !allowPrivate {
		dialer.Control = guardedControl
	}
	transport := &http.Transport{
		Proxy:                  nil,
		DialContext:            dialer.DialContext,
		ForceAttemptHTTP2:      true,
		TLSHandshakeTimeout:    5 * time.Second,
		ResponseHeaderTimeout:  6 * time.Second,
		MaxResponseHeaderBytes: 64 << 10,
		DisableKeepAlives:      true,
	}
	return &fetcher{client: &http.Client{
		Transport: transport,
		Timeout:   fetchTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return errors.New("link: too many redirects")
			}
			return checkURL(req.URL)
		},
	}}
}

type page struct {
	finalURL *url.URL
	body     []byte
}

// get downloads at most maxBodyBytes of an HTML page.
func (f *fetcher) get(ctx context.Context, u *url.URL) (page, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return page{}, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9,en;q=0.5")
	resp, err := f.client.Do(req)
	if err != nil {
		return page{}, fmt.Errorf("link: fetch: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return page{}, fmt.Errorf("link: fetch: status %d", resp.StatusCode)
	}
	if ct := strings.ToLower(resp.Header.Get("Content-Type")); !strings.Contains(ct, "html") {
		return page{}, fmt.Errorf("link: fetch: content type %q", ct)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return page{}, fmt.Errorf("link: read: %w", err)
	}
	return page{finalURL: resp.Request.URL, body: body}, nil
}
