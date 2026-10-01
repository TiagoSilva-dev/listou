package catalog

import "testing"

func TestMatchingScore(t *testing.T) {
	m := MatchingService{}
	cases := []struct {
		desired string
		cand    Candidate
		want    Verdict
	}{
		{"Air Fryer 5L", Candidate{Title: "Air Fryer Mondial 5 litros"}, Match},
		{"Air Fryer 5L", Candidate{Title: "Air fryer compacta 4 litros"}, NoMatch},
		{"Panela elétrica", Candidate{Title: "Panela elétrica de arroz"}, Match},
		{"Panela elétrica de arroz", Candidate{Title: "Jogo de panelas inox"}, NoMatch},
		{"Jogo de cama queen", Candidate{Title: "Jogo de cama king 400 fios"}, PartialMatch},
	}
	for _, tc := range cases {
		if _, got := m.Score(tc.desired, "", tc.cand); got != tc.want {
			s, _ := m.Score(tc.desired, "", tc.cand)
			t.Errorf("%q vs %q: got %s (%.2f) want %s", tc.desired, tc.cand.Title, got, s, tc.want)
		}
	}
	ranked := m.Rank("Air Fryer 5L", "Cozinha", []Candidate{
		{Title: "Liquidificador"}, {Title: "Air fryer compacta 4 litros"}, {Title: "Fritadeira air fryer 5 litros", Category: "Cozinha"},
	})
	if ranked[0].Index != 2 {
		t.Fatalf("expected 5L fryer first, got %+v", ranked)
	}
}
