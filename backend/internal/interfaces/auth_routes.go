package interfaces

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ilmsadmin/zplus-saas-base/internal/infrastructure/middleware"
	"go.uber.org/zap"
)

// AuthRoutes sets up authentication routes
func SetupAuthRoutes(app *fiber.App, authHandler *AuthHandler, authMiddleware fiber.Handler, logger *zap.Logger) {
	// Auth routes group
	auth := app.Group("/auth")

	// Public authentication endpoints
	auth.Post("/system-admin/login", authHandler.SystemAdminLogin)
	auth.Post("/tenant-admin/login", authHandler.TenantAdminLogin)
	auth.Post("/user/login", authHandler.UserLogin)
	auth.Post("/refresh", authHandler.RefreshToken)
	auth.Post("/logout", authHandler.Logout)
	auth.Get("/validate", authHandler.ValidateToken)

	// Protected routes (require authentication)
	authProtected := auth.Group("", authMiddleware)
	authProtected.Get("/profile", authHandler.GetProfile)
}

// LoginRedirectRoutes sets up role-based redirect routes
func SetupLoginRedirectRoutes(app *fiber.App, authMiddleware fiber.Handler, logger *zap.Logger) {
	// System Admin Routes (admin.zplus.io)
	systemAdminGroup := app.Group("/admin")
	systemAdminGroup.Use(authMiddleware)
	systemAdminGroup.Use(middleware.RequireSystemAdmin())

	// System admin dashboard redirect
	systemAdminGroup.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/admin/dashboard")
	})

	// System admin dashboard
	systemAdminGroup.Get("/dashboard", func(c *fiber.Ctx) error {
		user, _ := middleware.GetUserFromContext(c)
		return c.JSON(fiber.Map{
			"page":    "system_admin_dashboard",
			"user":    user.PreferredUsername,
			"role":    "system_admin",
			"access":  "full_system",
			"message": "Welcome to System Admin Dashboard",
		})
	})

	// Tenant management routes for system admin
	systemAdminGroup.Get("/tenants", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"page":    "tenant_management",
			"message": "Tenant Management Interface",
		})
	})

	// User management across all tenants
	systemAdminGroup.Get("/users", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"page":    "global_user_management",
			"message": "Global User Management Interface",
		})
	})

	// Tenant Admin Routes (tenant.zplus.io/admin)
	tenantAdminGroup := app.Group("/tenant/:tenant_slug/admin")
	tenantAdminGroup.Use(middleware.RequireTenant())
	tenantAdminGroup.Use(authMiddleware)
	tenantAdminGroup.Use(middleware.RequireTenantAdmin())

	// Tenant admin dashboard redirect
	tenantAdminGroup.Get("/", func(c *fiber.Ctx) error {
		tenantSlug := c.Params("tenant_slug")
		return c.Redirect("/tenant/" + tenantSlug + "/admin/dashboard")
	})

	// Tenant admin dashboard
	tenantAdminGroup.Get("/dashboard", func(c *fiber.Ctx) error {
		user, _ := middleware.GetUserFromContext(c)
		tenantCtx, _ := middleware.GetTenantFromContext(c)

		return c.JSON(fiber.Map{
			"page":    "tenant_admin_dashboard",
			"user":    user.PreferredUsername,
			"role":    "tenant_admin",
			"tenant":  tenantCtx.Subdomain,
			"access":  "tenant_management",
			"message": "Welcome to Tenant Admin Dashboard",
		})
	})

	// Tenant user management
	tenantAdminGroup.Get("/users", func(c *fiber.Ctx) error {
		tenantCtx, _ := middleware.GetTenantFromContext(c)
		return c.JSON(fiber.Map{
			"page":    "tenant_user_management",
			"tenant":  tenantCtx.Subdomain,
			"message": "Tenant User Management Interface",
		})
	})

	// Tenant settings
	tenantAdminGroup.Get("/settings", func(c *fiber.Ctx) error {
		tenantCtx, _ := middleware.GetTenantFromContext(c)
		return c.JSON(fiber.Map{
			"page":    "tenant_settings",
			"tenant":  tenantCtx.Subdomain,
			"message": "Tenant Settings Interface",
		})
	})

	// Custom domain management
	tenantAdminGroup.Get("/domains", func(c *fiber.Ctx) error {
		tenantCtx, _ := middleware.GetTenantFromContext(c)
		return c.JSON(fiber.Map{
			"page":    "domain_management",
			"tenant":  tenantCtx.Subdomain,
			"message": "Custom Domain Management Interface",
		})
	})

	// User Routes (tenant.zplus.io)
	userGroup := app.Group("/tenant/:tenant_slug")
	userGroup.Use(middleware.RequireTenant())
	userGroup.Use(authMiddleware)

	// User dashboard redirect
	userGroup.Get("/", func(c *fiber.Ctx) error {
		tenantSlug := c.Params("tenant_slug")
		return c.Redirect("/tenant/" + tenantSlug + "/dashboard")
	})

	// User dashboard
	userGroup.Get("/dashboard", func(c *fiber.Ctx) error {
		user, _ := middleware.GetUserFromContext(c)
		tenantCtx, _ := middleware.GetTenantFromContext(c)

		return c.JSON(fiber.Map{
			"page":    "user_dashboard",
			"user":    user.PreferredUsername,
			"role":    "user",
			"tenant":  tenantCtx.Subdomain,
			"access":  "tenant_services",
			"message": "Welcome to User Dashboard",
		})
	})

	// User profile
	userGroup.Get("/profile", func(c *fiber.Ctx) error {
		user, _ := middleware.GetUserFromContext(c)
		return c.JSON(fiber.Map{
			"page":    "user_profile",
			"user":    user.PreferredUsername,
			"message": "User Profile Management",
		})
	})

	// Module access routes
	moduleGroup := userGroup.Group("/modules")

	// File management module
	moduleGroup.Get("/files", func(c *fiber.Ctx) error {
		tenantCtx, _ := middleware.GetTenantFromContext(c)
		return c.JSON(fiber.Map{
			"module":  "file_management",
			"tenant":  tenantCtx.Subdomain,
			"message": "File Management Module",
		})
	})

	// POS module
	moduleGroup.Get("/pos", func(c *fiber.Ctx) error {
		tenantCtx, _ := middleware.GetTenantFromContext(c)
		return c.JSON(fiber.Map{
			"module":  "point_of_sale",
			"tenant":  tenantCtx.Subdomain,
			"message": "Point of Sale Module",
		})
	})
}

