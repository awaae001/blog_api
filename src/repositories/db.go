package repositories

import (
	"blog_api/src/model"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB initializes the database and runs migrations.
func InitDB(cfg *model.Config) (*sql.DB, error) {
	dbPath := cfg.Data.Database.Path
	if dbPath == "" {
		return nil, fmt.Errorf("database path is not configured")
	}

	log.Printf("初始化数据库于: %s", dbPath)

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
		log.Printf("运行迁移: %s\n", file)
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("could not read migration file %s: %w", file, err)
		}

		// Split the content by semicolon to execute multiple statements
		statements := strings.Split(string(content), ";")
		for _, stmt := range statements {
			// Trim whitespace and skip empty statements
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := db.Exec(stmt); err != nil {
				return nil, fmt.Errorf("could not execute migration statement in file %s: %w", file, err)
			}
		}
	}

	log.Println("Database migrations completed successfully.")
	return db, nil
}
