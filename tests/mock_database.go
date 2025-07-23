package tests

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"pf2.encounterbrew.com/internal/models"
)

// Test constants to avoid duplication across files
const (
	TestUserID        = 1
	TestPartyID       = 1
	TestEncounterID   = 1
	TestMonsterID     = 1
	TestPlayerID      = 1
	TestMonsterName   = "Test Monster"
	TestConditionName = "Test Condition"
	TestConditionID   = 1
	DBServiceNilError = "database service is nil"
)

var ErrNotFound = errors.New("not found")

// GroupedCondition represents a grouped condition for testing
type GroupedCondition struct {
	Category   string             `json:"category"`
	Conditions []models.Condition `json:"conditions"`
}

// MockDatabaseService implements the database.Service interface for testing
type MockDatabaseService struct {
	// Basic database operations
	HealthFunc            func() map[string]string
	CloseFunc             func() error
	InsertFunc            func(table string, columns []string, values ...interface{}) (sql.Result, error)
	QueryFunc             func(query string, args ...interface{}) (*sql.Rows, error)
	QueryRowFunc          func(query string, args ...interface{}) *sql.Row
	ExecFunc              func(query string, args ...interface{}) (sql.Result, error)
	BeginFunc             func() (*sql.Tx, error)
	InsertReturningIDFunc func(table string, columns []string, values ...interface{}) (int, error)

	// Model-specific operations
	GetAllPartiesFunc              func() ([]models.Party, error)
	GetPartyFunc                   func(partyID int) (models.Party, error)
	PartyExistsFunc                func(partyID int) (bool, error)
	UpdateWithPlayersFunc          func(party models.Party, playersToDelete []int) error
	DeleteFunc                     func(party models.Party) error
	PlayerDeleteFunc               func(playerID int) error
	GetAllEncountersFunc           func() ([]models.Encounter, error)
	GetEncounterFunc               func(encounterID int) (models.Encounter, error)
	GetEncounterWithCombatantsFunc func(encounterID int) (models.Encounter, error)
	CreateEncounterFunc            func(name string, partyID int) (int, error)
	UpdateEncounterFunc            func(encounterID int, name string, partyID int) error
	DeleteEncounterFunc            func(encounterID int) error
	SearchMonstersFunc             func(search string) ([]models.Monster, error)
	GetMonsterFunc                 func(monsterID int) (models.Monster, error)
	GetAllMonstersFunc             func() ([]models.Monster, error)
	AddMonsterToEncounterFunc      func(encounterID, monsterID, levelAdjustment, initiative int) (models.Encounter, error)
	RemoveMonsterFromEncounterFunc func(encounterID, associationID int) error
	RemovePlayerFromEncounterFunc  func(encounterID, associationID int) error
	GetGroupedConditionsFunc       func() ([]GroupedCondition, error)
	GetConditionFunc               func(conditionID int) (models.Condition, error)
	UpdateTurnAndRoundFunc         func(turn, round, encounterID int) error

	// Call counters for testing
	GetAllPartiesCallCount              int
	GetPartyCallCount                   int
	PartyExistsCallCount                int
	UpdateWithPlayersCallCount          int
	DeleteCallCount                     int
	PlayerDeleteCallCount               int
	GetAllEncountersCallCount           int
	GetEncounterCallCount               int
	GetEncounterWithCombatantsCallCount int
	CreateEncounterCallCount            int
	UpdateEncounterCallCount            int
	DeleteEncounterCallCount            int
	SearchMonstersCallCount             int
	GetMonsterCallCount                 int
	GetAllMonstersCallCount             int
	AddMonsterToEncounterCallCount      int
	RemoveMonsterFromEncounterCallCount int
	RemovePlayerFromEncounterCallCount  int
	GetGroupedConditionsCallCount       int
	GetConditionCallCount               int
	UpdateTurnAndRoundCallCount         int
}

// Basic database operations
func (m *MockDatabaseService) Health() map[string]string {
	if m.HealthFunc != nil {
		return m.HealthFunc()
	}
	return make(map[string]string)
}

func (m *MockDatabaseService) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func (m *MockDatabaseService) Insert(table string, columns []string, values ...interface{}) (sql.Result, error) {
	if m.InsertFunc != nil {
		return m.InsertFunc(table, columns, values...)
	}
	return nil, nil
}

