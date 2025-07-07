package tests

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"pf2.encounterbrew.com/internal/models"
)

// mockCombatantDatabaseService implements the database.Service interface for testing
type mockCombatantDatabaseService struct {
	db   *sql.DB
	mock sqlmock.Sqlmock
}

func (m *mockCombatantDatabaseService) Health() map[string]string {
	return make(map[string]string)
}

func (m *mockCombatantDatabaseService) Close() error {
	return m.db.Close()
}

func (m *mockCombatantDatabaseService) Insert(table string, columns []string, values ...interface{}) (sql.Result, error) {
	return nil, nil
}

func (m *mockCombatantDatabaseService) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return m.db.Query(query, args...)
}

func (m *mockCombatantDatabaseService) QueryRow(query string, args ...interface{}) *sql.Row {
	return m.db.QueryRow(query, args...)
}

func (m *mockCombatantDatabaseService) Exec(query string, args ...interface{}) (sql.Result, error) {
	return m.db.Exec(query, args...)
}

func (m *mockCombatantDatabaseService) Begin() (*sql.Tx, error) {
	return nil, nil
}

func (m *mockCombatantDatabaseService) InsertReturningID(table string, columns []string, values ...interface{}) (int, error) {
	return 0, nil
}

func setupCombatantMockDB(t *testing.T) (*mockCombatantDatabaseService, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	mockService := &mockCombatantDatabaseService{
		db:   db,
		mock: mock,
	}

	return mockService, func() {
		db.Close()
	}
}

// Test data helpers

func createSamplePlayer() models.Player {
	return models.Player{
		ID:            1,
		AssociationID: 100,
		Name:          "Test Player",
		Level:         5,
		Hp:            45,
		Ac:            18,
		Fort:          8,
		Ref:           6,
		Will:          7,
		Perception:    5,
		PartyID:       1,
		Initiative:    12,
		Conditions:    []models.Condition{},
		Enumeration:   1,
	}
}

func createSampleMonster() models.Monster {
	var monster models.Monster
	monster.ID = 1
	monster.AssociationID = 200
	monster.LevelAdjustment = 0
	monster.Enumeration = 1
	monster.Initiative = 15
	monster.Conditions = []models.Condition{}

	// Set up monster data structure
	monster.Data.Name = "Test Monster"
	monster.Data.System.Details.Level.Value = 3
	monster.Data.System.Attributes.Hp.Value = 35
	monster.Data.System.Attributes.Hp.Max = 35
	monster.Data.System.Attributes.Ac.Value = 16
	monster.Data.System.Attributes.Ac.Details = "natural armor"
	monster.Data.System.Perception.Mod = 8
	monster.Data.System.Abilities.Str.Mod = 4
	monster.Data.System.Abilities.Dex.Mod = 2
	monster.Data.System.Abilities.Con.Mod = 3
	monster.Data.System.Abilities.Int.Mod = -1
	monster.Data.System.Abilities.Wis.Mod = 1
	monster.Data.System.Abilities.Cha.Mod = 0
	monster.Data.System.Saves.Fortitude.Value = 12
	monster.Data.System.Saves.Reflex.Value = 8
	monster.Data.System.Saves.Will.Value = 6
	monster.Data.System.Traits.Size.Value = "Medium"
	monster.Data.System.Traits.Value = []string{"humanoid"}
	monster.Data.System.Attributes.Speed.Value = 25

	return monster
}

// Combatant Interface Compliance Tests

func TestPlayer_ImplementsCombatant(t *testing.T) {
	var _ models.Combatant = &models.Player{}
	var _ models.Combatant = (*models.Player)(nil)
}

func TestMonster_ImplementsCombatant(t *testing.T) {
	var _ models.Combatant = &models.Monster{}
	var _ models.Combatant = (*models.Monster)(nil)
}

// Player Tests

func TestPlayer_GetName(t *testing.T) {
	player := createSamplePlayer()
	name := player.GetName()
	if name != "Test Player" {
		t.Errorf("expected name 'Test Player', got '%s'", name)
	}
}

func TestPlayer_GetType(t *testing.T) {
	player := createSamplePlayer()
	playerType := player.GetType()
	if playerType != "player" {
		t.Errorf("expected type 'player', got '%s'", playerType)
	}
}

