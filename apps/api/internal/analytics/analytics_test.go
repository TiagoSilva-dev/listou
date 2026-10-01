package analytics

import "testing"

func TestSanitizeProps(t *testing.T) {
	if _, ok := sanitizeProps(map[string]any{"channel": "whatsapp", "n": 1.0, "ok": true}); !ok {
		t.Fatal("expected scalar props to pass")
	}
	if _, ok := sanitizeProps(map[string]any{"nested": map[string]any{"a": 1}}); ok {
		t.Fatal("expected nested props to be rejected")
	}
	if !clientEvents[ShareCreated] || clientEvents[ItemReserved] {
		t.Fatal("client event whitelist is wrong")
	}
}
