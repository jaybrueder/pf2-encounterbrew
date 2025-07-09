package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"pf2.encounterbrew.com/cmd/web"
	"pf2.encounterbrew.com/internal/models"
)

func TestHomeHandler(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*StandardMockDB)
		expectedStatus int
		expectError    bool
	}{
		{
			name: "successful home page load",
			mockSetup: func(mockDB *StandardMockDB) {
				monsters := []models.Monster{CreateSampleMonster()}
				monsters[0].ID = 1
				monsters[0].Data.Name = "Goblin"

				orc := CreateSampleMonster()
				orc.ID = 2
				orc.Data.Name = "Orc"
				monsters = append(monsters, orc)

				mockDB.SetupMockForGetAllMonsters(monsters)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "database error when fetching monsters",
			mockSetup: func(mockDB *StandardMockDB) {
				mockDB.Mock.ExpectQuery("SELECT id, data FROM monsters").
					WillReturnError(ErrNotFound)
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name: "empty monster list",
			mockSetup: func(mockDB *StandardMockDB) {
				mockDB.SetupMockForGetAllMonsters([]models.Monster{})
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, cleanup := NewStandardMockDB(t)
			defer cleanup()

			tt.mockSetup(mockDB)

			handler := web.HomeHandler(mockDB)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req = req.WithContext(context.Background())
			w := httptest.NewRecorder()

			handler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectError && w.Code != http.StatusInternalServerError {
				t.Errorf("Expected error status, got %d", w.Code)
			}

			if !tt.expectError && w.Code == http.StatusInternalServerError {
				t.Errorf("Unexpected error status: %d", w.Code)
			}

			// Verify mock expectations were met
			if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestHomeHandlerIntegration(t *testing.T) {
	mockDB, cleanup := NewStandardMockDB(t)
	defer cleanup()

	monsters := []models.Monster{CreateSampleMonster()}
	monsters[0].ID = 1
	monsters[0].Data.Name = "Goblin"

	orc := CreateSampleMonster()
	orc.ID = 2
	orc.Data.Name = "Orc"
	monsters = append(monsters, orc)

	mockDB.SetupMockForGetAllMonsters(monsters)

	handler := web.HomeHandler(mockDB)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.Background())
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Verify mock expectations were met
	if err := mockDB.Mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}
