package tests

import (
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"pf2.encounterbrew.com/internal/models"
)


// Use consolidated fixtures from mock_database.go

func TestGetCondition_Success(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	conditionID := 1
	conditionData := CreateSampleConditionData()
	jsonData, _ := json.Marshal(conditionData)

	// Set up the mock expectation
	rows := sqlmock.NewRows([]string{"id", "data"}).
		AddRow(conditionID, jsonData)

	mockDB.Mock.ExpectQuery(`SELECT id, data FROM conditions p WHERE id = \$1`).
		WithArgs(conditionID).
		WillReturnRows(rows)

	// Call the function
	condition, err := models.GetCondition(mockDB, conditionID)

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
	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetCondition_NotFound(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	conditionID := 999

	// Set up the mock expectation to return no rows
	mockDB.Mock.ExpectQuery(`SELECT id, data FROM conditions p WHERE id = \$1`).
		WithArgs(conditionID).
		WillReturnError(sql.ErrNoRows)

	// Call the function
	condition, err := models.GetCondition(mockDB, conditionID)

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
	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetCondition_DatabaseError(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	conditionID := 1
	expectedError := errors.New("database connection failed")

	// Set up the mock expectation to return an error
	mockDB.Mock.ExpectQuery(`SELECT id, data FROM conditions p WHERE id = \$1`).
		WithArgs(conditionID).
		WillReturnError(expectedError)

	// Call the function
	condition, err := models.GetCondition(mockDB, conditionID)

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
	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
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
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	conditionID := 1
	invalidJSON := []byte(`{"invalid": json}`)

	// Set up the mock expectation with invalid JSON
	rows := sqlmock.NewRows([]string{"id", "data"}).
		AddRow(conditionID, invalidJSON)

	mockDB.Mock.ExpectQuery(`SELECT id, data FROM conditions p WHERE id = \$1`).
		WithArgs(conditionID).
		WillReturnRows(rows)

	// Call the function
	condition, err := models.GetCondition(mockDB, conditionID)

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
	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetGroupedConditions_Success(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	// Create sample condition data for different groups
	statusCondition := CreateSampleConditionData()
	statusCondition["system"].(map[string]interface{})["group"] = "status"
	statusJSON, _ := json.Marshal(statusCondition)

	otherCondition := CreateSampleConditionData()
	otherCondition["name"] = "Other Condition"
	otherCondition["system"].(map[string]interface{})["group"] = nil
	otherJSON, _ := json.Marshal(otherCondition)

	// Set up the mock expectation
	rows := sqlmock.NewRows([]string{"id", "data"}).
		AddRow(1, statusJSON).
		AddRow(2, otherJSON)

	mockDB.Mock.ExpectQuery(`SELECT id, data FROM conditions`).
		WillReturnRows(rows)

	// Call the function
	groupedConditions, err := models.GetGroupedConditions(mockDB)

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
	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
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
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	expectedError := errors.New("database query failed")

	// Set up the mock expectation to return an error
	mockDB.Mock.ExpectQuery(`SELECT id, data FROM conditions`).
		WillReturnError(expectedError)

	// Call the function
	groupedConditions, err := models.GetGroupedConditions(mockDB)

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
	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetGroupedConditions_ScanError(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	// Set up the mock expectation with incorrect number of columns
	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(1) // Missing data column

	mockDB.Mock.ExpectQuery(`SELECT id, data FROM conditions`).
		WillReturnRows(rows)

	// Call the function
	groupedConditions, err := models.GetGroupedConditions(mockDB)

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
	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetGroupedConditions_JSONUnmarshalError(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	invalidJSON := []byte(`{"invalid": json}`)

	// Set up the mock expectation with invalid JSON
	rows := sqlmock.NewRows([]string{"id", "data"}).
		AddRow(1, invalidJSON)

	mockDB.Mock.ExpectQuery(`SELECT id, data FROM conditions`).
		WillReturnRows(rows)

	// Call the function
	groupedConditions, err := models.GetGroupedConditions(mockDB)

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
	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
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
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	// Create condition with empty string group
	conditionData := CreateSampleConditionData()
	conditionData["system"].(map[string]interface{})["group"] = ""
	jsonData, _ := json.Marshal(conditionData)

	// Set up the mock expectation
	rows := sqlmock.NewRows([]string{"id", "data"}).
		AddRow(1, jsonData)

	mockDB.Mock.ExpectQuery(`SELECT id, data FROM conditions`).
		WillReturnRows(rows)

	// Call the function
	groupedConditions, err := models.GetGroupedConditions(mockDB)

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
	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}