// Role-based Login Interface Routes
func SetupLoginInterfaceRoutes(app *fiber.App, logger *zap.Logger) {
	// System Admin Login Interface (admin.zplus.io)
	app.Get("/admin/login", func(c *fiber.Ctx) error {
		// Check if already authenticated and redirect
		if accessToken := c.Cookies("access_token"); accessToken != "" {
			// Could validate token here and redirect accordingly
			return c.Redirect("/admin/dashboard")
		}

		return c.JSON(fiber.Map{
			"login_type":   "system_admin",
			"page":         "system_admin_login",
			"client_id":    "zplus-admin-frontend",
			"endpoint":     "/auth/system-admin/login",
			"redirect_url": "/admin/dashboard",
			"title":        "System Administrator Login",
			"description":  "Login to manage the entire Zplus platform",
		})
	})

	// Tenant Admin Login Interface (tenant.zplus.io/admin/login)
	tenantAdminLogin := app.Group("/tenant/:tenant_slug/admin")
	tenantAdminLogin.Use(middleware.RequireTenant())

	tenantAdminLogin.Get("/login", func(c *fiber.Ctx) error {
		// Check if already authenticated and redirect
		if accessToken := c.Cookies("access_token"); accessToken != "" {
			tenantSlug := c.Params("tenant_slug")
			return c.Redirect("/tenant/" + tenantSlug + "/admin/dashboard")
		}

		tenantCtx, _ := middleware.GetTenantFromContext(c)
		tenantSlug := c.Params("tenant_slug")

		return c.JSON(fiber.Map{
			"login_type":   "tenant_admin",
			"page":         "tenant_admin_login",
			"client_id":    "zplus-tenant-frontend",
			"endpoint":     "/auth/tenant-admin/login",
			"redirect_url": "/tenant/" + tenantSlug + "/admin/dashboard",
			"tenant":       tenantCtx.Subdomain,
			"title":        "Tenant Administrator Login",
			"description":  "Login to manage " + tenantCtx.Subdomain + " tenant",
		})
	})

	// User Login Interface (tenant.zplus.io/login)
	userLogin := app.Group("/tenant/:tenant_slug")
	userLogin.Use(middleware.RequireTenant())

	userLogin.Get("/login", func(c *fiber.Ctx) error {
		// Check if already authenticated and redirect
		if accessToken := c.Cookies("access_token"); accessToken != "" {
			tenantSlug := c.Params("tenant_slug")
			return c.Redirect("/tenant/" + tenantSlug + "/dashboard")
		}

		tenantCtx, _ := middleware.GetTenantFromContext(c)
		tenantSlug := c.Params("tenant_slug")

		return c.JSON(fiber.Map{
			"login_type":   "user",
			"page":         "user_login",
			"client_id":    "zplus-tenant-frontend",
			"endpoint":     "/auth/user/login",
			"redirect_url": "/tenant/" + tenantSlug + "/dashboard",
			"tenant":       tenantCtx.Subdomain,
			"title":        "User Login",
			"description":  "Login to access " + tenantCtx.Subdomain + " services",
		})
	})

	// Root redirect based on domain
	app.Get("/", func(c *fiber.Ctx) error {
		host := c.Get("Host")

		// System admin domain
		if host == "admin.zplus.io" || host == "admin.localhost" {
			return c.Redirect("/admin/login")
		}

		// Tenant domain - try to extract tenant
		if tenantCtx, exists := middleware.GetTenantFromContext(c); exists {
			return c.Redirect("/tenant/" + tenantCtx.Subdomain + "/login")
		}

		// Default fallback
		return c.JSON(fiber.Map{
			"message": "Zplus SaaS Platform",
			"status":  "active",
			"host":    host,
		})
	})
}

