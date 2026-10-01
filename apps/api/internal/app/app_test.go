package app_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/listou/listou/apps/api/internal/app"
	"github.com/listou/listou/apps/api/internal/platform/config"
	"github.com/listou/listou/apps/api/internal/platform/database"
	"github.com/listou/listou/apps/api/migrations"
)

// These tests run the full HTTP stack against a real PostgreSQL database.
// Set TEST_DATABASE_URL (a throwaway database!) to enable them.

type client struct {
	t       *testing.T
	h       http.Handler
	cookies []*http.Cookie
}

type resp struct {
	code int
	body map[string]any
	hdr  http.Header
}

func (c *client) do(method, path string, body any, headers ...string) resp {
	c.t.Helper()
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, rd)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	for _, ck := range c.cookies {
		req.AddCookie(ck)
	}
	rec := httptest.NewRecorder()
	c.h.ServeHTTP(rec, req)
	for _, ck := range rec.Result().Cookies() {
		replaced := false
		for i, old := range c.cookies {
			if old.Name == ck.Name {
				c.cookies[i], replaced = ck, true
			}
		}
		if !replaced {
			c.cookies = append(c.cookies, ck)
		}
	}
	out := resp{code: rec.Code, hdr: rec.Header(), body: map[string]any{}}
	_ = json.Unmarshal(rec.Body.Bytes(), &out.body)
	return out
}

func (r resp) path(keys ...string) any {
	var cur any = r.body
	for _, k := range keys {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = m[k]
	}
	return cur
}

func (r resp) str(keys ...string) string { s, _ := r.path(keys...).(string); return s }
func (r resp) errCode() string           { return r.str("error", "code") }

func setup(t *testing.T) http.Handler { return setupWithFlags(t, "") }

