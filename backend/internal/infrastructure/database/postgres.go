package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type PostgresDB struct {
	*gorm.DB
}

type PostgresConfig struct {
	Host               string
	Port               int
	User               string
	Password           string
	DBName             string
	SSLMode            string
	MaxOpenConnections int
	MaxIdleConnections int
	ConnectionMaxAge   time.Duration
	LogLevel           logger.LogLevel
}

func NewPostgresDB(config PostgresConfig) (*PostgresDB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DBName,
		config.SSLMode,
	)

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(config.LogLevel),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(config.MaxOpenConnections)
	sqlDB.SetMaxIdleConns(config.MaxIdleConnections)
	sqlDB.SetConnMaxLifetime(config.ConnectionMaxAge)

	return &PostgresDB{DB: db}, nil
}

func (p *PostgresDB) GetTenantDB(tenantID string) *gorm.DB {
	// Implement schema-based multi-tenancy
	schemaName := fmt.Sprintf("tenant_%s", tenantID)
	return p.DB.Exec(fmt.Sprintf("SET search_path TO %s", schemaName))
}

func (p *PostgresDB) CreateTenantSchema(tenantID string) error {
	schemaName := fmt.Sprintf("tenant_%s", tenantID)
	return p.DB.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)).Error
}

func (p *PostgresDB) DropTenantSchema(tenantID string) error {
	schemaName := fmt.Sprintf("tenant_%s", tenantID)
	return p.DB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName)).Error
}

func (p *PostgresDB) Close() error {
	sqlDB, err := p.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
