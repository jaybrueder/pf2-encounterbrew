package server

import (
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"pf2.encounterbrew.com/cmd/web"
	"pf2.encounterbrew.com/cmd/web/encounter"
)

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(session.Middleware(s.sessionStore))
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))

	fileServer := http.FileServer(http.FS(web.Files))
	e.GET("/assets/*", echo.WrapHandler(fileServer))

	e.GET("/", encounter.EncounterListHandler(s.db))

	e.GET("/encounters", encounter.EncounterListHandler(s.db))
	e.GET("/encounters/:encounter_id", encounter.EncounterShowHandler(s.db))
	e.GET("/encounters/:encounter_id/edit/", encounter.EncounterEditHandler(s.db))
	// e.PATCH("/encounters/:encounter_id", encounter.EncounterUpdateHandler(s.db))
	e.POST("/encounters/:encounter_id/search_monsters", encounter.EncounterSearchMonster(s.db))
	e.POST("/encounters/:encounter_id/add_monster/:monster_id", encounter.EncounterAddMonster(s.db))
	e.POST("/encounters/:encounter_id/remove_monster/:monster_id", encounter.EncounterRemoveMonster(s.db))
	e.PATCH("/encounters/:encounter_id/combatant/:index/update", encounter.UpdateCombatant())
	e.POST("/encounters/:encounter_id/combatant/:index/search_conditions", encounter.SearchConditions(s.db))
	e.POST("/encounters/:encounter_id/combatant/:index/add_condition/:condition_id", encounter.AddCondition(s.db))
	e.POST("/encounters/:encounter_id/combatant/:index/remove_condition/:condition_id", encounter.RemoveCondition())

	e.POST("/encounters/:encounter_id/next_turn", encounter.ChangeTurn(true))
	e.POST("/encounters/:encounter_id/prev_turn", encounter.ChangeTurn(false))

	e.GET("/health", s.healthHandler)

	return e
}

func (s *Server) healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, s.db.Health())
}
