package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/listou/listou/apps/api/internal/affiliate"
	"github.com/listou/listou/apps/api/internal/affiliate/mock"
	"github.com/listou/listou/apps/api/internal/analytics"
	"github.com/listou/listou/apps/api/internal/auth"
	"github.com/listou/listou/apps/api/internal/catalog"
	"github.com/listou/listou/apps/api/internal/decision"
	"github.com/listou/listou/apps/api/internal/events"
	"github.com/listou/listou/apps/api/internal/lists"
	"github.com/listou/listou/apps/api/internal/platform/database"
	"github.com/listou/listou/apps/api/internal/recommendations"
	"github.com/listou/listou/apps/api/internal/reservations"
)

const (
	seedEmail    = "tiago@listou.dev"
	seedPassword = "listou123"
)

// seed loads development data through the real services (so it exercises the
// same validation and invariants as production). It is idempotent.
func seed(ctx context.Context, url string) error {
	pool, err := database.Connect(ctx, url)
	if err != nil {
		return err
	}
	defer pool.Close()

	var exists bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)`, seedEmail).Scan(&exists); err != nil {
		return err
	}
	if exists {
		log.Println("seed: already applied, skipping")
		return nil
	}

	rec := analytics.NewRecorder(pool)
	registry := affiliate.NewRegistry(mock.New("http://localhost:3000"))
	catSvc := catalog.NewService(pool, registry, rec)
	listSvc := lists.NewService(pool, catSvc, decision.RuleEngine{}, rec)
	evSvc := events.NewService(pool, listSvc, recommendations.RulesProvider{}, rec)
	authSvc := auth.NewService(pool, 24*time.Hour, rec)
	resSvc := reservations.NewService(pool, 7*24*time.Hour, rec)

	user, _, err := authSvc.Register(ctx, auth.RegisterInput{Name: "Tiago Silva", Email: seedEmail, Password: seedPassword})
	if err != nil {
		return fmt.Errorf("seed user: %w", err)
	}

	if _, err := pool.Exec(ctx, `UPDATE users SET role = 'ADMIN' WHERE id = $1`, user.ID); err != nil {
		return fmt.Errorf("seed admin: %w", err)
	}

	type spec struct {
		in    events.CreateInput
		title string
		extra []lists.ItemInput
	}
	wedding := "João & Maria"
	hostNames := wedding
	loc := "São Paulo, SP"
	date := time.Now().AddDate(0, 3, 12).Format(time.DateOnly)
	desc := "Estamos começando uma nova fase e vamos adorar ter vocês com a gente. Escolhemos alguns itens para facilitar — e qualquer carinho será muito bem-vindo! 💜"
	slugWedding, slugHouse := "joao-e-maria", "casa-nova-joao-maria"
	houseDesc := "Nosso primeiro apê juntos, 70 m² de recomeço."
	houseDate := time.Now().AddDate(0, 1, 5).Format(time.DateOnly)
	houseTitle := "Chá de casa nova do João e Maria"
	houseHosts := "João e Maria"

	specs := []spec{
		{in: events.CreateInput{Type: "WEDDING", Title: "Casamento João & Maria", HostNames: &hostNames, Description: &desc, EventDate: &date, Location: &loc, Slug: &slugWedding, Template: "SUGGESTED"},
			extra: []lists.ItemInput{{Title: ptr("Dinheiro para a lua de mel"), Emoji: ptr("✈️"), Description: ptr("Contribuição livre para a nossa viagem."), DesiredQuantity: ptrInt(1), PriceReferenceCents: ptr64(50000)}}},
		{in: events.CreateInput{Type: "HOUSEWARMING", Title: houseTitle, HostNames: &houseHosts, Description: &houseDesc, EventDate: &houseDate, Slug: &slugHouse, Template: "SUGGESTED"}},
	}
	for _, sp := range specs {
		ev, err := evSvc.Create(ctx, user.ID, sp.in)
		if err != nil {
			return fmt.Errorf("seed event %q: %w", sp.in.Title, err)
		}
		view, err := listSvc.ForEvent(ctx, user.ID, ev.ID)
		if err != nil {
			return err
		}
		attached := 0
		for _, it := range view.Items {
			ok, err := attachBestProduct(ctx, catSvc, listSvc, user.ID, it)
			if err != nil {
				return err
			}
			if ok {
				attached++
			}
		}
		for _, extra := range sp.extra {
			if _, err := listSvc.CreateItem(ctx, user.ID, ev.ListID, extra); err != nil {
				return fmt.Errorf("seed extra item: %w", err)
			}
		}
		published := "PUBLISHED"
		visibility := "PUBLIC"
		if _, err := evSvc.Update(ctx, user.ID, ev.ID, events.UpdateInput{Status: &published, Visibility: &visibility}); err != nil {
			return fmt.Errorf("seed publish: %w", err)
		}
		log.Printf("seed: event %s (%d items, %d with products)", ev.Slug, len(view.Items), attached)
	}

	// A couple of guest interactions so the public page shows real states.
	view, err := listSvc.ForEvent(ctx, user.ID, mustEventID(ctx, pool, slugWedding))
	if err != nil {
		return err
	}
	var withProduct []lists.Item
	for _, it := range view.Items {
		if it.Product != nil {
			withProduct = append(withProduct, it)
		}
	}
	if len(withProduct) > 7 {
		// Pick items further down so the first screen of the public list looks inviting.
		if _, err := resSvc.Create(ctx, slugWedding, withProduct[3].ID, nil, reservations.CreateInput{Kind: "RESERVATION", Quantity: 1, GuestName: "Ana (convidada de demonstração)"}); err != nil {
			return fmt.Errorf("seed reservation: %w", err)
		}
		if _, err := resSvc.Create(ctx, slugWedding, withProduct[6].ID, nil, reservations.CreateInput{Kind: "PURCHASE", Quantity: 1, GuestName: "Bruno (convidado de demonstração)"}); err != nil {
			return fmt.Errorf("seed purchase: %w", err)
		}
	}
	log.Printf("seed: done. login: %s / %s", seedEmail, seedPassword)
	return nil
}

func mustEventID(ctx context.Context, pool database.DBTX, slug string) uuid.UUID {
	var id uuid.UUID
	if err := pool.QueryRow(ctx, `SELECT id FROM events WHERE slug = $1`, slug).Scan(&id); err != nil {
		log.Fatalf("seed: event %s: %v", slug, err)
	}
	return id
}

// attachBestProduct searches the catalog for a desired item and links the
// best-ranked product when it is a MATCH or PARTIAL_MATCH.
func attachBestProduct(ctx context.Context, cat *catalog.Service, ls *lists.Service, userID uuid.UUID, it lists.Item) (bool, error) {
	res, err := cat.Search(ctx, catalog.SearchParams{Query: it.Title, Limit: 5})
	if err != nil || len(res) == 0 || res[0].Match == nil || res[0].Match.Verdict == catalog.NoMatch {
		return false, nil
	}
	p, _, err := cat.Import(ctx, res[0].ProviderCode, res[0].ExternalID)
	if err != nil {
		return false, fmt.Errorf("import %s: %w", res[0].ExternalID, err)
	}
	pid := p.ID.String()
	_, err = ls.UpdateItem(ctx, userID, it.ID, lists.ItemInput{ProductID: &pid})
	return err == nil, err
}

func ptr(s string) *string { return &s }
func ptrInt(i int) *int    { return &i }
func ptr64(i int64) *int64 { return &i }
