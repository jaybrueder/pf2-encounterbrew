package tests

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"pf2.encounterbrew.com/internal/models"
)

// mockDatabaseService implements the database.Service interface for testing
type mockDatabaseService struct {
	db      *sql.DB
	mock    sqlmock.Sqlmock
	queryFn func(query string, args ...interface{}) *sql.Row
}

func (m *mockDatabaseService) Health() map[string]string {
	return make(map[string]string)
}

func (m *mockDatabaseService) Close() error {
	return m.db.Close()
}

func (m *mockDatabaseService) Insert(table string, columns []string, values ...interface{}) (sql.Result, error) {
	return nil, nil
}

func (m *mockDatabaseService) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return nil, nil
}

func (m *mockDatabaseService) QueryRow(query string, args ...interface{}) *sql.Row {
	if m.queryFn != nil {
		return m.queryFn(query, args...)
	}
	return m.db.QueryRow(query, args...)
}

func (m *mockDatabaseService) Exec(query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}

func (m *mockDatabaseService) Begin() (*sql.Tx, error) {
	return nil, nil
}

func (m *mockDatabaseService) InsertReturningID(table string, columns []string, values ...interface{}) (int, error) {
	return 0, nil
}

func setupUserMockDB(t *testing.T) (*mockDatabaseService, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	mockService := &mockDatabaseService{
		db:   db,
		mock: mock,
	}

	return mockService, func() {
		db.Close()
	}
}

func TestGetUserByID_Success(t *testing.T) {
	mockService, cleanup := setupUserMockDB(t)
	defer cleanup()

	userID := 1
	expectedUser := &models.User{
		ID:            1,
		Name:          "John Doe",
		ActivePartyID: 5,
	}

	// Set up the mock expectation
	rows := sqlmock.NewRows([]string{"id", "name", "active_party_id"}).
		AddRow(expectedUser.ID, expectedUser.Name, expectedUser.ActivePartyID)

	mockService.mock.ExpectQuery(`SELECT id, name, active_party_id FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnRows(rows)

	// Call the function
	user, err := models.GetUserByID(mockService, userID)

	// Assertions
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if user == nil {
		t.Fatal("expected user to be non-nil")
	}

	if user.ID != expectedUser.ID {
		t.Errorf("expected user ID %d, got %d", expectedUser.ID, user.ID)
	}

	if user.Name != expectedUser.Name {
		t.Errorf("expected user name %s, got %s", expectedUser.Name, user.Name)
	}

	if user.ActivePartyID != expectedUser.ActivePartyID {
		t.Errorf("expected active party ID %d, got %d", expectedUser.ActivePartyID, user.ActivePartyID)
	}

	// Verify all expectations were met
	if err := mockService.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetUserByID_UserNotFound(t *testing.T) {
	mockService, cleanup := setupUserMockDB(t)
	defer cleanup()

	userID := 999

	// Set up the mock expectation to return no rows
	mockService.mock.ExpectQuery(`SELECT id, name, active_party_id FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	// Call the function
	user, err := models.GetUserByID(mockService, userID)

	// Assertions
	if err == nil {
		t.Error("expected error when user not found, got nil")
	}

	if user != nil {
		t.Error("expected user to be nil when not found")
	}

	expectedErrorMsg := "no user found with ID 999"
	if err.Error() != expectedErrorMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}

	// Verify all expectations were met
	if err := mockService.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetUserByID_DatabaseError(t *testing.T) {
	mockService, cleanup := setupUserMockDB(t)
	defer cleanup()

	userID := 1
	expectedError := errors.New("database connection failed")

	// Set up the mock expectation to return an error
	mockService.mock.ExpectQuery(`SELECT id, name, active_party_id FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnError(expectedError)

	// Call the function
	user, err := models.GetUserByID(mockService, userID)

	// Assertions
	if err == nil {
		t.Error("expected error when database fails, got nil")
	}

	if user != nil {
		t.Error("expected user to be nil when database fails")
	}

	expectedErrorMsg := "error getting user: database connection failed"
	if err.Error() != expectedErrorMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}

	// Verify all expectations were met
	if err := mockService.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetUserByID_NilDatabase(t *testing.T) {
	// Call the function with nil database
	user, err := models.GetUserByID(nil, 1)

	// Assertions
	if err == nil {
		t.Error("expected error when database is nil, got nil")
	}

	if user != nil {
		t.Error("expected user to be nil when database is nil")
	}

	expectedErrorMsg := "database service is nil"
	if err.Error() != expectedErrorMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}
}

func TestGetUserByID_ScanError(t *testing.T) {
	mockService, cleanup := setupUserMockDB(t)
	defer cleanup()

	userID := 1

	// Set up the mock expectation with incorrect number of columns to cause scan error
	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "John Doe") // Missing active_party_id column

	mockService.mock.ExpectQuery(`SELECT id, name, active_party_id FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnRows(rows)

	// Call the function
	user, err := models.GetUserByID(mockService, userID)

	// Assertions
	if err == nil {
		t.Error("expected error when scan fails, got nil")
	}

	if user != nil {
		t.Error("expected user to be nil when scan fails")
	}

	// The error should contain "error getting user"
	if err != nil && len(err.Error()) > 0 {
		expectedPrefix := "error getting user:"
		if len(err.Error()) < len(expectedPrefix) || err.Error()[:len(expectedPrefix)] != expectedPrefix {
			t.Errorf("expected error message to start with '%s', got '%s'", expectedPrefix, err.Error())
		}
	}

	// Verify all expectations were met
	if err := mockService.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetUserByID_WithZeroActivePartyID(t *testing.T) {
	mockService, cleanup := setupUserMockDB(t)
	defer cleanup()

	userID := 1
	expectedUser := &models.User{
		ID:            1,
		Name:          "Jane Doe",
		ActivePartyID: 0, // Zero value for active party ID
	}

	// Set up the mock expectation
	rows := sqlmock.NewRows([]string{"id", "name", "active_party_id"}).
		AddRow(expectedUser.ID, expectedUser.Name, expectedUser.ActivePartyID)

	mockService.mock.ExpectQuery(`SELECT id, name, active_party_id FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnRows(rows)

	// Call the function
	user, err := models.GetUserByID(mockService, userID)

	// Assertions
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if user == nil {
		t.Fatal("expected user to be non-nil")
	}

	if user.ID != expectedUser.ID {
		t.Errorf("expected user ID %d, got %d", expectedUser.ID, user.ID)
	}

	if user.Name != expectedUser.Name {
		t.Errorf("expected user name %s, got %s", expectedUser.Name, user.Name)
	}

	if user.ActivePartyID != expectedUser.ActivePartyID {
		t.Errorf("expected active party ID %d, got %d", expectedUser.ActivePartyID, user.ActivePartyID)
	}

	// Verify all expectations were met
	if err := mockService.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetUserByID_WithNegativeID(t *testing.T) {
	mockService, cleanup := setupUserMockDB(t)
	defer cleanup()

	userID := -1

	// Set up the mock expectation to return no rows for negative ID
	mockService.mock.ExpectQuery(`SELECT id, name, active_party_id FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	// Call the function
	user, err := models.GetUserByID(mockService, userID)

	// Assertions
	if err == nil {
		t.Error("expected error when user not found, got nil")
	}

	if user != nil {
		t.Error("expected user to be nil when not found")
	}

	expectedErrorMsg := "no user found with ID -1"
	if err.Error() != expectedErrorMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}

	// Verify all expectations were met
	if err := mockService.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}