package tests

import (
	"encoding/json"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"pf2.encounterbrew.com/internal/models"
)

func TestExportAllParties(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	userID := 1

	// Mock the parties query
	partyRows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "Test Party 1").
		AddRow(2, "Test Party 2")

	mockDB.Mock.ExpectQuery(`SELECT p\.id, p\.name FROM parties p WHERE p\.user_id = \$1 ORDER BY p\.id`).
		WithArgs(userID).
		WillReturnRows(partyRows)

	// Mock the players query for party 1
	playerRows1 := sqlmock.NewRows([]string{"name", "level", "hp", "ac", "fort", "ref", "will", "perception"}).
		AddRow("Player 1", 5, 45, 18, 8, 6, 7, 10).
		AddRow("Player 2", 4, 40, 17, 7, 5, 6, 9)

	mockDB.Mock.ExpectQuery(`SELECT name, level, hp, ac, fort, ref, will, perception FROM players WHERE party_id = \$1 ORDER BY id`).
		WithArgs(1).
		WillReturnRows(playerRows1)

	// Mock the players query for party 2
	playerRows2 := sqlmock.NewRows([]string{"name", "level", "hp", "ac", "fort", "ref", "will", "perception"}).
		AddRow("Player 3", 6, 50, 19, 9, 7, 8, 11)

	mockDB.Mock.ExpectQuery(`SELECT name, level, hp, ac, fort, ref, will, perception FROM players WHERE party_id = \$1 ORDER BY id`).
		WithArgs(2).
		WillReturnRows(playerRows2)

	// Execute the export
	exportData, err := models.ExportAllParties(mockDB, userID)
	if err != nil {
		t.Fatalf("ExportAllParties failed: %v", err)
	}

	// Verify the export data
	if len(exportData.Parties) != 2 {
		t.Errorf("Expected 2 parties, got %d", len(exportData.Parties))
	}

	// Verify party 1
	if exportData.Parties[0].Name != "Test Party 1" {
		t.Errorf("Expected party 1 name 'Test Party 1', got '%s'", exportData.Parties[0].Name)
	}
	if len(exportData.Parties[0].Players) != 2 {
		t.Errorf("Expected 2 players in party 1, got %d", len(exportData.Parties[0].Players))
	}

	// Verify player 1 data
	player1 := exportData.Parties[0].Players[0]
	if player1.Name != "Player 1" || player1.Level != 5 || player1.Hp != 45 {
		t.Errorf("Player 1 data mismatch: %+v", player1)
	}

	// Verify party 2
	if exportData.Parties[1].Name != "Test Party 2" {
		t.Errorf("Expected party 2 name 'Test Party 2', got '%s'", exportData.Parties[1].Name)
	}
	if len(exportData.Parties[1].Players) != 1 {
		t.Errorf("Expected 1 player in party 2, got %d", len(exportData.Parties[1].Players))
	}

	// Verify all expectations were met
	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %s", err)
	}
}

