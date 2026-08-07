package database

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jaybani/jb_cip/config"
)

var migrateInstance *migrate.Migrate

func RunMigrations(cfg *config.Config) error {
	db, err := GetDB()
	if err != nil {
		return fmt.Errorf("database not initialized: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migrate driver: %w", err)
	}

	migrationPath := cfg.Database.MigrationPath
	if migrationPath == "" {
		migrationPath = "migrations"
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	migrateInstance = m

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func RollbackMigration() error {
	if migrateInstance == nil {
		return fmt.Errorf("migration instance not initialized")
	}
	return migrateInstance.Steps(-1)
}

func MigrateDown() error {
	if migrateInstance == nil {
		return fmt.Errorf("migration instance not initialized")
	}
	return migrateInstance.Down()
}

func MigrateVersion() (uint, bool, error) {
	if migrateInstance == nil {
		return 0, false, fmt.Errorf("migration instance not initialized")
	}
	return migrateInstance.Version()
}

// GetMigrationFiles returns list of migration files
func GetMigrationFiles(migrationPath string) ([]string, error) {
	var files []string
	err := filepath.Walk(migrationPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".up.sql") {
			files = append(files, filepath.Base(path))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}
