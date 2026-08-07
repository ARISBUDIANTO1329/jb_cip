package database

import (
	"database/sql"
	"fmt"

	"github.com/jaybani/jb_cip/config"
	_ "github.com/lib/pq"
)

var db *sql.DB

func Init(cfg *config.DatabaseConfig) (*sql.DB, error) {
	var err error
	db, err = sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func GetDB() (*sql.DB, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return db, nil
}

func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

func Ping() error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	return db.Ping()
}
