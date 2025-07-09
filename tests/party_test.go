package tests

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"pf2.encounterbrew.com/internal/models"
)

// Test data helpers - use consolidated fixtures from mock_database.go

// Party Method Tests

func TestParty_GetLevel(t *testing.T) {
	tests := []struct {
		name          string
		party         models.Party
		expectedLevel float64
	}{
		{
			name:          "party with players",
			party:         CreateSampleParty(),
			expectedLevel: 4.5, // (5 + 4) / 2
		},
		{
			name: "party with no players",
			party: models.Party{
				ID:      1,
				Name:    "Empty Party",
				Players: []models.Player{},
			},
			expectedLevel: 0,
		},
		{
			name: "party with single player",
			party: models.Party{
				ID:   1,
				Name: "Single Player Party",
				Players: []models.Player{
					{Level: 7},
				},
			},
			expectedLevel: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := tt.party.GetLevel()
			if level != tt.expectedLevel {
				t.Errorf("expected level %.1f, got %.1f", tt.expectedLevel, level)
			}
		})
	}
}

func TestParty_GetNumbersOfPlayer(t *testing.T) {
	tests := []struct {
		name           string
		party          models.Party
		expectedNumber float64
	}{
		{
			name:           "party with players",
			party:          CreateSampleParty(),
			expectedNumber: 2,
		},
		{
			name: "party with no players",
			party: models.Party{
				Players: []models.Player{},
			},
			expectedNumber: 0,
		},
		{
			name: "party with single player",
			party: models.Party{
				Players: []models.Player{{}},
			},
			expectedNumber: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			number := tt.party.GetNumbersOfPlayer()
			if number != tt.expectedNumber {
				t.Errorf("expected number %.0f, got %.0f", tt.expectedNumber, number)
			}
		})
	}
}

// Party.Create Tests

func TestParty_Create_Success(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	party := models.Party{
		Name:   "New Party",
		UserID: 1,
	}

	expectedID := 1
	mockDB.SetupMockForPartyCreate(expectedID)

	id, err := party.Create(mockDB)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if id != expectedID {
		t.Errorf("expected ID %d, got %d", expectedID, id)
	}

	// Verify all expectations were met
	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestParty_Create_NilDatabase(t *testing.T) {
	party := models.Party{Name: "Test", UserID: 1}
	id, err := party.Create(nil)

	if err == nil {
		t.Error("expected error when database is nil, got nil")
	}

	if id != 0 {
		t.Errorf("expected ID 0, got %d", id)
	}

	expectedErrorMsg := "database service is nil"
	if err.Error() != expectedErrorMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}
}

// GetAllParties Tests

