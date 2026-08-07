package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jaybani/jb_cip/config"
	"github.com/jaybani/jb_cip/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger.Init(cfg.App.LogLevel, cfg.App.LogFormat)
	log := logger.Get()

	if len(os.Args) < 2 {
		fmt.Println("Usage: migrate [up|down|version|drop]")
		os.Exit(1)
	}

	command := os.Args[1]

	db, err := sql.Open("postgres", cfg.Database.DSN())
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.WithError(err).Fatal("Failed to create migrate driver")
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", cfg.Database.MigrationPath),
		"postgres",
		driver,
	)
	if err != nil {
		log.WithError(err).Fatal("Failed to create migrate instance")
	}

	switch command {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.WithError(err).Fatal("Migration failed")
		}
		version, dirty, err := m.Version()
		if err != nil {
			log.WithError(err).Error("Failed to get version")
		}
		log.Infof("Migration completed. Version: %d, Dirty: %v", version, dirty)

	case "down":
		if err := m.Down(); err != nil {
			log.WithError(err).Fatal("Migration down failed")
		}
		log.Info("Migration rolled back")

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.WithError(err).Error("Failed to get version")
		}
		fmt.Printf("Version: %d, Dirty: %v\n", version, dirty)

	case "drop":
		if err := m.Down(); err != nil {
			log.WithError(err).Fatal("Drop failed")
		}
		log.Info("All tables dropped")

	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}
