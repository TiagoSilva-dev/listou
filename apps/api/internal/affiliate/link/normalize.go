package link

import (
	"net/netip"
	"net/url"
	"strings"
)

// trackingParams are analytics identifiers with no effect on the product
// page. Anything else (including a partner tag) is kept as pasted.
var trackingParams = map[string]bool{
	"fbclid": true, "gclid": true, "msclkid": true, "igshid": true, "mc_cid": true, "mc_eid": true,
	"dclid": true, "yclid": true, "_ga": true,
}

// Normalize validates a pasted URL and returns its canonical form: scheme
// defaulted to https, lower-case host, no fragment, no default port, and no
// analytics parameters. Equal products therefore dedupe on the same string.
func Normalize(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > maxURLLength {
		return nil, errInvalidURL
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, errInvalidURL
	}
	if err := checkURL(u); err != nil {
		return nil, err
	}
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)
	if p := u.Port(); (p == "80" && u.Scheme == "http") || (p == "443" && u.Scheme == "https") {
		u.Host = u.Hostname()
	}
	u.Fragment = ""
	u.RawFragment = ""
	if u.RawQuery != "" {
		var keep []string
		for _, pair := range strings.Split(u.RawQuery, "&") {
			key, _, _ := strings.Cut(pair, "=")
			key = strings.ToLower(key)
			if pair == "" || strings.HasPrefix(key, "utm_") || trackingParams[key] {
				continue
			}
			keep = append(keep, pair)
		}
		u.RawQuery = strings.Join(keep, "&")
	}
	return u, nil
}

func parseHostIP(host string) (netip.Addr, error) { return netip.ParseAddr(host) }
