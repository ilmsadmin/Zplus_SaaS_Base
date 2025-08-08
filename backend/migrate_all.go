//go:build migrate_all

package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection
	dbURL := "postgres://postgres:postgres123@localhost:5432/zplus_saas_base?sslmode=disable"

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Connected to database successfully")

	// Create migration tracking table if it doesn't exist
	createMigrationTable := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(createMigrationTable); err != nil {
		log.Fatalf("Failed to create migration table: %v", err)
	}

	// Get list of migration files
	files, err := filepath.Glob("database/migrations/*.sql")
	if err != nil {
		log.Fatalf("Failed to list migration files: %v", err)
	}

	sort.Strings(files)

	// Check which migrations have been applied
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		log.Fatalf("Failed to query applied migrations: %v", err)
	}
	defer rows.Close()

	appliedMigrations := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			log.Fatalf("Failed to scan migration version: %v", err)
		}
		appliedMigrations[version] = true
	}

	// Apply pending migrations
	for _, file := range files {
		filename := filepath.Base(file)
		version := strings.TrimSuffix(filename, ".sql")

		if appliedMigrations[version] {
			fmt.Printf("Migration %s already applied, skipping\n", version)
			continue
		}

		fmt.Printf("Applying migration: %s\n", version)

		sqlBytes, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatalf("Failed to read migration file %s: %v", file, err)
		}

		// Execute migration
		if _, err := db.Exec(string(sqlBytes)); err != nil {
			log.Fatalf("Failed to execute migration %s: %v", version, err)
		}

		// Record migration as applied
		if _, err := db.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
			log.Fatalf("Failed to record migration %s: %v", version, err)
		}

		fmt.Printf("Successfully applied migration: %s\n", version)
	}

	fmt.Println("All migrations completed successfully!")
}
