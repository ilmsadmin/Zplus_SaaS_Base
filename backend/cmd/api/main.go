package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	_ "github.com/lib/pq"
)

type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
	TenantID  string `json:"tenant_id"`
	Status    string `json:"status"`
}

type Tenant struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Slug   string `json:"slug"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

type Role struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsSystem    bool   `json:"is_system"`
	TenantID    string `json:"tenant_id"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type CreateUserRequest struct {
	Email     string `json:"email"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"password"`
	Role      string `json:"role"`
	TenantID  string `json:"tenant_id"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

var db *sql.DB

func main() {
	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres123@localhost:5432/zplus?sslmode=disable"
	}

	var err error
	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Connected to database successfully")

	// Create Fiber app
	app := fiber.New(fiber.Config{
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	// Middleware
	app.Use(cors.New())
	app.Use(logger.New())

	// Routes
	setupRoutes(app)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}

func setupRoutes(app *fiber.App) {
	api := app.Group("/api/v1")

	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "API is running",
		})
	})

	// Auth routes
	auth := api.Group("/auth")
	auth.Post("/login", login)

	// User routes
	users := api.Group("/users")
	users.Get("/", getUsers)
	users.Get("/:id", getUser)
	users.Post("/", createUser)
	users.Put("/:id", updateUser)
	users.Put("/:id/password", changeUserPassword)
	users.Delete("/:id", deleteUser)

	// Tenant routes
	tenants := api.Group("/tenants")
	tenants.Get("/", getTenants)
	tenants.Get("/:id", getTenant)

	// Role routes
	roles := api.Group("/roles")
	roles.Get("/", getRoles)
	roles.Post("/", createRole)
}

// Auth handlers
func login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Check user exists and verify password
	var user User
	var passwordHash string
	err := db.QueryRow(`
		SELECT id, email, username, first_name, last_name, role, 
		       COALESCE(tenant_id, '') as tenant_id, status, password_hash
		FROM users WHERE email = $1 AND status = 'active'`,
		req.Email).Scan(
		&user.ID, &user.Email, &user.Username, &user.FirstName,
		&user.LastName, &user.Role, &user.TenantID, &user.Status, &passwordHash)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	// Verify password using database function
	var validPassword bool
	err = db.QueryRow("SELECT verify_password($1, $2)", req.Password, passwordHash).Scan(&validPassword)
	if err != nil || !validPassword {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Generate simple token (in production, use JWT)
	token := fmt.Sprintf("token_%s_%d", user.ID, 1234567890)

	return c.JSON(LoginResponse{
		Token: token,
		User:  user,
	})
}

// User handlers
func getUsers(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 10)
	offset := c.QueryInt("offset", 0)

	rows, err := db.Query(`
		SELECT id, email, username, first_name, last_name, role, 
		       COALESCE(tenant_id, '') as tenant_id, status 
		FROM users 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Email, &user.Username, &user.FirstName,
			&user.LastName, &user.Role, &user.TenantID, &user.Status)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Scan error"})
		}
		users = append(users, user)
	}

	return c.JSON(fiber.Map{"users": users})
}

func getUser(c *fiber.Ctx) error {
	id := c.Params("id")

	var user User
	err := db.QueryRow(`
		SELECT id, email, username, first_name, last_name, role, 
		       COALESCE(tenant_id, '') as tenant_id, status 
		FROM users WHERE id = $1`, id).Scan(
		&user.ID, &user.Email, &user.Username, &user.FirstName,
		&user.LastName, &user.Role, &user.TenantID, &user.Status)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "User not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	return c.JSON(user)
}

func createUser(c *fiber.Ctx) error {
	var req CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Hash password
	var passwordHash string
	err := db.QueryRow("SELECT hash_password($1)", req.Password).Scan(&passwordHash)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	var userID string
	err = db.QueryRow(`
		INSERT INTO users (email, username, first_name, last_name, role, tenant_id, status, password_hash, password_salt)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), 'active', $7, 'zplus_salt_2025')
		RETURNING id`,
		req.Email, req.Username, req.FirstName, req.LastName, req.Role, req.TenantID, passwordHash).Scan(&userID)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create user"})
	}

	return c.Status(201).JSON(fiber.Map{"id": userID, "message": "User created successfully"})
}

func updateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	var req CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	_, err := db.Exec(`
		UPDATE users 
		SET email = $1, username = $2, first_name = $3, last_name = $4, 
		    role = $5, tenant_id = NULLIF($6, ''), updated_at = CURRENT_TIMESTAMP
		WHERE id = $7`,
		req.Email, req.Username, req.FirstName, req.LastName, req.Role, req.TenantID, id)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update user"})
	}

	return c.JSON(fiber.Map{"message": "User updated successfully"})
}

func changeUserPassword(c *fiber.Ctx) error {
	id := c.Params("id")
	var req ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate input
	if req.CurrentPassword == "" || req.NewPassword == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Current password and new password are required"})
	}

	if len(req.NewPassword) < 6 {
		return c.Status(400).JSON(fiber.Map{"error": "New password must be at least 6 characters"})
	}

	// Get current password hash
	var currentPasswordHash string
	err := db.QueryRow("SELECT password_hash FROM users WHERE id = $1", id).Scan(&currentPasswordHash)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}

	// Verify current password
	var validPassword bool
	err = db.QueryRow("SELECT verify_password($1, $2)", req.CurrentPassword, currentPasswordHash).Scan(&validPassword)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to verify current password"})
	}

	if !validPassword {
		return c.Status(400).JSON(fiber.Map{"error": "Current password is incorrect"})
	}

	// Hash new password
	var newPasswordHash string
	err = db.QueryRow("SELECT hash_password($1)", req.NewPassword).Scan(&newPasswordHash)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to hash new password"})
	}

	// Update password
	_, err = db.Exec("UPDATE users SET password_hash = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2", newPasswordHash, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update password"})
	}

	return c.JSON(fiber.Map{"message": "Password changed successfully"})
}

func deleteUser(c *fiber.Ctx) error {
	id := c.Params("id")

	_, err := db.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete user"})
	}

	return c.JSON(fiber.Map{"message": "User deleted successfully"})
}

// Tenant handlers
func getTenants(c *fiber.Ctx) error {
	rows, err := db.Query(`
		SELECT id, name, slug, email, status 
		FROM tenants 
		ORDER BY created_at DESC`)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	defer rows.Close()

	var tenants []Tenant
	for rows.Next() {
		var tenant Tenant
		err := rows.Scan(&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.Email, &tenant.Status)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Scan error"})
		}
		tenants = append(tenants, tenant)
	}

	return c.JSON(fiber.Map{"tenants": tenants})
}

func getTenant(c *fiber.Ctx) error {
	id := c.Params("id")

	var tenant Tenant
	err := db.QueryRow(`
		SELECT id, name, slug, email, status 
		FROM tenants WHERE id = $1`, id).Scan(
		&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.Email, &tenant.Status)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "Tenant not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	return c.JSON(tenant)
}

// Role handlers
func getRoles(c *fiber.Ctx) error {
	tenantID := c.Query("tenant_id")

	var query string
	var args []interface{}

	if tenantID != "" {
		query = `SELECT id, name, description, is_system, COALESCE(tenant_id, '') as tenant_id 
		         FROM roles WHERE tenant_id = $1 OR is_system = true ORDER BY name`
		args = append(args, tenantID)
	} else {
		query = `SELECT id, name, description, is_system, COALESCE(tenant_id, '') as tenant_id 
		         FROM roles ORDER BY name`
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.IsSystem, &role.TenantID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Scan error"})
		}
		roles = append(roles, role)
	}

	return c.JSON(fiber.Map{"roles": roles})
}

func createRole(c *fiber.Ctx) error {
	var role Role
	if err := c.BodyParser(&role); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	var roleID string
	err := db.QueryRow(`
		INSERT INTO roles (name, description, is_system, tenant_id)
		VALUES ($1, $2, $3, NULLIF($4, ''))
		RETURNING id`,
		role.Name, role.Description, role.IsSystem, role.TenantID).Scan(&roleID)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create role"})
	}

	return c.Status(201).JSON(fiber.Map{"id": roleID, "message": "Role created successfully"})
}
