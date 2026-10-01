package admin

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/listou/listou/apps/api/internal/affiliate/curated"
	"github.com/listou/listou/apps/api/internal/auth"
	"github.com/listou/listou/apps/api/internal/platform/httpx"
)

type fakeCatalog struct{ created int }

func (f *fakeCatalog) List(context.Context, bool) ([]curated.Product, error) { return nil, nil }
func (f *fakeCatalog) Get(_ context.Context, id uuid.UUID) (curated.Product, error) {
	return curated.Product{ID: id}, nil
}
func (f *fakeCatalog) Merchants(context.Context) ([]curated.Merchant, error) {
	return []curated.Merchant{{Code: "MERCADO_LIVRE", Name: "Mercado Livre"}}, nil
}
func (f *fakeCatalog) Create(context.Context, curated.Input) (uuid.UUID, error) {
	f.created++
	return uuid.New(), nil
}
func (f *fakeCatalog) Update(context.Context, uuid.UUID, curated.Input) (bool, error) {
	return false, nil
}
func (f *fakeCatalog) Delete(context.Context, uuid.UUID) (bool, error) { return false, nil }

func TestCreateValidatesBeforeStoring(t *testing.T) {
	fc := &fakeCatalog{}
	svc := &Service{catalog: fc} // nil db: audit is best-effort and never reached on failure
	_, err := svc.Create(context.Background(), auth.User{}, curated.Input{Title: "ab"})
	if httpx.AsError(err).Code != "VALIDATION_FAILED" || fc.created != 0 {
		t.Fatalf("invalid input must be rejected before storage: %v created=%d", err, fc.created)
	}
}

func TestUpdateAndDeleteMissingAreNotFound(t *testing.T) {
	svc := &Service{catalog: &fakeCatalog{}}
	in := curated.Input{Title: "Produto bom", Offers: []curated.Offer{{Merchant: "MERCADO_LIVRE", URL: "https://meli.la/x"}}}
	if _, err := svc.Update(context.Background(), auth.User{}, uuid.New(), in); httpx.AsError(err).Code != "CURATED_PRODUCT_NOT_FOUND" {
		t.Fatalf("update: %v", err)
	}
	if err := svc.Delete(context.Background(), auth.User{}, uuid.New()); httpx.AsError(err).Code != "CURATED_PRODUCT_NOT_FOUND" {
		t.Fatalf("delete: %v", err)
	}
}
