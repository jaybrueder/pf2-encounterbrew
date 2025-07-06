package tests

import (
	"database/sql"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"pf2.encounterbrew.com/internal/seeder"
)

// SeederMockDatabase wraps sqlmock for our database.Service interface
type SeederMockDatabase struct {
	db   *sql.DB
	mock sqlmock.Sqlmock
}

func (m *SeederMockDatabase) Health() map[string]string {
	return map[string]string{"status": "ok"}
}

func (m *SeederMockDatabase) Close() error {
	return m.db.Close()
}

func (m *SeederMockDatabase) Insert(table string, columns []string, values ...interface{}) (sql.Result, error) {
	return nil, nil
}

func (m *SeederMockDatabase) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return m.db.Query(query, args...)
}

func (m *SeederMockDatabase) QueryRow(query string, args ...interface{}) *sql.Row {
	return m.db.QueryRow(query, args...)
}

func (m *SeederMockDatabase) Exec(query string, args ...interface{}) (sql.Result, error) {
	return m.db.Exec(query, args...)
}

func (m *SeederMockDatabase) Begin() (*sql.Tx, error) {
	return m.db.Begin()
}

func (m *SeederMockDatabase) InsertReturningID(table string, columns []string, values ...interface{}) (int, error) {
	return 1, nil
}

func setupMockDB(t *testing.T) (*SeederMockDatabase, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	
	mockDB := &SeederMockDatabase{
		db:   db,
		mock: mock,
	}
	
	cleanup := func() {
		// Only check expectations if test hasn't already failed
		if !t.Failed() {
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		}
		db.Close()
	}
	
	return mockDB, cleanup
}

func TestRun_NilDatabase(t *testing.T) {
	err := seeder.Run(nil)
	if err == nil {
		t.Error("Expected error when database service is nil")
	}
	if !strings.Contains(err.Error(), "database service cannot be nil") {
		t.Errorf("Expected error message about nil database service, got: %v", err)
	}
}

