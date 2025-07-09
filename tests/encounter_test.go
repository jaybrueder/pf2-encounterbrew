package tests

import (
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"pf2.encounterbrew.com/internal/models"
)

// Test data helpers - use consolidated fixtures from mock_database.go

// CreateEncounter Tests

func TestCreateEncounter_Success(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	encounterName := "Test Encounter"
	partyID := 1
	expectedEncounterID := 1

	// Mock encounter creation
	mockDB.Mock.ExpectQuery("INSERT INTO encounters").
		WithArgs(encounterName, partyID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(expectedEncounterID))

	// Mock player query
	playerRows := sqlmock.NewRows([]string{"id", "hp"}).
		AddRow(1, 25).
		AddRow(2, 30)
	mockDB.Mock.ExpectQuery("SELECT id, hp FROM players WHERE party_id = \\$1").
		WithArgs(partyID).
		WillReturnRows(playerRows)

	// Mock player insertions
	mockDB.Mock.ExpectExec("INSERT INTO encounter_players").
		WithArgs(expectedEncounterID, 1, 0, 25).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mockDB.Mock.ExpectExec("INSERT INTO encounter_players").
		WithArgs(expectedEncounterID, 2, 0, 30).
		WillReturnResult(sqlmock.NewResult(2, 1))

	encounter, err := models.CreateEncounter(mockDB, encounterName, partyID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if encounter.ID != expectedEncounterID {
		t.Errorf("expected encounter ID %d, got %d", expectedEncounterID, encounter.ID)
	}

	if encounter.Name != encounterName {
		t.Errorf("expected encounter name '%s', got '%s'", encounterName, encounter.Name)
	}

	if encounter.PartyID != partyID {
		t.Errorf("expected party ID %d, got %d", partyID, encounter.PartyID)
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestCreateEncounter_NilDatabase(t *testing.T) {
	encounter, err := models.CreateEncounter(nil, "Test", 1)
	if err == nil {
		t.Error("expected error when database is nil, got nil")
	}

	if encounter.ID != 0 {
		t.Error("expected empty encounter when database is nil")
	}

	expectedErrorMsg := "database service is nil"
	if err.Error() != expectedErrorMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}
}

func TestCreateEncounter_InsertError(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	mockDB.Mock.ExpectQuery("INSERT INTO encounters").
		WithArgs("Test", 1, 1).
		WillReturnError(sql.ErrConnDone)

	encounter, err := models.CreateEncounter(mockDB, "Test", 1)
	if err == nil {
		t.Error("expected error when insert fails, got nil")
	}

	if encounter.ID != 0 {
		t.Error("expected empty encounter when insert fails")
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestCreateEncounter_PlayerQueryError(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	mockDB.Mock.ExpectQuery("INSERT INTO encounters").
		WithArgs("Test", 1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mockDB.Mock.ExpectQuery("SELECT id, hp FROM players WHERE party_id = \\$1").
		WithArgs(1).
		WillReturnError(sql.ErrConnDone)

	_, err := models.CreateEncounter(mockDB, "Test", 1)
	if err == nil {
		t.Error("expected error when player query fails, got nil")
	}

	if !strings.Contains(err.Error(), "error getting party players") {
		t.Errorf("expected player query error, got: %v", err)
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// UpdateEncounter Tests

func TestUpdateEncounter_Success_NameOnly(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	encounterID := 1
	newName := "Updated Encounter"
	partyID := 1

	// Mock party existence check
	mockDB.Mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM parties WHERE id = \\$1\\)").
		WithArgs(partyID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Mock current party ID query
	mockDB.Mock.ExpectQuery("SELECT party_id FROM encounters WHERE id = \\$1").
		WithArgs(encounterID).
		WillReturnRows(sqlmock.NewRows([]string{"party_id"}).AddRow(partyID))

	// Mock encounter update (name only, same party)
	mockDB.Mock.ExpectExec("UPDATE encounters SET name = \\$1 WHERE id = \\$2").
		WithArgs(newName, encounterID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := models.UpdateEncounter(mockDB, encounterID, newName, partyID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestUpdateEncounter_Success_PartyChange(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	encounterID := 1
	newName := "Updated Encounter"
	newPartyID := 2
	oldPartyID := 1

	// Mock party existence check
	mockDB.Mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM parties WHERE id = \\$1\\)").
		WithArgs(newPartyID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Mock current party ID query
	mockDB.Mock.ExpectQuery("SELECT party_id FROM encounters WHERE id = \\$1").
		WithArgs(encounterID).
		WillReturnRows(sqlmock.NewRows([]string{"party_id"}).AddRow(oldPartyID))

	// Mock transaction
	mockDB.Mock.ExpectBegin()

	// Mock deleting existing players
	mockDB.Mock.ExpectExec("DELETE FROM encounter_players WHERE encounter_id = \\$1").
		WithArgs(encounterID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock new party players query
	playerRows := sqlmock.NewRows([]string{"id", "hp"}).
		AddRow(3, 35).
		AddRow(4, 40)
	mockDB.Mock.ExpectQuery("SELECT id, hp FROM players WHERE party_id = \\$1").
		WithArgs(newPartyID).
		WillReturnRows(playerRows)

	// Mock adding new players
	mockDB.Mock.ExpectExec("INSERT INTO encounter_players").
		WithArgs(encounterID, 3, 0, 35).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mockDB.Mock.ExpectExec("INSERT INTO encounter_players").
		WithArgs(encounterID, 4, 0, 40).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock encounter update
	mockDB.Mock.ExpectExec("UPDATE encounters SET name = \\$1, party_id = \\$2 WHERE id = \\$3").
		WithArgs(newName, newPartyID, encounterID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock transaction commit
	mockDB.Mock.ExpectCommit()

	err := models.UpdateEncounter(mockDB, encounterID, newName, newPartyID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestUpdateEncounter_NilDatabase(t *testing.T) {
	err := models.UpdateEncounter(nil, 1, "Test", 1)
	if err == nil {
		t.Error("expected error when database is nil, got nil")
	}

	expectedErrorMsg := "database service is nil"
	if err.Error() != expectedErrorMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}
}

func TestUpdateEncounter_PartyNotExists(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	mockDB.Mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM parties WHERE id = \\$1\\)").
		WithArgs(999).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	err := models.UpdateEncounter(mockDB, 1, "Test", 999)
	if err == nil {
		t.Error("expected error when party doesn't exist, got nil")
	}

	expectedError := "party with ID 999 does not exist"
	if err.Error() != expectedError {
		t.Errorf("expected error message '%s', got '%s'", expectedError, err.Error())
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// UpdateTurnAndRound Tests

func TestUpdateTurnAndRound_Success(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	encounterID := 1
	turn := 2
	round := 3

	mockDB.Mock.ExpectExec("UPDATE encounters SET turn = \\$1, round = \\$2 WHERE id = \\$3").
		WithArgs(turn, round, encounterID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := models.UpdateTurnAndRound(mockDB, turn, round, encounterID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestUpdateTurnAndRound_InvalidEncounterID(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	err := models.UpdateTurnAndRound(mockDB, 1, 1, 0)
	if err == nil {
		t.Error("expected error for invalid encounter ID, got nil")
	}

	expectedError := "invalid encounter ID"
	if err.Error() != expectedError {
		t.Errorf("expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

func TestUpdateTurnAndRound_InvalidTurnOrRound(t *testing.T) {
	tests := []struct {
		name  string
		turn  int
		round int
	}{
		{"negative turn", -1, 1},
		{"negative round", 1, -1},
		{"both negative", -1, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, cleanup := NewStandardMockDB(t)
			defer cleanup()

			err := models.UpdateTurnAndRound(mockDB, tt.turn, tt.round, 1)
			if err == nil {
				t.Error("expected error for invalid turn or round, got nil")
			}

			expectedError := "invalid turn or round"
			if err.Error() != expectedError {
				t.Errorf("expected error message '%s', got '%s'", expectedError, err.Error())
			}
		})
	}
}

// DeleteEncounter Tests
//
// NOTE: These tests use sqlmock which cannot simulate real database constraints
// like foreign keys or CASCADE deletes. In a real database, deleting an encounter
// that has related records in combatant_conditions (without CASCADE) would fail
// with a foreign key constraint violation. This limitation of sqlmock means that
// CASCADE delete behavior cannot be properly tested without an integration test
// against a real database.

func TestDeleteEncounter_Success(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	encounterID := 1

	// In the real implementation after migration 000013, the database will
	// automatically CASCADE delete related records in:
	// - encounter_monsters (has CASCADE)
	// - encounter_players (has CASCADE)
	// - combatant_conditions (has CASCADE after migration 000013)
	mockDB.Mock.ExpectExec("DELETE FROM encounters WHERE id = \\$1").
		WithArgs(encounterID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := models.DeleteEncounter(mockDB, encounterID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestDeleteEncounter_NilDatabase(t *testing.T) {
	err := models.DeleteEncounter(nil, 1)
	if err == nil {
		t.Error("expected error when database is nil, got nil")
	}

	expectedErrorMsg := "database service is nil"
	if err.Error() != expectedErrorMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}
}

// GetEncounter Tests

func TestGetEncounter_Success(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	encounterID := 1

	// Mock main encounter query
	encounterRows := sqlmock.NewRows([]string{"id", "name", "user_id", "party_id", "turn", "round", "user_name", "party_name"}).
		AddRow(1, "Test Encounter", 1, 1, 0, 1, "Test User", "Test Party")
	mockDB.Mock.ExpectQuery("SELECT e.id, e.name, e.user_id, e.party_id, e.turn, e.round, u.name AS user_name, p.name AS party_name").
		WithArgs(1, encounterID).
		WillReturnRows(encounterRows)

	// Mock monsters query (empty)
	mockDB.Mock.ExpectQuery("SELECT m.id, m.data, em.level_adjustment, em.id, em.initiative, em.hp as current_hp, em.enumeration").
		WithArgs(encounterID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "data", "level_adjustment", "id", "initiative", "current_hp", "enumeration"}))

	// Mock players query
	playerRows := sqlmock.NewRows([]string{"id", "name", "level", "hp", "ac", "fort", "ref", "will", "initiative", "association_id", "current_hp"}).
		AddRow(1, "Test Player", 5, 25, 18, 8, 6, 7, 12, 100, 25)
	mockDB.Mock.ExpectQuery("SELECT p.id, p.name, p.level, p.hp, p.ac, p.fort, p.ref, p.will, ep.initiative, ep.id as association_id, ep.hp as current_hp").
		WithArgs(encounterID).
		WillReturnRows(playerRows)

	encounter, err := models.GetEncounter(mockDB, encounterID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if encounter.ID != encounterID {
		t.Errorf("expected encounter ID %d, got %d", encounterID, encounter.ID)
	}

	if encounter.Name != "Test Encounter" {
		t.Errorf("expected encounter name 'Test Encounter', got '%s'", encounter.Name)
	}

	if len(encounter.Players) != 1 {
		t.Errorf("expected 1 player, got %d", len(encounter.Players))
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestGetEncounter_NotFound(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	encounterID := 999

	mockDB.Mock.ExpectQuery("SELECT e.id, e.name, e.user_id, e.party_id, e.turn, e.round, u.name AS user_name, p.name AS party_name").
		WithArgs(1, encounterID).
		WillReturnError(sql.ErrNoRows)

	encounter, err := models.GetEncounter(mockDB, encounterID)
	if err == nil {
		t.Error("expected error when encounter not found, got nil")
	}

	if encounter.ID != 0 {
		t.Error("expected empty encounter when not found")
	}

	expectedError := "no encounter found with ID 999"
	if err.Error() != expectedError {
		t.Errorf("expected error message '%s', got '%s'", expectedError, err.Error())
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestGetEncounter_NilDatabase(t *testing.T) {
	encounter, err := models.GetEncounter(nil, 1)
	if err == nil {
		t.Error("expected error when database is nil, got nil")
	}

	if encounter.ID != 0 {
		t.Error("expected empty encounter when database is nil")
	}

	expectedErrorMsg := "database service is nil"
	if err.Error() != expectedErrorMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}
}

// GetAllEncounters Tests

func TestGetAllEncounters_Success(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	encounterRows := sqlmock.NewRows([]string{"id", "name", "user_id", "party_id", "user_name", "party_name"}).
		AddRow(1, "Encounter 1", 1, 1, "Test User", "Party 1").
		AddRow(2, "Encounter 2", 1, 2, "Test User", "Party 2")

	mockDB.Mock.ExpectQuery("SELECT e.id, e.name, e.user_id, e.party_id, u.name AS user_name, p.name AS party_name").
		WithArgs(1).
		WillReturnRows(encounterRows)

	encounters, err := models.GetAllEncounters(mockDB)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(encounters) != 2 {
		t.Errorf("expected 2 encounters, got %d", len(encounters))
	}

	if encounters[0].Name != "Encounter 1" {
		t.Errorf("expected first encounter name 'Encounter 1', got '%s'", encounters[0].Name)
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestGetAllEncounters_NilDatabase(t *testing.T) {
	encounters, err := models.GetAllEncounters(nil)
	if err == nil {
		t.Error("expected error when database is nil, got nil")
	}

	if encounters != nil {
		t.Error("expected nil encounters when database is nil")
	}

	expectedErrorMsg := "database service is nil"
	if err.Error() != expectedErrorMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}
}

// AddMonsterToEncounter Tests

func TestAddMonsterToEncounter_Success(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	encounterID := 1
	monsterID := 1
	levelAdjustment := 0
	initiative := 15

	monster := CreateSampleMonster()
	jsonData, _ := json.Marshal(monster.Data)

	// Mock transaction
	mockDB.Mock.ExpectBegin()

	// Mock monster data query
	mockDB.Mock.ExpectQuery("SELECT data FROM monsters WHERE id = \\$1").
		WithArgs(monsterID).
		WillReturnRows(sqlmock.NewRows([]string{"data"}).AddRow(jsonData))

	// Mock max enumeration query
	mockDB.Mock.ExpectQuery("SELECT COALESCE\\(MAX\\(em.enumeration\\), 0\\)").
		WithArgs(encounterID, "Test Monster").
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(0))

	// Mock monster insertion
	mockDB.Mock.ExpectExec("INSERT INTO encounter_monsters").
		WithArgs(encounterID, monsterID, levelAdjustment, initiative, 35, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock transaction commit
	mockDB.Mock.ExpectCommit()

	// Mock GetEncounter call
	encounterRows := sqlmock.NewRows([]string{"id", "name", "user_id", "party_id", "turn", "round", "user_name", "party_name"}).
		AddRow(1, "Test Encounter", 1, 1, 0, 1, "Test User", "Test Party")
	mockDB.Mock.ExpectQuery("SELECT e.id, e.name, e.user_id, e.party_id, e.turn, e.round, u.name AS user_name, p.name AS party_name").
		WithArgs(1, encounterID).
		WillReturnRows(encounterRows)

	// Mock monsters query for GetEncounter
	mockDB.Mock.ExpectQuery("SELECT m.id, m.data, em.level_adjustment, em.id, em.initiative, em.hp as current_hp, em.enumeration").
		WithArgs(encounterID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "data", "level_adjustment", "id", "initiative", "current_hp", "enumeration"}))

	// Mock players query for GetEncounter
	mockDB.Mock.ExpectQuery("SELECT p.id, p.name, p.level, p.hp, p.ac, p.fort, p.ref, p.will, ep.initiative, ep.id as association_id, ep.hp as current_hp").
		WithArgs(encounterID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "level", "hp", "ac", "fort", "ref", "will", "initiative", "association_id", "current_hp"}))

	encounter, err := models.AddMonsterToEncounter(mockDB, encounterID, monsterID, levelAdjustment, initiative)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if encounter.ID != encounterID {
		t.Errorf("expected encounter ID %d, got %d", encounterID, encounter.ID)
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestAddMonsterToEncounter_TransactionError(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	mockDB.Mock.ExpectBegin().WillReturnError(sql.ErrConnDone)

	encounter, err := models.AddMonsterToEncounter(mockDB, 1, 1, 0, 15)
	if err == nil {
		t.Error("expected error when transaction fails, got nil")
	}

	if encounter.ID != 0 {
		t.Error("expected empty encounter when transaction fails")
	}

	if !strings.Contains(err.Error(), "Error starting transaction") {
		t.Errorf("expected transaction error, got: %v", err)
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// RemoveMonsterFromEncounter Tests

func TestRemoveMonsterFromEncounter_Success(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	encounterID := 1
	associationID := 100

	// Mock transaction
	mockDB.Mock.ExpectBegin()

	// Mock monster deletion
	mockDB.Mock.ExpectExec("DELETE FROM encounter_monsters WHERE id = \\$1").
		WithArgs(associationID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock transaction commit
	mockDB.Mock.ExpectCommit()

	err := models.RemoveMonsterFromEncounter(mockDB, encounterID, associationID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// RemovePlayerFromEncounter Tests

func TestRemovePlayerFromEncounter_Success(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	encounterID := 1
	associationID := 100

	// Mock transaction
	mockDB.Mock.ExpectBegin()

	// Mock player deletion
	mockDB.Mock.ExpectExec("DELETE FROM encounter_players WHERE id = \\$1").
		WithArgs(associationID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock transaction commit
	mockDB.Mock.ExpectCommit()

	err := models.RemovePlayerFromEncounter(mockDB, encounterID, associationID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// Encounter Methods Tests

func TestEncounter_GetPartyName(t *testing.T) {
	encounter := CreateSampleEncounter()
	partyName := encounter.GetPartyName()
	if partyName != "Test Party" {
		t.Errorf("expected party name 'Test Party', got '%s'", partyName)
	}
}

func TestEncounter_GetDifficulty(t *testing.T) {
	encounter := CreateSampleEncounter()

	// Add players
	player1 := models.Player{ID: 1, Level: 5}
	player2 := models.Player{ID: 2, Level: 5}
	encounter.Players = []*models.Player{&player1, &player2}

	// Add combatants (for calculation)
	encounter.Combatants = []models.Combatant{&player1, &player2}

	difficulty := encounter.GetDifficulty()
	if difficulty != 0 { // No monsters = trivial
		t.Errorf("expected difficulty 0 (trivial), got %d", difficulty)
	}
}

// GetCombatantConditions Tests

func TestGetCombatantConditions_Success_Monster(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	encounterID := 1
	associationID := 100
	isMonster := true

	conditionData := map[string]interface{}{
		"name": "Test Condition",
		"system": map[string]interface{}{
			"value": map[string]interface{}{
				"value": 0,
			},
		},
	}
	jsonData, _ := json.Marshal(conditionData)

	conditionRows := sqlmock.NewRows([]string{"id", "data", "condition_value"}).
		AddRow(1, jsonData, 5)

	mockDB.Mock.ExpectQuery("SELECT c.id, c.data, cc.condition_value FROM combatant_conditions cc JOIN conditions c ON cc.condition_id = c.id WHERE cc.encounter_id = \\$1 AND cc.encounter_monster_id = \\$2").
		WithArgs(encounterID, associationID).
		WillReturnRows(conditionRows)

	conditions, err := models.GetCombatantConditions(mockDB, encounterID, associationID, isMonster)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(conditions) != 1 {
		t.Errorf("expected 1 condition, got %d", len(conditions))
	}

	if conditions[0].Data.System.Value.Value != 5 {
		t.Errorf("expected condition value 5, got %d", conditions[0].Data.System.Value.Value)
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestGetCombatantConditions_Success_Player(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	encounterID := 1
	associationID := 100
	isMonster := false

	conditionRows := sqlmock.NewRows([]string{"id", "data", "condition_value"})

	mockDB.Mock.ExpectQuery("SELECT c.id, c.data, cc.condition_value FROM combatant_conditions cc JOIN conditions c ON cc.condition_id = c.id WHERE cc.encounter_id = \\$1 AND cc.encounter_player_id = \\$2").
		WithArgs(encounterID, associationID).
		WillReturnRows(conditionRows)

	conditions, err := models.GetCombatantConditions(mockDB, encounterID, associationID, isMonster)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(conditions) != 0 {
		t.Errorf("expected 0 conditions, got %d", len(conditions))
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestGetCombatantConditions_QueryError(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	mockDB.Mock.ExpectQuery("SELECT c.id, c.data, cc.condition_value FROM combatant_conditions cc JOIN conditions c ON cc.condition_id = c.id WHERE cc.encounter_id = \\$1 AND cc.encounter_monster_id = \\$2").
		WithArgs(1, 100).
		WillReturnError(sql.ErrConnDone)

	conditions, err := models.GetCombatantConditions(mockDB, 1, 100, true)
	if err == nil {
		t.Error("expected error when query fails, got nil")
	}

	if conditions != nil {
		t.Error("expected nil conditions when query fails")
	}

	if !strings.Contains(err.Error(), "error querying combatant conditions") {
		t.Errorf("expected query error, got: %v", err)
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// GetEncounterWithCombatants Tests

func TestGetEncounterWithCombatants_Success(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	encounterID := 1

	// Mock GetEncounter call
	encounterRows := sqlmock.NewRows([]string{"id", "name", "user_id", "party_id", "turn", "round", "user_name", "party_name"}).
		AddRow(1, "Test Encounter", 1, 1, 0, 1, "Test User", "Test Party")
	mockDB.Mock.ExpectQuery("SELECT e.id, e.name, e.user_id, e.party_id, e.turn, e.round, u.name AS user_name, p.name AS party_name").
		WithArgs(1, encounterID).
		WillReturnRows(encounterRows)

	// Mock monsters query (empty)
	mockDB.Mock.ExpectQuery("SELECT m.id, m.data, em.level_adjustment, em.id, em.initiative, em.hp as current_hp, em.enumeration").
		WithArgs(encounterID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "data", "level_adjustment", "id", "initiative", "current_hp", "enumeration"}))

	// Mock players query
	playerRows := sqlmock.NewRows([]string{"id", "name", "level", "hp", "ac", "fort", "ref", "will", "initiative", "association_id", "current_hp"}).
		AddRow(1, "Test Player", 5, 25, 18, 8, 6, 7, 12, 100, 25)
	mockDB.Mock.ExpectQuery("SELECT p.id, p.name, p.level, p.hp, p.ac, p.fort, p.ref, p.will, ep.initiative, ep.id as association_id, ep.hp as current_hp").
		WithArgs(encounterID).
		WillReturnRows(playerRows)

	// Mock condition queries for each combatant
	mockDB.Mock.ExpectQuery("SELECT c.id, c.data, cc.condition_value FROM combatant_conditions cc JOIN conditions c ON cc.condition_id = c.id WHERE cc.encounter_id = \\$1 AND cc.encounter_player_id = \\$2").
		WithArgs(encounterID, 100).
		WillReturnRows(sqlmock.NewRows([]string{"id", "data", "condition_value"}))

	encounter, err := models.GetEncounterWithCombatants(mockDB, encounterID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if encounter.ID != encounterID {
		t.Errorf("expected encounter ID %d, got %d", encounterID, encounter.ID)
	}

	if len(encounter.Combatants) != 1 {
		t.Errorf("expected 1 combatant, got %d", len(encounter.Combatants))
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestGetEncounterWithCombatants_GetEncounterError(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	encounterID := 999

	mockDB.Mock.ExpectQuery("SELECT e.id, e.name, e.user_id, e.party_id, e.turn, e.round, u.name AS user_name, p.name AS party_name").
		WithArgs(1, encounterID).
		WillReturnError(sql.ErrNoRows)

	encounter, err := models.GetEncounterWithCombatants(mockDB, encounterID)
	if err == nil {
		t.Error("expected error when GetEncounter fails, got nil")
	}

	if encounter.ID != 0 {
		t.Error("expected empty encounter when GetEncounter fails")
	}

	if !strings.Contains(err.Error(), "error fetching encounter") {
		t.Errorf("expected GetEncounter error, got: %v", err)
	}

	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}
