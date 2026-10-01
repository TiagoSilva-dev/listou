package decision

import (
	"context"
	"testing"
)

func TestRuleEngineClassify(t *testing.T) {
	labels := []string{"Cozinha", "Quarto", "Banheiro", "Limpeza e organização"}
	cases := map[string]string{
		"Air Fryer Mondial":        "Cozinha",
		"Jogo de cama queen":       "Quarto",
		"Kit toalhas":              "Banheiro",
		"Aspirador de pó":          "Limpeza e organização",
		"Dinheiro para lua de mel": "",
	}
	for text, want := range cases {
		d, err := RuleEngine{}.Classify(context.Background(), text, labels)
		if err != nil || d.Label != want {
			t.Errorf("%q: got %q want %q", text, d.Label, want)
		}
	}
}