func (m *MockDatabaseService) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if m.QueryFunc != nil {
		return m.QueryFunc(query, args...)
	}
	return nil, ErrNotFound
}

func (m *MockDatabaseService) QueryRow(query string, args ...interface{}) *sql.Row {
	if m.QueryRowFunc != nil {
		return m.QueryRowFunc(query, args...)
	}
	return nil
}

func (m *MockDatabaseService) Exec(query string, args ...interface{}) (sql.Result, error) {
	if m.ExecFunc != nil {
		return m.ExecFunc(query, args...)
	}
	return nil, nil
}

func (m *MockDatabaseService) Begin() (*sql.Tx, error) {
	if m.BeginFunc != nil {
		return m.BeginFunc()
	}
	return nil, nil
}

func (m *MockDatabaseService) InsertReturningID(table string, columns []string, values ...interface{}) (int, error) {
	if m.InsertReturningIDFunc != nil {
		return m.InsertReturningIDFunc(table, columns, values...)
	}
	return 0, nil
}

// Model-specific operations
func (m *MockDatabaseService) GetAllParties() ([]models.Party, error) {
	m.GetAllPartiesCallCount++
	if m.GetAllPartiesFunc != nil {
		return m.GetAllPartiesFunc()
	}
	return nil, ErrNotFound
}

func (m *MockDatabaseService) GetParty(partyID int) (models.Party, error) {
	m.GetPartyCallCount++
	if m.GetPartyFunc != nil {
		return m.GetPartyFunc(partyID)
	}
	return models.Party{}, ErrNotFound
}

func (m *MockDatabaseService) PartyExists(partyID int) (bool, error) {
	m.PartyExistsCallCount++
	if m.PartyExistsFunc != nil {
		return m.PartyExistsFunc(partyID)
	}
	return false, ErrNotFound
}

func (m *MockDatabaseService) UpdateWithPlayers(party models.Party, playersToDelete []int) error {
	m.UpdateWithPlayersCallCount++
	if m.UpdateWithPlayersFunc != nil {
		return m.UpdateWithPlayersFunc(party, playersToDelete)
	}
	return ErrNotFound
}

func (m *MockDatabaseService) Delete(party models.Party) error {
	m.DeleteCallCount++
	if m.DeleteFunc != nil {
		return m.DeleteFunc(party)
	}
	return ErrNotFound
}

func (m *MockDatabaseService) PlayerDelete(playerID int) error {
	m.PlayerDeleteCallCount++
	if m.PlayerDeleteFunc != nil {
		return m.PlayerDeleteFunc(playerID)
	}
	return ErrNotFound
}

func (m *MockDatabaseService) GetAllEncounters() ([]models.Encounter, error) {
	m.GetAllEncountersCallCount++
	if m.GetAllEncountersFunc != nil {
		return m.GetAllEncountersFunc()
	}
	return nil, ErrNotFound
}

func (m *MockDatabaseService) GetEncounter(encounterID int) (models.Encounter, error) {
	m.GetEncounterCallCount++
	if m.GetEncounterFunc != nil {
		return m.GetEncounterFunc(encounterID)
	}
	return models.Encounter{}, ErrNotFound
}

func (m *MockDatabaseService) GetEncounterWithCombatants(encounterID int) (models.Encounter, error) {
	m.GetEncounterWithCombatantsCallCount++
	if m.GetEncounterWithCombatantsFunc != nil {
		return m.GetEncounterWithCombatantsFunc(encounterID)
	}
	return models.Encounter{}, ErrNotFound
}

func (m *MockDatabaseService) CreateEncounter(name string, partyID int) (int, error) {
	m.CreateEncounterCallCount++
	if m.CreateEncounterFunc != nil {
		return m.CreateEncounterFunc(name, partyID)
	}
	return 0, ErrNotFound
}

func (m *MockDatabaseService) UpdateEncounter(encounterID int, name string, partyID int) error {
	m.UpdateEncounterCallCount++
	if m.UpdateEncounterFunc != nil {
		return m.UpdateEncounterFunc(encounterID, name, partyID)
	}
	return ErrNotFound
}

func (m *MockDatabaseService) DeleteEncounter(encounterID int) error {
	m.DeleteEncounterCallCount++
	if m.DeleteEncounterFunc != nil {
		return m.DeleteEncounterFunc(encounterID)
	}
	return ErrNotFound
}

