// Package auth handles accounts, password login and server-side sessions.
package auth

import (
	"context"
	"net/mail"
	"strings"

	"github.com/google/uuid"

	"github.com/listou/listou/apps/api/internal/platform/validate"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	AvatarURL *string   `json:"avatarUrl"`
	Role      string    `json:"role"`
}

func (u User) IsAdmin() bool { return u.Role == "ADMIN" }

type RegisterInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (in *RegisterInput) Validate() error {
	in.Name = strings.TrimSpace(in.Name)
	in.Email = NormalizeEmail(in.Email)
	errs := validate.Errors{}
	errs.Length("name", in.Name, 2, 80, "Conte seu nome (2 a 80 caracteres).")
	if addr, err := mail.ParseAddress(in.Email); err != nil || addr.Address != in.Email || len(in.Email) > 254 {
		errs.Add("email", "E-mail inválido.")
	}
	if n := len(in.Password); n < 8 || n > 128 {
		errs.Add("password", "Use uma senha com 8 a 128 caracteres.")
	}
	return errs.Err()
}

type ctxKey struct{}

func WithUser(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, ctxKey{}, u)
}

// UserFrom returns the authenticated user, if any.
func UserFrom(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(ctxKey{}).(User)
	return u, ok
}
