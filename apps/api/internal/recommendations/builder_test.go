package recommendations

import (
	"context"
	"testing"
)

func TestBuilderSkipsExistingAndKeepsStructure(t *testing.T) {
	b := NewBuilder(RulesProvider{})
	all, err := b.Suggest(context.Background(), BuildRequest{EventType: "HOUSEWARMING"})
	if err != nil || len(all) == 0 {
		t.Fatalf("expected suggestions, got %v %v", all, err)
	}
	first := all[0].Desires[0].Title

	got, err := b.Suggest(context.Background(), BuildRequest{EventType: "HOUSEWARMING", Existing: []string{first + " 5L"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range got {
		for _, d := range c.Desires {
			if d.Title == first {
				t.Fatalf("%q should have been filtered out", first)
			}
		}
	}
}

func TestBuilderPromptBoostsMatchingCategory(t *testing.T) {
	b := NewBuilder(RulesProvider{})
	got, err := b.Suggest(context.Background(), BuildRequest{EventType: "HOUSEWARMING", Prompt: "banheiro toalhas"})
	if err != nil || len(got) == 0 {
		t.Fatalf("expected suggestions, got %v %v", got, err)
	}
	if got[0].Name != "Banheiro" {
		t.Fatalf("expected bathroom category first, got %q", got[0].Name)
	}
}

func TestBuilderLimit(t *testing.T) {
	b := NewBuilder(RulesProvider{})
	got, _ := b.Suggest(context.Background(), BuildRequest{EventType: "HOUSEWARMING"})
	n := 0
	for _, c := range got {
		n += len(c.Desires)
	}
	if n > 24 {
		t.Fatalf("limit exceeded: %d", n)
	}
}
