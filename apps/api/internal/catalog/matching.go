package catalog

import (
	"sort"

	"github.com/listou/listou/apps/api/internal/catalog/textnorm"
)

type Verdict string

const (
	Match        Verdict = "MATCH"
	PartialMatch Verdict = "PARTIAL_MATCH"
	NoMatch      Verdict = "NO_MATCH"
)

type Candidate struct {
	Title    string
	Brand    string
	Category string
}

type Ranked struct {
	Index   int     `json:"index"`
	Score   float64 `json:"score"`
	Verdict Verdict `json:"verdict"`
}

// MatchingService ranks catalog candidates against a user's desired item
// ("Air Fryer 5L" vs "Air fryer Mondial 5 litros"). It is deterministic and
// explainable; a DecisionEngine (e.g. Jev) can replace Score later behind
// the JEV_MATCHING flag if tests show a real gain.
type MatchingService struct{}

func (MatchingService) Score(desired, desiredCategory string, c Candidate) (float64, Verdict) {
	want := textnorm.Tokens(desired)
	if len(want) == 0 {
		return 0, NoMatch
	}
	have := map[string]bool{}
	for _, t := range textnorm.Tokens(c.Title + " " + c.Brand) {
		have[t] = true
	}
	hits := 0
	penalty := 0.0
	for _, t := range want {
		if have[t] {
			hits++
			continue
		}
		// A different size/quantity ("5l" vs "4l") is a strong negative signal.
		if isMeasure(t) && hasOtherMeasure(have, t) {
			penalty += 0.35
		}
	}
	score := float64(hits)/float64(len(want)) - penalty
	if desiredCategory != "" && c.Category != "" && textnorm.Normalize(desiredCategory) == textnorm.Normalize(c.Category) {
		score += 0.1
	}
	if score > 1 {
		score = 1
	}
	if score < 0 {
		score = 0
	}
	switch {
	case score >= 0.75:
		return score, Match
	case score >= 0.4:
		return score, PartialMatch
	default:
		return score, NoMatch
	}
}

// Rank returns candidates ordered by score, best first.
func (m MatchingService) Rank(desired, desiredCategory string, cands []Candidate) []Ranked {
	out := make([]Ranked, len(cands))
	for i, c := range cands {
		s, v := m.Score(desired, desiredCategory, c)
		out[i] = Ranked{Index: i, Score: s, Verdict: v}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	return out
}

func isMeasure(t string) bool {
	if len(t) < 2 || t[0] < '0' || t[0] > '9' {
		return false
	}
	last := t[len(t)-1]
	return last < '0' || last > '9'
}

func hasOtherMeasure(have map[string]bool, t string) bool {
	suffix := trimLeadingDigits(t)
	for h := range have {
		if h != t && isMeasure(h) && trimLeadingDigits(h) == suffix {
			return true
		}
	}
	return false
}

func trimLeadingDigits(s string) string {
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	return s[i:]
}
