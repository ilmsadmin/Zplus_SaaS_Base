package application

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// DNSValidationService handles DNS validation for custom domains
type DNSValidationService struct {
	logger     *zap.Logger
	httpClient *http.Client
}

// NewDNSValidationService creates a new DNS validation service
func NewDNSValidationService(logger *zap.Logger) *DNSValidationService {
	// Configure HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				d := &net.Dialer{
					Timeout: 5 * time.Second,
				}
				return d.DialContext(ctx, network, addr)
			},
		},
	}

	return &DNSValidationService{
		logger:     logger,
		httpClient: httpClient,
	}
}

// ValidationResult contains the result of domain validation
type ValidationResult struct {
	Valid        bool              `json:"valid"`
	Method       string            `json:"method"`
	Domain       string            `json:"domain"`
	Token        string            `json:"token"`
	ErrorMessage string            `json:"error_message,omitempty"`
	Details      map[string]string `json:"details,omitempty"`
	CheckedAt    time.Time         `json:"checked_at"`
}

// DNSRecordInfo contains information about DNS records
type DNSRecordInfo struct {
	Type     string   `json:"type"`
	Name     string   `json:"name"`
	Values   []string `json:"values"`
	TTL      int      `json:"ttl,omitempty"`
	Priority int      `json:"priority,omitempty"`
}

// ValidateDNSRecord validates DNS TXT record for domain ownership
func (s *DNSValidationService) ValidateDNSRecord(ctx context.Context, domain, expectedToken string) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:     false,
		Method:    "dns",
		Domain:    domain,
		Token:     expectedToken,
		CheckedAt: time.Now(),
		Details:   make(map[string]string),
	}

	// Construct the verification record name
	recordName := "_zplus-verify." + domain

	s.logger.Info("Starting DNS validation",
		zap.String("domain", domain),
		zap.String("record_name", recordName),
		zap.String("expected_token", expectedToken),
	)

	// Perform DNS lookup with timeout
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// Use a custom resolver for more control
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 5 * time.Second,
			}
			return d.DialContext(ctx, network, "8.8.8.8:53") // Use Google DNS
		},
	}

	txtRecords, err := resolver.LookupTXT(ctx, recordName)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("DNS lookup failed: %v", err)
		result.Details["dns_error"] = err.Error()

		s.logger.Warn("DNS lookup failed",
			zap.String("domain", domain),
			zap.String("record_name", recordName),
			zap.Error(err),
		)

		return result, nil // Return result with error, not an error itself
	}

	result.Details["records_found"] = fmt.Sprintf("%d", len(txtRecords))
	result.Details["all_records"] = strings.Join(txtRecords, ", ")

	// Check if any of the TXT records match our expected token
	for _, record := range txtRecords {
		trimmedRecord := strings.TrimSpace(record)

		s.logger.Debug("Checking TXT record",
			zap.String("domain", domain),
			zap.String("record", trimmedRecord),
			zap.String("expected", expectedToken),
		)

		if trimmedRecord == expectedToken {
			result.Valid = true
			result.Details["matching_record"] = trimmedRecord

			s.logger.Info("DNS validation successful",
				zap.String("domain", domain),
				zap.String("record_name", recordName),
				zap.String("token", expectedToken),
			)

			return result, nil
		}
	}

	// If we get here, no matching record was found
	result.ErrorMessage = "Verification token not found in DNS records"
	result.Details["expected_token"] = expectedToken

	s.logger.Warn("DNS validation failed - token not found",
		zap.String("domain", domain),
		zap.String("record_name", recordName),
		zap.String("expected_token", expectedToken),
		zap.Strings("found_records", txtRecords),
	)

	return result, nil
}

