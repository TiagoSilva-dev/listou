// Package textnorm normalizes Portuguese product text for matching and search.
package textnorm

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var stopwords = map[string]bool{
	"de": true, "da": true, "do": true, "das": true, "dos": true, "e": true, "com": true,
	"para": true, "a": true, "o": true, "as": true, "os": true, "em": true, "um": true, "uma": true,
	"kit": false,
}

var unitAliases = map[string]string{
	"litros": "l", "litro": "l", "lts": "l", "lt": "l",
	"polegadas": "pol", "polegada": "pol", "pecas": "pc", "peca": "pc", "pcs": "pc",
}

// Normalize lowercases, strips accents and collapses punctuation to spaces.
func Normalize(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	plain, _, err := transform.String(t, strings.ToLower(s))
	if err != nil {
		plain = strings.ToLower(s)
	}
	var b strings.Builder
	for _, r := range plain {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			b.WriteByte(' ')
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

// Tokens returns normalized, de-duplicated tokens without stopwords. Numbers
// glued to units ("5l", "5 litros") become "5l".
func Tokens(s string) []string {
	fields := strings.Fields(Normalize(s))
	seen := map[string]bool{}
	var out []string
	add := func(t string) {
		if t != "" && !stopwords[t] && !seen[t] {
			seen[t] = true
			out = append(out, t)
		}
	}
	for i := 0; i < len(fields); i++ {
		f := fields[i]
		if alias, ok := unitAliases[f]; ok {
			f = alias
		}
		if isNumber(f) && i+1 < len(fields) {
			next := fields[i+1]
			if alias, ok := unitAliases[next]; ok {
				next = alias
			}
			if next == "l" || next == "pol" || next == "pc" || next == "w" || next == "kg" {
				add(f + next)
				i++
				continue
			}
		}
		add(f)
	}
	return out
}

func isNumber(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != ""
}