func TestPlayer_GetInitiative(t *testing.T) {
	player := createSamplePlayer()
	initiative := player.GetInitiative()
	if initiative != 12 {
		t.Errorf("expected initiative 12, got %d", initiative)
	}
}

func TestPlayer_SetInitiative_Success(t *testing.T) {
	mockService, cleanup := setupCombatantMockDB(t)
	defer cleanup()

	player := createSamplePlayer()
	newInitiative := 20

	mockService.mock.ExpectExec("UPDATE encounter_players").
		WithArgs(newInitiative, player.AssociationID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := player.SetInitiative(mockService, newInitiative)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if player.GetInitiative() != newInitiative {
		t.Errorf("expected initiative %d, got %d", newInitiative, player.GetInitiative())
	}

	if err := mockService.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestPlayer_SetInitiative_DatabaseError(t *testing.T) {
	mockService, cleanup := setupCombatantMockDB(t)
	defer cleanup()

	player := createSamplePlayer()
	newInitiative := 20

	mockService.mock.ExpectExec("UPDATE encounter_players").
		WithArgs(newInitiative, player.AssociationID).
		WillReturnError(sql.ErrConnDone)

	err := player.SetInitiative(mockService, newInitiative)
	if err == nil {
		t.Error("expected error when database fails, got nil")
	}

	if !strings.Contains(err.Error(), "error updating player initiative") {
		t.Errorf("expected initiative update error, got: %v", err)
	}

	if err := mockService.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestPlayer_GetHp(t *testing.T) {
	player := createSamplePlayer()
	hp := player.GetHp()
	if hp != 45 {
		t.Errorf("expected HP 45, got %d", hp)
	}
}

func TestPlayer_SetHp_Success(t *testing.T) {
	mockService, cleanup := setupCombatantMockDB(t)
	defer cleanup()

	player := createSamplePlayer()
	damage := 10
	expectedNewHp := player.Hp - damage

	mockService.mock.ExpectExec("UPDATE encounter_players").
		WithArgs(expectedNewHp, player.ID, player.AssociationID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := player.SetHp(mockService, damage)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if player.GetHp() != expectedNewHp {
		t.Errorf("expected HP %d, got %d", expectedNewHp, player.GetHp())
	}

	if err := mockService.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestPlayer_GetAc(t *testing.T) {
	player := createSamplePlayer()
	ac := player.GetAc()
	if ac != 18 {
		t.Errorf("expected AC 18, got %d", ac)
	}
}

func TestPlayer_GetLevel(t *testing.T) {
	player := createSamplePlayer()
	level := player.GetLevel()
	if level != 5 {
		t.Errorf("expected level 5, got %d", level)
	}
}

func TestPlayer_IsMonster(t *testing.T) {
	player := createSamplePlayer()
	if player.IsMonster() {
		t.Error("expected player.IsMonster() to return false")
	}
}

func TestPlayer_GenerateInitiative(t *testing.T) {
	player := createSamplePlayer()
	initiative := player.GenerateInitiative()
	
	// Initiative should be between 1+perception (6) and 20+perception (25)
	expectedMin := 1 + player.GetPerceptionMod()
	expectedMax := 20 + player.GetPerceptionMod()
	
	if initiative < expectedMin || initiative > expectedMax {
		t.Errorf("expected initiative between %d and %d, got %d", expectedMin, expectedMax, initiative)
	}
}

func TestPlayer_GetAssociationID(t *testing.T) {
	player := createSamplePlayer()
	associationID := player.GetAssociationID()
	if associationID != 100 {
		t.Errorf("expected association ID 100, got %d", associationID)
	}
}

func TestPlayer_Update_Success(t *testing.T) {
	mockService, cleanup := setupCombatantMockDB(t)
	defer cleanup()

	player := createSamplePlayer()

	mockService.mock.ExpectQuery("UPDATE players").
		WithArgs(player.Name, player.Level, player.Ac, player.Hp, player.Fort, player.Ref, player.Will, player.ID, player.PartyID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(player.ID))

	err := player.Update(mockService)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if err := mockService.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestPlayer_Update_NilDatabase(t *testing.T) {
	player := createSamplePlayer()
	err := player.Update(nil)
	if err == nil {
		t.Error("expected error when database is nil, got nil")
	}

	expectedErrorMsg := "database service is nil"
	if err.Error() != expectedErrorMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}
}

func TestPlayerDelete_Success(t *testing.T) {
	mockService, cleanup := setupCombatantMockDB(t)
	defer cleanup()

	playerID := 1

	mockService.mock.ExpectExec("DELETE FROM players").
		WithArgs(playerID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := models.PlayerDelete(mockService, playerID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if err := mockService.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestPlayerDelete_NilDatabase(t *testing.T) {
	err := models.PlayerDelete(nil, 1)
	if err == nil {
		t.Error("expected error when database is nil, got nil")
	}

	expectedErrorMsg := "database service is nil"
	if err.Error() != expectedErrorMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}
}

// Monster Tests

func TestMonster_GetName(t *testing.T) {
	tests := []struct {
		name            string
		monster         models.Monster
		expectedName    string
	}{
		{
			name:         "basic monster",
			monster:      createSampleMonster(),
			expectedName: "Test Monster 1",
		},
		{
			name: "elite monster",
			monster: func() models.Monster {
				m := createSampleMonster()
				m.LevelAdjustment = 1
				return m
			}(),
			expectedName: "Elite Test Monster 1",
		},
		{
			name: "weak monster",
			monster: func() models.Monster {
				m := createSampleMonster()
				m.LevelAdjustment = -1
				return m
			}(),
			expectedName: "Weak Test Monster 1",
		},
		{
			name: "monster without enumeration",
			monster: func() models.Monster {
				m := createSampleMonster()
				m.Enumeration = 0
				return m
			}(),
			expectedName: "Test Monster",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name := tt.monster.GetName()
			if name != tt.expectedName {
				t.Errorf("expected name '%s', got '%s'", tt.expectedName, name)
			}
		})
	}
}

func TestMonster_GetType(t *testing.T) {
	monster := createSampleMonster()
	monsterType := monster.GetType()
	if monsterType != "monster" {
		t.Errorf("expected type 'monster', got '%s'", monsterType)
	}
}

func TestMonster_GetLevel(t *testing.T) {
	tests := []struct {
		name            string
		monster         models.Monster
		expectedLevel   int
	}{
		{
			name:          "basic monster",
			monster:       createSampleMonster(),
			expectedLevel: 3,
		},
		{
			name: "elite monster",
			monster: func() models.Monster {
				m := createSampleMonster()
				m.LevelAdjustment = 1
				return m
			}(),
			expectedLevel: 4,
		},
		{
			name: "weak monster",
			monster: func() models.Monster {
				m := createSampleMonster()
				m.LevelAdjustment = -1
				return m
			}(),
			expectedLevel: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := tt.monster.GetLevel()
			if level != tt.expectedLevel {
				t.Errorf("expected level %d, got %d", tt.expectedLevel, level)
			}
		})
	}
}

func TestMonster_GetHp(t *testing.T) {
	tests := []struct {
		name        string
		monster     models.Monster
		expectedHp  int
	}{
		{
			name:       "basic monster",
			monster:    createSampleMonster(),
			expectedHp: 35,
		},
		{
			name: "elite monster level 3",
			monster: func() models.Monster {
				m := createSampleMonster()
				m.LevelAdjustment = 1
				return m
			}(),
			expectedHp: 50, // 35 + 15 for elite level 3
		},
		{
			name: "weak monster level 3",
			monster: func() models.Monster {
				m := createSampleMonster()
				m.LevelAdjustment = -1
				return m
			}(),
			expectedHp: 20, // 35 - 15 for weak level 3
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hp := tt.monster.GetHp()
			if hp != tt.expectedHp {
				t.Errorf("expected HP %d, got %d", tt.expectedHp, hp)
			}
		})
	}
}

func TestMonster_GetAc(t *testing.T) {
	monster := createSampleMonster()
	ac := monster.GetAc()
	if ac != 16 {
		t.Errorf("expected AC 16, got %d", ac)
	}
}

func TestMonster_GetAcDetails(t *testing.T) {
	monster := createSampleMonster()
	acDetails := monster.GetAcDetails()
	expected := " natural armor"
	if acDetails != expected {
		t.Errorf("expected AC details '%s', got '%s'", expected, acDetails)
	}
}

func TestMonster_IsMonster(t *testing.T) {
	monster := createSampleMonster()
	if !monster.IsMonster() {
		t.Error("expected monster.IsMonster() to return true")
	}
}

func TestMonster_GetSize(t *testing.T) {
	monster := createSampleMonster()
	size := monster.GetSize()
	if size != "Medium" {
		t.Errorf("expected size 'Medium', got '%s'", size)
	}
}

func TestMonster_GetTraits(t *testing.T) {
	monster := createSampleMonster()
	traits := monster.GetTraits()
	if len(traits) != 1 || traits[0] != "humanoid" {
		t.Errorf("expected traits ['humanoid'], got %v", traits)
	}
}

func TestMonster_GetSpeed(t *testing.T) {
	monster := createSampleMonster()
	speed := monster.GetSpeed()
	expected := "25 feet"
	if speed != expected {
		t.Errorf("expected speed '%s', got '%s'", expected, speed)
	}
}

func TestMonster_SetInitiative_Success(t *testing.T) {
	mockService, cleanup := setupCombatantMockDB(t)
	defer cleanup()

	monster := createSampleMonster()
	newInitiative := 18

	mockService.mock.ExpectExec("UPDATE encounter_monsters").
		WithArgs(newInitiative, monster.AssociationID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := monster.SetInitiative(mockService, newInitiative)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if monster.GetInitiative() != newInitiative {
		t.Errorf("expected initiative %d, got %d", newInitiative, monster.GetInitiative())
	}

	if err := mockService.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestMonster_SetHp_Success(t *testing.T) {
	mockService, cleanup := setupCombatantMockDB(t)
	defer cleanup()

	monster := createSampleMonster()
	damage := 10
	expectedNewHp := monster.Data.System.Attributes.Hp.Value - damage

	mockService.mock.ExpectExec("UPDATE encounter_monsters").
		WithArgs(expectedNewHp, monster.AssociationID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := monster.SetHp(mockService, damage)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if monster.Data.System.Attributes.Hp.Value != expectedNewHp {
		t.Errorf("expected HP %d, got %d", expectedNewHp, monster.Data.System.Attributes.Hp.Value)
	}

	if err := mockService.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestGetMonster_Success(t *testing.T) {
	mockService, cleanup := setupCombatantMockDB(t)
	defer cleanup()

	monsterID := 1
	monster := createSampleMonster()
	jsonData, _ := json.Marshal(monster.Data)

	rows := sqlmock.NewRows([]string{"id", "data"}).
		AddRow(monsterID, jsonData)

	mockService.mock.ExpectQuery("SELECT id, data FROM monsters WHERE id = \\$1").
		WithArgs(monsterID).
		WillReturnRows(rows)

	result, err := models.GetMonster(mockService, monsterID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result.ID != monsterID {
		t.Errorf("expected monster ID %d, got %d", monsterID, result.ID)
	}

	if result.Data.Name != "Test Monster" {
		t.Errorf("expected monster name 'Test Monster', got '%s'", result.Data.Name)
	}

	if err := mockService.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestGetMonster_NotFound(t *testing.T) {
	mockService, cleanup := setupCombatantMockDB(t)
	defer cleanup()

	monsterID := 999

	mockService.mock.ExpectQuery("SELECT id, data FROM monsters WHERE id = \\$1").
		WithArgs(monsterID).
		WillReturnError(sql.ErrNoRows)

	monster, err := models.GetMonster(mockService, monsterID)
	if err == nil {
		t.Error("expected error when monster not found, got nil")
	}

	if monster.ID != 0 {
		t.Error("expected empty monster when not found")
	}

	expectedErrorMsg := "no monster found with ID 999"
	if err.Error() != expectedErrorMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}

	if err := mockService.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestGetMonster_NilDatabase(t *testing.T) {
	monster, err := models.GetMonster(nil, 1)
	if err == nil {
		t.Error("expected error when database is nil, got nil")
	}

	if monster.ID != 0 {
		t.Error("expected empty monster when database is nil")
	}

	expectedErrorMsg := "database service is nil"
	if err.Error() != expectedErrorMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}
}

func TestGetMonster_InvalidJSON(t *testing.T) {
	mockService, cleanup := setupCombatantMockDB(t)
	defer cleanup()

	monsterID := 1
	invalidJSON := []byte(`{"invalid": json}`)

	rows := sqlmock.NewRows([]string{"id", "data"}).
		AddRow(monsterID, invalidJSON)

	mockService.mock.ExpectQuery("SELECT id, data FROM monsters WHERE id = \\$1").
		WithArgs(monsterID).
		WillReturnRows(rows)

	monster, err := models.GetMonster(mockService, monsterID)
	if err == nil {
		t.Error("expected error when JSON is invalid, got nil")
	}

	if monster.ID != 0 {
		t.Error("expected empty monster when JSON is invalid")
	}

	expectedPrefix := "error unmarshaling monster data:"
	if !strings.HasPrefix(err.Error(), expectedPrefix) {
		t.Errorf("expected error message to start with '%s', got '%s'", expectedPrefix, err.Error())
	}

	if err := mockService.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestSearchMonsters_Success(t *testing.T) {
	mockService, cleanup := setupCombatantMockDB(t)
	defer cleanup()

	searchTerm := "Test"
	monster1 := createSampleMonster()
	monster1.Data.Name = "Test Monster 1"
	jsonData1, _ := json.Marshal(monster1.Data)

	monster2 := createSampleMonster()
	monster2.Data.Name = "Test Monster 2"
	jsonData2, _ := json.Marshal(monster2.Data)

	rows := sqlmock.NewRows([]string{"id", "data"}).
		AddRow(1, jsonData1).
		AddRow(2, jsonData2)

	mockService.mock.ExpectQuery("SELECT id, data FROM monsters WHERE LOWER\\(data->>'name'\\) LIKE LOWER\\(\\$1\\) LIMIT 20").
		WithArgs("%"+searchTerm+"%").
		WillReturnRows(rows)

	results, err := models.SearchMonsters(mockService, searchTerm)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 monsters, got %d", len(results))
	}

	if results[0].Data.Name != "Test Monster 1" {
		t.Errorf("expected first monster name 'Test Monster 1', got '%s'", results[0].Data.Name)
	}

	if err := mockService.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestSearchMonsters_DatabaseError(t *testing.T) {
	mockService, cleanup := setupCombatantMockDB(t)
	defer cleanup()

	searchTerm := "Test"
	expectedError := errors.New("database connection failed")

	mockService.mock.ExpectQuery("SELECT id, data FROM monsters WHERE LOWER\\(data->>'name'\\) LIKE LOWER\\(\\$1\\) LIMIT 20").
		WithArgs("%"+searchTerm+"%").
		WillReturnError(expectedError)

	results, err := models.SearchMonsters(mockService, searchTerm)
	if err == nil {
		t.Error("expected error when database fails, got nil")
	}

	if results != nil {
		t.Error("expected nil results when database fails")
	}

	expectedErrorMsg := "database query error:"
	if !strings.Contains(err.Error(), expectedErrorMsg) {
		t.Errorf("expected error message to contain '%s', got '%s'", expectedErrorMsg, err.Error())
	}

	if err := mockService.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestGetAllMonsters_Success(t *testing.T) {
	mockService, cleanup := setupCombatantMockDB(t)
	defer cleanup()

	monster1 := createSampleMonster()
	jsonData1, _ := json.Marshal(monster1.Data)

	monster2 := createSampleMonster()
	monster2.Data.Name = "Another Monster"
	jsonData2, _ := json.Marshal(monster2.Data)

	rows := sqlmock.NewRows([]string{"id", "data"}).
		AddRow(1, jsonData1).
		AddRow(2, jsonData2)

	mockService.mock.ExpectQuery("SELECT id, data FROM monsters").
		WillReturnRows(rows)

	results, err := models.GetAllMonsters(mockService)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 monsters, got %d", len(results))
	}

	if results[0].Data.Name != "Test Monster" {
		t.Errorf("expected first monster name 'Test Monster', got '%s'", results[0].Data.Name)
	}

	if results[1].Data.Name != "Another Monster" {
		t.Errorf("expected second monster name 'Another Monster', got '%s'", results[1].Data.Name)
	}

	if err := mockService.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

// Combatant Interface Tests

func TestSortCombatantsByInitiative(t *testing.T) {
	player1 := createSamplePlayer()
	player1.Initiative = 10

	player2 := createSamplePlayer()
	player2.Initiative = 15

	monster1 := createSampleMonster()
	monster1.Initiative = 12

	combatants := []models.Combatant{&player1, &monster1, &player2}
	
	models.SortCombatantsByInitiative(combatants)

	// Should be sorted by initiative descending: player2 (15), monster1 (12), player1 (10)
	if combatants[0].GetInitiative() != 15 {
		t.Errorf("expected first combatant initiative 15, got %d", combatants[0].GetInitiative())
	}
	if combatants[1].GetInitiative() != 12 {
		t.Errorf("expected second combatant initiative 12, got %d", combatants[1].GetInitiative())
	}
	if combatants[2].GetInitiative() != 10 {
		t.Errorf("expected third combatant initiative 10, got %d", combatants[2].GetInitiative())
	}
}

// Condition Tests for Both Combatants

func TestCombatant_ConditionManagement(t *testing.T) {
	tests := []struct {
		name      string
		combatant models.Combatant
	}{
		{
			name: "player conditions",
			combatant: func() models.Combatant {
				player := createSamplePlayer()
				return &player
			}(),
		},
		{
			name: "monster conditions",
			combatant: func() models.Combatant {
				monster := createSampleMonster()
				return &monster
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test getting empty conditions
			conditions := tt.combatant.GetConditions()
			if len(conditions) != 0 {
				t.Errorf("expected 0 conditions, got %d", len(conditions))
			}

			// Test HasCondition with no conditions
			if tt.combatant.HasCondition(1) {
				t.Error("expected HasCondition to return false for non-existent condition")
			}

			// Test GetConditionValue with no conditions
			value := tt.combatant.GetConditionValue(1)
			if value != 0 {
				t.Errorf("expected condition value 0, got %d", value)
			}

			// Test IsOffGuard with no conditions
			if tt.combatant.IsOffGuard() {
				t.Error("expected IsOffGuard to return false with no conditions")
			}

			// Test SetEnumeration
			tt.combatant.SetEnumeration(5)
			// Note: We can't easily test the result since the interface doesn't have GetEnumeration()
		})
	}
}

func TestCombatant_InterfaceMethods(t *testing.T) {
	player := createSamplePlayer()
	monster := createSampleMonster()

	combatants := []models.Combatant{&player, &monster}

	for i, combatant := range combatants {
		t.Run(fmt.Sprintf("combatant_%d", i), func(t *testing.T) {
			// Test all interface methods don't panic
			_ = combatant.GetName()
			_ = combatant.GetInitiative()
			_ = combatant.GetHp()
			_ = combatant.GetMaxHp()
			_ = combatant.GetAc()
			_ = combatant.GetAcDetails()
			_ = combatant.GetType()
			_ = combatant.GetLevel()
			_ = combatant.GetSize()
			_ = combatant.GetTraits()
			_ = combatant.GetPerceptionMod()
			_ = combatant.GetPerceptionSenses()
			_ = combatant.GetLanguages()
			_ = combatant.GetSkills()
			_ = combatant.GetLores()
			_ = combatant.GetStr()
			_ = combatant.GetDex()
			_ = combatant.GetCon()
			_ = combatant.GetInt()
			_ = combatant.GetWis()
			_ = combatant.GetCha()
			_ = combatant.GetFort()
			_ = combatant.GetRef()
			_ = combatant.GetWill()
			_ = combatant.GetImmunities()
			_ = combatant.GetResistances()
			_ = combatant.GetWeaknesses()
			_ = combatant.GetSpeed()
			_ = combatant.GetOtherSpeeds()
			_ = combatant.GetAttacks()
			_ = combatant.GetSpellSchool()
			_ = combatant.GetSpells()
			_ = combatant.GetDefensiveActions()
			_ = combatant.GetOffensiveActions()
			_ = combatant.GetInteractions()
			_ = combatant.GetInventory()
			_ = combatant.GetConditions()
			_ = combatant.HasCondition(1)
			_ = combatant.GetConditionValue(1)
			_ = combatant.SetConditionValue(1, 5)
			_ = combatant.GetAdjustmentModifier()
			_ = combatant.IsMonster()
			_ = combatant.IsOffGuard()
			_ = combatant.AdjustConditions()
			_ = combatant.GetAssociationID()
			_ = combatant.GenerateInitiative()

			combatant.SetEnumeration(1)
			combatant.SetConditions([]models.Condition{})
		})
	}
}