func (m *MockDatabaseService) SearchMonsters(search string) ([]models.Monster, error) {
	m.SearchMonstersCallCount++
	if m.SearchMonstersFunc != nil {
		return m.SearchMonstersFunc(search)
	}
	return nil, ErrNotFound
}

func (m *MockDatabaseService) GetMonster(monsterID int) (models.Monster, error) {
	m.GetMonsterCallCount++
	if m.GetMonsterFunc != nil {
		return m.GetMonsterFunc(monsterID)
	}
	return models.Monster{}, ErrNotFound
}

func (m *MockDatabaseService) GetAllMonsters() ([]models.Monster, error) {
	m.GetAllMonstersCallCount++
	if m.GetAllMonstersFunc != nil {
		return m.GetAllMonstersFunc()
	}
	return nil, ErrNotFound
}

func (m *MockDatabaseService) AddMonsterToEncounter(encounterID, monsterID, levelAdjustment, initiative int) (models.Encounter, error) {
	m.AddMonsterToEncounterCallCount++
	if m.AddMonsterToEncounterFunc != nil {
		return m.AddMonsterToEncounterFunc(encounterID, monsterID, levelAdjustment, initiative)
	}
	return models.Encounter{}, ErrNotFound
}

func (m *MockDatabaseService) RemoveMonsterFromEncounter(encounterID, associationID int) error {
	m.RemoveMonsterFromEncounterCallCount++
	if m.RemoveMonsterFromEncounterFunc != nil {
		return m.RemoveMonsterFromEncounterFunc(encounterID, associationID)
	}
	return ErrNotFound
}

func (m *MockDatabaseService) RemovePlayerFromEncounter(encounterID, associationID int) error {
	m.RemovePlayerFromEncounterCallCount++
	if m.RemovePlayerFromEncounterFunc != nil {
		return m.RemovePlayerFromEncounterFunc(encounterID, associationID)
	}
	return ErrNotFound
}

func (m *MockDatabaseService) GetGroupedConditions() ([]GroupedCondition, error) {
	m.GetGroupedConditionsCallCount++
	if m.GetGroupedConditionsFunc != nil {
		return m.GetGroupedConditionsFunc()
	}
	return nil, ErrNotFound
}

func (m *MockDatabaseService) GetCondition(conditionID int) (models.Condition, error) {
	m.GetConditionCallCount++
	if m.GetConditionFunc != nil {
		return m.GetConditionFunc(conditionID)
	}
	return models.Condition{}, ErrNotFound
}

func (m *MockDatabaseService) UpdateTurnAndRound(turn, round, encounterID int) error {
	m.UpdateTurnAndRoundCallCount++
	if m.UpdateTurnAndRoundFunc != nil {
		return m.UpdateTurnAndRoundFunc(turn, round, encounterID)
	}
	return ErrNotFound
}

// =============================================
// SQLMOCK-BASED STANDARDIZED MOCK
// =============================================

// StandardMockDB provides a standardized sqlmock-based database service for testing
type StandardMockDB struct {
	DB   *sql.DB
	Mock sqlmock.Sqlmock
}

// NewStandardMockDB creates a new standardized mock database for testing
func NewStandardMockDB(t *testing.T) (*StandardMockDB, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	return &StandardMockDB{
			DB:   db,
			Mock: mock,
		}, func() {
			_ = db.Close()
		}
}

// Implement database.Service interface for StandardMockDB
func (s *StandardMockDB) Health() map[string]string {
	return make(map[string]string)
}

func (s *StandardMockDB) Close() error {
	return s.DB.Close()
}

func (s *StandardMockDB) Insert(table string, columns []string, values ...interface{}) (sql.Result, error) {
	return s.DB.Exec("INSERT", values...)
}

func (s *StandardMockDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return s.DB.Query(query, args...)
}

func (s *StandardMockDB) QueryRow(query string, args ...interface{}) *sql.Row {
	return s.DB.QueryRow(query, args...)
}

func (s *StandardMockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return s.DB.Exec(query, args...)
}

func (s *StandardMockDB) Begin() (*sql.Tx, error) {
	return s.DB.Begin()
}

