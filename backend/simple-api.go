package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// Response structures
type HealthResponse struct {
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

type DomainResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type Domain struct {
	ID                 string    `json:"id"`
	TenantID           string    `json:"tenant_id"`
	Domain             string    `json:"domain"`
	IsCustom           bool      `json:"is_custom"`
	Verified           bool      `json:"verified"`
	SSLEnabled         bool      `json:"ssl_enabled"`
	VerificationToken  string    `json:"verification_token,omitempty"`
	VerificationMethod string    `json:"verification_method"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"created_at"`
}

func main() {
	log.Println("Starting Zplus SaaS Backend Test Server...")

	// Set up routes
	http.HandleFunc("/health", corsHandler(healthHandler))
	http.HandleFunc("/api/v1/health", corsHandler(healthHandler))
	http.HandleFunc("/api/v1/", corsHandler(apiHandler))

	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("Server starting on http://localhost:8080")
	log.Printf("Health endpoint: http://localhost:8080/health")
	log.Printf("API base: http://localhost:8080/api/v1")

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func corsHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	method := r.Method

	// Parse tenant ID and domain ID from path
	parts := strings.Split(strings.TrimPrefix(path, "/api/v1/"), "/")

	switch {
	case strings.HasPrefix(path, "/api/v1/tenants/") && strings.Contains(path, "/domains"):
		handleDomainRoutes(w, r, parts)
	case strings.HasPrefix(path, "/api/v1/domains/"):
		handlePublicDomainRoutes(w, r, parts)
	case strings.HasPrefix(path, "/api/v1/admin/domains") && method == "GET":
		getAdminDomainsHandler(w, r)
	default:
		http.NotFound(w, r)
	}
}

func handleDomainRoutes(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) < 3 {
		http.NotFound(w, r)
		return
	}

	tenantID := parts[1] // tenants/{tenant_id}/domains
	method := r.Method

	switch method {
	case "GET":
		if len(parts) == 3 { // GET /tenants/{tenant_id}/domains
			getDomainsHandler(w, r, tenantID)
		} else if len(parts) == 5 && parts[4] == "instructions" { // GET /tenants/{tenant_id}/domains/{id}/instructions
			getDomainInstructionsHandler(w, r, tenantID, parts[3])
		}
	case "POST":
		if len(parts) == 3 { // POST /tenants/{tenant_id}/domains
			addDomainHandler(w, r, tenantID)
		} else if len(parts) == 5 && parts[4] == "verify" { // POST /tenants/{tenant_id}/domains/{id}/verify
			verifyDomainHandler(w, r, tenantID, parts[3])
		}
	case "DELETE":
		if len(parts) == 4 { // DELETE /tenants/{tenant_id}/domains/{id}
			deleteDomainHandler(w, r, tenantID, parts[3])
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(DomainResponse{
			Success: false,
			Message: "Method not allowed",
		})
	}
}

func handlePublicDomainRoutes(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) < 2 {
		http.NotFound(w, r)
		return
	}

	domain := parts[1] // domains/{domain}/...

	if len(parts) == 3 {
		switch parts[2] {
		case "status":
			getDomainStatusHandler(w, r, domain)
		case "metrics":
			getDomainMetricsHandler(w, r, domain)
		default:
			http.NotFound(w, r)
		}
	} else {
		http.NotFound(w, r)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "healthy",
		Message:   "Zplus SaaS Backend Test Server is running",
		Timestamp: time.Now(),
		Version:   "1.0.0-test",
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func getDomainsHandler(w http.ResponseWriter, r *http.Request, tenantID string) {
	// Mock domain data
	domains := []Domain{
		{
			ID:                 "550e8400-e29b-41d4-a716-446655440001",
			TenantID:           tenantID,
			Domain:             fmt.Sprintf("%s.zplus.io", tenantID),
			IsCustom:           false,
			Verified:           true,
			SSLEnabled:         true,
			VerificationMethod: "dns",
			Status:             "active",
			CreatedAt:          time.Now().Add(-24 * time.Hour),
		},
	}

	if tenantID == "acme" {
		domains = append(domains, Domain{
			ID:                 "550e8400-e29b-41d4-a716-446655440002",
			TenantID:           tenantID,
			Domain:             "app.acme.com",
			IsCustom:           true,
			Verified:           false,
			SSLEnabled:         false,
			VerificationToken:  "zplus-verify-abc123def456",
			VerificationMethod: "dns",
			Status:             "pending_verification",
			CreatedAt:          time.Now().Add(-1 * time.Hour),
		})
	}

	response := DomainResponse{
		Success: true,
		Message: "Domains retrieved successfully",
		Data:    domains,
	}

	json.NewEncoder(w).Encode(response)
}

func addDomainHandler(w http.ResponseWriter, r *http.Request, tenantID string) {
	var req struct {
		Domain             string `json:"domain"`
		VerificationMethod string `json:"verification_method"`
		AutoSSL            bool   `json:"auto_ssl"`
		Priority           int    `json:"priority"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := DomainResponse{
		Success: true,
		Message: "Domain added successfully",
		Data: map[string]interface{}{
			"domain":              req.Domain,
			"verification_token":  "zplus-verify-abc123def456",
			"verification_method": req.VerificationMethod,
			"dns_record": map[string]interface{}{
				"type":  "TXT",
				"name":  "_zplus-verify." + req.Domain,
				"value": "zplus-verify-abc123def456",
				"ttl":   300,
			},
			"instructions": "Add a TXT record to your DNS with the provided values",
			"expires_at":   time.Now().Add(24 * time.Hour),
		},
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func verifyDomainHandler(w http.ResponseWriter, r *http.Request, tenantID, domainID string) {
	response := DomainResponse{
		Success: true,
		Message: "Domain verified successfully",
		Data: map[string]interface{}{
			"domain":        "app.acme.com",
			"verified":      true,
			"ssl_enabled":   false,
			"status":        "active",
			"last_checked":  time.Now(),
			"health_status": "healthy",
		},
	}

	json.NewEncoder(w).Encode(response)
}

func getDomainInstructionsHandler(w http.ResponseWriter, r *http.Request, tenantID, domainID string) {
	response := DomainResponse{
		Success: true,
		Message: "Domain verification instructions",
		Data: map[string]interface{}{
			"verification_method": "dns",
			"dns_record": map[string]interface{}{
				"type":  "TXT",
				"name":  "_zplus-verify.app.acme.com",
				"value": "zplus-verify-abc123def456",
				"ttl":   300,
			},
			"instructions": []string{
				"1. Go to your DNS provider (e.g., Cloudflare, Namecheap, GoDaddy)",
				"2. Add a new TXT record with the provided name and value",
				"3. Wait for DNS propagation (usually 5-15 minutes)",
				"4. Click 'Verify Domain' to complete the process",
			},
			"expires_at": time.Now().Add(24 * time.Hour),
		},
	}

	json.NewEncoder(w).Encode(response)
}

func deleteDomainHandler(w http.ResponseWriter, r *http.Request, tenantID, domainID string) {
	response := DomainResponse{
		Success: true,
		Message: "Domain deleted successfully",
	}

	json.NewEncoder(w).Encode(response)
}

func getDomainStatusHandler(w http.ResponseWriter, r *http.Request, domain string) {
	response := DomainResponse{
		Success: true,
		Message: "Domain status retrieved",
		Data: map[string]interface{}{
			"domain":        domain,
			"verified":      true,
			"ssl_enabled":   true,
			"status":        "active",
			"health_status": "healthy",
			"last_checked":  time.Now(),
			"response_time": "120ms",
		},
	}

	json.NewEncoder(w).Encode(response)
}

func getDomainMetricsHandler(w http.ResponseWriter, r *http.Request, domain string) {
	response := DomainResponse{
		Success: true,
		Message: "Domain metrics retrieved",
		Data: map[string]interface{}{
			"domain": domain,
			"metrics": map[string]interface{}{
				"requests": map[string]interface{}{
					"total": 12500,
					"2xx":   11875,
					"4xx":   500,
					"5xx":   125,
				},
				"response_time": map[string]interface{}{
					"avg": 120,
					"p50": 100,
					"p95": 180,
					"p99": 250,
				},
				"ssl": map[string]interface{}{
					"enabled":    true,
					"expires_at": time.Now().Add(30 * 24 * time.Hour),
					"auto_renew": true,
					"issuer":     "Let's Encrypt",
				},
			},
		},
	}

	json.NewEncoder(w).Encode(response)
}

func getAdminDomainsHandler(w http.ResponseWriter, r *http.Request) {
	domains := []Domain{
		{
			ID:         "550e8400-e29b-41d4-a716-446655440001",
			TenantID:   "admin",
			Domain:     "admin.zplus.io",
			IsCustom:   false,
			Verified:   true,
			SSLEnabled: true,
			Status:     "active",
			CreatedAt:  time.Now().Add(-30 * 24 * time.Hour),
		},
		{
			ID:         "550e8400-e29b-41d4-a716-446655440002",
			TenantID:   "acme",
			Domain:     "acme.zplus.io",
			IsCustom:   false,
			Verified:   true,
			SSLEnabled: true,
			Status:     "active",
			CreatedAt:  time.Now().Add(-7 * 24 * time.Hour),
		},
		{
			ID:         "550e8400-e29b-41d4-a716-446655440003",
			TenantID:   "acme",
			Domain:     "app.acme.com",
			IsCustom:   true,
			Verified:   false,
			SSLEnabled: false,
			Status:     "pending_verification",
			CreatedAt:  time.Now().Add(-1 * time.Hour),
		},
	}

	response := DomainResponse{
		Success: true,
		Message: "All domains retrieved successfully",
		Data:    domains,
	}

	json.NewEncoder(w).Encode(response)
}
