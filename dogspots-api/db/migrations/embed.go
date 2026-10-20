package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var Files embed.FS

func Up(ctx context.Context, database *sql.DB) error {
	database.SetMaxOpenConns(1)
	database.SetMaxIdleConns(1)

	for _, pragma := range []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA busy_timeout = 5000",
	} {
		if _, err := database.ExecContext(ctx, pragma); err != nil {
			return fmt.Errorf("configure sqlite before migrations: %w", err)
		}
	}

	provider, err := goose.NewProvider(goose.DialectSQLite3, database, Files)
	if err != nil {
		return fmt.Errorf("create goose provider: %w", err)
	}
	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("apply goose migrations: %w", err)
	}
	return nil
}
