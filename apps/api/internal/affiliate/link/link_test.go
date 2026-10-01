package link

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"strings"
	"testing"

	"github.com/listou/listou/apps/api/internal/affiliate"
)

func TestNormalize(t *testing.T) {
	cases := []struct{ in, want string }{
		{"loja.com.br/p/1", "https://loja.com.br/p/1"},
		{"  HTTPS://Loja.COM/p?id=7&utm_source=x&fbclid=abc#frag ", "https://loja.com/p?id=7"},
		{"https://loja.com:443/a", "https://loja.com/a"},
		{"http://loja.com:80/a?b=1&c=2", "http://loja.com/a?b=1&c=2"},
		{"https://loja.com/a?tag=meu-20&ref=x", "https://loja.com/a?tag=meu-20&ref=x"},
	}
	for _, c := range cases {
		u, err := Normalize(c.in)
		if err != nil || u.String() != c.want {
			t.Errorf("Normalize(%q) = %v, %v; want %q", c.in, u, err, c.want)
		}
	}
	for _, bad := range []string{"", "ftp://x.com/a", "javascript:alert(1)", "https://user:pw@x.com/", "https://x.com:8080/", "https:///nohost", "file:///etc/passwd", "https://x.com/" + strings.Repeat("a", 2100)} {
		if _, err := Normalize(bad); err == nil {
			t.Errorf("Normalize(%q) should fail", bad)
		}
	}
}

func TestIPAllowed(t *testing.T) {
	for ip, want := range map[string]bool{
		"8.8.8.8": true, "2606:4700:4700::1111": true,
		"127.0.0.1": false, "10.1.2.3": false, "172.16.0.1": false, "192.168.1.1": false,
		"169.254.169.254": false, "100.64.0.1": false, "0.0.0.0": false, "::1": false,
		"fe80::1": false, "fd00::1": false, "::ffff:127.0.0.1": false, "224.0.0.1": false, "240.0.0.1": false,
	} {
		if got := ipAllowed(netip.MustParseAddr(ip)); got != want {
			t.Errorf("ipAllowed(%s) = %v, want %v", ip, got, want)
		}
	}
}

const productPage = `<!doctype html><html><head>
<title>  Fallback   title </title>
<meta property="og:title" content="Air Fryer 5L Preta">
<meta property="og:image" content="/img/af.jpg">
<meta property="og:site_name" content="Loja Teste">
<meta property="og:description" content="Fritadeira sem óleo">
</head><body></body></html>`

func TestParseMetaOpenGraph(t *testing.T) {
	base, _ := url.Parse("https://loja.com/produto/1")
	m := parseMeta([]byte(productPage), base)
	if m.Title != "Air Fryer 5L Preta" || m.SiteName != "Loja Teste" || m.Image != "https://loja.com/img/af.jpg" || m.Description != "Fritadeira sem óleo" {
		t.Fatalf("unexpected meta: %+v", m)
	}
}

func TestParseMetaJSONLDWins(t *testing.T) {
	page := `<html><head><title>T</title><meta property="og:title" content="OG">
<script type="application/ld+json">{"@context":"https://schema.org","@graph":[{"@type":"WebSite","name":"Site"},
{"@type":["Product"],"name":"Panela de Pressão","image":["https://c.com/a.jpg"],"brand":{"@type":"Brand","name":"Marca"},"offers":{"price":"99.90"}}]}</script>
</head></html>`
	base, _ := url.Parse("https://loja.com/")
	m := parseMeta([]byte(page), base)
	if m.Title != "Panela de Pressão" || m.Brand != "Marca" || m.Image != "https://c.com/a.jpg" {
		t.Fatalf("unexpected meta: %+v", m)
	}
}

func TestParseMetaFallbacksAndSafety(t *testing.T) {
	base, _ := url.Parse("https://loja.com/")
	m := parseMeta([]byte(`<html><head><title>Só título</title><meta property="og:image" content="javascript:alert(1)"></head></html>`), base)
	if m.Title != "Só título" || m.Image != "" {
		t.Fatalf("unexpected meta: %+v", m)
	}
	long := strings.Repeat("ab", 300)
	m = parseMeta([]byte(`<title>`+long+`</title>`), base)
	if len([]rune(m.Title)) != 200 {
		t.Fatalf("title not clipped: %d", len([]rune(m.Title)))
	}
	if m := parseMeta([]byte(`<html><body>garbage <<<`), base); m.Title != "" {
		t.Fatalf("expected empty meta, got %+v", m)
	}
}

func testProvider() *Provider { return &Provider{f: newFetcher(true)} }

func TestGetProductReadsPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(productPage))
	}))
	defer srv.Close()
	// httptest listens on a random port; bypass the port rule via the fetcher directly.
	u, _ := url.Parse(srv.URL + "/p?utm_source=x")
	pg, err := testProvider().f.get(context.Background(), u)
	if err != nil {
		t.Fatal(err)
	}
	m := parseMeta(pg.body, pg.finalURL)
	pd := build(pg.finalURL, m)
	if pd.Title != "Air Fryer 5L Preta" || len(pd.Offers) != 1 || pd.Offers[0].PriceCents != nil ||
		pd.Offers[0].Availability != affiliate.Unknown || *pd.Offers[0].StoreName != "Loja Teste" {
		t.Fatalf("unexpected product: %+v", pd)
	}
}

func TestFetchRejections(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/forbidden", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusForbidden) })
	mux.HandleFunc("/json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	})
	mux.HandleFunc("/loop", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/loop", http.StatusFound) })
	mux.HandleFunc("/internal", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://169.254.169.254:80/latest", http.StatusFound)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	f := newFetcher(true)
	for _, path := range []string{"/forbidden", "/json", "/loop", "/internal"} {
		u, _ := url.Parse(srv.URL + path)
		if _, err := f.get(context.Background(), u); err == nil {
			t.Errorf("%s should fail", path)
		}
	}
}

func TestGuardedDialerBlocksLoopback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(productPage))
	}))
	defer srv.Close()
	u, _ := url.Parse(srv.URL)
	_, err := newFetcher(false).get(context.Background(), u)
	if err == nil || !errors.Is(err, errBlockedIP) {
		t.Fatalf("loopback must be blocked at dial time, got %v", err)
	}
}

func TestGetProductRejectsInternalTargets(t *testing.T) {
	p := New()
	for _, raw := range []string{"http://localhost/x", "https://127.0.0.1/x", "http://10.0.0.5/", "https://[::1]/", "http://169.254.169.254/latest/meta-data", "ftp://x.com", ""} {
		if _, err := p.GetProduct(context.Background(), raw); !errors.Is(err, affiliate.ErrInvalidURL) {
			t.Errorf("GetProduct(%q) = %v, want ErrInvalidURL", raw, err)
		}
	}
}

func TestBuildAffiliateURLKeepsPastedURL(t *testing.T) {
	got, err := New().BuildAffiliateURL(context.Background(), MerchantCode, "https://loja.com/p?id=1", affiliate.ClickContext{ClickID: "c"})
	if err != nil || got != "https://loja.com/p?id=1" {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestIsGenericTitle(t *testing.T) {
	u, _ := url.Parse("https://www.amazon.com.br/dp/B0X")
	for title, want := range map[string]bool{
		"Amazon.com.br": true, "Amazon": true, "AMAZON": true, "Loja X": true,
		"Fritadeira Air Fryer 5L": false, "": false,
	} {
		site := ""
		if title == "Loja X" {
			site = "Loja X"
		}
		if got := isGenericTitle(title, u, site); got != want {
			t.Errorf("isGenericTitle(%q) = %v, want %v", title, got, want)
		}
	}
}
