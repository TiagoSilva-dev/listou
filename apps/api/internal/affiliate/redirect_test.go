package affiliate

import "testing"

func TestRedirectHelpers(t *testing.T) {
	if !safeTarget("https://example.com/p") || safeTarget("javascript:alert(1)") || safeTarget("//evil.com") || safeTarget("ftp://x.com") {
		t.Fatal("safeTarget wrong")
	}
	if h := referrerHost("https://wa.me/path?secret=1"); h == nil || *h != "wa.me" {
		t.Fatalf("referrerHost: %v", h)
	}
	if referrerHost("") != nil || referrerHost("::bad") != nil {
		t.Fatal("expected nil")
	}
	cases := map[string]string{
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17) Mobile": "MOBILE",
		"Mozilla/5.0 (iPad)":                            "TABLET",
		"Mozilla/5.0 (X11; Linux x86_64)":               "DESKTOP",
		"":                                              "UNKNOWN",
	}
	for ua, want := range cases {
		if got := deviceClass(ua); got != want {
			t.Errorf("%q: %s", ua, got)
		}
	}
}