func setupWithFlags(t *testing.T, flagSpec string) http.Handler {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	db, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	goose.SetBaseFS(migrations.FS)
	_ = goose.SetDialect("postgres")
	if _, err := db.Exec(`DROP SCHEMA public CASCADE; CREATE SCHEMA public`); err != nil {
		t.Fatal(err)
	}
	if err := goose.Up(db, "."); err != nil {
		t.Fatal(err)
	}
	pool, err := database.Connect(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	cfg := config.Config{Env: "test", PublicWebURL: "http://localhost:3000", SessionTTL: 3600e9, ReservationTTL: 3600e9, FeatureFlags: flagSpec}
	return app.New(cfg, pool, slog.New(slog.NewTextHandler(io.Discard, nil)), "test").Handler()
}

func register(t *testing.T, h http.Handler, email string) *client {
	t.Helper()
	c := &client{t: t, h: h}
	r := c.do("POST", "/api/v1/auth/register", map[string]any{"name": "Pessoa Teste", "email": email, "password": "senha-segura-1"})
	if r.code != http.StatusCreated {
		t.Fatalf("register: %d %v", r.code, r.body)
	}
	return c
}

func TestCreatorToGuestFlow(t *testing.T) {
	h := setup(t)
	owner := register(t, h, "owner@example.com")

	if r := owner.do("GET", "/api/v1/me", nil); r.code != 200 || r.str("user", "email") != "owner@example.com" {
		t.Fatalf("me: %v", r.body)
	}
	if r := (&client{t: t, h: h}).do("GET", "/api/v1/me", nil); r.code != 401 || r.errCode() != "UNAUTHORIZED" {
		t.Fatalf("anon me: %d %v", r.code, r.body)
	}
	if r := (&client{t: t, h: h}).do("POST", "/api/v1/auth/register", map[string]any{"name": "X", "email": "owner@example.com", "password": "senha-segura-1"}); r.code != 409 && r.code != 422 {
		t.Fatalf("duplicate email: %d", r.code)
	}

	// Event + list from template
	r := owner.do("POST", "/api/v1/events", map[string]any{"type": "HOUSEWARMING", "title": "Casa nova", "hostNames": "Tiago & Júlia", "template": "SUGGESTED"})
	if r.code != 201 {
		t.Fatalf("create event: %d %v", r.code, r.body)
	}
	eventID, slug, listID := r.str("event", "id"), r.str("event", "slug"), r.str("event", "listId")
	if slug != "tiago-e-julia" {
		t.Fatalf("slug = %q", slug)
	}

	// Draft is not public
	anon := &client{t: t, h: h}
	if r := anon.do("GET", "/api/v1/public/lists/"+slug, nil); r.code != 404 || r.errCode() != "LIST_NOT_FOUND" {
		t.Fatalf("draft must be hidden: %d %v", r.code, r.body)
	}

	// Manual item + product association
	r = owner.do("POST", "/api/v1/lists/"+listID+"/items", map[string]any{"title": "Dinheiro para a lua de mel", "desiredQuantity": 1})
	if r.code != 201 || r.path("item", "product") != nil {
		t.Fatalf("manual item: %d %v", r.code, r.body)
	}
	s := owner.do("GET", "/api/v1/products/search?q=air+fryer", nil)
	results, _ := s.path("results").([]any)
	if s.code != 200 || len(results) == 0 {
		t.Fatalf("search: %d %v", s.code, s.body)
	}
	first := results[0].(map[string]any)
	imp := owner.do("POST", "/api/v1/products/import", map[string]any{"providerCode": first["providerCode"], "externalId": first["externalId"]})
	if imp.code != 200 {
		t.Fatalf("import: %d %v", imp.code, imp.body)
	}
	productID := imp.str("product", "id")
	offers, _ := imp.path("offers").([]any)
	if len(offers) < 2 {
		t.Fatalf("expected multi-marketplace offers, got %d", len(offers))
	}
	if strings.Contains(string(mustJSON(offers)), "product_url") || strings.Contains(string(mustJSON(offers)), "/demo/loja") {
		t.Fatal("offers must never expose merchant URLs")
	}
	r = owner.do("POST", "/api/v1/lists/"+listID+"/items", map[string]any{"productId": productID, "desiredQuantity": 2})
	if r.code != 201 {
		t.Fatalf("item with product: %d %v", r.code, r.body)
	}
	itemID := r.str("item", "id")
	if r.str("item", "title") == "" || r.str("item", "status") != "AVAILABLE" {
		t.Fatalf("item defaults: %v", r.body)
	}

	// Publish and read publicly
	if r := owner.do("PATCH", "/api/v1/events/"+eventID, map[string]any{"status": "PUBLISHED", "visibility": "PUBLIC"}); r.code != 200 {
		t.Fatalf("publish: %d %v", r.code, r.body)
	}
	pub := anon.do("GET", "/api/v1/public/lists/"+slug, nil)
	if pub.code != 200 || strings.Contains(string(mustJSON(pub.body)), "\"notes\"") {
		t.Fatalf("public: %d %v", pub.code, pub.body)
	}

	// Guest reserves 1 of 2, then the rest; third is rejected
	res := anon.do("POST", "/api/v1/public/lists/"+slug+"/items/"+itemID+"/reservations", map[string]any{"guestName": "Ana", "quantity": 1})
	if res.code != 201 || res.str("reservation", "manageToken") == "" {
		t.Fatalf("reserve: %d %v", res.code, res.body)
	}
	token, resID := res.str("reservation", "manageToken"), res.str("reservation", "id")
	pub = anon.do("GET", "/api/v1/public/lists/"+slug, nil)
	if !strings.Contains(string(mustJSON(pub.body)), "PARTIALLY_RESERVED") || strings.Contains(string(mustJSON(pub.body)), "Ana") {
		t.Fatalf("public must show partial state without guest names: %v", pub.body)
	}
	if r := anon.do("POST", "/api/v1/public/lists/"+slug+"/items/"+itemID+"/reservations", map[string]any{"guestName": "Bia", "quantity": 2}); r.code != 409 || r.errCode() != "ITEM_NOT_AVAILABLE" {
		t.Fatalf("over-reserve: %d %v", r.code, r.body)
	}
	// Cancel needs the token
	if r := anon.do("DELETE", "/api/v1/reservations/"+resID, nil); r.code != 404 {
		t.Fatalf("cancel w/o token: %d", r.code)
	}
	if r := anon.do("DELETE", "/api/v1/reservations/"+resID, nil, "X-Reservation-Token", token); r.code != 204 {
		t.Fatalf("cancel: %d %v", r.code, r.body)
	}
	if r := anon.do("POST", "/api/v1/public/lists/"+slug+"/items/"+itemID+"/reservations", map[string]any{"guestName": "Bia", "quantity": 2}); r.code != 201 {
		t.Fatalf("reserve after cancel: %d %v", r.code, r.body)
	}

	// Outbound redirect records a click and never trusts query ids blindly
	offerID := offers[0].(map[string]any)["id"].(string)
	g := anon.do("GET", "/go/"+offerID+"?item="+itemID+"&utm_source=whatsapp", nil, "Referer", "https://wa.me/abc?x=1")
	loc := g.hdr.Get("Location")
	if g.code != 302 || !strings.Contains(loc, "/demo/loja/") || !strings.Contains(loc, "demo_tag=listou-") {
		t.Fatalf("go: %d loc=%q", g.code, loc)
	}
	if r := anon.do("GET", "/go/00000000-0000-7000-8000-000000000000", nil); r.code != 404 {
		t.Fatalf("go unknown offer: %d", r.code)
	}

	// Client analytics whitelist
	if r := anon.do("POST", "/api/v1/analytics/track", map[string]any{"name": "ITEM_RESERVED", "slug": slug}); r.code != 400 {
		t.Fatalf("clients must not forge business events: %d", r.code)
	}
	if r := anon.do("POST", "/api/v1/analytics/track", map[string]any{"name": "LIST_VIEWED", "slug": slug}); r.code != 202 {
		t.Fatalf("track: %d", r.code)
	}

	// Dashboard
	d := owner.do("GET", "/api/v1/events/"+eventID+"/dashboard", nil)
	if d.code != 200 || d.path("stats", "reservedUnits").(float64) != 2 || d.path("stats", "outboundClicks").(float64) != 1 || d.path("stats", "views").(float64) != 1 {
		t.Fatalf("dashboard: %d %v", d.code, d.body)
	}
}

func TestAuthorizationIsolation(t *testing.T) {
	h := setup(t)
	a, b := register(t, h, "a@example.com"), register(t, h, "b@example.com")
	r := a.do("POST", "/api/v1/events", map[string]any{"type": "WEDDING", "title": "Casamento A"})
	eventID, listID := r.str("event", "id"), r.str("event", "listId")
	it := a.do("POST", "/api/v1/lists/"+listID+"/items", map[string]any{"title": "Item A"})
	itemID := it.str("item", "id")

	for name, resp := range map[string]resp{
		"get event":    b.do("GET", "/api/v1/events/"+eventID, nil),
		"patch event":  b.do("PATCH", "/api/v1/events/"+eventID, map[string]any{"title": "hack"}),
		"delete event": b.do("DELETE", "/api/v1/events/"+eventID, nil),
		"list items":   b.do("GET", "/api/v1/lists/"+listID+"/items", nil),
		"create item":  b.do("POST", "/api/v1/lists/"+listID+"/items", map[string]any{"title": "x"}),
		"patch item":   b.do("PATCH", "/api/v1/items/"+itemID, map[string]any{"title": "hack"}),
		"delete item":  b.do("DELETE", "/api/v1/items/"+itemID, nil),
		"dashboard":    b.do("GET", "/api/v1/events/"+eventID+"/dashboard", nil),
	} {
		if resp.code != 404 {
			t.Errorf("%s: expected 404 (no existence leak), got %d %v", name, resp.code, resp.body)
		}
	}
	if r := b.do("GET", "/api/v1/events", nil); len(r.path("events").([]any)) != 0 {
		t.Fatal("B must not see A's events")
	}
	// A foreign category cannot be attached to an item
	bEv := b.do("POST", "/api/v1/events", map[string]any{"type": "WEDDING", "title": "Casamento B", "template": "SUGGESTED"})
	bList := bEv.str("event", "listId")
	cats := b.do("GET", "/api/v1/events/"+bEv.str("event", "id")+"/list", nil).path("categories").([]any)
	foreign := cats[0].(map[string]any)["id"].(string)
	if r := a.do("PATCH", "/api/v1/items/"+itemID, map[string]any{"categoryId": foreign}); r.code != 404 {
		t.Fatalf("foreign category accepted: %d", r.code)
	}
	_ = bList
	// Wrong login looks the same for unknown email and wrong password
	anon := &client{t: t, h: h}
	r1 := anon.do("POST", "/api/v1/auth/login", map[string]any{"email": "a@example.com", "password": "errada-errada"})
	r2 := anon.do("POST", "/api/v1/auth/login", map[string]any{"email": "nobody@example.com", "password": "errada-errada"})
	if r1.code != 401 || r2.code != 401 || r1.str("error", "message") != r2.str("error", "message") {
		t.Fatalf("login responses leak: %v %v", r1.body, r2.body)
	}
	// Logout revokes the session
	if r := a.do("POST", "/api/v1/auth/logout", nil); r.code != 204 {
		t.Fatalf("logout: %d", r.code)
	}
}

func TestConcurrentReservationsNeverOversell(t *testing.T) {
	h := setup(t)
	owner := register(t, h, "owner@example.com")
	r := owner.do("POST", "/api/v1/events", map[string]any{"type": "WEDDING", "title": "Concorrência"})
	slug, listID, eventID := r.str("event", "slug"), r.str("event", "listId"), r.str("event", "id")
	item := owner.do("POST", "/api/v1/lists/"+listID+"/items", map[string]any{"title": "Último item", "desiredQuantity": 3}).str("item", "id")
	owner.do("PATCH", "/api/v1/events/"+eventID, map[string]any{"status": "PUBLISHED", "visibility": "PUBLIC"})

	var ok, conflict int32
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			g := &client{t: t, h: h}
			switch g.do("POST", "/api/v1/public/lists/"+slug+"/items/"+item+"/reservations", map[string]any{"guestName": "Convidado", "quantity": 1}).code {
			case 201:
				atomic.AddInt32(&ok, 1)
			case 409:
				atomic.AddInt32(&conflict, 1)
			}
		}()
	}
	wg.Wait()
	if ok != 3 || conflict != 17 {
		t.Fatalf("oversold: %d succeeded, %d conflicts (want 3/17)", ok, conflict)
	}
	d := owner.do("GET", "/api/v1/events/"+eventID+"/dashboard", nil)
	if d.path("stats", "reservedUnits").(float64) != 3 {
		t.Fatalf("dashboard: %v", d.body)
	}
}

