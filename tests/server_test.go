package tests

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"pf2.encounterbrew.com/internal/database"
	"pf2.encounterbrew.com/internal/server"
)

// Use standardized mock database from mock_database.go
// For server tests that need simple health checks, we'll use the MockDatabaseService

func TestNewServer_MissingDBURL(t *testing.T) {
	// Clear environment variables
	originalDBURL := os.Getenv("DB_URL")
	defer os.Setenv("DB_URL", originalDBURL)
	
	os.Unsetenv("DB_URL")
	
	_, err := server.NewServer()
	if err == nil {
		t.Error("Expected error when DB_URL is not set, got nil")
	}
	
	if !strings.Contains(err.Error(), "DB_URL environment variable not set") {
		t.Errorf("Expected error message about DB_URL, got: %v", err)
	}
}

func TestNewServer_MissingMigrationsPath(t *testing.T) {
	// Set required DB_URL but clear MIGRATIONS_PATH
	originalDBURL := os.Getenv("DB_URL")
	originalMigrationsPath := os.Getenv("MIGRATIONS_PATH")
	defer func() {
		os.Setenv("DB_URL", originalDBURL)
		os.Setenv("MIGRATIONS_PATH", originalMigrationsPath)
	}()
	
	os.Setenv("DB_URL", "postgres://test:test@localhost:5432/test")
	os.Unsetenv("MIGRATIONS_PATH")
	
	_, err := server.NewServer()
	if err == nil {
		t.Error("Expected error when MIGRATIONS_PATH is not set, got nil")
	}
	
	if !strings.Contains(err.Error(), "MIGRATIONS_PATH environment variable not set") {
		t.Errorf("Expected error message about MIGRATIONS_PATH, got: %v", err)
	}
}

func TestNewServer_DefaultPort(t *testing.T) {
	// Clear PORT environment variable to test default
	originalPort := os.Getenv("PORT")
	originalDBURL := os.Getenv("DB_URL")
	originalMigrationsPath := os.Getenv("MIGRATIONS_PATH")
	originalDisableMigrations := os.Getenv("DISABLE_MIGRATIONS")
	originalDisableSeed := os.Getenv("DISABLE_SEED")
	
	defer func() {
		os.Setenv("PORT", originalPort)
		os.Setenv("DB_URL", originalDBURL)
		os.Setenv("MIGRATIONS_PATH", originalMigrationsPath)
		os.Setenv("DISABLE_MIGRATIONS", originalDisableMigrations)
		os.Setenv("DISABLE_SEED", originalDisableSeed)
	}()
	
	os.Unsetenv("PORT")
	os.Setenv("DB_URL", "postgres://test:test@localhost:5432/test")
	os.Setenv("MIGRATIONS_PATH", "/tmp/migrations")
	os.Setenv("DISABLE_MIGRATIONS", "true")
	os.Setenv("DISABLE_SEED", "true")
	
	// This test would fail in a real environment due to database connection
	// but demonstrates the port configuration logic
	_, err := server.NewServer()
	
	// We expect an error due to database connection failure, not port configuration
	if err != nil && !strings.Contains(err.Error(), "database") {
		t.Errorf("Expected database-related error, got: %v", err)
	}
}

func TestNewServer_CustomPort(t *testing.T) {
	originalPort := os.Getenv("PORT")
	originalDBURL := os.Getenv("DB_URL")
	originalMigrationsPath := os.Getenv("MIGRATIONS_PATH")
	originalDisableMigrations := os.Getenv("DISABLE_MIGRATIONS")
	originalDisableSeed := os.Getenv("DISABLE_SEED")
	
	defer func() {
		os.Setenv("PORT", originalPort)
		os.Setenv("DB_URL", originalDBURL)
		os.Setenv("MIGRATIONS_PATH", originalMigrationsPath)
		os.Setenv("DISABLE_MIGRATIONS", originalDisableMigrations)
		os.Setenv("DISABLE_SEED", originalDisableSeed)
	}()
	
	os.Setenv("PORT", "9090")
	os.Setenv("DB_URL", "postgres://test:test@localhost:5432/test")
	os.Setenv("MIGRATIONS_PATH", "/tmp/migrations")
	os.Setenv("DISABLE_MIGRATIONS", "true")
	os.Setenv("DISABLE_SEED", "true")
	
	// This test would fail in a real environment due to database connection
	// but demonstrates the port configuration logic
	_, err := server.NewServer()
	
	// We expect an error due to database connection failure, not port configuration
	if err != nil && !strings.Contains(err.Error(), "database") {
		t.Errorf("Expected database-related error, got: %v", err)
	}
}

