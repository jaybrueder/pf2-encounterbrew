package server

import (
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"pf2.encounterbrew.com/cmd/web"
	"pf2.encounterbrew.com/cmd/web/encounter"
	"pf2.encounterbrew.com/cmd/web/party"
)

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(session.Middleware(s.sessionStore))
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))

	// Basic HTTP Auth until we have a proper auth system
	e.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		// Define your credentials
		if username == "dragon" && password == "hobgoblin" {
			return true, nil
		}
		return false, nil
	}))

	fileServer := http.FileServer(http.FS(web.Files))
	e.GET("/assets/*", echo.WrapHandler(fileServer))

	e.GET("/", encounter.EncounterListHandler(s.db))

	// Encounter routes
	e.GET("/encounters/new", encounter.EncounterNewHandler(s.db))
	e.POST("/encounters", encounter.EncounterCreateHandler(s.db))
	e.GET("/encounters/:encounter_id/edit", encounter.EncounterEditHandler(s.db))
	e.PUT("/encounters/:encounter_id", encounter.EncounterUpdateHandler(s.db))
	e.DELETE("/encounters/:encounter_id", encounter.EncounterDeleteHandler(s.db))
	e.GET("/encounters", encounter.EncounterListHandler(s.db))
	e.GET("/encounters/:encounter_id", encounter.EncounterShowHandler(s.db))
	e.POST("/encounters/:encounter_id/search_monsters", encounter.EncounterSearchMonster(s.db))
	e.POST("/encounters/:encounter_id/add_monster/:monster_id", encounter.EncounterAddMonster(s.db))
	e.POST("/encounters/:encounter_id/remove_monster/:association_id", encounter.EncounterRemoveMonster(s.db))
	e.DELETE("/encounters/:encounter_id/remove_combatant/:association_id/:is_monster", encounter.EncounterRemoveCombatant(s.db))
	e.PATCH("/encounters/:encounter_id/combatant/:index/update", encounter.UpdateCombatant(s.db))
	e.PATCH("/encounters/:encounter_id/bulk_update_initiative", encounter.BulkUpdateInitiative(s.db))
	e.POST("/encounters/:encounter_id/combatant/:index/add_condition/:condition_id", encounter.AddCondition(s.db))
	e.POST("/encounters/:encounter_id/combatant/:index/remove_condition/:condition_id", encounter.RemoveCondition())
	e.POST("/encounters/:encounter_id/next_turn", encounter.ChangeTurn(true))
	e.POST("/encounters/:encounter_id/prev_turn", encounter.ChangeTurn(false))

	// Party routes
	e.GET("/parties", party.PartyListHandler(s.db))
	e.GET("/parties/new", party.PartyNewHandler)
	e.POST("/parties", party.PartyCreateHandler(s.db))
	e.GET("/parties/:party_id/edit", party.PartyEditHandler(s.db))
	e.PATCH("/parties/:party_id", party.PartyUpdateHandler(s.db))
	e.DELETE("/parties/:party_id", party.DeletePartyHandler(s.db))

	// Player routes
	e.GET("/parties/:party_id/player/new", party.PlayerNewHandler(s.db))
	e.DELETE("/parties/:party_id/:player_id", party.PlayerDeleteHandler(s.db))

	e.GET("/health", s.healthHandler)

	return e
}

func (s *Server) healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, s.db.Health())
}