// ValidateHTTPRecord validates HTTP file for domain ownership
func (s *DNSValidationService) ValidateHTTPRecord(ctx context.Context, domain, expectedToken string) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:     false,
		Method:    "http",
		Domain:    domain,
		Token:     expectedToken,
		CheckedAt: time.Now(),
		Details:   make(map[string]string),
	}

	// Construct the verification URL
	verificationPath := "/.well-known/zplus-verification"
	urls := []string{
		fmt.Sprintf("http://%s%s", domain, verificationPath),
		fmt.Sprintf("https://%s%s", domain, verificationPath),
	}

	s.logger.Info("Starting HTTP validation",
		zap.String("domain", domain),
		zap.Strings("urls", urls),
		zap.String("expected_token", expectedToken),
	)

	// Try both HTTP and HTTPS
	for _, url := range urls {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			cancel()
			continue
		}

		req.Header.Set("User-Agent", "Zplus-Domain-Validator/1.0")

		resp, err := s.httpClient.Do(req)
		if err != nil {
			result.Details[fmt.Sprintf("error_%s", strings.Split(url, "://")[0])] = err.Error()
			cancel()
			continue
		}

		if resp.StatusCode != http.StatusOK {
			result.Details[fmt.Sprintf("status_%s", strings.Split(url, "://")[0])] = fmt.Sprintf("%d", resp.StatusCode)
			resp.Body.Close()
			cancel()
			continue
		}

		// Read the response body
		body := make([]byte, 1024) // Limit to 1KB
		n, err := resp.Body.Read(body)
		resp.Body.Close()
		cancel()

		if err != nil && n == 0 {
			result.Details[fmt.Sprintf("read_error_%s", strings.Split(url, "://")[0])] = err.Error()
			continue
		}

		content := strings.TrimSpace(string(body[:n]))
		result.Details[fmt.Sprintf("content_%s", strings.Split(url, "://")[0])] = content

		s.logger.Debug("HTTP validation response",
			zap.String("url", url),
			zap.String("content", content),
			zap.String("expected", expectedToken),
		)

		if content == expectedToken {
			result.Valid = true
			result.Details["verified_url"] = url
			result.Details["matching_content"] = content

			s.logger.Info("HTTP validation successful",
				zap.String("domain", domain),
				zap.String("url", url),
				zap.String("token", expectedToken),
			)

			return result, nil
		}
	}

	// If we get here, validation failed
	result.ErrorMessage = "Verification token not found via HTTP"
	result.Details["expected_token"] = expectedToken

	s.logger.Warn("HTTP validation failed",
		zap.String("domain", domain),
		zap.String("expected_token", expectedToken),
	)

	return result, nil
}

// CheckDomainReachability checks if a domain is reachable
func (s *DNSValidationService) CheckDomainReachability(ctx context.Context, domain string) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:     false,
		Method:    "reachability",
		Domain:    domain,
		CheckedAt: time.Now(),
		Details:   make(map[string]string),
	}

	s.logger.Info("Checking domain reachability", zap.String("domain", domain))

	// Check DNS resolution
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: 5 * time.Second}
			return d.DialContext(ctx, network, "8.8.8.8:53")
		},
	}

	// Check A records
	aRecords, err := resolver.LookupIPAddr(ctx, domain)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("DNS resolution failed: %v", err)
		result.Details["dns_error"] = err.Error()
		return result, nil
	}

	ips := make([]string, len(aRecords))
	for i, ip := range aRecords {
		ips[i] = ip.IP.String()
	}
	result.Details["ip_addresses"] = strings.Join(ips, ", ")
	result.Details["ip_count"] = fmt.Sprintf("%d", len(ips))

	// Check HTTP connectivity
	urls := []string{
		fmt.Sprintf("http://%s", domain),
		fmt.Sprintf("https://%s", domain),
	}

	for _, url := range urls {
		protocol := strings.Split(url, "://")[0]

		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
		if err != nil {
			cancel()
			continue
		}

		req.Header.Set("User-Agent", "Zplus-Domain-Validator/1.0")

		resp, err := s.httpClient.Do(req)
		if err != nil {
			result.Details[fmt.Sprintf("%s_error", protocol)] = err.Error()
			cancel()
			continue
		}

		result.Details[fmt.Sprintf("%s_status", protocol)] = fmt.Sprintf("%d", resp.StatusCode)
		result.Details[fmt.Sprintf("%s_reachable", protocol)] = "true"
		resp.Body.Close()
		cancel()

		if resp.StatusCode < 500 { // Consider 4xx as reachable (domain responds)
			result.Valid = true
		}
	}

	if result.Valid {
		s.logger.Info("Domain reachability check passed", zap.String("domain", domain))
	} else {
		s.logger.Warn("Domain reachability check failed", zap.String("domain", domain))
	}

	return result, nil
}