// Login Interface HTML Routes (if serving HTML directly)
func SetupLoginHTMLRoutes(app *fiber.App, logger *zap.Logger) {
	// System Admin Login Page
	app.Get("/admin/login.html", func(c *fiber.Ctx) error {
		return c.SendString(`
<!DOCTYPE html>
<html>
<head>
    <title>System Admin Login - Zplus</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body { font-family: Arial, sans-serif; max-width: 400px; margin: 100px auto; padding: 20px; }
        .form-group { margin-bottom: 15px; }
        label { display: block; margin-bottom: 5px; }
        input { width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; }
        button { width: 100%; padding: 10px; background: #007bff; color: white; border: none; border-radius: 4px; }
        .error { color: red; margin-top: 10px; }
    </style>
</head>
<body>
    <h2>System Administrator Login</h2>
    <form id="loginForm">
        <div class="form-group">
            <label>Username:</label>
            <input type="text" id="username" name="username" required>
        </div>
        <div class="form-group">
            <label>Password:</label>
            <input type="password" id="password" name="password" required>
        </div>
        <button type="submit">Login</button>
        <div id="error" class="error"></div>
    </form>

    <script>
        document.getElementById('loginForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            const formData = new FormData(e.target);
            const data = Object.fromEntries(formData);
            
            try {
                const response = await fetch('/auth/system-admin/login', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                });
                
                const result = await response.json();
                if (result.success) {
                    window.location.href = result.redirect_url || '/admin/dashboard';
                } else {
                    document.getElementById('error').textContent = result.message || 'Login failed';
                }
            } catch (error) {
                document.getElementById('error').textContent = 'Network error occurred';
            }
        });
    </script>
</body>
</html>
		`)
	})

	// Tenant Admin Login Page
	app.Get("/tenant/:tenant_slug/admin/login.html", func(c *fiber.Ctx) error {
		tenantSlug := c.Params("tenant_slug")
		return c.SendString(`
<!DOCTYPE html>
<html>
<head>
    <title>Tenant Admin Login - ` + tenantSlug + `</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body { font-family: Arial, sans-serif; max-width: 400px; margin: 100px auto; padding: 20px; }
        .form-group { margin-bottom: 15px; }
        label { display: block; margin-bottom: 5px; }
        input { width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; }
        button { width: 100%; padding: 10px; background: #28a745; color: white; border: none; border-radius: 4px; }
        .error { color: red; margin-top: 10px; }
    </style>
</head>
<body>
    <h2>Tenant Administrator Login</h2>
    <p>Login to manage ` + tenantSlug + ` tenant</p>
    <form id="loginForm">
        <div class="form-group">
            <label>Username:</label>
            <input type="text" id="username" name="username" required>
        </div>
        <div class="form-group">
            <label>Password:</label>
            <input type="password" id="password" name="password" required>
        </div>
        <button type="submit">Login</button>
        <div id="error" class="error"></div>
    </form>

    <script>
        document.getElementById('loginForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            const formData = new FormData(e.target);
            const data = Object.fromEntries(formData);
            
            try {
                const response = await fetch('/auth/tenant-admin/login', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                });
                
                const result = await response.json();
                if (result.success) {
                    window.location.href = result.redirect_url || '/tenant/` + tenantSlug + `/admin/dashboard';
                } else {
                    document.getElementById('error').textContent = result.message || 'Login failed';
                }
            } catch (error) {
                document.getElementById('error').textContent = 'Network error occurred';
            }
        });
    </script>
</body>
</html>
		`)
	})

	// User Login Page
	app.Get("/tenant/:tenant_slug/login.html", func(c *fiber.Ctx) error {
		tenantSlug := c.Params("tenant_slug")
		return c.SendString(`
<!DOCTYPE html>
<html>
<head>
    <title>User Login - ` + tenantSlug + `</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body { font-family: Arial, sans-serif; max-width: 400px; margin: 100px auto; padding: 20px; }
        .form-group { margin-bottom: 15px; }
        label { display: block; margin-bottom: 5px; }
        input { width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; }
        button { width: 100%; padding: 10px; background: #17a2b8; color: white; border: none; border-radius: 4px; }
        .error { color: red; margin-top: 10px; }
    </style>
</head>
<body>
    <h2>User Login</h2>
    <p>Login to access ` + tenantSlug + ` services</p>
    <form id="loginForm">
        <div class="form-group">
            <label>Username:</label>
            <input type="text" id="username" name="username" required>
        </div>
        <div class="form-group">
            <label>Password:</label>
            <input type="password" id="password" name="password" required>
        </div>
        <button type="submit">Login</button>
        <div id="error" class="error"></div>
    </form>

    <script>
        document.getElementById('loginForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            const formData = new FormData(e.target);
            const data = Object.fromEntries(formData);
            
            try {
                const response = await fetch('/auth/user/login', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                });
                
                const result = await response.json();
                if (result.success) {
                    window.location.href = result.redirect_url || '/tenant/` + tenantSlug + `/dashboard';
                } else {
                    document.getElementById('error').textContent = result.message || 'Login failed';
                }
            } catch (error) {
                document.getElementById('error').textContent = 'Network error occurred';
            }
        });
    </script>
</body>
</html>
		`)
	})
}
