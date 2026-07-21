package migration

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"lockbox/internal/config"

	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Migrator struct {
	env *config.Env
}

func New(env *config.Env) *Migrator {
	return &Migrator{env: env}
}

func (m *Migrator) Up() error {
	mg, err := m.newMigrate()
	if err != nil {
		return fmt.Errorf("migrate up: %v", err)
	}

	defer closeMigrate(mg)

	if err := mg.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

func (m *Migrator) Down() error {
	mg, err := m.newMigrate()
	if err != nil {
		return fmt.Errorf("migrate up: %v", err)
	}

	defer closeMigrate(mg)

	if err := mg.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

func (m *Migrator) Force(version int) error {
	mg, err := m.newMigrate()
	if err != nil {
		return fmt.Errorf("migrate version: %v", err)
	}

	defer closeMigrate(mg)

	if err := mg.Force(version); err != nil {
		return err
	}

	return nil
}

func (m *Migrator) Version() (uint, bool, error) {
	mg, err := m.newMigrate()
	if err != nil {
		return 0, false, fmt.Errorf("migrate version: %v", err)
	}

	defer closeMigrate(mg)

	v, s, err := mg.Version()
	if err != nil {
		return 0, false, err
	}

	return v, s, err
}

func (m *Migrator) Bootstrap() error {
	v, dirty, err := m.Version()

	switch {
	case err == nil:
		if dirty {
			return fmt.Errorf("database is dirty at version %d", v)
		}
		return m.Up()

	case errors.Is(err, migrate.ErrNilVersion):
		hasSchema, err := m.HasSchema()
		if err != nil {
			return err
		}

		if hasSchema {
			if err := m.Force(1); err != nil {
				return err
			}
		}

		return m.Up()

	default:
		return err
	}
}

func (m *Migrator) HasSchema() (bool, error) {
	db, err := sql.Open("postgres", m.env.Database.URL)
	if err != nil {
		return false, err
	}
	defer func() {
		if dbErr := db.Close(); dbErr != nil {
			fmt.Printf("db error: %v", dbErr)
		}
	}()

	var exists bool

	err = db.QueryRow(`
		SELECT EXISTS (
		SELECT 1
		FROM information_schema.tables
		WHERE table_schema = 'public'
			AND table_name = 'users'
		)
	`).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (m *Migrator) newMigrate() (*migrate.Migrate, error) {
	return migrate.New(
		"file://migrations",
		m.env.Database.URL,
	)
}

func closeMigrate(m *migrate.Migrate) {
	srcErr, dbErr := m.Close()

	if srcErr != nil {
		log.Printf("close migration source: %v", srcErr)
	}

	if dbErr != nil {
		log.Printf("close migration database: %v", dbErr)
	}
}
