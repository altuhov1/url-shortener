package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upGrantPrivileges, downGrantPrivileges)
}

func quotePostgresIdentifier(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

func upGrantPrivileges(ctx context.Context, tx *sql.Tx) error {
	if err := godotenv.Load(); err != nil {
		slog.Warn("Файл .env не найден, использую системные переменные окружения")
	}
	username := os.Getenv("PG_USERNAME_FOR_APP")
	if username == "" {
		return fmt.Errorf("PG_USERNAME_FOR_APP is not set")
	}
	quotedUser := quotePostgresIdentifier(username)

	_, err := tx.ExecContext(ctx, fmt.Sprintf(`
		GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE urls TO %s;
	`, quotedUser))
	if err != nil {
		return err
	}
	fmt.Printf("Granted privileges to urls: %s\n", username)
	return nil
}

func downGrantPrivileges(ctx context.Context, tx *sql.Tx) error {
	if err := godotenv.Load(); err != nil {
		slog.Warn("Файл .env не найден, использую системные переменные окружения")
	}
	username := os.Getenv("PG_USERNAME_FOR_APP")
	if username == "" {
		username = "myapp_user"
	}
	quotedUser := quotePostgresIdentifier(username)

	_, err := tx.ExecContext(ctx, fmt.Sprintf(`
		REVOKE ALL PRIVILEGES ON TABLE urls FROM %s;
	`, quotedUser))
	if err != nil {
		if strings.Contains(err.Error(), "undefined_object") {
			fmt.Printf("Privileges already revoked for urls: %s\n", username)
			return nil
		}
		return err
	}

	fmt.Printf("Revoked privileges from urls: %s\n", username)
	return nil
}