func (s *StandardMockDB) InsertReturningID(table string, columns []string, values ...interface{}) (int, error) {
	// Build the same query as the real implementation
	// Build the query string with RETURNING id (same as real implementation)
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING id", // #nosec G201 - Test code with controlled inputs
		table,
		strings.Join(columns, ", "),
		strings.Join(strings.Split(strings.Repeat("?", len(columns)), ""), ", "))

	// Replace ? with $1, $2, etc. for PostgreSQL (same as real implementation)
	for i := 1; strings.Contains(query, "?"); i++ {
		query = strings.Replace(query, "?", fmt.Sprintf("$%d", i), 1)
	}

	row := s.DB.QueryRow(query, values...)
	var id int
	err := row.Scan(&id)
	return id, err
}

// =============================================
// TEST HELPER FUNCTIONS
// =============================================

// SetupMockForGetAllParties sets up mock expectations for models.GetAllParties
func (s *StandardMockDB) SetupMockForGetAllParties(parties []models.Party) {
	rows := sqlmock.NewRows([]string{"id", "name", "user_id", "user_name"})
	for _, party := range parties {
		userName := ""
		if party.User != nil {
			userName = party.User.Name
		}
		rows.AddRow(party.ID, party.Name, party.UserID, userName)
	}

	s.Mock.ExpectQuery(`SELECT p\.id, p\.name, p\.user_id, u\.name AS user_name FROM parties p JOIN users u ON p\.user_id = u\.id WHERE p\.user_id = \$1 ORDER BY p\.id`).
		WithArgs(1).
		WillReturnRows(rows)

	// For each party, mock the player query
	for _, party := range parties {
		playerRows := sqlmock.NewRows([]string{"id", "name", "level", "hp", "ac", "fort", "ref", "will"})
		for _, player := range party.Players {
			playerRows.AddRow(player.ID, player.Name, player.Level, player.Hp, player.Ac, player.Fort, player.Ref, player.Will)
		}
		s.Mock.ExpectQuery(`SELECT id, name, level, hp, ac, fort, ref, will FROM players WHERE party_id = \$1`).
			WithArgs(party.ID).
			WillReturnRows(playerRows)
	}
}

// SetupMockForGetParty sets up mock expectations for models.GetParty
func (s *StandardMockDB) SetupMockForGetParty(party models.Party) {
	userName := ""
	userID := 1
	if party.User != nil {
		userName = party.User.Name
		userID = party.User.ID
	}

	rows := sqlmock.NewRows([]string{"id", "name", "user_id", "user_name"}).
		AddRow(party.ID, party.Name, userID, userName)

	s.Mock.ExpectQuery(`SELECT p.id, p.name, p.user_id, u.name AS user_name`).
		WithArgs(1, party.ID).
		WillReturnRows(rows)

	// Mock the players query - include perception column
	playerRows := sqlmock.NewRows([]string{"id", "name", "level", "hp", "ac", "fort", "ref", "will", "perception"})
	for _, player := range party.Players {
		playerRows.AddRow(player.ID, player.Name, player.Level, player.Hp, player.Ac, player.Fort, player.Ref, player.Will, player.Perception)
	}
	s.Mock.ExpectQuery(`SELECT id, name, level, hp, ac, fort, ref, will, perception`).
		WithArgs(party.ID).
		WillReturnRows(playerRows)
}

