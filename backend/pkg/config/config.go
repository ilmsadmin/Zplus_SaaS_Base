package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	MongoDB  MongoDBConfig
	JWT      JWTConfig
	Logger   LoggerConfig
}

type AppConfig struct {
	Name        string
	Environment string
	Version     string
	Debug       bool
}

type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	Host               string
	Port               int
	User               string
	Password           string
	DBName             string
	SSLMode            string
	MaxOpenConnections int
	MaxIdleConnections int
	ConnectionMaxAge   time.Duration
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type MongoDBConfig struct {
	URI      string
	Database string
	Timeout  time.Duration
}

type JWTConfig struct {
	Secret         string
	ExpirationTime time.Duration
	Issuer         string
}

type LoggerConfig struct {
	Level      string
	Format     string
	OutputPath string
}

func Load() (*Config, error) {
	// Load .env file if exists
	if err := godotenv.Load(); err != nil {
		// Don't fail if .env file doesn't exist
		fmt.Printf("Warning: .env file not found: %v\n", err)
	}

	config := &Config{
		App: AppConfig{
			Name:        getEnv("APP_NAME", "Zplus SaaS Base"),
			Environment: getEnv("APP_ENV", "development"),
			Version:     getEnv("APP_VERSION", "1.0.0"),
			Debug:       getEnvAsBool("APP_DEBUG", true),
		},
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "localhost"),
			Port:         getEnvAsInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getEnvAsDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),
		},
		Database: DatabaseConfig{
			Host:               getEnv("DB_HOST", "localhost"),
			Port:               getEnvAsInt("DB_PORT", 5432),
			User:               getEnv("DB_USER", "postgres"),
			Password:           getEnv("DB_PASSWORD", ""),
			DBName:             getEnv("DB_NAME", "zplus_saas"),
			SSLMode:            getEnv("DB_SSLMODE", "disable"),
			MaxOpenConnections: getEnvAsInt("DB_MAX_OPEN_CONNECTIONS", 25),
			MaxIdleConnections: getEnvAsInt("DB_MAX_IDLE_CONNECTIONS", 25),
			ConnectionMaxAge:   getEnvAsDuration("DB_CONNECTION_MAX_AGE", 5*time.Minute),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		MongoDB: MongoDBConfig{
			URI:      getEnv("MONGO_URI", "mongodb://localhost:27017"),
			Database: getEnv("MONGO_DATABASE", "zplus_saas"),
			Timeout:  getEnvAsDuration("MONGO_TIMEOUT", 10*time.Second),
		},
		JWT: JWTConfig{
			Secret:         getEnv("JWT_SECRET", "your-secret-key"),
			ExpirationTime: getEnvAsDuration("JWT_EXPIRATION", 24*time.Hour),
			Issuer:         getEnv("JWT_ISSUER", "zplus-saas-base"),
		},
		Logger: LoggerConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			Format:     getEnv("LOG_FORMAT", "json"),
			OutputPath: getEnv("LOG_OUTPUT_PATH", "stdout"),
		},
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := strconv.ParseBool(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := time.ParseDuration(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}

func (c *Config) DatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}
