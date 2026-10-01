package events

import "testing"

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Tiago & Júlia":            "tiago-e-julia",
		"  Chá de Bebê da Ana!!  ": "cha-de-bebe-da-ana",
		"Casamento — Ção / 2026":   "casamento-cao-2026",
	}
	for in, want := range cases {
		if got := Slugify(in); got != want {
			t.Errorf("Slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestValidSlug(t *testing.T) {
	for s, want := range map[string]bool{
		"tiago-e-julia": true, "ab": false, "-abc": false, "abc-": false, "a--b": false,
		"Tiago": false, "dashboard": false, "lista-2026": true,
	} {
		if got := ValidSlug(s); got != want {
			t.Errorf("ValidSlug(%q) = %v", s, got)
		}
	}
}

func TestCreateValidate(t *testing.T) {
	blank := "  "
	in := CreateInput{Type: "WEDDING", Title: " Nosso casamento ", HostNames: &blank}
	if _, err := in.Validate(); err != nil {
		t.Fatal(err)
	}
	if in.HostNames != nil || in.Title != "Nosso casamento" {
		t.Fatalf("not normalized: %+v", in)
	}
	bad := CreateInput{Type: "PARTY", Title: ""}
	if _, err := bad.Validate(); err == nil {
		t.Fatal("expected error")
	}
	date := "2026-13-40"
	bad2 := CreateInput{Type: "WEDDING", Title: "x", EventDate: &date}
	if _, err := bad2.Validate(); err == nil {
		t.Fatal("expected date error")
	}
}