// SetupMockForPartyExists sets up mock expectations for models.PartyExists
func (s *StandardMockDB) SetupMockForPartyExists(partyID int, exists bool) {
	rows := sqlmock.NewRows([]string{"exists"}).AddRow(exists)
	s.Mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM parties WHERE id = \$1\)`).
		WithArgs(partyID).
		WillReturnRows(rows)
}

// SetupMockForGetAllMonsters sets up mock expectations for models.GetAllMonsters
func (s *StandardMockDB) SetupMockForGetAllMonsters(monsters []models.Monster) {
	rows := sqlmock.NewRows([]string{"id", "data"})
	for _, monster := range monsters {
		data := `{"name":"` + monster.Data.Name + `"}`
		rows.AddRow(monster.ID, data)
	}
	s.Mock.ExpectQuery("SELECT id, data FROM monsters").
		WillReturnRows(rows)
}

// SetupMockForGetAllEncounters sets up mock expectations for models.GetAllEncounters
func (s *StandardMockDB) SetupMockForGetAllEncounters(encounters []models.Encounter) {
	rows := sqlmock.NewRows([]string{"id", "name", "user_id", "party_id", "user_name", "party_name"})
	for _, encounter := range encounters {
		userName := ""
		partyName := ""
		if encounter.User != nil {
			userName = encounter.User.Name
		}
		if encounter.Party != nil {
			partyName = encounter.Party.Name
		}
		rows.AddRow(encounter.ID, encounter.Name, encounter.UserID, encounter.PartyID, userName, partyName)
	}
	s.Mock.ExpectQuery(`SELECT e\.id, e\.name, e\.user_id, e\.party_id, u\.name AS user_name, p\.name AS party_name FROM encounters e JOIN users u ON e\.user_id = u\.id JOIN parties p ON e\.party_id = p\.id WHERE e\.user_id = \$1 ORDER BY e\.id`).
		WithArgs(1).
		WillReturnRows(rows)
}

// SetupMockForCreateEncounter sets up mock expectations for models.CreateEncounter
func (s *StandardMockDB) SetupMockForCreateEncounter(encounterID int, partyID int, players []models.Player) {
	// Mock the encounter creation
	s.Mock.ExpectQuery(`INSERT INTO encounters \(name, party_id, user_id\) VALUES \(\$1, \$2, \$3\) RETURNING id`).
		WithArgs(sqlmock.AnyArg(), partyID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(encounterID))

	// Mock getting players for the party
	playerRows := sqlmock.NewRows([]string{"id", "hp"})
	for _, player := range players {
		playerRows.AddRow(player.ID, player.Hp)
	}
	s.Mock.ExpectQuery(`SELECT id, hp FROM players WHERE party_id = \$1`).
		WithArgs(partyID).
		WillReturnRows(playerRows)

	// Mock inserting each player into the encounter
	for _, player := range players {
		s.Mock.ExpectExec(`INSERT INTO encounter_players \(encounter_id, player_id, initiative, hp\) VALUES \(\$1, \$2, \$3, \$4\)`).
			WithArgs(encounterID, player.ID, 0, player.Hp).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}
}

// SetupMockForGetEncounter sets up mock expectations for models.GetEncounter
func (s *StandardMockDB) SetupMockForGetEncounter(encounter models.Encounter) {
	userName := ""
	partyName := ""
	if encounter.User != nil {
		userName = encounter.User.Name
	}
	if encounter.Party != nil {
		partyName = encounter.Party.Name
	}

	rows := sqlmock.NewRows([]string{"id", "name", "user_id", "party_id", "turn", "round", "user_name", "party_name"}).
		AddRow(encounter.ID, encounter.Name, encounter.UserID, encounter.PartyID, encounter.Turn, encounter.Round, userName, partyName)

	s.Mock.ExpectQuery(`SELECT e\.id, e\.name, e\.user_id, e\.party_id, e\.turn, e\.round, u\.name AS user_name, p\.name AS party_name FROM encounters e JOIN users u ON e\.user_id = u\.id JOIN parties p ON e\.party_id = p\.id WHERE e\.user_id = \$1 AND e\.id = \$2`).
		WithArgs(1, encounter.ID).
		WillReturnRows(rows)
}

// SetupMockForGetEncounterWithCombatants sets up mock expectations for models.GetEncounterWithCombatants
func (s *StandardMockDB) SetupMockForGetEncounterWithCombatants(encounter models.Encounter) {
	// First the encounter query (same as GetEncounter)
	userName := ""
	partyName := ""
	if encounter.User != nil {
		userName = encounter.User.Name
	}
	if encounter.Party != nil {
		partyName = encounter.Party.Name
	}

	rows := sqlmock.NewRows([]string{"id", "name", "user_id", "party_id", "turn", "round", "user_name", "party_name"}).
		AddRow(encounter.ID, encounter.Name, encounter.UserID, encounter.PartyID, encounter.Turn, encounter.Round, userName, partyName)

	s.Mock.ExpectQuery(`SELECT e\.id, e\.name, e\.user_id, e\.party_id, e\.turn, e\.round, u\.name AS user_name, p\.name AS party_name FROM encounters e JOIN users u ON e\.user_id = u\.id JOIN parties p ON e\.party_id = p\.id WHERE e\.user_id = \$1 AND e\.id = \$2`).
		WithArgs(1, encounter.ID).
		WillReturnRows(rows)

	// Mock the monsters query
	monsterRows := sqlmock.NewRows([]string{"id", "data", "level_adjustment", "id", "initiative", "current_hp", "enumeration"})
	s.Mock.ExpectQuery(`SELECT m\.id, m\.data, em\.level_adjustment, em\.id, em\.initiative, em\.hp as current_hp, em\.enumeration FROM monsters m JOIN encounter_monsters em ON m\.id = em\.monster_id WHERE em\.encounter_id = \$1`).
		WithArgs(encounter.ID).
		WillReturnRows(monsterRows)

	// Mock the players query
	playerRows := sqlmock.NewRows([]string{"id", "name", "level", "hp", "ac", "fort", "ref", "will", "initiative", "association_id", "current_hp"})
	s.Mock.ExpectQuery(`SELECT p\.id, p\.name, p\.level, p\.hp, p\.ac, p\.fort, p\.ref, p\.will, ep\.initiative, ep\.id as association_id, ep\.hp as current_hp FROM players p JOIN encounter_players ep ON p\.id = ep\.player_id WHERE ep\.encounter_id = \$1`).
		WithArgs(encounter.ID).
		WillReturnRows(playerRows)
}

// SetupMockForUpdateEncounter sets up mock expectations for models.UpdateEncounter
func (s *StandardMockDB) SetupMockForUpdateEncounter(partyID int, encounterID int, name string) {
	// First expect the party existence check (exact query from the model)
	s.Mock.ExpectQuery("SELECT EXISTS").
		WithArgs(partyID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Then expect the current party ID check
	s.Mock.ExpectQuery("SELECT party_id FROM encounters").
		WithArgs(encounterID).
		WillReturnRows(sqlmock.NewRows([]string{"party_id"}).AddRow(partyID))

	// Since party_id is the same, expect only name update
	s.Mock.ExpectExec("UPDATE encounters SET name").
		WithArgs(name, encounterID).
		WillReturnResult(sqlmock.NewResult(1, 1))
}

// SetupMockForDeleteEncounter sets up mock expectations for models.DeleteEncounter
func (s *StandardMockDB) SetupMockForDeleteEncounter() {
	s.Mock.ExpectExec("DELETE FROM encounters").
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
}

// SetupMockForSearchMonsters sets up mock expectations for models.SearchMonsters
func (s *StandardMockDB) SetupMockForSearchMonsters(monsters []models.Monster) {
	rows := sqlmock.NewRows([]string{"id", "data", "priority", "name_lower"})
	for i, monster := range monsters {
		data := `{"name":"` + monster.Data.Name + `"}`
		// Simulate priority ordering: first monster gets priority 1, others get priority 2 or 3
		priority := i + 1
		if priority > 3 {
			priority = 3
		}
		rows.AddRow(monster.ID, data, priority, strings.ToLower(monster.Data.Name))
	}
	s.Mock.ExpectQuery("WITH search_results AS").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(rows)
}

// SetupMockForGetMonster sets up mock expectations for models.GetMonster
func (s *StandardMockDB) SetupMockForGetMonster(monster models.Monster) {
	data := `{"name":"` + monster.Data.Name + `"}`
	rows := sqlmock.NewRows([]string{"id", "data"}).
		AddRow(monster.ID, data)
	s.Mock.ExpectQuery("SELECT id, data FROM monsters").
		WithArgs(monster.ID).
		WillReturnRows(rows)
}

// SetupMockForAddMonsterToEncounter sets up mock expectations for models.AddMonsterToEncounter
func (s *StandardMockDB) SetupMockForAddMonsterToEncounter(encounter models.Encounter) {
	// 1. Transaction begin (though the code doesn't use it properly)
	s.Mock.ExpectBegin()

	// 2. Get monster data query
	s.Mock.ExpectQuery(`SELECT data FROM monsters WHERE id = \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"data"}).AddRow(`{"name":"` + TestMonsterName + `","system":{"attributes":{"hp":{"max":25,"value":25}}}}`))

	// 3. Get max enumeration query
	s.Mock.ExpectQuery(`SELECT COALESCE\(MAX\(em\.enumeration\), 0\)`).
		WithArgs(encounter.ID, TestMonsterName).
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(0))

	// 4. Insert into encounter_monsters
	s.Mock.ExpectExec(`INSERT INTO encounter_monsters \(encounter_id, monster_id, level_adjustment, initiative, hp, enumeration\) VALUES \(\$1, \$2, \$3, \$4, \$5, \$6\)`).
		WithArgs(encounter.ID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 5. Transaction commit
	s.Mock.ExpectCommit()

	// 6. Mock the return encounter query
	s.SetupMockForGetEncounterWithCombatants(encounter)
}

