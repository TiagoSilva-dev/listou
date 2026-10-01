// Package validate collects field errors into a VALIDATION_FAILED response.
package validate

import (
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/listou/listou/apps/api/internal/platform/httpx"
)

type Errors map[string]string

func (e Errors) Add(field, msg string) {
	if _, exists := e[field]; !exists {
		e[field] = msg
	}
}

// Err returns a *httpx.Error when any field failed, otherwise nil.
func (e Errors) Err() error {
	if len(e) == 0 {
		return nil
	}
	return httpx.Validation(e)
}

func (e Errors) Length(field, value string, min, max int, msg string) {
	n := utf8.RuneCountInString(value)
	if n < min || n > max {
		e.Add(field, msg)
	}
}

// OptionalURL checks that a non-empty value is an absolute http(s) URL.
func (e Errors) OptionalURL(field string, value *string) {
	if value == nil || *value == "" {
		return
	}
	u, err := url.Parse(*value)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || len(*value) > 2048 {
		e.Add(field, "Link inválido.")
	}
}

// Trim trims a string pointer in place and turns blanks into nil.
func Trim(s *string) *string {
	if s == nil {
		return nil
	}
	t := strings.TrimSpace(*s)
	if t == "" {
		return nil
	}
	return &t
}