// GetDNSRecords retrieves various DNS records for a domain
func (s *DNSValidationService) GetDNSRecords(ctx context.Context, domain string) (map[string][]DNSRecordInfo, error) {
	records := make(map[string][]DNSRecordInfo)

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: 5 * time.Second}
			return d.DialContext(ctx, network, "8.8.8.8:53")
		},
	}

	// Get A records
	if aRecords, err := resolver.LookupIPAddr(ctx, domain); err == nil {
		var aInfo []DNSRecordInfo
		for _, record := range aRecords {
			aInfo = append(aInfo, DNSRecordInfo{
				Type:   "A",
				Name:   domain,
				Values: []string{record.IP.String()},
			})
		}
		if len(aInfo) > 0 {
			records["A"] = aInfo
		}
	}

	// Get CNAME records
	if cname, err := resolver.LookupCNAME(ctx, domain); err == nil && cname != domain+"." {
		records["CNAME"] = []DNSRecordInfo{{
			Type:   "CNAME",
			Name:   domain,
			Values: []string{strings.TrimSuffix(cname, ".")},
		}}
	}

	// Get TXT records
	if txtRecords, err := resolver.LookupTXT(ctx, domain); err == nil && len(txtRecords) > 0 {
		records["TXT"] = []DNSRecordInfo{{
			Type:   "TXT",
			Name:   domain,
			Values: txtRecords,
		}}
	}

	// Get MX records
	if mxRecords, err := resolver.LookupMX(ctx, domain); err == nil && len(mxRecords) > 0 {
		var mxInfo []DNSRecordInfo
		for _, record := range mxRecords {
			mxInfo = append(mxInfo, DNSRecordInfo{
				Type:     "MX",
				Name:     domain,
				Values:   []string{strings.TrimSuffix(record.Host, ".")},
				Priority: int(record.Pref),
			})
		}
		records["MX"] = mxInfo
	}

	s.logger.Debug("Retrieved DNS records",
		zap.String("domain", domain),
		zap.Int("record_types", len(records)),
	)

	return records, nil
}

// ValidateDomainConfiguration performs comprehensive domain validation
func (s *DNSValidationService) ValidateDomainConfiguration(ctx context.Context, domain, token, method string) (*ValidationResult, error) {
	s.logger.Info("Starting comprehensive domain validation",
		zap.String("domain", domain),
		zap.String("method", method),
	)

	var result *ValidationResult
	var err error

	switch method {
	case "dns":
		result, err = s.ValidateDNSRecord(ctx, domain, token)
	case "http":
		result, err = s.ValidateHTTPRecord(ctx, domain, token)
	default:
		return nil, fmt.Errorf("unsupported validation method: %s", method)
	}

	if err != nil {
		return nil, err
	}

	// If primary validation passed, also check reachability
	if result.Valid {
		reachabilityResult, _ := s.CheckDomainReachability(ctx, domain)
		if reachabilityResult != nil {
			// Merge reachability details
			for key, value := range reachabilityResult.Details {
				result.Details["reachability_"+key] = value
			}
			result.Details["reachability_check"] = fmt.Sprintf("%t", reachabilityResult.Valid)
		}
	}

	return result, nil
}
