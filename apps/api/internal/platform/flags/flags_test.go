package flags

import "testing"

func TestParse(t *testing.T) {
	s := Parse("AI_LIST_BUILDER=true, multi_marketplace=false,UNKNOWN=true,BROKEN,GROUP_GIFTS=maybe")
	cases := map[Flag]bool{
		AIListBuilder:    true,
		MultiMarketplace: false,
		GroupGifts:       false,
		JevMatching:      false,
	}
	for f, want := range cases {
		if got := s.Enabled(f); got != want {
			t.Errorf("%s: got %v want %v", f, got, want)
		}
	}
	if _, ok := s.All()["UNKNOWN"]; ok {
		t.Error("unknown flag should be ignored")
	}
}
