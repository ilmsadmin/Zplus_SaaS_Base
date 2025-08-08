//go:build run_migration

package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

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

	// Read and execute migration file
	migrationFile := "database/migrations/012_create_domain_management_tables.sql"
	sqlBytes, err := ioutil.ReadFile(migrationFile)
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	sqlContent := string(sqlBytes)
	fmt.Printf("Executing migration: %s\n", filepath.Base(migrationFile))

	// Execute migration
	if _, err := db.Exec(sqlContent); err != nil {
		log.Fatalf("Failed to execute migration: %v", err)
	}

	fmt.Println("Migration executed successfully!")
}
