// Package flags implements simple configuration-backed feature flags.
package flags

import (
	"strconv"
	"strings"
)

type Flag string

const (
	AIListBuilder    Flag = "AI_LIST_BUILDER"
	GroupGifts       Flag = "GROUP_GIFTS"
	MultiMarketplace Flag = "MULTI_MARKETPLACE"
	JevMatching      Flag = "JEV_MATCHING"
	PriceAlerts      Flag = "PRICE_ALERTS"
)

var defaults = map[Flag]bool{
	AIListBuilder:    false,
	GroupGifts:       false,
	MultiMarketplace: true,
	JevMatching:      false,
	PriceAlerts:      false,
}

type Set struct{ values map[Flag]bool }

// Parse reads "FLAG_A=true,FLAG_B=false" on top of the defaults. Unknown or
// malformed entries are ignored so a typo never takes the API down.
func Parse(spec string) Set {
	values := make(map[Flag]bool, len(defaults))
	for k, v := range defaults {
		values[k] = v
	}
	for _, part := range strings.Split(spec, ",") {
		name, raw, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			continue
		}
		f := Flag(strings.ToUpper(strings.TrimSpace(name)))
		if _, known := defaults[f]; !known {
			continue
		}
		if b, err := strconv.ParseBool(strings.TrimSpace(raw)); err == nil {
			values[f] = b
		}
	}
	return Set{values: values}
}

func (s Set) Enabled(f Flag) bool { return s.values[f] }

func (s Set) All() map[Flag]bool {
	out := make(map[Flag]bool, len(s.values))
	for k, v := range s.values {
		out[k] = v
	}
	return out
}
