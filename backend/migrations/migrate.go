package migrations

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"sort"
)

//go:embed *.sql
var files embed.FS

const migrationsTable = "schema_migrations_cr45_reduced"

func Up(db *sql.DB) error {
	if db == nil {
		return errors.New("db is required")
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS ` + migrationsTable + ` (name TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`); err != nil {
		return err
	}

	names, err := fs.Glob(files, "*.sql")
	if err != nil {
		return fmt.Errorf("glob migrations: %w", err)
	}
	sort.Strings(names)

	for _, name := range names {
		var exists bool
		if err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM `+migrationsTable+` WHERE name = $1)`, name).Scan(&exists); err != nil {
			return err
		}
		if exists {
			continue
		}

		sqlBytes, err := files.ReadFile(name)
		if err != nil {
			return err
		}
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(string(sqlBytes)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply %s: %w", name, err)
		}
		if _, err := tx.Exec(`INSERT INTO `+migrationsTable+` (name) VALUES ($1)`, name); err != nil {
			_ = tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}

	return nil
}