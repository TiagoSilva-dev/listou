package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/listou/listou/apps/api/internal/platform/database"
)

type Repository struct{}

const userColumns = `u.id, u.email, u.name, u.avatar_url, u.role`

func scanUser(row interface{ Scan(...any) error }) (User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.Name, &u.AvatarURL, &u.Role)
	return u, err
}

func (Repository) CreateUser(ctx context.Context, db database.DBTX, u User, passwordHash string) error {
	if _, err := db.Exec(ctx,
		`INSERT INTO users (id, email, name, role) VALUES ($1, $2, $3, $4)`,
		u.ID, u.Email, u.Name, u.Role); err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	if _, err := db.Exec(ctx,
		`INSERT INTO auth_identities (id, user_id, provider, provider_subject, password_hash)
		 VALUES ($1, $2, 'PASSWORD', $3, $4)`,
		uuid.Must(uuid.NewV7()), u.ID, u.Email, passwordHash); err != nil {
		return fmt.Errorf("insert identity: %w", err)
	}
	return nil
}

// FindPasswordIdentity returns the user and password hash for an email login.
func (Repository) FindPasswordIdentity(ctx context.Context, db database.DBTX, email string) (User, string, error) {
	row := db.QueryRow(ctx, `
		SELECT `+userColumns+`, i.password_hash
		FROM auth_identities i JOIN users u ON u.id = i.user_id
		WHERE i.provider = 'PASSWORD' AND i.provider_subject = $1`, email)
	var u User
	var hash string
	err := row.Scan(&u.ID, &u.Email, &u.Name, &u.AvatarURL, &u.Role, &hash)
	return u, hash, err
}

func (Repository) CreateSession(ctx context.Context, db database.DBTX, id, userID uuid.UUID, tokenHash []byte, expiresAt time.Time) error {
	_, err := db.Exec(ctx,
		`INSERT INTO sessions (id, user_id, token_hash, expires_at) VALUES ($1, $2, $3, $4)`,
		id, userID, tokenHash, expiresAt)
	return err
}

func (Repository) UserBySession(ctx context.Context, db database.DBTX, tokenHash []byte) (User, error) {
	return scanUser(db.QueryRow(ctx, `
		SELECT `+userColumns+`
		FROM sessions s JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1 AND s.revoked_at IS NULL AND s.expires_at > now()`, tokenHash))
}

func (Repository) RevokeSession(ctx context.Context, db database.DBTX, tokenHash []byte) error {
	_, err := db.Exec(ctx, `UPDATE sessions SET revoked_at = now() WHERE token_hash = $1 AND revoked_at IS NULL`, tokenHash)
	return err
}