// SetupMockForPartyCreate sets up mock expectations for Party.Create
func (s *StandardMockDB) SetupMockForPartyCreate(partyID int) {
	s.Mock.ExpectQuery(`INSERT INTO parties \(name, user_id\) VALUES \(\$1, \$2\) RETURNING id`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(partyID))
}

// =============================================
// CONSOLIDATED TEST FIXTURES
// =============================================

// CreateSampleUser creates a sample user for testing
func CreateSampleUser() models.User {
	return models.User{
		ID:            TestUserID,
		Name:          "Test User",
		ActivePartyID: TestPartyID,
	}
}

// CreateSampleParty creates a sample party for testing
func CreateSampleParty() models.Party {
	return models.Party{
		ID:     TestPartyID,
		Name:   "Test Party",
		UserID: TestUserID,
		User: &models.User{
			ID:   TestUserID,
			Name: "Test User",
		},
		Players: []models.Player{
			CreateSamplePlayer(),
			{
				ID:            TestPlayerID + 1,
				AssociationID: 101,
				Name:          "Test Player 2",
				Level:         4,
				Hp:            40,
				Ac:            17,
				Fort:          7,
				Ref:           5,
				Will:          6,
				Perception:    9,
				PartyID:       TestPartyID,
				Initiative:    10,
				Conditions:    []models.Condition{},
				Enumeration:   2,
			},
		},
	}
}

