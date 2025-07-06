package tests

import (
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"pf2.encounterbrew.com/internal/models"
)

// mockConditionDatabaseService implements the database.Service interface for testing
type mockConditionDatabaseService struct {
	db      *sql.DB
	mock    sqlmock.Sqlmock
	queryFn func(query string, args ...interface{}) *sql.Row
}

func (m *mockConditionDatabaseService) Health() map[string]string {
	return make(map[string]string)
}

func (m *mockConditionDatabaseService) Close() error {
	return m.db.Close()
}

func (m *mockConditionDatabaseService) Insert(table string, columns []string, values ...interface{}) (sql.Result, error) {
	return nil, nil
}

func (m *mockConditionDatabaseService) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return m.db.Query(query, args...)
}

func (m *mockConditionDatabaseService) QueryRow(query string, args ...interface{}) *sql.Row {
	if m.queryFn != nil {
		return m.queryFn(query, args...)
	}
	return m.db.QueryRow(query, args...)
}

func (m *mockConditionDatabaseService) Exec(query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}

func (m *mockConditionDatabaseService) Begin() (*sql.Tx, error) {
	return nil, nil
}

func (m *mockConditionDatabaseService) InsertReturningID(table string, columns []string, values ...interface{}) (int, error) {
	return 0, nil
}

func setupConditionMockDB(t *testing.T) (*mockConditionDatabaseService, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	mockService := &mockConditionDatabaseService{
		db:   db,
		mock: mock,
	}

	return mockService, func() {
		db.Close()
	}
}

func createSampleConditionData() map[string]interface{} {
	return map[string]interface{}{
		"_id":  "condition-test-id",
		"img":  "path/to/condition.jpg",
		"name": "Test Condition",
		"system": map[string]interface{}{
			"description": map[string]interface{}{
				"value": "This is a test condition description",
			},
			"duration": map[string]interface{}{
				"expiry": "turn-start",
				"unit":   "rounds",
				"value":  3,
			},
			"group": "status",
			"overrides": []interface{}{},
			"publication": map[string]interface{}{
				"license":  "OGL",
				"remaster": true,
				"title":    "Test Publication",
			},
			"references": map[string]interface{}{
				"children":     []interface{}{},
				"immunityFrom": []interface{}{},
				"overriddenBy": []interface{}{},
				"overrides":    []interface{}{},
			},
			"rules": []interface{}{},
			"traits": map[string]interface{}{
				"value": []interface{}{},
			},
			"value": map[string]interface{}{
				"isValued": true,
				"value":    5,
			},
		},
		"type": "condition",
	}
}

func TestGetCondition_Success(t *testing.T) {
	mockService, cleanup := setupConditionMockDB(t)
	defer cleanup()

	conditionID := 1
	conditionData := createSampleConditionData()
	jsonData, _ := json.Marshal(conditionData)

	// Set up the mock expectation
	rows := sqlmock.NewRows([]string{"id", "data"}).
		AddRow(conditionID, jsonData)

	mockService.mock.ExpectQuery(`SELECT id, data FROM conditions p WHERE id = \$1`).
		WithArgs(conditionID).
		WillReturnRows(rows)

	// Call the function
	condition, err := models.GetCondition(mockService, conditionID)

	// Assertions
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if condition.ID != conditionID {
		t.Errorf("expected condition ID %d, got %d", conditionID, condition.ID)
	}

	if condition.Data.Name != "Test Condition" {
		t.Errorf("expected condition name 'Test Condition', got '%s'", condition.Data.Name)
	}

	if condition.Data.System.Value.Value != 5 {
		t.Errorf("expected condition value 5, got %d", condition.Data.System.Value.Value)
	}

	if !condition.Data.System.Value.IsValued {
		t.Error("expected condition to be valued")
	}

	// Verify all expectations were met
	if err := mockService.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetCondition_NotFound(t *testing.T) {
	mockService, cleanup := setupConditionMockDB(t)
	defer cleanup()

	conditionID := 999

	// Set up the mock expectation to return no rows
	mockService.mock.ExpectQuery(`SELECT id, data FROM conditions p WHERE id = \$1`).
		WithArgs(conditionID).
		WillReturnError(sql.ErrNoRows)

	// Call the function
	condition, err := models.GetCondition(mockService, conditionID)

	// Assertions
	if err == nil {
		t.Error("expected error when condition not found, got nil")
	}

	if condition.ID != 0 {
		t.Error("expected empty condition when not found")
	}

	expectedErrorMsg := "no condition found with ID 999"
	if err.Error() != expectedErrorMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}

	// Verify all expectations were met
	if err := mockService.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetCondition_DatabaseError(t *testing.T) {
	mockService, cleanup := setupConditionMockDB(t)
	defer cleanup()

	conditionID := 1
	expectedError := errors.New("database connection failed")

	// Set up the mock expectation to return an error
	mockService.mock.ExpectQuery(`SELECT id, data FROM conditions p WHERE id = \$1`).
		WithArgs(conditionID).
		WillReturnError(expectedError)

	// Call the function
	condition, err := models.GetCondition(mockService, conditionID)

	// Assertions
	if err == nil {
		t.Error("expected error when database fails, got nil")
	}

	if condition.ID != 0 {
		t.Error("expected empty condition when database fails")
	}

	expectedErrorMsg := "error scanning condition row: database connection failed"
	if err.Error() != expectedErrorMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}

	// Verify all expectations were met
	if err := mockService.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetCondition_NilDatabase(t *testing.T) {
	// Call the function with nil database
	condition, err := models.GetCondition(nil, 1)

	// Assertions
	if err == nil {
		t.Error("expected error when database is nil, got nil")
	}

	if condition.ID != 0 {
		t.Error("expected empty condition when database is nil")
	}

	expectedErrorMsg := "database service is nil"
	if err.Error() != expectedErrorMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}
}

