package store

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	migratepgx "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	// Registers the "pgx/v5" driver with database/sql, which golang-migrate needs.
	// The blank import is for that side effect alone — nothing here calls it.
	_ "github.com/jackc/pgx/v5/stdlib"
)

// migrationsDir is where the .sql files sit inside the embedded filesystem.
const migrationsDir = "sql/migrations"

// Migrate applies every migration that has not run yet, and returns once the
// database schema matches the files.
//
// Migrations are passed in as a filesystem rather than read from disk, because
// main embeds them into the binary with go:embed. That means a deployed binary
// carries its own schema: there is no second artefact to ship and no chance of
// running new code against an old database.
//
// It is called at startup, before the server accepts requests. Applying the
// schema by hand — which is what step 2a did — means every environment drifts
// separately and nothing records which changes have run where.
func Migrate(migrations fs.FS, databaseURL string) error {
	source, err := iofs.New(migrations, migrationsDir)
	if err != nil {
		return fmt.Errorf("read embedded migrations: %w", err)
	}

	// golang-migrate works through database/sql, not pgxpool, so this is a second
	// short-lived connection — opened, used, and closed before the pool the
	// application actually serves from.
	db, err := sql.Open("pgx/v5", databaseURL)
	if err != nil {
		return fmt.Errorf("open database for migrations: %w", err)
	}
	defer db.Close()

	driver, err := migratepgx.WithInstance(db, &migratepgx.Config{})
	if err != nil {
		return fmt.Errorf("prepare migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "pgx5", driver)
	if err != nil {
		return fmt.Errorf("prepare migrations: %w", err)
	}

	// ErrNoChange means the database was already up to date, which is the normal
	// case on every restart after the first.
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}
