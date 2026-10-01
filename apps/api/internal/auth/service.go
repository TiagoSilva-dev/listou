package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/listou/listou/apps/api/internal/platform/database"
	"github.com/listou/listou/apps/api/internal/platform/httpx"
	"github.com/listou/listou/apps/api/internal/platform/ids"
)

var (
	ErrEmailTaken         = httpx.NewError(http.StatusConflict, "EMAIL_TAKEN", "Já existe uma conta com este e-mail.")
	ErrInvalidCredentials = httpx.NewError(http.StatusUnauthorized, "INVALID_CREDENTIALS", "E-mail ou senha incorretos.")
)

// Tracker records product analytics without auth depending on the analytics module.
type Tracker interface {
	TrackUser(ctx context.Context, name string, userID string, props map[string]any)
}

type Service struct {
	pool       *pgxpool.Pool
	repo       Repository
	sessionTTL time.Duration
	tracker    Tracker
	now        func() time.Time
	// dummyHash keeps login timing similar whether or not the email exists.
	dummyHash string
}

func NewService(pool *pgxpool.Pool, sessionTTL time.Duration, tracker Tracker) *Service {
	dummy, _ := HashPassword("timing-equalizer-password")
	return &Service{pool: pool, sessionTTL: sessionTTL, tracker: tracker, now: time.Now, dummyHash: dummy}
}

type Session struct {
	Token     string
	ExpiresAt time.Time
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (User, Session, error) {
	if err := in.Validate(); err != nil {
		return User{}, Session{}, err
	}
	hash, err := HashPassword(in.Password)
	if err != nil {
		return User{}, Session{}, err
	}
	u := User{ID: ids.New(), Email: in.Email, Name: in.Name, Role: "USER"}
	var sess Session
	err = database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		if err := s.repo.CreateUser(ctx, tx, u, hash); err != nil {
			if database.IsUniqueViolation(err, "") {
				return ErrEmailTaken
			}
			return err
		}
		sess, err = s.createSession(ctx, tx, u)
		return err
	})
	if err != nil {
		return User{}, Session{}, err
	}
	if s.tracker != nil {
		s.tracker.TrackUser(ctx, "USER_REGISTERED", u.ID.String(), nil)
	}
	return u, sess, nil
}

func (s *Service) Login(ctx context.Context, in LoginInput) (User, Session, error) {
	email := NormalizeEmail(in.Email)
	u, hash, err := s.repo.FindPasswordIdentity(ctx, s.pool, email)
	if err != nil {
		if !database.IsNoRows(err) {
			return User{}, Session{}, fmt.Errorf("login: %w", err)
		}
		_, _ = VerifyPassword(in.Password, s.dummyHash)
		return User{}, Session{}, ErrInvalidCredentials
	}
	ok, err := VerifyPassword(in.Password, hash)
	if err != nil {
		return User{}, Session{}, fmt.Errorf("login: %w", err)
	}
	if !ok {
		return User{}, Session{}, ErrInvalidCredentials
	}
	sess, err := s.createSession(ctx, s.pool, u)
	return u, sess, err
}

func (s *Service) createSession(ctx context.Context, db database.DBTX, u User) (Session, error) {
	token, hash := ids.Token()
	expires := s.now().Add(s.sessionTTL)
	if err := s.repo.CreateSession(ctx, db, ids.New(), u.ID, hash, expires); err != nil {
		return Session{}, fmt.Errorf("create session: %w", err)
	}
	return Session{Token: token, ExpiresAt: expires}, nil
}

func (s *Service) Authenticate(ctx context.Context, token string) (User, error) {
	if token == "" || len(token) > 128 {
		return User{}, httpx.ErrUnauthorized
	}
	u, err := s.repo.UserBySession(ctx, s.pool, ids.HashToken(token))
	if err != nil {
		if database.IsNoRows(err) {
			return User{}, httpx.ErrUnauthorized
		}
		return User{}, fmt.Errorf("authenticate: %w", err)
	}
	return u, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.repo.RevokeSession(ctx, s.pool, ids.HashToken(token))
}
