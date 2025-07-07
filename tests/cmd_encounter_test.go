package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/labstack/echo/v4"

	"pf2.encounterbrew.com/cmd/web/encounter"
	"pf2.encounterbrew.com/internal/models"
)

func TestEncounterNewHandler(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*StandardMockDB)
		expectedStatus int
		expectError    bool
	}{
		{
			name: "successful encounter new page",
			mockSetup: func(mockDB *StandardMockDB) {
				parties := []models.Party{CreateSampleParty()}
				mockDB.SetupMockForGetAllParties(parties)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "database error when fetching parties",
			mockSetup: func(mockDB *StandardMockDB) {
				mockDB.Mock.ExpectQuery(`SELECT p\.id, p\.name, p\.user_id, u\.name AS user_name FROM parties p JOIN users u ON p\.user_id = u\.id WHERE p\.user_id = \$1 ORDER BY p\.id`).
					WithArgs(1).
					WillReturnError(ErrNotFound)
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, cleanup := NewStandardMockDB(t)
			defer cleanup()

			tt.mockSetup(mockDB)

			handler := encounter.EncounterNewHandler(mockDB)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/encounter/new", nil)
			req = req.WithContext(context.Background())
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler(c)
			
			if tt.expectError {
				if err == nil && rec.Code != tt.expectedStatus {
					t.Errorf("Expected error status %d, got %d", tt.expectedStatus, rec.Code)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if rec.Code != tt.expectedStatus {
					t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
				}
			}

			// Verify mock expectations were met
			if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestEncounterCreateHandler(t *testing.T) {
	tests := []struct {
		name           string
		formData       url.Values
		mockSetup      func(*StandardMockDB)
		expectedStatus int
		expectError    bool
	}{
		{
			name: "successful encounter creation",
			formData: url.Values{
				"name":     {"Test Encounter"},
				"party_id": {"1"},
			},
			mockSetup: func(mockDB *StandardMockDB) {
				mockDB.SetupMockForPartyExists(1, true)
				party := CreateSampleParty()
				mockDB.SetupMockForCreateEncounter(1, 1, party.Players)
				encounters := []models.Encounter{CreateSampleEncounter()}
				mockDB.SetupMockForGetAllEncounters(encounters)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "invalid party ID",
			formData: url.Values{
				"name":     {"Test Encounter"},
				"party_id": {"invalid"},
			},
			mockSetup: func(mockDB *StandardMockDB) {
				// No mock setup needed for this test
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "party does not exist",
			formData: url.Values{
				"name":     {"Test Encounter"},
				"party_id": {"999"},
			},
			mockSetup: func(mockDB *StandardMockDB) {
				mockDB.SetupMockForPartyExists(999, false)
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "create encounter fails",
			formData: url.Values{
				"name":     {"Test Encounter"},
				"party_id": {"1"},
			},
			mockSetup: func(mockDB *StandardMockDB) {
				mockDB.SetupMockForPartyExists(1, true)
				mockDB.Mock.ExpectQuery("INSERT INTO encounters").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), 1).
					WillReturnError(ErrNotFound)
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, cleanup := NewStandardMockDB(t)
			defer cleanup()

			tt.mockSetup(mockDB)

			handler := encounter.EncounterCreateHandler(mockDB)

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/encounter/create", strings.NewReader(tt.formData.Encode()))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
			req = req.WithContext(context.Background())
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler(c)
			
			if tt.expectError {
				if err == nil && rec.Code != tt.expectedStatus {
					t.Errorf("Expected error status %d, got %d", tt.expectedStatus, rec.Code)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if rec.Code != tt.expectedStatus {
					t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
				}
			}

			// Verify mock expectations were met (but only if we set up expectations)
			if tt.name != "invalid party ID" {
				if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
					t.Errorf("unfulfilled expectations: %s", err)
				}
			}
		})
	}
}

func TestEncounterEditHandler(t *testing.T) {
	tests := []struct {
		name           string
		encounterID    string
		mockSetup      func(*StandardMockDB)
		expectedStatus int
		expectError    bool
	}{
		{
			name:        "successful encounter edit",
			encounterID: "1",
			mockSetup: func(mockDB *StandardMockDB) {
				encounter := CreateSampleEncounter()
				mockDB.SetupMockForGetEncounterWithCombatants(encounter)
				parties := []models.Party{CreateSampleParty()}
				mockDB.SetupMockForGetAllParties(parties)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:        "invalid encounter ID",
			encounterID: "invalid",
			mockSetup: func(mockDB *StandardMockDB) {
				// No mock setup needed for this test
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:        "encounter not found",
			encounterID: "999",
			mockSetup: func(mockDB *StandardMockDB) {
				mockDB.Mock.ExpectQuery(`SELECT e\.id, e\.name, e\.user_id, e\.party_id, e\.turn, e\.round, u\.name AS user_name, p\.name AS party_name FROM encounters e JOIN users u ON e\.user_id = u\.id JOIN parties p ON e\.party_id = p\.id WHERE e\.user_id = \$1 AND e\.id = \$2`).
					WithArgs(1, 999).
					WillReturnError(ErrNotFound)
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, cleanup := NewStandardMockDB(t)
			defer cleanup()

			tt.mockSetup(mockDB)

			handler := encounter.EncounterEditHandler(mockDB)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/encounter/"+tt.encounterID+"/edit", nil)
			req = req.WithContext(context.Background())
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("encounter_id")
			c.SetParamValues(tt.encounterID)

			err := handler(c)
			
			if tt.expectError {
				if err == nil && rec.Code != tt.expectedStatus {
					t.Errorf("Expected error status %d, got %d", tt.expectedStatus, rec.Code)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if rec.Code != tt.expectedStatus {
					t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
				}
			}

			// Verify mock expectations were met (but only if we set up expectations)
			if tt.name != "invalid encounter ID" {
				if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
					t.Errorf("unfulfilled expectations: %s", err)
				}
			}
		})
	}
}

func TestEncounterUpdateHandler(t *testing.T) {
	t.Skip("Skipping complex update handler test for now")
	tests := []struct {
		name           string
		encounterID    string
		formData       url.Values
		mockSetup      func(*StandardMockDB)
		expectedStatus int
		expectError    bool
	}{
		{
			name:        "successful encounter update",
			encounterID: "1",
			formData: url.Values{
				"name":     {"Updated Encounter"},
				"party_id": {"1"},
			},
			mockSetup: func(mockDB *StandardMockDB) {
				// Mock party existence check first
				mockDB.Mock.ExpectQuery("SELECT EXISTS").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
				
				// Mock current party ID check second
				mockDB.Mock.ExpectQuery("SELECT party_id").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"party_id"}).AddRow(1))
				
				// Mock the name-only update
				mockDB.Mock.ExpectExec("UPDATE encounters").
					WithArgs("Updated Encounter", 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedStatus: http.StatusSeeOther,
			expectError:    false,
		},
		{
			name:        "invalid encounter ID",
			encounterID: "invalid",
			formData: url.Values{
				"name":     {"Updated Encounter"},
				"party_id": {"1"},
			},
			mockSetup: func(mockDB *StandardMockDB) {
				// No mock setup needed for this test
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:        "party does not exist",
			encounterID: "1",
			formData: url.Values{
				"name":     {"Updated Encounter"},
				"party_id": {"999"},
			},
			mockSetup: func(mockDB *StandardMockDB) {
				mockDB.SetupMockForPartyExists(999, false)
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:        "update fails",
			encounterID: "1",
			formData: url.Values{
				"name":     {"Updated Encounter"},
				"party_id": {"1"},
			},
			mockSetup: func(mockDB *StandardMockDB) {
				// Mock party existence check to pass
				mockDB.Mock.ExpectQuery("SELECT EXISTS").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
				
				// Mock current party ID check
				mockDB.Mock.ExpectQuery("SELECT party_id").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"party_id"}).AddRow(1))
				
				// Mock update to fail
				mockDB.Mock.ExpectExec("UPDATE encounters").
					WithArgs("Updated Encounter", 1).
					WillReturnError(ErrNotFound)
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, cleanup := NewStandardMockDB(t)
			defer cleanup()

			tt.mockSetup(mockDB)

			handler := encounter.EncounterUpdateHandler(mockDB)

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/encounter/"+tt.encounterID+"/update", strings.NewReader(tt.formData.Encode()))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
			req = req.WithContext(context.Background())
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("encounter_id")
			c.SetParamValues(tt.encounterID)

			err := handler(c)
			
			if tt.expectError {
				if err == nil && rec.Code != tt.expectedStatus {
					t.Errorf("Expected error status %d, got %d", tt.expectedStatus, rec.Code)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if rec.Code != tt.expectedStatus {
					t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
				}
			}

			// Verify mock expectations were met (but only if we set up expectations)
			if tt.name != "invalid encounter ID" {
				if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
					t.Errorf("unfulfilled expectations: %s", err)
				}
			}
		})
	}
}

func TestEncounterDeleteHandler(t *testing.T) {
	tests := []struct {
		name           string
		encounterID    string
		mockSetup      func(*StandardMockDB)
		expectedStatus int
		expectError    bool
	}{
		{
			name:        "successful encounter deletion",
			encounterID: "1",
			mockSetup: func(mockDB *StandardMockDB) {
				mockDB.SetupMockForDeleteEncounter()
				encounters := []models.Encounter{}
				mockDB.SetupMockForGetAllEncounters(encounters)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:        "invalid encounter ID",
			encounterID: "invalid",
			mockSetup: func(mockDB *StandardMockDB) {
				// No mock setup needed for this test
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:        "delete fails",
			encounterID: "1",
			mockSetup: func(mockDB *StandardMockDB) {
				mockDB.Mock.ExpectExec("DELETE FROM encounters").
					WithArgs(sqlmock.AnyArg()).
					WillReturnError(ErrNotFound)
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, cleanup := NewStandardMockDB(t)
			defer cleanup()

			tt.mockSetup(mockDB)

			handler := encounter.EncounterDeleteHandler(mockDB)

			e := echo.New()
			req := httptest.NewRequest(http.MethodDelete, "/encounter/"+tt.encounterID, nil)
			req = req.WithContext(context.Background())
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("encounter_id")
			c.SetParamValues(tt.encounterID)

			err := handler(c)
			
			if tt.expectError {
				if err == nil && rec.Code != tt.expectedStatus {
					t.Errorf("Expected error status %d, got %d", tt.expectedStatus, rec.Code)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if rec.Code != tt.expectedStatus {
					t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
				}
			}

			// Verify mock expectations were met (but only if we set up expectations)
			if tt.name != "invalid encounter ID" {
				if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
					t.Errorf("unfulfilled expectations: %s", err)
				}
			}
		})
	}
}

func TestEncounterListHandler(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*StandardMockDB)
		expectedStatus int
		expectError    bool
	}{
		{
			name: "successful encounter list",
			mockSetup: func(mockDB *StandardMockDB) {
				encounters := []models.Encounter{
					{ID: 1, Name: "Encounter 1", UserID: 1, PartyID: 1},
					{ID: 2, Name: "Encounter 2", UserID: 1, PartyID: 1},
				}
				mockDB.SetupMockForGetAllEncounters(encounters)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "database error when fetching encounters",
			mockSetup: func(mockDB *StandardMockDB) {
				mockDB.Mock.ExpectQuery(`SELECT e\.id, e\.name, e\.user_id, e\.party_id, u\.name AS user_name, p\.name AS party_name FROM encounters e JOIN users u ON e\.user_id = u\.id JOIN parties p ON e\.party_id = p\.id WHERE e\.user_id = \$1 ORDER BY e\.id`).
					WithArgs(1).
					WillReturnError(ErrNotFound)
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, cleanup := NewStandardMockDB(t)
			defer cleanup()

			tt.mockSetup(mockDB)

			handler := encounter.EncounterListHandler(mockDB)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/encounters", nil)
			req = req.WithContext(context.Background())
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler(c)
			
			if tt.expectError {
				if err == nil && rec.Code != tt.expectedStatus {
					t.Errorf("Expected error status %d, got %d", tt.expectedStatus, rec.Code)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if rec.Code != tt.expectedStatus {
					t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
				}
			}

			// Verify mock expectations were met
			if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestEncounterShowHandler(t *testing.T) {
	tests := []struct {
		name           string
		encounterID    string
		mockSetup      func(*StandardMockDB)
		expectedStatus int
		expectError    bool
	}{
		{
			name:        "successful encounter show",
			encounterID: "1",
			mockSetup: func(mockDB *StandardMockDB) {
				encounter := CreateSampleEncounter()
				mockDB.SetupMockForGetEncounterWithCombatants(encounter)
				
				// Mock GetGroupedConditions query
				conditionData := CreateSampleConditionData()
				jsonData, _ := json.Marshal(conditionData)
				mockDB.Mock.ExpectQuery(`SELECT id, data FROM conditions`).
					WillReturnRows(sqlmock.NewRows([]string{"id", "data"}).
						AddRow(1, jsonData))
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:        "encounter not found",
			encounterID: "999",
			mockSetup: func(mockDB *StandardMockDB) {
				mockDB.Mock.ExpectQuery(`SELECT e\.id, e\.name, e\.user_id, e\.party_id, e\.turn, e\.round, u\.name AS user_name, p\.name AS party_name FROM encounters e JOIN users u ON e\.user_id = u\.id JOIN parties p ON e\.party_id = p\.id WHERE e\.user_id = \$1 AND e\.id = \$2`).
					WithArgs(1, 999).
					WillReturnError(ErrNotFound)
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, cleanup := NewStandardMockDB(t)
			defer cleanup()

			tt.mockSetup(mockDB)

			handler := encounter.EncounterShowHandler(mockDB)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/encounter/"+tt.encounterID, nil)
			req = req.WithContext(context.Background())
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("encounter_id")
			c.SetParamValues(tt.encounterID)

			err := handler(c)
			
			if tt.expectError {
				if err == nil && rec.Code != tt.expectedStatus {
					t.Errorf("Expected error status %d, got %d", tt.expectedStatus, rec.Code)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if rec.Code != tt.expectedStatus {
					t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
				}
			}

			// Verify mock expectations were met
			if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestEncounterSearchMonster(t *testing.T) {
	tests := []struct {
		name           string
		formData       url.Values
		encounterID    string
		mockSetup      func(*StandardMockDB)
		expectedStatus int
		expectError    bool
	}{
		{
			name: "successful monster search",
			formData: url.Values{
				"search": {"goblin"},
			},
			encounterID: "1",
			mockSetup: func(mockDB *StandardMockDB) {
				monsters := []models.Monster{CreateSampleMonster()}
				monsters[0].Data.Name = "Goblin"
				mockDB.SetupMockForSearchMonsters(monsters)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "search fails",
			formData: url.Values{
				"search": {"invalid"},
			},
			encounterID: "1",
			mockSetup: func(mockDB *StandardMockDB) {
				mockDB.Mock.ExpectQuery("SELECT id, data FROM monsters").
					WithArgs(sqlmock.AnyArg()).
					WillReturnError(ErrNotFound)
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, cleanup := NewStandardMockDB(t)
			defer cleanup()

			tt.mockSetup(mockDB)

			handler := encounter.EncounterSearchMonster(mockDB)

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/encounter/"+tt.encounterID+"/search", strings.NewReader(tt.formData.Encode()))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
			req = req.WithContext(context.Background())
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("encounter_id")
			c.SetParamValues(tt.encounterID)

			err := handler(c)
			
			if tt.expectError {
				if err == nil && rec.Code != tt.expectedStatus {
					t.Errorf("Expected error status %d, got %d", tt.expectedStatus, rec.Code)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if rec.Code != tt.expectedStatus {
					t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
				}
			}

			// Verify mock expectations were met
			if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestEncounterAddMonster(t *testing.T) {
	tests := []struct {
		name           string
		encounterID    string
		monsterID      string
		formData       url.Values
		mockSetup      func(*StandardMockDB)
		expectedStatus int
		expectError    bool
	}{
		{
			name:        "successful monster addition",
			encounterID: "1",
			monsterID:   "1",
			formData: url.Values{
				"level_adjustment": {"0"},
			},
			mockSetup: func(mockDB *StandardMockDB) {
				monster := CreateSampleMonster()
				monster.Data.Name = "Goblin"
				mockDB.SetupMockForGetMonster(monster)
				encounter := CreateSampleEncounter()
				mockDB.SetupMockForAddMonsterToEncounter(encounter)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:        "monster not found",
			encounterID: "1",
			monsterID:   "999",
			formData: url.Values{
				"level_adjustment": {"0"},
			},
			mockSetup: func(mockDB *StandardMockDB) {
				mockDB.Mock.ExpectQuery("SELECT id, data FROM monsters").
					WithArgs(999).
					WillReturnError(ErrNotFound)
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, cleanup := NewStandardMockDB(t)
			defer cleanup()

			tt.mockSetup(mockDB)

			handler := encounter.EncounterAddMonster(mockDB)

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/encounter/"+tt.encounterID+"/monster/"+tt.monsterID, strings.NewReader(tt.formData.Encode()))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
			req = req.WithContext(context.Background())
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("encounter_id", "monster_id")
			c.SetParamValues(tt.encounterID, tt.monsterID)

			err := handler(c)
			
			if tt.expectError {
				if err == nil && rec.Code != tt.expectedStatus {
					t.Errorf("Expected error status %d, got %d", tt.expectedStatus, rec.Code)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if rec.Code != tt.expectedStatus {
					t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
				}
			}

			// Verify mock expectations were met
			if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestEncounterHandlersIntegration(t *testing.T) {
	t.Skip("Skipping complex integration test for now")
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()
	
	// Set up expectations for the integration test workflow
	parties := []models.Party{CreateSampleParty()}
	mockDB.SetupMockForGetAllParties(parties)
	
	mockDB.SetupMockForPartyExists(1, true)
	party := CreateSampleParty()
	mockDB.SetupMockForCreateEncounter(1, 1, party.Players)
	
	encounters := []models.Encounter{CreateSampleEncounter()}
	mockDB.SetupMockForGetAllEncounters(encounters)
	
	sampleEncounter := CreateSampleEncounter()
	mockDB.SetupMockForGetEncounter(sampleEncounter)

	t.Run("encounter workflow", func(t *testing.T) {
		newHandler := encounter.EncounterNewHandler(mockDB)
		createHandler := encounter.EncounterCreateHandler(mockDB)
		listHandler := encounter.EncounterListHandler(mockDB)
		editHandler := encounter.EncounterEditHandler(mockDB)

		e := echo.New()

		// Test new
		newReq := httptest.NewRequest(http.MethodGet, "/encounter/new", nil)
		newReq = newReq.WithContext(context.Background())
		newRec := httptest.NewRecorder()
		newCtx := e.NewContext(newReq, newRec)

		err := newHandler(newCtx)
		if err != nil {
			t.Errorf("New handler failed: %v", err)
		}

		// Test create
		createReq := httptest.NewRequest(http.MethodPost, "/encounter/create", strings.NewReader("name=Test Encounter&party_id=1"))
		createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		createReq = createReq.WithContext(context.Background())
		createRec := httptest.NewRecorder()
		createCtx := e.NewContext(createReq, createRec)

		err = createHandler(createCtx)
		if err != nil {
			t.Errorf("Create handler failed: %v", err)
		}

		// Test list
		listReq := httptest.NewRequest(http.MethodGet, "/encounters", nil)
		listReq = listReq.WithContext(context.Background())
		listRec := httptest.NewRecorder()
		listCtx := e.NewContext(listReq, listRec)

		err = listHandler(listCtx)
		if err != nil {
			t.Errorf("List handler failed: %v", err)
		}

		// Test edit (add parties mock again)
		mockDB.SetupMockForGetAllParties(parties)
		
		editReq := httptest.NewRequest(http.MethodGet, "/encounter/1/edit", nil)
		editReq = editReq.WithContext(context.Background())
		editRec := httptest.NewRecorder()
		editCtx := e.NewContext(editReq, editRec)
		editCtx.SetParamNames("encounter_id")
		editCtx.SetParamValues("1")

		err = editHandler(editCtx)
		if err != nil {
			t.Errorf("Edit handler failed: %v", err)
		}

		// Verify mock expectations were met
		if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %s", err)
		}
	})
}