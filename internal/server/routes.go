package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/echo-contrib/session"

	"pf2.encounterbrew.com/cmd/web"
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

 	e.GET("/", web.EncounterListHandler(s.db))

  	e.GET("/encounters", web.EncounterListHandler(s.db))
    e.GET("/encounters/:encounter_id", web.EncounterShowHandler(s.db))
    e.GET("/encounters/:encounter_id/edit", web.EncounterEditHandler(s.db))
    // e.PATCH("/encounters/:encounter_id", web.EncounterUpdateHandler(s.db))
	e.POST("/encounters/:encounter_id/search_monsters", web.EncounterSearchMonster(s.db))
    e.POST("/encounters/:encounter_id/add_monster/:monster_id", web.EncounterAddMonster(s.db))
    e.POST("/encounters/:encounter_id/remove_monster/:monster_id", web.EncounterRemoveMonster(s.db))
    e.PATCH("/encounters/:encounter_id/combatant/:index/update", web.UpdateCombatant(s.db))

	e.GET("/health", s.healthHandler)

	return e
}

func (s *Server) healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, s.db.Health())
}
