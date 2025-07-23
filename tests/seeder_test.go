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

// Use standardized mock database from mock_database.go

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
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Save current working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer func() { _ = os.Chdir(originalWd) }()

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create mock database
	mockDB, cleanup := NewStandardMockDB(t)
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
	defer func() { _ = os.Remove(tempFile.Name()) }()

	// Write test data
	testData := map[string]interface{}{
		"name":        "Test Item",
		"description": "A test item",
	}
	jsonData, _ := json.Marshal(testData)
	if _, err := tempFile.Write(jsonData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	_ = tempFile.Close()

	// Create mock database
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	// Expect the upsert query
	mockDB.Mock.ExpectExec("INSERT INTO test_table").
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
	defer func() { _ = os.Remove(tempFile.Name()) }()

	if _, err := tempFile.WriteString("invalid json"); err != nil {
		t.Fatalf("Failed to write invalid JSON: %v", err)
	}
	_ = tempFile.Close()

	mockDB, cleanup := NewStandardMockDB(t)
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
	defer func() { _ = os.Remove(tempFile.Name()) }()

	testData := map[string]interface{}{
		"description": "A test item without name",
	}
	jsonData, _ := json.Marshal(testData)
	if _, err := tempFile.Write(jsonData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	_ = tempFile.Close()

	mockDB, cleanup := NewStandardMockDB(t)
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
	defer func() { _ = os.Remove(tempFile.Name()) }()

	testData := map[string]interface{}{
		"name":        "   ",
		"description": "A test item with empty name",
	}
	jsonData, _ := json.Marshal(testData)
	if _, err := tempFile.Write(jsonData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	_ = tempFile.Close()

	mockDB, cleanup := NewStandardMockDB(t)
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
	defer func() { _ = os.Remove(tempFile.Name()) }()

	testData := map[string]interface{}{
		"name":        "Test Item",
		"description": "A test item",
	}
	jsonData, _ := json.Marshal(testData)
	if _, err := tempFile.Write(jsonData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	_ = tempFile.Close()

	// Create mock database that returns error
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	// Expect the upsert query to fail
	mockDB.Mock.ExpectExec("INSERT INTO test_table").
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
	defer func() { _ = os.Remove(tempFile.Name()) }()

	testData := map[string]interface{}{
		"name":        "Test Item",
		"description": "A test item",
	}
	jsonData, _ := json.Marshal(testData)
	if _, err := tempFile.Write(jsonData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	_ = tempFile.Close()

	// Create mock database that returns no rows affected
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	// Expect the upsert query to return 0 rows affected
	mockDB.Mock.ExpectExec("INSERT INTO test_table").
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
	mockDB, cleanup := NewStandardMockDB(t)
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
