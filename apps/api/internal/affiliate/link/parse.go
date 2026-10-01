package link

import (
	"bytes"
	"encoding/json"
	"net/url"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/html"
)

// meta is what a page publicly declares about itself for link previews
// (Open Graph, Twitter cards, schema.org Product). Prices are deliberately
// not read: they would be unverified data presented as real.
type meta struct {
	Title       string
	Image       string
	Description string
	Brand       string
	SiteName    string
}

func parseMeta(body []byte, base *url.URL) meta {
	var (
		og      = map[string]string{}
		titleEl string
		ld      meta
	)
	z := html.NewTokenizer(bytes.NewReader(body))
	inTitle, inLD := false, false
	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		}
		switch tt {
		case html.StartTagToken, html.SelfClosingTagToken:
			name, hasAttr := z.TagName()
			switch string(name) {
			case "meta":
				var key, content string
				for hasAttr {
					var k, v []byte
					k, v, hasAttr = z.TagAttr()
					switch string(k) {
					case "property", "name":
						if key == "" {
							key = strings.ToLower(string(v))
						}
					case "content":
						content = string(v)
					}
				}
				if key != "" && content != "" {
					if _, seen := og[key]; !seen {
						og[key] = content
					}
				}
			case "title":
				inTitle = tt == html.StartTagToken && titleEl == ""
			case "script":
				inLD = false
				for hasAttr {
					var k, v []byte
					k, v, hasAttr = z.TagAttr()
					if string(k) == "type" && strings.EqualFold(string(v), "application/ld+json") {
						inLD = true
					}
				}
			}
		case html.TextToken:
			if inTitle {
				titleEl += string(z.Text())
			} else if inLD && ld.Title == "" {
				if m, ok := productFromJSONLD(z.Text()); ok {
					ld = m
				}
			}
		case html.EndTagToken:
			name, _ := z.TagName()
			switch string(name) {
			case "title":
				inTitle = false
			case "script":
				inLD = false
			}
		}
	}

	pick := func(vals ...string) string {
		for _, v := range vals {
			if v = clean(v); v != "" {
				return v
			}
		}
		return ""
	}
	m := meta{
		Title:       clip(pick(ld.Title, og["og:title"], og["twitter:title"], titleEl), 200),
		Description: clip(pick(ld.Description, og["og:description"], og["twitter:description"], og["description"]), 500),
		Brand:       clip(pick(ld.Brand), 80),
		SiteName:    clip(pick(og["og:site_name"]), 80),
	}
	m.Image = absoluteHTTP(pick(ld.Image, og["og:image"], og["og:image:url"], og["twitter:image"]), base)
	return m
}

// productFromJSONLD extracts a schema.org Product from a JSON-LD block, which
// may be an object, an array, or an @graph.
func productFromJSONLD(raw []byte) (meta, bool) {
	var v any
	if err := json.Unmarshal(bytes.TrimSpace(raw), &v); err != nil {
		return meta{}, false
	}
	return findProduct(v, 0)
}

func findProduct(v any, depth int) (meta, bool) {
	if depth > 4 {
		return meta{}, false
	}
	switch t := v.(type) {
	case []any:
		for _, e := range t {
			if m, ok := findProduct(e, depth+1); ok {
				return m, true
			}
		}
	case map[string]any:
		if isProduct(t["@type"]) {
			m := meta{
				Title:       str(t["name"]),
				Description: str(t["description"]),
				Image:       imageOf(t["image"]),
				Brand:       str(t["brand"]),
			}
			if m.Title != "" {
				return m, true
			}
		}
		if g, ok := t["@graph"]; ok {
			return findProduct(g, depth+1)
		}
	}
	return meta{}, false
}

func isProduct(t any) bool {
	switch v := t.(type) {
	case string:
		return v == "Product"
	case []any:
		for _, e := range v {
			if s, ok := e.(string); ok && s == "Product" {
				return true
			}
		}
	}
	return false
}

func str(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case map[string]any:
		return str(t["name"])
	}
	return ""
}

func imageOf(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case []any:
		if len(t) > 0 {
			return imageOf(t[0])
		}
	case map[string]any:
		return str(t["url"])
	}
	return ""
}

func clean(s string) string { return strings.Join(strings.Fields(s), " ") }

func clip(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	r := []rune(s)
	return strings.TrimSpace(string(r[:max-1])) + "…"
}

func absoluteHTTP(raw string, base *url.URL) string {
	if raw == "" {
		return ""
	}
	u, err := base.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || len(u.String()) > maxURLLength {
		return ""
	}
	return u.String()
}