func TestRun_MissingDirectories(t *testing.T) {
	// Create temporary directory without data subdirectories
	tempDir, err := os.MkdirTemp("", "seeder_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Save current working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create mock database
	mockDB, cleanup := setupMockDB(t)
	defer cleanup()

	// Run seeder - should succeed even with missing directories (they are handled gracefully)
	err = seeder.Run(mockDB)
	if err != nil {
		t.Errorf("Expected no error with missing directories, got: %v", err)
	}
}

func TestUpsertSeedFile_Success(t *testing.T) {
	// Create temporary file
	tempFile, err := os.CreateTemp("", "test_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write test data
	testData := map[string]interface{}{
		"name":        "Test Item",
		"description": "A test item",
	}
	jsonData, _ := json.Marshal(testData)
	if _, err := tempFile.Write(jsonData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tempFile.Close()

	// Create mock database
	mockDB, cleanup := setupMockDB(t)
	defer cleanup()

	// Expect the upsert query
	mockDB.mock.ExpectExec("INSERT INTO test_table").
		WithArgs("Test Item", jsonData).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Test upsert
	changed, err := seeder.UpsertSeedFile(mockDB, tempFile.Name(), "test_table")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !changed {
		t.Error("Expected changed to be true")
	}
}

func TestUpsertSeedFile_InvalidJSON(t *testing.T) {
	// Create temporary file with invalid JSON
	tempFile, err := os.CreateTemp("", "test_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.WriteString("invalid json"); err != nil {
		t.Fatalf("Failed to write invalid JSON: %v", err)
	}
	tempFile.Close()

	mockDB, cleanup := setupMockDB(t)
	defer cleanup()

	// Test upsert
	_, err = seeder.UpsertSeedFile(mockDB, tempFile.Name(), "test_table")
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "unable to parse JSON") {
		t.Errorf("Expected JSON parse error, got: %v", err)
	}
}

func TestUpsertSeedFile_MissingName(t *testing.T) {
	// Create temporary file without name field
	tempFile, err := os.CreateTemp("", "test_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	testData := map[string]interface{}{
		"description": "A test item without name",
	}
	jsonData, _ := json.Marshal(testData)
	if _, err := tempFile.Write(jsonData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tempFile.Close()

	mockDB, cleanup := setupMockDB(t)
	defer cleanup()

	// Test upsert
	_, err = seeder.UpsertSeedFile(mockDB, tempFile.Name(), "test_table")
	if err == nil {
		t.Error("Expected error for missing name field")
	}
	if !strings.Contains(err.Error(), "missing 'name' field") {
		t.Errorf("Expected missing name error, got: %v", err)
	}
}

func TestUpsertSeedFile_EmptyName(t *testing.T) {
	// Create temporary file with empty name
	tempFile, err := os.CreateTemp("", "test_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	testData := map[string]interface{}{
		"name":        "   ",
		"description": "A test item with empty name",
	}
	jsonData, _ := json.Marshal(testData)
	if _, err := tempFile.Write(jsonData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tempFile.Close()

	mockDB, cleanup := setupMockDB(t)
	defer cleanup()

	// Test upsert
	_, err = seeder.UpsertSeedFile(mockDB, tempFile.Name(), "test_table")
	if err == nil {
		t.Error("Expected error for empty name field")
	}
	if !strings.Contains(err.Error(), "cannot be empty or only whitespace") {
		t.Errorf("Expected empty name error, got: %v", err)
	}
}

func TestUpsertSeedFile_DatabaseError(t *testing.T) {
	// Create temporary file
	tempFile, err := os.CreateTemp("", "test_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	testData := map[string]interface{}{
		"name":        "Test Item",
		"description": "A test item",
	}
	jsonData, _ := json.Marshal(testData)
	if _, err := tempFile.Write(jsonData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tempFile.Close()

	// Create mock database that returns error
	mockDB, cleanup := setupMockDB(t)
	defer cleanup()

	// Expect the upsert query to fail
	mockDB.mock.ExpectExec("INSERT INTO test_table").
		WithArgs("Test Item", jsonData).
		WillReturnError(sql.ErrConnDone)

	// Test upsert
	_, err = seeder.UpsertSeedFile(mockDB, tempFile.Name(), "test_table")
	if err == nil {
		t.Error("Expected database error")
	}
	if !strings.Contains(err.Error(), "unable to upsert data") {
		t.Errorf("Expected upsert error, got: %v", err)
	}
}

func TestUpsertSeedFile_NoRowsAffected(t *testing.T) {
	// Create temporary file
	tempFile, err := os.CreateTemp("", "test_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	testData := map[string]interface{}{
		"name":        "Test Item",
		"description": "A test item",
	}
	jsonData, _ := json.Marshal(testData)
	if _, err := tempFile.Write(jsonData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tempFile.Close()

	// Create mock database that returns no rows affected
	mockDB, cleanup := setupMockDB(t)
	defer cleanup()

	// Expect the upsert query to return 0 rows affected
	mockDB.mock.ExpectExec("INSERT INTO test_table").
		WithArgs("Test Item", jsonData).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Test upsert
	changed, err := seeder.UpsertSeedFile(mockDB, tempFile.Name(), "test_table")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if changed {
		t.Error("Expected changed to be false when no rows affected")
	}
}

func TestUpsertSeedFile_NonExistentFile(t *testing.T) {
	mockDB, cleanup := setupMockDB(t)
	defer cleanup()

	// Test with non-existent file
	_, err := seeder.UpsertSeedFile(mockDB, "nonexistent.json", "test_table")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
	if !strings.Contains(err.Error(), "unable to read file") {
		t.Errorf("Expected file read error, got: %v", err)
	}
}

func TestUpsertSeedParties_Success(t *testing.T) {
	// Create temporary parties file
	tempFile, err := os.CreateTemp("", "parties_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	partiesData := seeder.PartiesData{
		Parties: []struct {
			Name    string `json:"name"`
			Players []struct {
				Name       string `json:"name"`
				Level      int    `json:"level"`
				Hp         int    `json:"hp"`
				Ac         int    `json:"ac"`
				Fort       int    `json:"for"`
				Ref        int    `json:"ref"`
				Will       int    `json:"wil"`
				Perception int    `json:"perception"`
			} `json:"players"`
		}{
			{
				Name: "Test Party",
				Players: []struct {
					Name       string `json:"name"`
					Level      int    `json:"level"`
					Hp         int    `json:"hp"`
					Ac         int    `json:"ac"`
					Fort       int    `json:"for"`
					Ref        int    `json:"ref"`
					Will       int    `json:"wil"`
					Perception int    `json:"perception"`
				}{
					{
						Name:       "Test Player",
						Level:      1,
						Hp:         10,
						Ac:         15,
						Fort:       5,
						Ref:        5,
						Will:       5,
						Perception: 5,
					},
				},
			},
		},
	}
	jsonData, _ := json.Marshal(partiesData)
	if _, err := tempFile.Write(jsonData); err != nil {
		t.Fatalf("Failed to write parties data: %v", err)
	}
	tempFile.Close()

	// Create mock database with transaction support
	mockDB, cleanup := setupMockDB(t)
	defer cleanup()

	// Expect transaction to begin
	mockDB.mock.ExpectBegin()

	// Expect party insert
	mockDB.mock.ExpectExec("INSERT INTO parties").
		WithArgs("Test Party", 1).
		WillReturnResult(sqlmock.NewResult(1, 0))

	// Expect party ID query
	mockDB.mock.ExpectQuery("SELECT id FROM parties").
		WithArgs("Test Party", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Expect player insert
	mockDB.mock.ExpectExec("INSERT INTO players").
		WithArgs("Test Player", 1, 10, 15, 5, 5, 5, 5, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect transaction to commit
	mockDB.mock.ExpectCommit()

	// Test upsert
	err = seeder.UpsertSeedParties(mockDB, tempFile.Name())
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestUpsertSeedParties_FileNotFound(t *testing.T) {
	mockDB, cleanup := setupMockDB(t)
	defer cleanup()

	// Test with non-existent file
	err := seeder.UpsertSeedParties(mockDB, "nonexistent.json")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
	if !strings.Contains(err.Error(), "unable to read parties file") {
		t.Errorf("Expected file read error, got: %v", err)
	}
}

func TestUpsertSeedParties_InvalidJSON(t *testing.T) {
	// Create temporary file with invalid JSON
	tempFile, err := os.CreateTemp("", "parties_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.WriteString("invalid json"); err != nil {
		t.Fatalf("Failed to write invalid JSON: %v", err)
	}
	tempFile.Close()

	mockDB, cleanup := setupMockDB(t)
	defer cleanup()

	// Test upsert
	err = seeder.UpsertSeedParties(mockDB, tempFile.Name())
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "unable to parse") {
		t.Errorf("Expected JSON parse error, got: %v", err)
	}
}

func TestUpsertSeedParties_TransactionError(t *testing.T) {
	// Create temporary parties file
	tempFile, err := os.CreateTemp("", "parties_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	partiesData := seeder.PartiesData{
		Parties: []struct {
			Name    string `json:"name"`
			Players []struct {
				Name       string `json:"name"`
				Level      int    `json:"level"`
				Hp         int    `json:"hp"`
				Ac         int    `json:"ac"`
				Fort       int    `json:"for"`
				Ref        int    `json:"ref"`
				Will       int    `json:"wil"`
				Perception int    `json:"perception"`
			} `json:"players"`
		}{
			{
				Name: "Test Party",
				Players: []struct {
					Name       string `json:"name"`
					Level      int    `json:"level"`
					Hp         int    `json:"hp"`
					Ac         int    `json:"ac"`
					Fort       int    `json:"for"`
					Ref        int    `json:"ref"`
					Will       int    `json:"wil"`
					Perception int    `json:"perception"`
				}{
					{
						Name:       "Test Player",
						Level:      1,
						Hp:         10,
						Ac:         15,
						Fort:       5,
						Ref:        5,
						Will:       5,
						Perception: 5,
					},
				},
			},
		},
	}
	jsonData, _ := json.Marshal(partiesData)
	if _, err := tempFile.Write(jsonData); err != nil {
		t.Fatalf("Failed to write parties data: %v", err)
	}
	tempFile.Close()

	// Create mock database that fails to begin transaction
	mockDB, cleanup := setupMockDB(t)
	defer cleanup()

	// Expect transaction to fail
	mockDB.mock.ExpectBegin().WillReturnError(sql.ErrConnDone)

	// Test upsert
	err = seeder.UpsertSeedParties(mockDB, tempFile.Name())
	if err == nil {
		t.Error("Expected transaction error")
	}
	if !strings.Contains(err.Error(), "error starting transaction") {
		t.Errorf("Expected transaction error, got: %v", err)
	}
}

func TestUpsertSeedParties_EmptyPartyName(t *testing.T) {
	// Create temporary parties file with empty party name
	tempFile, err := os.CreateTemp("", "parties_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	partiesData := seeder.PartiesData{
		Parties: []struct {
			Name    string `json:"name"`
			Players []struct {
				Name       string `json:"name"`
				Level      int    `json:"level"`
				Hp         int    `json:"hp"`
				Ac         int    `json:"ac"`
				Fort       int    `json:"for"`
				Ref        int    `json:"ref"`
				Will       int    `json:"wil"`
				Perception int    `json:"perception"`
			} `json:"players"`
		}{
			{
				Name: "   ", // Empty name
				Players: []struct {
					Name       string `json:"name"`
					Level      int    `json:"level"`
					Hp         int    `json:"hp"`
					Ac         int    `json:"ac"`
					Fort       int    `json:"for"`
					Ref        int    `json:"ref"`
					Will       int    `json:"wil"`
					Perception int    `json:"perception"`
				}{
					{
						Name:       "Test Player",
						Level:      1,
						Hp:         10,
						Ac:         15,
						Fort:       5,
						Ref:        5,
						Will:       5,
						Perception: 5,
					},
				},
			},
		},
	}
	jsonData, _ := json.Marshal(partiesData)
	if _, err := tempFile.Write(jsonData); err != nil {
		t.Fatalf("Failed to write parties data: %v", err)
	}
	tempFile.Close()

	// Create mock database
	mockDB, cleanup := setupMockDB(t)
	defer cleanup()

	// Expect transaction to begin and commit (no actual operations)
	mockDB.mock.ExpectBegin()
	mockDB.mock.ExpectCommit()

	// Test upsert - should succeed but skip empty party
	err = seeder.UpsertSeedParties(mockDB, tempFile.Name())
	if err != nil {
		t.Errorf("Expected no error for empty party name, got: %v", err)
	}
}

func TestUpsertSeedParties_PlayerInsertError(t *testing.T) {
	// Create temporary parties file
	tempFile, err := os.CreateTemp("", "parties_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	partiesData := seeder.PartiesData{
		Parties: []struct {
			Name    string `json:"name"`
			Players []struct {
				Name       string `json:"name"`
				Level      int    `json:"level"`
				Hp         int    `json:"hp"`
				Ac         int    `json:"ac"`
				Fort       int    `json:"for"`
				Ref        int    `json:"ref"`
				Will       int    `json:"wil"`
				Perception int    `json:"perception"`
			} `json:"players"`
		}{
			{
				Name: "Test Party",
				Players: []struct {
					Name       string `json:"name"`
					Level      int    `json:"level"`
					Hp         int    `json:"hp"`
					Ac         int    `json:"ac"`
					Fort       int    `json:"for"`
					Ref        int    `json:"ref"`
					Will       int    `json:"wil"`
					Perception int    `json:"perception"`
				}{
					{
						Name:       "Test Player",
						Level:      1,
						Hp:         10,
						Ac:         15,
						Fort:       5,
						Ref:        5,
						Will:       5,
						Perception: 5,
					},
				},
			},
		},
	}
	jsonData, _ := json.Marshal(partiesData)
	if _, err := tempFile.Write(jsonData); err != nil {
		t.Fatalf("Failed to write parties data: %v", err)
	}
	tempFile.Close()

	// Create mock database
	mockDB, cleanup := setupMockDB(t)
	defer cleanup()

	// Expect transaction to begin
	mockDB.mock.ExpectBegin()

	// Expect party insert
	mockDB.mock.ExpectExec("INSERT INTO parties").
		WithArgs("Test Party", 1).
		WillReturnResult(sqlmock.NewResult(1, 0))

	// Expect party ID query
	mockDB.mock.ExpectQuery("SELECT id FROM parties").
		WithArgs("Test Party", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Expect player insert to fail
	mockDB.mock.ExpectExec("INSERT INTO players").
		WithArgs("Test Player", 1, 10, 15, 5, 5, 5, 5, 1).
		WillReturnError(sql.ErrConnDone)

	// Expect transaction rollback
	mockDB.mock.ExpectRollback()

	// Test upsert
	err = seeder.UpsertSeedParties(mockDB, tempFile.Name())
	if err == nil {
		t.Error("Expected error for player insert failure")
	}
	if !strings.Contains(err.Error(), "error upserting player") {
		t.Errorf("Expected player upsert error, got: %v", err)
	}
}

func TestUpsertSeedFile_NameFieldWrongType(t *testing.T) {
	// Create temporary file with name field as number instead of string
	tempFile, err := os.CreateTemp("", "test_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	testData := map[string]interface{}{
		"name":        123, // Name as number instead of string
		"description": "A test item",
	}
	jsonData, _ := json.Marshal(testData)
	if _, err := tempFile.Write(jsonData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tempFile.Close()

	mockDB, cleanup := setupMockDB(t)
	defer cleanup()

	// Test upsert
	_, err = seeder.UpsertSeedFile(mockDB, tempFile.Name(), "test_table")
	if err == nil {
		t.Error("Expected error for name field with wrong type")
	}
	if !strings.Contains(err.Error(), "is not a string") {
		t.Errorf("Expected type error for name field, got: %v", err)
	}
}

func TestUpsertSeedParties_EmptyPlayerName(t *testing.T) {
	// Create temporary parties file with empty player name
	tempFile, err := os.CreateTemp("", "parties_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	partiesData := seeder.PartiesData{
		Parties: []struct {
			Name    string `json:"name"`
			Players []struct {
				Name       string `json:"name"`
				Level      int    `json:"level"`
				Hp         int    `json:"hp"`
				Ac         int    `json:"ac"`
				Fort       int    `json:"for"`
				Ref        int    `json:"ref"`
				Will       int    `json:"wil"`
				Perception int    `json:"perception"`
			} `json:"players"`
		}{
			{
				Name: "Test Party",
				Players: []struct {
					Name       string `json:"name"`
					Level      int    `json:"level"`
					Hp         int    `json:"hp"`
					Ac         int    `json:"ac"`
					Fort       int    `json:"for"`
					Ref        int    `json:"ref"`
					Will       int    `json:"wil"`
					Perception int    `json:"perception"`
				}{
					{
						Name:       "   ", // Empty player name
						Level:      1,
						Hp:         10,
						Ac:         15,
						Fort:       5,
						Ref:        5,
						Will:       5,
						Perception: 5,
					},
				},
			},
		},
	}
	jsonData, _ := json.Marshal(partiesData)
	if _, err := tempFile.Write(jsonData); err != nil {
		t.Fatalf("Failed to write parties data: %v", err)
	}
	tempFile.Close()

	// Create mock database
	mockDB, cleanup := setupMockDB(t)
	defer cleanup()

	// Expect transaction to begin
	mockDB.mock.ExpectBegin()

	// Expect party insert
	mockDB.mock.ExpectExec("INSERT INTO parties").
		WithArgs("Test Party", 1).
		WillReturnResult(sqlmock.NewResult(1, 0))

	// Expect party ID query
	mockDB.mock.ExpectQuery("SELECT id FROM parties").
		WithArgs("Test Party", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// No player insert expected since player name is empty
	// Expect transaction to commit
	mockDB.mock.ExpectCommit()

	// Test upsert - should succeed but skip empty player name
	err = seeder.UpsertSeedParties(mockDB, tempFile.Name())
	if err != nil {
		t.Errorf("Expected no error for empty player name, got: %v", err)
	}
}



func TestUpsertSeedParties_PartyQueryError(t *testing.T) {
	// Create temporary parties file
	tempFile, err := os.CreateTemp("", "parties_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	partiesData := seeder.PartiesData{
		Parties: []struct {
			Name    string `json:"name"`
			Players []struct {
				Name       string `json:"name"`
				Level      int    `json:"level"`
				Hp         int    `json:"hp"`
				Ac         int    `json:"ac"`
				Fort       int    `json:"for"`
				Ref        int    `json:"ref"`
				Will       int    `json:"wil"`
				Perception int    `json:"perception"`
			} `json:"players"`
		}{
			{
				Name: "Test Party",
				Players: []struct {
					Name       string `json:"name"`
					Level      int    `json:"level"`
					Hp         int    `json:"hp"`
					Ac         int    `json:"ac"`
					Fort       int    `json:"for"`
					Ref        int    `json:"ref"`
					Will       int    `json:"wil"`
					Perception int    `json:"perception"`
				}{
					{
						Name:       "Test Player",
						Level:      1,
						Hp:         10,
						Ac:         15,
						Fort:       5,
						Ref:        5,
						Will:       5,
						Perception: 5,
					},
				},
			},
		},
	}
	jsonData, _ := json.Marshal(partiesData)
	if _, err := tempFile.Write(jsonData); err != nil {
		t.Fatalf("Failed to write parties data: %v", err)
	}
	tempFile.Close()

	// Create mock database
	mockDB, cleanup := setupMockDB(t)
	defer cleanup()

	// Expect transaction to begin
	mockDB.mock.ExpectBegin()

	// Expect party insert
	mockDB.mock.ExpectExec("INSERT INTO parties").
		WithArgs("Test Party", 1).
		WillReturnResult(sqlmock.NewResult(1, 0))

	// Expect party ID query to fail
	mockDB.mock.ExpectQuery("SELECT id FROM parties").
		WithArgs("Test Party", 1).
		WillReturnError(sql.ErrNoRows)

	// Expect transaction rollback
	mockDB.mock.ExpectRollback()

	// Test upsert
	err = seeder.UpsertSeedParties(mockDB, tempFile.Name())
	if err == nil {
		t.Error("Expected error for party query failure")
	}
	if !strings.Contains(err.Error(), "error fetching ID for party") {
		t.Errorf("Expected party query error, got: %v", err)
	}
}