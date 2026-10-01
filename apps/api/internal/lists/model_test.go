package lists

import "testing"

func TestDeriveStatus(t *testing.T) {
	cases := []struct {
		desired, purchased, reserved int
		archived                     bool
		want                         ItemStatus
	}{
		{6, 0, 0, false, Available},
		{6, 2, 1, false, PartiallyReserved},
		{6, 2, 4, false, Reserved},
		{6, 6, 0, false, Purchased},
		{1, 0, 1, false, Reserved},
		{1, 0, 0, true, Archived},
	}
	for _, c := range cases {
		if got := DeriveStatus(c.desired, c.purchased, c.reserved, c.archived); got != c.want {
			t.Errorf("%+v: got %s", c, got)
		}
	}
	it := Item{DesiredQuantity: 6, PurchasedQuantity: 2, ReservedQuantity: 1}
	it.finalize()
	if it.AvailableQuantity != 3 || it.Status != PartiallyReserved || it.Offers == nil {
		t.Fatalf("finalize: %+v", it)
	}
}

func TestItemInputValidate(t *testing.T) {
	blank := "   "
	in := ItemInput{Title: &blank}
	if err := in.Validate(true); err == nil {
		t.Fatal("expected title error")
	}
	title, qty := "Toalhas", 0
	in = ItemInput{Title: &title, DesiredQuantity: &qty}
	if err := in.Validate(true); err == nil {
		t.Fatal("expected quantity error")
	}
	qty = 6
	if err := in.Validate(true); err != nil {
		t.Fatal(err)
	}
}
