package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/ilmsadmin/zplus-saas-base/internal/application"
	"github.com/ilmsadmin/zplus-saas-base/internal/infrastructure/auth"
	"github.com/ilmsadmin/zplus-saas-base/internal/infrastructure/database"
	"github.com/ilmsadmin/zplus-saas-base/internal/infrastructure/middleware"
	"github.com/ilmsadmin/zplus-saas-base/internal/interfaces"
	"github.com/ilmsadmin/zplus-saas-base/pkg/config"
	"github.com/ilmsadmin/zplus-saas-base/pkg/logger"
)

type Application struct {
	Config            *config.Config
	Logger            *logger.Logger
	Postgres          *database.PostgresDB
	Redis             *database.RedisClient
	Mongo             *database.MongoClient
	Fiber             *fiber.App
	KeycloakValidator *auth.KeycloakValidator
	KeycloakClient    *auth.KeycloakClient
	AuthService       *application.AuthService
	AuthHandler       *interfaces.AuthHandler
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	loggerInstance, err := logger.NewLogger(logger.Config{
		Level:      cfg.Logger.Level,
		Format:     cfg.Logger.Format,
		OutputPath: cfg.Logger.OutputPath,
	})
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Initialize application
	app := &Application{
		Config: cfg,
		Logger: loggerInstance,
	}

	// Setup databases
	if err := app.setupDatabases(); err != nil {
		log.Fatalf("Failed to setup databases: %v", err)
	}

	// Setup authentication
	if err := app.setupAuth(); err != nil {
		log.Fatalf("Failed to setup authentication: %v", err)
	}

	// Setup Fiber app
	app.setupFiber()

	// Setup routes
	app.setupRoutes()

	// Start server
	app.start()
}

func (app *Application) setupDatabases() error {
	var err error

	// Setup PostgreSQL
	app.Postgres, err = database.NewPostgresDB(database.PostgresConfig{
		Host:               app.Config.Database.Host,
		Port:               app.Config.Database.Port,
		User:               app.Config.Database.User,
		Password:           app.Config.Database.Password,
		DBName:             app.Config.Database.DBName,
		SSLMode:            app.Config.Database.SSLMode,
		MaxOpenConnections: app.Config.Database.MaxOpenConnections,
		MaxIdleConnections: app.Config.Database.MaxIdleConnections,
		ConnectionMaxAge:   app.Config.Database.ConnectionMaxAge,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	log.Println("Connected to PostgreSQL")

	// Setup Redis
	app.Redis, err = database.NewRedisClient(database.RedisConfig{
		Host:     app.Config.Redis.Host,
		Port:     app.Config.Redis.Port,
		Password: app.Config.Redis.Password,
		DB:       app.Config.Redis.DB,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	log.Println("Connected to Redis")

	// Setup MongoDB
	app.Mongo, err = database.NewMongoClient(database.MongoConfig{
		URI:      app.Config.MongoDB.URI,
		Database: app.Config.MongoDB.Database,
		Timeout:  app.Config.MongoDB.Timeout,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	log.Println("Connected to MongoDB")

	return nil
}

func (app *Application) setupAuth() error {
	// Load auth config
	authConfig := config.LoadAuthConfig()

	// Initialize Keycloak validator
	app.KeycloakValidator = auth.NewKeycloakValidator(auth.KeycloakConfig{
		URL:      authConfig.Keycloak.URL,
		Realm:    authConfig.Keycloak.Realm,
		ClientID: authConfig.Keycloak.BackendClientID,
		Secret:   authConfig.Keycloak.BackendSecret,
	})

	// Initialize Keycloak client
	app.KeycloakClient = auth.NewKeycloakClient(auth.KeycloakConfig{
		URL:      authConfig.Keycloak.URL,
		Realm:    authConfig.Keycloak.Realm,
		ClientID: authConfig.Keycloak.BackendClientID,
		Secret:   authConfig.Keycloak.BackendSecret,
	})

	// Initialize auth service
	app.AuthService = application.NewAuthService(
		app.KeycloakClient,
		app.KeycloakValidator,
		app.Logger.Desugar(), // Get raw zap.Logger
	)

	// Initialize auth handler
	app.AuthHandler = interfaces.NewAuthHandler(
		app.AuthService,
		app.Logger.Desugar(), // Get raw zap.Logger
	)

	log.Println("Authentication system initialized")
	return nil
}

func (app *Application) setupFiber() {
	// Create Fiber app
	app.Fiber = fiber.New(fiber.Config{
		ReadTimeout:  app.Config.Server.ReadTimeout,
		WriteTimeout: app.Config.Server.WriteTimeout,
		IdleTimeout:  app.Config.Server.IdleTimeout,
		ErrorHandler: middleware.ErrorHandler,
	})

	// Setup middleware
	middleware.SetupMiddleware(app.Fiber)
}

func (app *Application) setupRoutes() {
	// Create auth middleware
	authMiddleware := middleware.AuthMiddleware(app.KeycloakValidator, app.Logger.Desugar())

	// Setup authentication routes
	interfaces.SetupAuthRoutes(app.Fiber, app.AuthHandler, authMiddleware, app.Logger.Desugar())

	// Setup login interface routes (JSON API)
	interfaces.SetupLoginInterfaceRoutes(app.Fiber, app.Logger.Desugar())

	// Setup HTML login pages (optional)
	interfaces.SetupLoginHTMLRoutes(app.Fiber, app.Logger.Desugar())

	// Setup role-based redirect routes
	interfaces.SetupLoginRedirectRoutes(app.Fiber, authMiddleware, app.Logger.Desugar())

	// Health check
	app.Fiber.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"time":    time.Now().UTC(),
			"version": app.Config.App.Version,
		})
	})

	// API v1 routes
	v1 := app.Fiber.Group("/api/v1")

	// Tenant routes (require tenant middleware)
	tenantRoutes := v1.Group("/tenant", middleware.RequireTenant())
	tenantRoutes.Get("/info", func(c *fiber.Ctx) error {
		tenantCtx, _ := middleware.GetTenantFromContext(c)
		return c.JSON(fiber.Map{
			"tenant": tenantCtx,
		})
	})

	// System admin routes
	adminRoutes := v1.Group("/admin")
	adminRoutes.Use(authMiddleware)
	adminRoutes.Use(middleware.RequireSystemAdmin())
	adminRoutes.Get("/tenants", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Tenant listing endpoint",
		})
	})
}

func (app *Application) start() {
	// Create a channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		addr := fmt.Sprintf("%s:%d", app.Config.Server.Host, app.Config.Server.Port)
		log.Printf("Starting server on %s", addr)

		if err := app.Fiber.Listen(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown Fiber server
	if err := app.Fiber.ShutdownWithContext(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Close database connections
	if app.Postgres != nil {
		if err := app.Postgres.Close(); err != nil {
			log.Printf("Failed to close PostgreSQL connection: %v", err)
		}
	}

	if app.Redis != nil {
		if err := app.Redis.Close(); err != nil {
			log.Printf("Failed to close Redis connection: %v", err)
		}
	}

	if app.Mongo != nil {
		if err := app.Mongo.Close(ctx); err != nil {
			log.Printf("Failed to close MongoDB connection: %v", err)
		}
	}

	log.Println("Server exited")
}
