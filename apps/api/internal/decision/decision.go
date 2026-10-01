// Package decision abstracts high-frequency structured decisions
// (classification, choice, scoring). The app depends only on Engine; a
// Jev-backed engine can be added later without touching callers. Use a
// deterministic rule whenever it solves the problem correctly.
package decision

import (
	"context"
	"strings"

	"github.com/listou/listou/apps/api/internal/catalog/textnorm"
)

type Decision struct {
	Label      string  `json:"label"`
	Confidence float64 `json:"confidence"`
	Source     string  `json:"source"`
}

type Engine interface {
	// Classify picks the best label for text among labels, or "" when unsure.
	Classify(ctx context.Context, text string, labels []string) (Decision, error)
}

// RuleEngine classifies items into the list's own categories by keyword.
type RuleEngine struct{}

var keywords = map[string][]string{
	"cozinha":     {"fryer", "panela", "faqueiro", "talher", "prato", "jantar", "liquidificador", "cafeteira", "cafe", "micro", "batedeira", "geladeira", "copo", "taca", "forno", "mixer", "chaleira", "torradeira"},
	"mesa":        {"prato", "jantar", "taca", "copo", "sousplat", "guardanapo", "talher"},
	"quarto":      {"cama", "travesseiro", "edredom", "lencol", "colchao", "cobertor", "fronha"},
	"banheiro":    {"toalha", "roupao", "banheiro"},
	"sala":        {"tv", "televisao", "luminaria", "sofa", "manta", "almofada", "quadro", "tapete"},
	"limpeza":     {"aspirador", "vassoura", "mop", "balde", "limpeza"},
	"organizacao": {"organizador", "caixa", "cabide", "prateleira"},
	"higiene":     {"fralda", "lenco", "banheira", "pomada", "shampoo"},
	"roupinhas":   {"body", "bodies", "macacao", "meia", "touca", "roupa"},
	"passeio":     {"carrinho", "bebe conforto", "cadeirinha", "bolsa", "canguru"},
	"bebe":        {"berco", "mobile", "baba"},
	"bagagem":     {"mala", "mochila", "necessaire"},
	"lua de mel":  {"lua de mel", "viagem", "passeio", "jantar"},
}

func (RuleEngine) Classify(_ context.Context, text string, labels []string) (Decision, error) {
	normText := " " + textnorm.Normalize(text) + " "
	best, bestHits := "", 0
	for _, label := range labels {
		normLabel := textnorm.Normalize(label)
		hits := 0
		for key, words := range keywords {
			if !strings.Contains(normLabel, key) {
				continue
			}
			for _, w := range words {
				if strings.Contains(normText, " "+w) {
					hits++
				}
			}
		}
		if hits > bestHits {
			best, bestHits = label, hits
		}
	}
	if best == "" {
		return Decision{Source: "RULES"}, nil
	}
	return Decision{Label: best, Confidence: min(1, 0.6+0.2*float64(bestHits-1)), Source: "RULES"}, nil
}

// MockEngine always answers with a fixed label; for tests.
type MockEngine struct{ Label string }

func (m MockEngine) Classify(context.Context, string, []string) (Decision, error) {
	return Decision{Label: m.Label, Confidence: 1, Source: "MOCK"}, nil
}
