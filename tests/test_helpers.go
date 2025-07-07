package tests

import (
	"context"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/labstack/echo/v4"
)

// Test constants are defined in mock_database.go

// Common test assertions
func requireNoError(t *testing.T, err error, msg ...string) {
	t.Helper()
	if err != nil {
		if len(msg) > 0 {
			t.Errorf("%s: unexpected error: %v", msg[0], err)
		} else {
			t.Errorf("unexpected error: %v", err)
		}
	}
}

func requireError(t *testing.T, err error, msg ...string) {
	t.Helper()
	if err == nil {
		if len(msg) > 0 {
			t.Errorf("%s: expected error but got none", msg[0])
		} else {
			t.Error("expected error but got none")
		}
	}
}

func requireErrorContains(t *testing.T, err error, substring string) {
	t.Helper()
	if err == nil {
		t.Error("expected error but got none")
		return
	}
	if !contains(err.Error(), substring) {
		t.Errorf("expected error to contain '%s', got: %v", substring, err)
	}
}

// HTTP test helpers
func newTestContext(t *testing.T, method, path string, body io.Reader) (echo.Context, *httptest.ResponseRecorder) {
	t.Helper()
	e := echo.New()
	req := httptest.NewRequest(method, path, body)
	req = req.WithContext(context.Background())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

func requireHTTPStatus(t *testing.T, rec *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if rec.Code != expected {
		t.Errorf("expected status %d, got %d", expected, rec.Code)
	}
}

// Database test helpers
func setupMockDB(t *testing.T) (*StandardMockDB, func()) {
	t.Helper()
	return NewStandardMockDB(t)
}

func requireMockExpectationsMet(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled mock expectations: %s", err)
	}
}

// String utility helper (used by requireErrorContains)
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}