func TestGetCondition_InvalidJSON(t *testing.T) {
	mockService, cleanup := setupConditionMockDB(t)
	defer cleanup()

	conditionID := 1
	invalidJSON := []byte(`{"invalid": json}`)

	// Set up the mock expectation with invalid JSON
	rows := sqlmock.NewRows([]string{"id", "data"}).
		AddRow(conditionID, invalidJSON)

	mockService.mock.ExpectQuery(`SELECT id, data FROM conditions p WHERE id = \$1`).
		WithArgs(conditionID).
		WillReturnRows(rows)

	// Call the function
	condition, err := models.GetCondition(mockService, conditionID)

	// Assertions
	if err == nil {
		t.Error("expected error when JSON is invalid, got nil")
	}

	if condition.ID != 0 {
		t.Error("expected empty condition when JSON is invalid")
	}

	// The error should contain "error unmarshaling condition data"
	if err != nil && len(err.Error()) > 0 {
		expectedPrefix := "error unmarshaling condition data:"
		if len(err.Error()) < len(expectedPrefix) || err.Error()[:len(expectedPrefix)] != expectedPrefix {
			t.Errorf("expected error message to start with '%s', got '%s'", expectedPrefix, err.Error())
		}
	}

	// Verify all expectations were met
	if err := mockService.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetGroupedConditions_Success(t *testing.T) {
	mockService, cleanup := setupConditionMockDB(t)
	defer cleanup()

	// Create sample condition data for different groups
	statusCondition := createSampleConditionData()
	statusCondition["system"].(map[string]interface{})["group"] = "status"
	statusJSON, _ := json.Marshal(statusCondition)

	otherCondition := createSampleConditionData()
	otherCondition["name"] = "Other Condition"
	otherCondition["system"].(map[string]interface{})["group"] = nil
	otherJSON, _ := json.Marshal(otherCondition)

	// Set up the mock expectation
	rows := sqlmock.NewRows([]string{"id", "data"}).
		AddRow(1, statusJSON).
		AddRow(2, otherJSON)

	mockService.mock.ExpectQuery(`SELECT id, data FROM conditions`).
		WillReturnRows(rows)

	// Call the function
	groupedConditions, err := models.GetGroupedConditions(mockService)

	// Assertions
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if groupedConditions == nil {
		t.Fatal("expected grouped conditions to be non-nil")
	}

	// Check status group
	if _, exists := groupedConditions["status"]; !exists {
		t.Error("expected 'status' group to exist")
	}

	if len(groupedConditions["status"]) != 1 {
		t.Errorf("expected 1 condition in 'status' group, got %d", len(groupedConditions["status"]))
	}

	// Check other group
	if _, exists := groupedConditions["other"]; !exists {
		t.Error("expected 'other' group to exist")
	}

	if len(groupedConditions["other"]) != 1 {
		t.Errorf("expected 1 condition in 'other' group, got %d", len(groupedConditions["other"]))
	}

	// Verify all expectations were met
	if err := mockService.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetGroupedConditions_NilDatabase(t *testing.T) {
	// Call the function with nil database
	groupedConditions, err := models.GetGroupedConditions(nil)

	// Assertions
	if err == nil {
		t.Error("expected error when database is nil, got nil")
	}

	if groupedConditions != nil {
		t.Error("expected nil grouped conditions when database is nil")
	}

	expectedErrorMsg := "database service is nil"
	if err.Error() != expectedErrorMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}
}

func TestGetGroupedConditions_DatabaseError(t *testing.T) {
	mockService, cleanup := setupConditionMockDB(t)
	defer cleanup()

	expectedError := errors.New("database query failed")

	// Set up the mock expectation to return an error
	mockService.mock.ExpectQuery(`SELECT id, data FROM conditions`).
		WillReturnError(expectedError)

	// Call the function
	groupedConditions, err := models.GetGroupedConditions(mockService)

	// Assertions
	if err == nil {
		t.Error("expected error when database query fails, got nil")
	}

	if groupedConditions != nil {
		t.Error("expected nil grouped conditions when database query fails")
	}

	expectedErrorMsg := "error querying conditions: database query failed"
	if err.Error() != expectedErrorMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}

	// Verify all expectations were met
	if err := mockService.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetGroupedConditions_ScanError(t *testing.T) {
	mockService, cleanup := setupConditionMockDB(t)
	defer cleanup()

	// Set up the mock expectation with incorrect number of columns
	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(1) // Missing data column

	mockService.mock.ExpectQuery(`SELECT id, data FROM conditions`).
		WillReturnRows(rows)

	// Call the function
	groupedConditions, err := models.GetGroupedConditions(mockService)

	// Assertions
	if err == nil {
		t.Error("expected error when scan fails, got nil")
	}

	if groupedConditions != nil {
		t.Error("expected nil grouped conditions when scan fails")
	}

	// The error should contain "error scanning row"
	if err != nil && len(err.Error()) > 0 {
		expectedPrefix := "error scanning row:"
		if len(err.Error()) < len(expectedPrefix) || err.Error()[:len(expectedPrefix)] != expectedPrefix {
			t.Errorf("expected error message to start with '%s', got '%s'", expectedPrefix, err.Error())
		}
	}

	// Verify all expectations were met
	if err := mockService.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetGroupedConditions_JSONUnmarshalError(t *testing.T) {
	mockService, cleanup := setupConditionMockDB(t)
	defer cleanup()

	invalidJSON := []byte(`{"invalid": json}`)

	// Set up the mock expectation with invalid JSON
	rows := sqlmock.NewRows([]string{"id", "data"}).
		AddRow(1, invalidJSON)

	mockService.mock.ExpectQuery(`SELECT id, data FROM conditions`).
		WillReturnRows(rows)

	// Call the function
	groupedConditions, err := models.GetGroupedConditions(mockService)

	// Assertions
	if err == nil {
		t.Error("expected error when JSON unmarshal fails, got nil")
	}

	if groupedConditions != nil {
		t.Error("expected nil grouped conditions when JSON unmarshal fails")
	}

	// The error should contain "error unmarshaling JSON"
	if err != nil && len(err.Error()) > 0 {
		expectedPrefix := "error unmarshaling JSON:"
		if len(err.Error()) < len(expectedPrefix) || err.Error()[:len(expectedPrefix)] != expectedPrefix {
			t.Errorf("expected error message to start with '%s', got '%s'", expectedPrefix, err.Error())
		}
	}

	// Verify all expectations were met
	if err := mockService.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCondition_GetValue(t *testing.T) {
	condition := models.Condition{}
	condition.Data.System.Value.Value = 10

	value := condition.GetValue()
	if value != 10 {
		t.Errorf("expected value 10, got %d", value)
	}
}

func TestCondition_SetValue(t *testing.T) {
	condition := models.Condition{}
	condition.SetValue(15)

	if condition.Data.System.Value.Value != 15 {
		t.Errorf("expected value 15, got %d", condition.Data.System.Value.Value)
	}
}

func TestCondition_GetName(t *testing.T) {
	condition := models.Condition{}
	condition.Data.Name = "Test <hr />Condition"

	name := condition.GetName()
	expectedName := "Test Condition"
	if name != expectedName {
		t.Errorf("expected name '%s', got '%s'", expectedName, name)
	}
}

func TestCondition_IsValued(t *testing.T) {
	condition := models.Condition{}
	condition.Data.System.Value.IsValued = true

	if !condition.IsValued() {
		t.Error("expected condition to be valued")
	}

	condition.Data.System.Value.IsValued = false
	if condition.IsValued() {
		t.Error("expected condition to not be valued")
	}
}

func TestCondition_GetDescription(t *testing.T) {
	condition := models.Condition{}
	condition.Data.System.Description.Value = "Test <p>description</p> with @Localize[text] pattern"

	description := condition.GetDescription()
	// The description should have HTML removed and patterns processed
	if description == "" {
		t.Error("expected non-empty description")
	}
}

func TestGetGroupedConditions_EmptyGroup(t *testing.T) {
	mockService, cleanup := setupConditionMockDB(t)
	defer cleanup()

	// Create condition with empty string group
	conditionData := createSampleConditionData()
	conditionData["system"].(map[string]interface{})["group"] = ""
	jsonData, _ := json.Marshal(conditionData)

	// Set up the mock expectation
	rows := sqlmock.NewRows([]string{"id", "data"}).
		AddRow(1, jsonData)

	mockService.mock.ExpectQuery(`SELECT id, data FROM conditions`).
		WillReturnRows(rows)

	// Call the function
	groupedConditions, err := models.GetGroupedConditions(mockService)

	// Assertions
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Empty string group should be treated as "other"
	if _, exists := groupedConditions["other"]; !exists {
		t.Error("expected 'other' group to exist for empty string group")
	}

	if len(groupedConditions["other"]) != 1 {
		t.Errorf("expected 1 condition in 'other' group, got %d", len(groupedConditions["other"]))
	}

	// Verify all expectations were met
	if err := mockService.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}