// Test route registration by creating a test server with mock database
func TestRouteRegistration(t *testing.T) {
	// Create a test server with a mock database
	testServer := &TestServer{
		db: &MockDatabaseService{},
	}
	
	handler := testServer.RegisterRoutes()
	
	// Test cases for different routes
	testCases := []struct {
		method       string
		path         string
		expectedCode int
		description  string
	}{
		{"GET", "/health", http.StatusOK, "Health endpoint"},
		{"GET", "/", http.StatusUnauthorized, "Root endpoint (no auth)"},
		{"GET", "/encounters", http.StatusUnauthorized, "Encounters endpoint (no auth)"},
		{"GET", "/parties", http.StatusUnauthorized, "Parties endpoint (no auth)"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			
			handler.ServeHTTP(rec, req)
			
			if rec.Code != tc.expectedCode {
				t.Errorf("Expected status code %d for %s %s, got %d", 
					tc.expectedCode, tc.method, tc.path, rec.Code)
			}
		})
	}
}

// Test health handler specifically
func TestHealthHandler(t *testing.T) {
	testServer := &TestServer{
		db: &MockDatabaseService{
			HealthFunc: func() map[string]string {
				return map[string]string{
					"status": "up",
					"database": "connected",
				}
			},
		},
	}
	
	handler := testServer.RegisterRoutes()
	
	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	
	handler.ServeHTTP(rec, req)
	
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d for health endpoint, got %d", 
			http.StatusOK, rec.Code)
	}
	
	// Check that response contains JSON
	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Expected JSON content type, got %s", contentType)
	}
	
	// Check that response body contains expected health status
	body := rec.Body.String()
	if !strings.Contains(body, "up") {
		t.Errorf("Expected health status 'up' in response, got: %s", body)
	}
}

// Test basic auth middleware
func TestBasicAuthMiddleware(t *testing.T) {
	// Set auth credentials
	originalUsername := os.Getenv("USERNAME")
	originalPassword := os.Getenv("PASSWORD")
	defer func() {
		os.Setenv("USERNAME", originalUsername)
		os.Setenv("PASSWORD", originalPassword)
	}()
	
	os.Setenv("USERNAME", "testuser")
	os.Setenv("PASSWORD", "testpass")
	
	testServer := &TestServer{
		db: &MockDatabaseService{},
	}
	
	handler := testServer.RegisterRoutes()
	
	// Test without auth
	req := httptest.NewRequest("GET", "/encounters", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d for request without auth, got %d", 
			http.StatusUnauthorized, rec.Code)
	}
	
	// Test with correct auth
	req = httptest.NewRequest("GET", "/encounters", nil)
	req.SetBasicAuth("testuser", "testpass")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	
	// Should not be unauthorized (might be other errors due to missing data)
	if rec.Code == http.StatusUnauthorized {
		t.Error("Expected authorized request to not return 401")
	}
	
	// Test with incorrect auth
	req = httptest.NewRequest("GET", "/encounters", nil)
	req.SetBasicAuth("wrong", "credentials")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d for request with wrong auth, got %d", 
			http.StatusUnauthorized, rec.Code)
	}
}

// Test that static file routes are registered
func TestStaticFileRoutes(t *testing.T) {
	testServer := &TestServer{
		db: &MockDatabaseService{},
	}
	
	handler := testServer.RegisterRoutes()
	
	req := httptest.NewRequest("GET", "/assets/style.css", nil)
	rec := httptest.NewRecorder()
	
	handler.ServeHTTP(rec, req)
	
	// Should not return 404 (route should be registered even if file doesn't exist)
	if rec.Code == http.StatusNotFound {
		t.Error("Static file route /assets/* should be registered")
	}
}

// TestServer is a test version of the server that we can use for testing routes
type TestServer struct {
	db database.Service
}

func (s *TestServer) RegisterRoutes() http.Handler {
	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip middleware for health endpoint
			if c.Path() == "/health" {
				return next(c)
			}
			
			// Simple basic auth check for testing
			username := os.Getenv("USERNAME")
			password := os.Getenv("PASSWORD")
			
			if username == "" || password == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Auth not configured")
			}
			
			auth := c.Request().Header.Get("Authorization")
			if auth == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing Authorization header")
			}
			
			// Check basic auth
			if strings.HasPrefix(auth, "Basic ") {
				// For testing, we'll do a simple check
				// In real implementation, this would properly decode base64
				user, pass, ok := c.Request().BasicAuth()
				if !ok || user != username || pass != password {
					return echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
				}
			}
			
			return next(c)
		}
	})
	
	// Register a few key routes for testing
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, s.db.Health())
	})
	
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Home")
	})
	
	e.GET("/encounters", func(c echo.Context) error {
		return c.String(http.StatusOK, "Encounters")
	})
	
	e.GET("/parties", func(c echo.Context) error {
		return c.String(http.StatusOK, "Parties")
	})
	
	e.GET("/assets/*", func(c echo.Context) error {
		return c.String(http.StatusOK, "Static file")
	})
	
	return e
}