package recommendations

import (
	"context"
	"sort"
	"strings"

	"github.com/listou/listou/apps/api/internal/catalog/textnorm"
)

// BuildRequest asks for desires that complement what a list already has.
type BuildRequest struct {
	EventType string
	// Prompt is optional free text ("apartamento pequeno, casal que adora cozinhar").
	Prompt string
	// Existing item titles; suggestions that match them are dropped.
	Existing []string
}

// Builder turns a Provider's suggestions into a short, ranked list of desires
// the owner does not have yet. It never produces products, prices or URLs.
type Builder struct {
	provider Provider
	limit    int
}

func NewBuilder(p Provider) Builder { return Builder{provider: p, limit: 24} }

func (b Builder) Suggest(ctx context.Context, req BuildRequest) ([]CategorySuggestion, error) {
	base, err := b.provider.SuggestCategories(ctx, Request{EventType: req.EventType, Prompt: req.Prompt})
	if err != nil {
		return nil, err
	}
	have := make([]map[string]bool, 0, len(req.Existing))
	for _, t := range req.Existing {
		have = append(have, tokenSet(t))
	}
	promptTokens := tokenSet(req.Prompt)

	type scored struct {
		cat   CategorySuggestion
		score int
	}
	var cats []scored
	remaining := b.limit
	for _, c := range base {
		out := CategorySuggestion{Name: c.Name, Emoji: c.Emoji, Reason: c.Reason}
		best := 0
		var ds []Desire
		for _, d := range c.Desires {
			if alreadyHave(tokenSet(d.Title), have) {
				continue
			}
			ds = append(ds, d)
		}
		sort.SliceStable(ds, func(i, j int) bool { return desireScore(ds[i], c, promptTokens) > desireScore(ds[j], c, promptTokens) })
		for _, d := range ds {
			if remaining == 0 {
				break
			}
			out.Desires = append(out.Desires, d)
			remaining--
			if s := desireScore(d, c, promptTokens); s > best {
				best = s
			}
		}
		if len(out.Desires) > 0 {
			cats = append(cats, scored{out, best})
		}
	}
	sort.SliceStable(cats, func(i, j int) bool { return cats[i].score > cats[j].score })
	res := make([]CategorySuggestion, 0, len(cats))
	for _, c := range cats {
		res = append(res, c.cat)
	}
	return res, nil
}

func importanceScore(i Importance) int {
	switch i {
	case Essential:
		return 3
	case Recommended:
		return 2
	}
	return 1
}

// desireScore ranks by importance, then boosts desires and categories the
// prompt mentions (each shared word counts, desire words more than category words).
func desireScore(d Desire, c CategorySuggestion, prompt map[string]bool) int {
	s := importanceScore(d.Importance)
	for t := range tokenSet(d.Title) {
		if prompt[t] {
			s += 4
		}
	}
	for t := range tokenSet(c.Name) {
		if prompt[t] {
			s += 2
		}
	}
	return s
}

// alreadyHave reports whether a list item covers the suggestion: all tokens of
// the shorter title appear in the longer one ("Air fryer" ~ "Air fryer 5L").
func alreadyHave(sug map[string]bool, have []map[string]bool) bool {
	if len(sug) == 0 {
		return false
	}
	for _, h := range have {
		if len(h) == 0 {
			continue
		}
		short, long := sug, h
		if len(h) < len(sug) {
			short, long = h, sug
		}
		all := true
		for t := range short {
			if !long[t] {
				all = false
				break
			}
		}
		if all {
			return true
		}
	}
	return false
}

func tokenSet(s string) map[string]bool {
	out := map[string]bool{}
	for _, t := range textnorm.Tokens(strings.TrimSpace(s)) {
		out[t] = true
	}
	return out
}
