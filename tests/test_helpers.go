package tests

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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

// HTTP test helpers comment removed - functions were unused

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

// String utility helper removed - was unused
