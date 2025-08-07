package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"crypto/sha256"
	"encoding/hex"

	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
)

type Migration struct {
	Version       string
	Filename      string
	Checksum      string
	AppliedAt     *time.Time
	ExecutionTime *int
}

type Migrator struct {
	db             *sql.DB
	migrationsPath string
}

func NewMigrator(db *sql.DB, migrationsPath string) *Migrator {
	return &Migrator{
		db:             db,
		migrationsPath: migrationsPath,
	}
}

// ensureMigrationsTable creates the schema_migrations table if it doesn't exist
func (m *Migrator) ensureMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			checksum VARCHAR(64),
			execution_time_ms INTEGER
		);
	`
	_, err := m.db.Exec(query)
	return err
}

// getAppliedMigrations returns a map of applied migrations
func (m *Migrator) getAppliedMigrations() (map[string]Migration, error) {
	query := `
		SELECT version, applied_at, checksum, execution_time_ms 
		FROM schema_migrations 
		ORDER BY version
	`
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]Migration)
	for rows.Next() {
		var m Migration
		var checksum sql.NullString
		var executionTime sql.NullInt32

		err := rows.Scan(&m.Version, &m.AppliedAt, &checksum, &executionTime)
		if err != nil {
			return nil, err
		}

		if checksum.Valid {
			m.Checksum = checksum.String
		}
		if executionTime.Valid {
			execTime := int(executionTime.Int32)
			m.ExecutionTime = &execTime
		}

		applied[m.Version] = m
	}

	return applied, nil
}

// getAvailableMigrations returns all migration files
func (m *Migrator) getAvailableMigrations() ([]Migration, error) {
	files, err := filepath.Glob(filepath.Join(m.migrationsPath, "*.sql"))
	if err != nil {
		return nil, err
	}

	var migrations []Migration
	for _, file := range files {
		if strings.HasSuffix(file, "_rollback.sql") {
			continue // Skip rollback files
		}

		filename := filepath.Base(file)
		version := strings.TrimSuffix(filename, ".sql")

		// Calculate checksum
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		hash := sha256.Sum256(content)
		checksum := hex.EncodeToString(hash[:])

		migrations = append(migrations, Migration{
			Version:  version,
			Filename: filename,
			Checksum: checksum,
		})
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// applyMigration applies a single migration
func (m *Migrator) applyMigration(migration Migration) error {
	migrationPath := filepath.Join(m.migrationsPath, migration.Filename)

	log.Printf("Applying migration: %s", migration.Version)

	content, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	start := time.Now()
	_, err = m.db.Exec(string(content))
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}
	executionTime := int(time.Since(start).Milliseconds())

	// Record the migration
	_, err = m.db.Exec(`
		INSERT INTO schema_migrations (version, checksum, execution_time_ms) 
		VALUES ($1, $2, $3)
	`, migration.Version, migration.Checksum, executionTime)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	log.Printf("Migration %s applied successfully (%dms)", migration.Version, executionTime)
	return nil
}

// rollbackMigration rolls back a single migration
func (m *Migrator) rollbackMigration(version string) error {
	rollbackFile := filepath.Join(m.migrationsPath, version+"_rollback.sql")

	if _, err := os.Stat(rollbackFile); os.IsNotExist(err) {
		return fmt.Errorf("rollback file not found: %s", rollbackFile)
	}

	log.Printf("Rolling back migration: %s", version)

	content, err := os.ReadFile(rollbackFile)
	if err != nil {
		return fmt.Errorf("failed to read rollback file: %w", err)
	}

	_, err = m.db.Exec(string(content))
	if err != nil {
		return fmt.Errorf("failed to execute rollback: %w", err)
	}

	// Remove from migrations table
	_, err = m.db.Exec("DELETE FROM schema_migrations WHERE version = $1", version)
	if err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	log.Printf("Migration %s rolled back successfully", version)
	return nil
}

// Up applies all pending migrations
func (m *Migrator) Up() error {
	if err := m.ensureMigrationsTable(); err != nil {
		return fmt.Errorf("failed to ensure migrations table: %w", err)
	}

	applied, err := m.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	available, err := m.getAvailableMigrations()
	if err != nil {
		return fmt.Errorf("failed to get available migrations: %w", err)
	}

	pendingCount := 0
	for _, migration := range available {
		if appliedMigration, exists := applied[migration.Version]; exists {
			// Check if checksum has changed
			if appliedMigration.Checksum != migration.Checksum {
				log.Printf("WARNING: Migration %s checksum has changed!", migration.Version)
				log.Printf("Applied: %s", appliedMigration.Checksum)
				log.Printf("Current: %s", migration.Checksum)
			}
			continue
		}

		if err := m.applyMigration(migration); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.Version, err)
		}
		pendingCount++
	}

	if pendingCount == 0 {
		log.Println("No pending migrations")
	} else {
		log.Printf("Applied %d migrations", pendingCount)
	}

	return nil
}

// Down rolls back the last migration
func (m *Migrator) Down() error {
	applied, err := m.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if len(applied) == 0 {
		log.Println("No migrations to rollback")
		return nil
	}

	// Find the latest migration
	var latestVersion string
	var latestTime time.Time
	for version, migration := range applied {
		if migration.AppliedAt != nil && migration.AppliedAt.After(latestTime) {
			latestVersion = version
			latestTime = *migration.AppliedAt
		}
	}

	if latestVersion == "" {
		log.Println("No migrations to rollback")
		return nil
	}

	return m.rollbackMigration(latestVersion)
}

// Status shows migration status
func (m *Migrator) Status() error {
	if err := m.ensureMigrationsTable(); err != nil {
		return fmt.Errorf("failed to ensure migrations table: %w", err)
	}

	applied, err := m.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	available, err := m.getAvailableMigrations()
	if err != nil {
		return fmt.Errorf("failed to get available migrations: %w", err)
	}

	fmt.Println("Migration Status:")
	fmt.Println("=================")
	fmt.Printf("%-40s %-10s %-20s %-15s\n", "Version", "Status", "Applied At", "Execution Time")
	fmt.Println(strings.Repeat("-", 85))

	for _, migration := range available {
		status := "PENDING"
		appliedAt := ""
		executionTime := ""

		if appliedMigration, exists := applied[migration.Version]; exists {
			status = "APPLIED"
			if appliedMigration.AppliedAt != nil {
				appliedAt = appliedMigration.AppliedAt.Format("2006-01-02 15:04:05")
			}
			if appliedMigration.ExecutionTime != nil {
				executionTime = fmt.Sprintf("%dms", *appliedMigration.ExecutionTime)
			}

			// Check checksum
			if appliedMigration.Checksum != migration.Checksum {
				status = "MODIFIED"
			}
		}

		fmt.Printf("%-40s %-10s %-20s %-15s\n", migration.Version, status, appliedAt, executionTime)
	}

	fmt.Printf("\nSummary:\n")
	fmt.Printf("Total migrations: %d\n", len(available))
	fmt.Printf("Applied migrations: %d\n", len(applied))
	fmt.Printf("Pending migrations: %d\n", len(available)-len(applied))

	return nil
}

// CreateTenantSchema creates a new tenant schema
func (m *Migrator) CreateTenantSchema(tenantSlug string) error {
	log.Printf("Creating schema for tenant: %s", tenantSlug)

	_, err := m.db.Exec("SELECT create_tenant_schema($1)", tenantSlug)
	if err != nil {
		return fmt.Errorf("failed to create tenant schema: %w", err)
	}

	log.Printf("Tenant schema created successfully for: %s", tenantSlug)
	return nil
}

// DropTenantSchema drops a tenant schema
func (m *Migrator) DropTenantSchema(tenantSlug string) error {
	log.Printf("Dropping schema for tenant: %s", tenantSlug)

	_, err := m.db.Exec("SELECT drop_tenant_schema($1)", tenantSlug)
	if err != nil {
		return fmt.Errorf("failed to drop tenant schema: %w", err)
	}

	log.Printf("Tenant schema dropped successfully for: %s", tenantSlug)
	return nil
}

func main() {
	var dbURL string
	var migrationsPath string

	rootCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Database migration tool for Zplus SaaS Base",
		Long:  "A database migration tool that supports PostgreSQL schema-per-tenant architecture",
	}

	rootCmd.PersistentFlags().StringVar(&dbURL, "database-url", os.Getenv("DATABASE_URL"), "Database connection URL")
	rootCmd.PersistentFlags().StringVar(&migrationsPath, "migrations-path", "./migrations", "Path to migrations directory")

	upCmd := &cobra.Command{
		Use:   "up",
		Short: "Apply all pending migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := sql.Open("postgres", dbURL)
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer db.Close()

			migrator := NewMigrator(db, migrationsPath)
			return migrator.Up()
		},
	}

	downCmd := &cobra.Command{
		Use:   "down",
		Short: "Rollback the last migration",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := sql.Open("postgres", dbURL)
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer db.Close()

			migrator := NewMigrator(db, migrationsPath)
			return migrator.Down()
		},
	}

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show migration status",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := sql.Open("postgres", dbURL)
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer db.Close()

			migrator := NewMigrator(db, migrationsPath)
			return migrator.Status()
		},
	}

	createTenantCmd := &cobra.Command{
		Use:   "create-tenant [tenant-slug]",
		Short: "Create schema for a new tenant",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := sql.Open("postgres", dbURL)
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer db.Close()

			migrator := NewMigrator(db, migrationsPath)
			return migrator.CreateTenantSchema(args[0])
		},
	}

	dropTenantCmd := &cobra.Command{
		Use:   "drop-tenant [tenant-slug]",
		Short: "Drop schema for a tenant (DANGEROUS)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Are you sure you want to drop all data for tenant '%s'? (yes/no): ", args[0])
			var confirm string
			fmt.Scanln(&confirm)

			if confirm != "yes" {
				fmt.Println("Operation cancelled")
				return nil
			}

			db, err := sql.Open("postgres", dbURL)
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer db.Close()

			migrator := NewMigrator(db, migrationsPath)
			return migrator.DropTenantSchema(args[0])
		},
	}

	rootCmd.AddCommand(upCmd, downCmd, statusCmd, createTenantCmd, dropTenantCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
