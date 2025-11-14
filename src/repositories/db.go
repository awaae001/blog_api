package repositories

import (
	"blog_api/src/model"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB initializes the database and runs migrations.
func InitDB(cfg *model.Config) (*sql.DB, error) {
	dbPath := cfg.Data.Database.Path
	if dbPath == "" {
		return nil, fmt.Errorf("database path is not configured")
	}

	log.Printf("Initializing database at: %s", dbPath)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("could not open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("could not connect to database: %w", err)
	}

	// Run migrations
	migrationFiles, err := filepath.Glob("migrations/*.sql")
	if err != nil {
		return nil, fmt.Errorf("could not find migration files: %w", err)
	}
	sort.Strings(migrationFiles)

	for _, file := range migrationFiles {
		log.Printf("Running migration: %s\n", file)
		content, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("could not read migration file %s: %w", file, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			return nil, fmt.Errorf("could not execute migration file %s: %w", file, err)
		}
	}

	log.Println("Database migrations completed successfully.")
	return db, nil
}
