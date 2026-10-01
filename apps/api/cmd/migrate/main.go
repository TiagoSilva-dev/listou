// Command migrate applies versioned schema migrations and development seeds.
//
//	go run ./cmd/migrate up|down|status|reset|seed|admin <email> [...]
//
// Several commands may be chained, e.g. `migrate up seed`.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/listou/listou/apps/api/migrations"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "migrate:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: migrate up|down|status|reset|seed|admin <email>")
	}
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	db, err := sql.Open("pgx", url)
	if err != nil {
		return err
	}
	defer db.Close()

	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	ctx := context.Background()
	for i := 0; i < len(args); i++ {
		cmd := args[i]
		if cmd == "admin" { // takes the user's e-mail as the next argument
			if i+1 >= len(args) {
				return fmt.Errorf("usage: migrate admin <email>")
			}
			i++
			if err := makeAdmin(ctx, db, args[i]); err != nil {
				return err
			}
			continue
		}
		if err := runCommand(ctx, db, url, cmd); err != nil {
			return err
		}
	}
	return nil
}

// makeAdmin grants the ADMIN role to an existing account (the only way to get one).
func makeAdmin(ctx context.Context, db *sql.DB, email string) error {
	res, err := db.ExecContext(ctx, `UPDATE users SET role = 'ADMIN' WHERE email = $1`, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("no user with e-mail %q", email)
	}
	fmt.Println("migrate: admin role granted to", email)
	return nil
}

func runCommand(ctx context.Context, db *sql.DB, url, cmd string) error {
	switch cmd {
	case "up":
		return goose.UpContext(ctx, db, ".")
	case "down":
		return goose.DownContext(ctx, db, ".")
	case "status":
		return goose.StatusContext(ctx, db, ".")
	case "reset":
		// Dev helper: down-migrations cannot remove reference rows that data points to,
		// so wipe the schema and re-apply from scratch. Never allowed in production.
		if os.Getenv("APP_ENV") == "production" {
			return fmt.Errorf("reset is disabled when APP_ENV=production")
		}
		if _, err := db.ExecContext(ctx, "DROP SCHEMA public CASCADE; CREATE SCHEMA public"); err != nil {
			return err
		}
		return goose.UpContext(ctx, db, ".")
	case "seed":
		return seed(ctx, url)
	default:
		return fmt.Errorf("unknown command %q", cmd)
	}
}