func TestSurpriseModeHidesWhoBoughtWhat(t *testing.T) {
	h := setup(t)
	owner := register(t, h, "owner@example.com")
	r := owner.do("POST", "/api/v1/events", map[string]any{"type": "BIRTHDAY", "title": "Surpresa"})
	slug, listID, eventID := r.str("event", "slug"), r.str("event", "listId"), r.str("event", "id")
	item := owner.do("POST", "/api/v1/lists/"+listID+"/items", map[string]any{"title": "Relógio"}).str("item", "id")
	owner.do("PATCH", "/api/v1/events/"+eventID, map[string]any{"status": "PUBLISHED", "visibility": "PUBLIC", "surpriseMode": true})
	(&client{t: t, h: h}).do("POST", "/api/v1/public/lists/"+slug+"/items/"+item+"/reservations", map[string]any{"guestName": "Carla", "kind": "PURCHASE"})
	d := owner.do("GET", "/api/v1/events/"+eventID+"/dashboard", nil)
	raw := string(mustJSON(d.body))
	if d.path("stats", "purchasedUnits").(float64) != 1 || strings.Contains(raw, "Carla") || strings.Contains(raw, "Relógio") {
		t.Fatalf("surprise mode leaked: %s", raw)
	}
	// Per-item state is hidden from the owner too, but guests still see availability.
	ownerList := owner.do("GET", "/api/v1/events/"+eventID+"/list", nil)
	items := ownerList.path("items").([]any)
	if it := items[0].(map[string]any); it["purchasedQuantity"].(float64) != 0 || it["status"] != "AVAILABLE" {
		t.Fatalf("owner list leaked per-item state: %v", it)
	}
	pub := (&client{t: t, h: h}).do("GET", "/api/v1/public/lists/"+slug, nil)
	if it := pub.path("items").([]any)[0].(map[string]any); it["status"] != "PURCHASED" {
		t.Fatalf("guests must still see availability: %v", it)
	}
}