// CreateSamplePlayer creates a sample player for testing
func CreateSamplePlayer() models.Player {
	return models.Player{
		ID:            TestPlayerID,
		AssociationID: 100,
		Name:          "Test Player",
		Level:         5,
		Hp:            45,
		Ac:            18,
		Fort:          8,
		Ref:           6,
		Will:          7,
		Perception:    10,
		PartyID:       TestPartyID,
		Initiative:    12,
		Conditions:    []models.Condition{},
		Enumeration:   1,
	}
}

// CreateSampleMonster creates a sample monster for testing
func CreateSampleMonster() models.Monster {
	monster := models.Monster{
		ID:              TestMonsterID,
		AssociationID:   200,
		LevelAdjustment: 0,
		Enumeration:     1,
		Initiative:      15,
		Conditions:      []models.Condition{},
	}
	monster.Data.Name = TestMonsterName
	monster.Data.System.Details.Level.Value = 3
	monster.Data.System.Attributes.Hp.Value = 35
	monster.Data.System.Attributes.Hp.Max = 35
	monster.Data.System.Attributes.Ac.Value = 16
	monster.Data.System.Attributes.Ac.Details = "natural armor"
	monster.Data.System.Traits.Size.Value = "Medium"
	monster.Data.System.Attributes.Speed.Value = 25
	monster.Data.System.Traits.Value = []string{"humanoid"}
	return monster
}

// CreateSampleEncounter creates a sample encounter for testing
func CreateSampleEncounter() models.Encounter {
	return models.Encounter{
		ID:      TestEncounterID,
		Name:    "Test Encounter",
		UserID:  TestUserID,
		PartyID: TestPartyID,
		Round:   1,
		Turn:    0,
		User: &models.User{
			ID:   TestUserID,
			Name: "Test User",
		},
		Party: &models.Party{
			ID:   TestPartyID,
			Name: "Test Party",
		},
	}
}

// CreateSampleCondition creates a sample condition for testing
func CreateSampleCondition() models.Condition {
	condition := models.Condition{
		ID: TestConditionID,
	}
	condition.Data.Name = TestConditionName
	condition.Data.System.Value.Value = 0
	return condition
}

// CreateSampleConditionData creates sample condition data for testing
func CreateSampleConditionData() map[string]interface{} {
	return map[string]interface{}{
		"_id":  "condition-test-id",
		"img":  "path/to/condition.jpg",
		"name": TestConditionName,
		"system": map[string]interface{}{
			"description": map[string]interface{}{
				"value": "Test condition description",
			},
			"value": map[string]interface{}{
				"isValued": true,
				"value":    5,
			},
		},
		"type": "condition",
	}
}

// CreateSampleItem creates a sample item for testing
func CreateSampleItem() models.Item {
	return models.Item{
		ID:   "test-item-id",
		Name: "Test Item",
	}
}