func TestGetAllParties_Success(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	// Mock parties query
	partyRows := sqlmock.NewRows([]string{"id", "name", "user_id", "user_name"}).
		AddRow(1, "Party 1", 1, "Test User").
		AddRow(2, "Party 2", 1, "Test User")

	mockDB.Mock.ExpectQuery("SELECT p.id, p.name, p.user_id, u.name AS user_name FROM parties p JOIN users u ON p.user_id = u.id WHERE p.user_id = \\$1 ORDER BY p.id").
		WithArgs(1).
		WillReturnRows(partyRows)

	// Mock players query for party 1
	playerRows1 := sqlmock.NewRows([]string{"id", "name", "level", "hp", "ac", "fort", "ref", "will"}).
		AddRow(1, "Player 1", 5, 45, 18, 8, 6, 7).
		AddRow(2, "Player 2", 4, 35, 16, 6, 8, 5)
	mockDB.Mock.ExpectQuery("SELECT id, name, level, hp, ac, fort, ref, will FROM players WHERE party_id = \\$1").
		WithArgs(1).
		WillReturnRows(playerRows1)

	// Mock players query for party 2 (empty)
	playerRows2 := sqlmock.NewRows([]string{"id", "name", "level", "hp", "ac", "fort", "ref", "will"})
	mockDB.Mock.ExpectQuery("SELECT id, name, level, hp, ac, fort, ref, will FROM players WHERE party_id = \\$1").
		WithArgs(2).
		WillReturnRows(playerRows2)

	parties, err := models.GetAllParties(mockDB)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(parties) != 2 {
		t.Errorf("expected 2 parties, got %d", len(parties))
	}

	if parties[0].Name != "Party 1" {
		t.Errorf("expected first party name 'Party 1', got '%s'", parties[0].Name)
	}

	if len(parties[0].Players) != 2 {
		t.Errorf("expected 2 players in first party, got %d", len(parties[0].Players))
	}

	if len(parties[1].Players) != 0 {
		t.Errorf("expected 0 players in second party, got %d", len(parties[1].Players))
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestGetAllParties_NilDatabase(t *testing.T) {
	parties, err := models.GetAllParties(nil)
	if err == nil {
		t.Error("expected error when database is nil, got nil")
	}

	if parties != nil {
		t.Error("expected nil parties when database is nil")
	}

	expectedErrorMsg := "database service is nil"
	if err.Error() != expectedErrorMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}
}

func TestGetAllParties_QueryError(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	mockDB.Mock.ExpectQuery("SELECT p.id, p.name, p.user_id, u.name AS user_name FROM parties p JOIN users u ON p.user_id = u.id WHERE p.user_id = \\$1 ORDER BY p.id").
		WithArgs(1).
		WillReturnError(sql.ErrConnDone)

	parties, err := models.GetAllParties(mockDB)
	if err == nil {
		t.Error("expected error when query fails, got nil")
	}

	if parties != nil {
		t.Error("expected nil parties when query fails")
	}

	if !strings.Contains(err.Error(), "error querying parties") {
		t.Errorf("expected query error, got: %v", err)
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// GetParty Tests

func TestGetParty_Success(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	partyID := 1

	// Mock party query
	partyRows := sqlmock.NewRows([]string{"id", "name", "user_id", "user_name"}).
		AddRow(1, "Test Party", 1, "Test User")
	mockDB.Mock.ExpectQuery("SELECT p.id, p.name, p.user_id, u.name AS user_name FROM parties p JOIN users u ON p.user_id = u.id WHERE p.user_id = \\$1 AND p.id = \\$2").
		WithArgs(1, partyID).
		WillReturnRows(partyRows)

	// Mock players query
	playerRows := sqlmock.NewRows([]string{"id", "name", "level", "hp", "ac", "fort", "ref", "will", "perception"}).
		AddRow(1, "Player 1", 5, 45, 18, 8, 6, 7, 5)
	mockDB.Mock.ExpectQuery("SELECT id, name, level, hp, ac, fort, ref, will, perception FROM players WHERE party_id = \\$1").
		WithArgs(partyID).
		WillReturnRows(playerRows)

	party, err := models.GetParty(mockDB, partyID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if party.ID != partyID {
		t.Errorf("expected party ID %d, got %d", partyID, party.ID)
	}

	if party.Name != "Test Party" {
		t.Errorf("expected party name 'Test Party', got '%s'", party.Name)
	}

	if len(party.Players) != 1 {
		t.Errorf("expected 1 player, got %d", len(party.Players))
	}

	if party.Players[0].PartyID != partyID {
		t.Errorf("expected player party ID %d, got %d", partyID, party.Players[0].PartyID)
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestGetParty_NotFound(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	partyID := 999

	mockDB.Mock.ExpectQuery("SELECT p.id, p.name, p.user_id, u.name AS user_name FROM parties p JOIN users u ON p.user_id = u.id WHERE p.user_id = \\$1 AND p.id = \\$2").
		WithArgs(1, partyID).
		WillReturnError(sql.ErrNoRows)

	party, err := models.GetParty(mockDB, partyID)
	if err == nil {
		t.Error("expected error when party not found, got nil")
	}

	if party.ID != 0 {
		t.Error("expected empty party when not found")
	}

	expectedError := "no party found with ID 999"
	if err.Error() != expectedError {
		t.Errorf("expected error message '%s', got '%s'", expectedError, err.Error())
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestGetParty_NilDatabase(t *testing.T) {
	party, err := models.GetParty(nil, 1)
	if err == nil {
		t.Error("expected error when database is nil, got nil")
	}

	if party.ID != 0 {
		t.Error("expected empty party when database is nil")
	}

	expectedErrorMsg := "database service is nil"
	if err.Error() != expectedErrorMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}
}

// PartyExists Tests

func TestPartyExists_True(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	partyID := 1

	mockDB.Mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM parties WHERE id = \\$1\\)").
		WithArgs(partyID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := models.PartyExists(mockDB, partyID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if !exists {
		t.Error("expected party to exist")
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestPartyExists_False(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	partyID := 999

	mockDB.Mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM parties WHERE id = \\$1\\)").
		WithArgs(partyID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	exists, err := models.PartyExists(mockDB, partyID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if exists {
		t.Error("expected party to not exist")
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestPartyExists_QueryError(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	mockDB.Mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM parties WHERE id = \\$1\\)").
		WithArgs(1).
		WillReturnError(sql.ErrConnDone)

	exists, err := models.PartyExists(mockDB, 1)
	if err == nil {
		t.Error("expected error when query fails, got nil")
	}

	if exists {
		t.Error("expected exists to be false when query fails")
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// Party.Update Tests

func TestParty_Update_Success(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	party := CreateSampleParty()
	party.Name = "Updated Party Name"

	mockDB.Mock.ExpectExec("UPDATE parties SET name = \\$1 WHERE id = \\$2 AND user_id = \\$3").
		WithArgs(party.Name, party.ID, party.UserID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := party.Update(mockDB)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestParty_Update_NilDatabase(t *testing.T) {
	party := CreateSampleParty()
	err := party.Update(nil)

	if err == nil {
		t.Error("expected error when database is nil, got nil")
	}

	expectedErrorMsg := "database service is nil"
	if err.Error() != expectedErrorMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}
}

func TestParty_Update_NoRowsAffected(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	party := CreateSampleParty()

	mockDB.Mock.ExpectExec("UPDATE parties SET name = \\$1 WHERE id = \\$2 AND user_id = \\$3").
		WithArgs(party.Name, party.ID, party.UserID).
		WillReturnResult(sqlmock.NewResult(1, 0))

	err := party.Update(mockDB)
	if err == nil {
		t.Error("expected error when no rows affected, got nil")
	}

	expectedError := "party not found or user not authorized"
	if err.Error() != expectedError {
		t.Errorf("expected error message '%s', got '%s'", expectedError, err.Error())
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// Party.Delete Tests

func TestParty_Delete_Success(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	party := CreateSampleParty()

	mockDB.Mock.ExpectExec("DELETE FROM parties WHERE id = \\$1 AND user_id = \\$2").
		WithArgs(party.ID, party.UserID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := party.Delete(mockDB)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestParty_Delete_NilDatabase(t *testing.T) {
	party := CreateSampleParty()
	err := party.Delete(nil)

	if err == nil {
		t.Error("expected error when database is nil, got nil")
	}

	expectedErrorMsg := "database service is nil"
	if err.Error() != expectedErrorMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}
}

func TestParty_Delete_NoRowsAffected(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	party := CreateSampleParty()

	mockDB.Mock.ExpectExec("DELETE FROM parties WHERE id = \\$1 AND user_id = \\$2").
		WithArgs(party.ID, party.UserID).
		WillReturnResult(sqlmock.NewResult(1, 0))

	err := party.Delete(mockDB)
	if err == nil {
		t.Error("expected error when no rows affected, got nil")
	}

	expectedError := "party not found or user not authorized"
	if err.Error() != expectedError {
		t.Errorf("expected error message '%s', got '%s'", expectedError, err.Error())
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// Party.UpdateWithPlayers Tests

func TestParty_UpdateWithPlayers_Success_NewPlayer(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	party := CreateSampleParty()
	// Add a new player (ID = 0)
	newPlayer := models.Player{
		ID:         0, // New player
		Name:       "New Player",
		Level:      3,
		Hp:         30,
		Ac:         15,
		Fort:       5,
		Ref:        5,
		Will:       4,
		Perception: 3,
	}
	party.Players = append(party.Players, newPlayer)

	// Mock transaction
	mockDB.Mock.ExpectBegin()

	// Mock party update
	mockDB.Mock.ExpectExec("UPDATE parties SET name = \\$1 WHERE id = \\$2 AND user_id = \\$3").
		WithArgs(party.Name, party.ID, party.UserID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock existing player updates
	for _, player := range party.Players[:2] { // First two are existing
		mockDB.Mock.ExpectExec("UPDATE players SET name = \\$1, level = \\$2, ac = \\$3, hp = \\$4, fort = \\$5, ref = \\$6, will = \\$7, perception = \\$8 WHERE id = \\$9 AND party_id = \\$10").
			WithArgs(player.Name, player.Level, player.Ac, player.Hp, player.Fort, player.Ref, player.Will, player.Perception, player.ID, party.ID).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	// Mock new player insert
	mockDB.Mock.ExpectQuery("INSERT INTO players \\(name, level, ac, hp, fort, ref, will, perception, party_id\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, \\$5, \\$6, \\$7, \\$8, \\$9\\) RETURNING id").
		WithArgs(newPlayer.Name, newPlayer.Level, newPlayer.Ac, newPlayer.Hp, newPlayer.Fort, newPlayer.Ref, newPlayer.Will, newPlayer.Perception, party.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))

	// Mock transaction commit
	mockDB.Mock.ExpectCommit()

	err := party.UpdateWithPlayers(mockDB, []int{})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestParty_UpdateWithPlayers_Success_DeletePlayers(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	party := CreateSampleParty()
	playersToDelete := []int{3, 4}

	// Mock transaction
	mockDB.Mock.ExpectBegin()

	// Mock party update
	mockDB.Mock.ExpectExec("UPDATE parties SET name = \\$1 WHERE id = \\$2 AND user_id = \\$3").
		WithArgs(party.Name, party.ID, party.UserID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock player deletion
	mockDB.Mock.ExpectExec("DELETE FROM players WHERE id IN \\(\\$1,\\$2\\)").
		WithArgs(3, 4).
		WillReturnResult(sqlmock.NewResult(1, 2))

	// Mock existing player updates
	for _, player := range party.Players {
		mockDB.Mock.ExpectExec("UPDATE players SET name = \\$1, level = \\$2, ac = \\$3, hp = \\$4, fort = \\$5, ref = \\$6, will = \\$7, perception = \\$8 WHERE id = \\$9 AND party_id = \\$10").
			WithArgs(player.Name, player.Level, player.Ac, player.Hp, player.Fort, player.Ref, player.Will, player.Perception, player.ID, party.ID).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	// Mock transaction commit
	mockDB.Mock.ExpectCommit()

	err := party.UpdateWithPlayers(mockDB, playersToDelete)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestParty_UpdateWithPlayers_NilDatabase(t *testing.T) {
	party := CreateSampleParty()
	err := party.UpdateWithPlayers(nil, []int{})

	if err == nil {
		t.Error("expected error when database is nil, got nil")
	}

	expectedErrorMsg := "database service is nil"
	if err.Error() != expectedErrorMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}
}

func TestParty_UpdateWithPlayers_TransactionError(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	party := CreateSampleParty()

	mockDB.Mock.ExpectBegin().WillReturnError(sql.ErrConnDone)

	err := party.UpdateWithPlayers(mockDB, []int{})
	if err == nil {
		t.Error("expected error when transaction fails, got nil")
	}

	if !strings.Contains(err.Error(), "error starting transaction") {
		t.Errorf("expected transaction error, got: %v", err)
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestParty_UpdateWithPlayers_PlayerNotFound(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	party := CreateSampleParty()

	// Mock transaction
	mockDB.Mock.ExpectBegin()

	// Mock party update
	mockDB.Mock.ExpectExec("UPDATE parties SET name = \\$1 WHERE id = \\$2 AND user_id = \\$3").
		WithArgs(party.Name, party.ID, party.UserID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock first player update (success)
	mockDB.Mock.ExpectExec("UPDATE players SET name = \\$1, level = \\$2, ac = \\$3, hp = \\$4, fort = \\$5, ref = \\$6, will = \\$7, perception = \\$8 WHERE id = \\$9 AND party_id = \\$10").
		WithArgs(party.Players[0].Name, party.Players[0].Level, party.Players[0].Ac, party.Players[0].Hp, party.Players[0].Fort, party.Players[0].Ref, party.Players[0].Will, party.Players[0].Perception, party.Players[0].ID, party.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock second player update (no rows affected)
	mockDB.Mock.ExpectExec("UPDATE players SET name = \\$1, level = \\$2, ac = \\$3, hp = \\$4, fort = \\$5, ref = \\$6, will = \\$7, perception = \\$8 WHERE id = \\$9 AND party_id = \\$10").
		WithArgs(party.Players[1].Name, party.Players[1].Level, party.Players[1].Ac, party.Players[1].Hp, party.Players[1].Fort, party.Players[1].Ref, party.Players[1].Will, party.Players[1].Perception, party.Players[1].ID, party.ID).
		WillReturnResult(sqlmock.NewResult(1, 0))

	// Mock transaction rollback
	mockDB.Mock.ExpectRollback()

	err := party.UpdateWithPlayers(mockDB, []int{})
	if err == nil {
		t.Error("expected error when player not found, got nil")
	}

	expectedError := "player 2 not found or not associated with party"
	if err.Error() != expectedError {
		t.Errorf("expected error message '%s', got '%s'", expectedError, err.Error())
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}