func TestAIListBuilder(t *testing.T) {
	// Flag off: the feature does not exist.
	off := register(t, setup(t), "off@example.com")
	ev := off.do("POST", "/api/v1/events", map[string]any{"type": "HOUSEWARMING", "title": "Casa nova"})
	if r := off.do("POST", "/api/v1/events/"+ev.str("event", "id")+"/suggestions", map[string]any{}); r.code != 404 || r.errCode() != "FEATURE_DISABLED" {
		t.Fatalf("flag off must hide the feature, got %d %v", r.code, r.body)
	}

	h := setupWithFlags(t, "AI_LIST_BUILDER=true")
	owner, other := register(t, h, "owner@example.com"), register(t, h, "other@example.com")
	r := owner.do("POST", "/api/v1/events", map[string]any{"type": "HOUSEWARMING", "title": "Casa nova"})
	eventID, listID := r.str("event", "id"), r.str("event", "listId")
	owner.do("POST", "/api/v1/lists/"+listID+"/items", map[string]any{"title": "Air fryer 5L"})

	if r := other.do("POST", "/api/v1/events/"+eventID+"/suggestions", map[string]any{}); r.code != 404 {
		t.Fatalf("strangers must get 404, got %d", r.code)
	}
	sug := owner.do("POST", "/api/v1/events/"+eventID+"/suggestions", map[string]any{"prompt": "banheiro"})
	if sug.code != 200 {
		t.Fatalf("suggest: %d %v", sug.code, sug.body)
	}
	raw := string(mustJSON(sug.body))
	if strings.Contains(raw, "Air fryer") || strings.Contains(raw, "http") || strings.Contains(raw, "price") {
		t.Fatalf("suggestions must skip existing items and carry no products/links: %s", raw)
	}
	cats := sug.path("categories").([]any)
	if cats[0].(map[string]any)["name"] != "Banheiro" {
		t.Fatalf("prompt should rank the bathroom first: %v", cats[0])
	}

	apply := map[string]any{"categories": []any{map[string]any{"name": "Banheiro", "emoji": "🛁", "desires": []any{
		map[string]any{"title": "Jogo de toalhas", "emoji": "🧺", "quantity": 2, "priority": "HIGH"},
	}}}}
	a := owner.do("POST", "/api/v1/events/"+eventID+"/suggestions/apply", apply)
	if a.code != 201 || a.path("addedItems").(float64) != 1 {
		t.Fatalf("apply: %d %v", a.code, a.body)
	}
	// A second apply reuses the category instead of duplicating it.
	owner.do("POST", "/api/v1/events/"+eventID+"/suggestions/apply", apply)
	view := owner.do("GET", "/api/v1/events/"+eventID+"/list", nil)
	banheiros := 0
	for _, c := range view.path("categories").([]any) {
		if c.(map[string]any)["name"] == "Banheiro" {
			banheiros++
		}
	}
	if banheiros != 1 || len(view.path("items").([]any)) != 3 {
		t.Fatalf("expected 1 Banheiro category and 3 items: %v", view.body)
	}
	if r := other.do("POST", "/api/v1/events/"+eventID+"/suggestions/apply", apply); r.code != 404 {
		t.Fatalf("strangers cannot apply, got %d", r.code)
	}
	if r := owner.do("POST", "/api/v1/events/"+eventID+"/suggestions/apply", map[string]any{"categories": []any{}}); r.code != 422 {
		t.Fatalf("empty apply must be a validation error, got %d", r.code)
	}
}

func mustJSON(v any) []byte { b, _ := json.Marshal(v); return b }