func TestImportParties(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	userID := 1

	// Create test import data
	importData := &models.PartiesExportData{
		Parties: []models.PartyExportData{
			{
				Name: "Imported Party 1",
				Players: []models.PlayerExportData{
					{
						Name:       "Imported Player 1",
						Level:      5,
						Hp:         45,
						Ac:         18,
						Fort:       8,
						Ref:        6,
						Will:       7,
						Perception: 10,
					},
				},
			},
			{
				Name: "Imported Party 2",
				Players: []models.PlayerExportData{
					{
						Name:       "Imported Player 2",
						Level:      6,
						Hp:         50,
						Ac:         19,
						Fort:       9,
						Ref:        7,
						Will:       8,
						Perception: 11,
					},
				},
			},
		},
	}

	// Mock transaction
	mockDB.Mock.ExpectBegin()

	// Mock delete existing parties
	mockDB.Mock.ExpectExec(`DELETE FROM parties WHERE user_id = \$1`).
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 2))

	// Mock insert party 1
	mockDB.Mock.ExpectQuery(`INSERT INTO parties \(name, user_id\) VALUES \(\$1, \$2\) RETURNING id`).
		WithArgs("Imported Party 1", userID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Mock insert player for party 1
	mockDB.Mock.ExpectExec(`INSERT INTO players \( name, level, hp, ac, fort, ref, will, perception, party_id \) VALUES \(\$1, \$2, \$3, \$4, \$5, \$6, \$7, \$8, \$9\)`).
		WithArgs("Imported Player 1", 5, 45, 18, 8, 6, 7, 10, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock insert party 2
	mockDB.Mock.ExpectQuery(`INSERT INTO parties \(name, user_id\) VALUES \(\$1, \$2\) RETURNING id`).
		WithArgs("Imported Party 2", userID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

	// Mock insert player for party 2
	mockDB.Mock.ExpectExec(`INSERT INTO players \( name, level, hp, ac, fort, ref, will, perception, party_id \) VALUES \(\$1, \$2, \$3, \$4, \$5, \$6, \$7, \$8, \$9\)`).
		WithArgs("Imported Player 2", 6, 50, 19, 9, 7, 8, 11, 2).
		WillReturnResult(sqlmock.NewResult(2, 1))

	// Mock commit
	mockDB.Mock.ExpectCommit()

	// Execute the import
	err := models.ImportParties(mockDB, userID, importData)
	if err != nil {
		t.Fatalf("ImportParties failed: %v", err)
	}

	// Verify all expectations were met
	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %s", err)
	}
}

func TestImportParties_EmptyData(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	userID := 1

	// Create empty import data
	importData := &models.PartiesExportData{
		Parties: []models.PartyExportData{},
	}

	// Mock transaction
	mockDB.Mock.ExpectBegin()

	// Mock delete existing parties
	mockDB.Mock.ExpectExec(`DELETE FROM parties WHERE user_id = \$1`).
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Mock commit (no inserts expected)
	mockDB.Mock.ExpectCommit()

	// Execute the import
	err := models.ImportParties(mockDB, userID, importData)
	if err != nil {
		t.Fatalf("ImportParties failed: %v", err)
	}

	// Verify all expectations were met
	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %s", err)
	}
}

func TestDeleteAllUserParties(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	userID := 1

	// Mock delete query
	mockDB.Mock.ExpectExec(`DELETE FROM parties WHERE user_id = \$1`).
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 3))

	// Execute the delete
	err := models.DeleteAllUserParties(mockDB, userID)
	if err != nil {
		t.Fatalf("DeleteAllUserParties failed: %v", err)
	}

	// Verify all expectations were met
	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %s", err)
	}
}

func TestExportImportRoundTrip(t *testing.T) {
	// Test that data can be exported and reimported successfully
	testData := &models.PartiesExportData{
		Parties: []models.PartyExportData{
			{
				Name: "Test Party",
				Players: []models.PlayerExportData{
					{
						Name:       "Test Player",
						Level:      5,
						Hp:         45,
						Ac:         18,
						Fort:       8,
						Ref:        6,
						Will:       7,
						Perception: 10,
					},
				},
			},
		},
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(testData, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	// Unmarshal back
	var importedData models.PartiesExportData
	err = json.Unmarshal(jsonData, &importedData)
	if err != nil {
		t.Fatalf("Failed to unmarshal test data: %v", err)
	}

	// Verify data integrity
	if len(importedData.Parties) != 1 {
		t.Errorf("Expected 1 party, got %d", len(importedData.Parties))
	}

	if importedData.Parties[0].Name != "Test Party" {
		t.Errorf("Expected party name 'Test Party', got '%s'", importedData.Parties[0].Name)
	}

	if len(importedData.Parties[0].Players) != 1 {
		t.Errorf("Expected 1 player, got %d", len(importedData.Parties[0].Players))
	}

	player := importedData.Parties[0].Players[0]
	if player.Name != "Test Player" || player.Level != 5 || player.Hp != 45 {
		t.Errorf("Player data mismatch: %+v", player)
	}
}
