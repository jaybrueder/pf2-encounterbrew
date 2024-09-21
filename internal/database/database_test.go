package database

import (
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("DB_DATABASE", "testdb")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_USERNAME", "testuser")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_SCHEMA", "public")

	// Reset dbInstance
	dbInstance = nil

	// Call New() and check if it returns a non-nil Service
	service := New()
	assert.NotNil(t, service, "New() should return a non-nil Service")

	// Call New() again and check if it returns the same instance
	service2 := New()
	assert.Equal(t, service, service2, "New() should return the same instance on subsequent calls")
}

func TestHealth(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Create a service with the mock database
	s := &service{db: db}

	// Expect a ping
	mock.ExpectPing()

	// Call Health() and check the result
	health := s.Health()
	assert.Equal(t, "up", health["status"], "Health status should be 'up'")
	assert.Contains(t, health, "open_connections", "Health should contain 'open_connections'")
	assert.Contains(t, health, "in_use", "Health should contain 'in_use'")
	assert.Contains(t, health, "idle", "Health should contain 'idle'")

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestInsert(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Create a service with the mock database
	s := &service{db: db}

	// Set up expectations
	mock.ExpectPrepare("INSERT INTO test_table").
		ExpectExec().
		WithArgs("value1", "value2").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Call Insert() and check the result
	result, err := s.Insert("test_table", []string{"column1", "column2"}, "value1", "value2")
	assert.NoError(t, err, "Insert should not return an error")
	assert.NotNil(t, result, "Insert should return a non-nil result")

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestQuery(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Create a service with the mock database
	s := &service{db: db}

	// Set up expectations
	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "Test1").
		AddRow(2, "Test2")
	mock.ExpectQuery("SELECT \\* FROM test_table").WillReturnRows(rows)

	// Call Query() and check the result
	result, err := s.Query("SELECT * FROM test_table")
	assert.NoError(t, err, "Query should not return an error")
	assert.NotNil(t, result, "Query should return non-nil rows")

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
