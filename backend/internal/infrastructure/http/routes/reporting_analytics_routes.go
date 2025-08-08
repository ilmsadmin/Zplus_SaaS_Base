package routes

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/ilmsadmin/zplus-saas-base/internal/application/services"
	"github.com/ilmsadmin/zplus-saas-base/internal/infrastructure/http/handlers"
)

// SetupReportingAnalyticsRoutes sets up reporting and analytics routes
func SetupReportingAnalyticsRoutes(
	app *fiber.App,
	reportingService services.ReportingAnalyticsService,
	logger *zap.Logger,
) {
	// Create handler
	handler := handlers.NewReportingAnalyticsHandler(reportingService, logger)

	// API routes group
	api := app.Group("/api")

	// Reports routes
	reports := api.Group("/reports")
	{
		reports.Post("/", handler.CreateReport)               // POST /api/reports
		reports.Get("/", handler.ListReports)                 // GET /api/reports
		reports.Get("/:id", handler.GetReport)                // GET /api/reports/:id
		reports.Put("/:id", handler.UpdateReport)             // PUT /api/reports/:id
		reports.Delete("/:id", handler.DeleteReport)          // DELETE /api/reports/:id
		reports.Get("/:id/download", handler.DownloadReport)  // GET /api/reports/:id/download
		reports.Post("/:id/generate", handler.GenerateReport) // POST /api/reports/:id/generate
	}

	// Analytics routes
	analytics := api.Group("/analytics")
	{
		// Activity tracking
		activity := analytics.Group("/activity")
		{
			activity.Post("/", handler.RecordUserActivity)                          // POST /api/analytics/activity
			activity.Get("/", handler.GetUserActivityMetrics)                       // GET /api/analytics/activity
			activity.Get("/users/:user_id/summary", handler.GetUserActivitySummary) // GET /api/analytics/activity/users/:user_id/summary
		}

		// System metrics
		system := analytics.Group("/system")
		{
			system.Post("/metrics", handler.RecordSystemMetric) // POST /api/analytics/system/metrics
			system.Get("/metrics", handler.GetSystemMetrics)    // GET /api/analytics/system/metrics
			system.Get("/overview", handler.GetSystemOverview)  // GET /api/analytics/system/overview
		}

		// Dashboard
		analytics.Get("/dashboard", handler.GetDashboardStats) // GET /api/analytics/dashboard

		// Health check
		analytics.Get("/health", handler.HealthCheck) // GET /api/analytics/health
	}

	logger.Info("Reporting and Analytics routes configured",
		zap.String("base_path", "/api"),
		zap.Strings("endpoints", []string{
			"POST /api/reports",
			"GET /api/reports",
			"GET /api/reports/:id",
			"PUT /api/reports/:id",
			"DELETE /api/reports/:id",
			"GET /api/reports/:id/download",
			"POST /api/reports/:id/generate",
			"POST /api/analytics/activity",
			"GET /api/analytics/activity",
			"GET /api/analytics/activity/users/:user_id/summary",
			"POST /api/analytics/system/metrics",
			"GET /api/analytics/system/metrics",
			"GET /api/analytics/system/overview",
			"GET /api/analytics/dashboard",
			"GET /api/analytics/health",
		}),
	)
}
