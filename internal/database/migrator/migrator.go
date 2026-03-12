package migrator

import (
	"fmt"

	"github.com/PavelsDenisovs/feature-flag-gatekeeper/db/migrations"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
)

func ApplyMigrations(dbURL string) error {
	d, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("failed to get migrations from file system: %v", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dbURL)
	if err != nil {
		return fmt.Errorf("failed to set up migration: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	return nil
}