package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"pf2.encounterbrew.com/cmd/web"
)

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
  		Level: 5,
	}))

	fileServer := http.FileServer(http.FS(web.Files))
	e.GET("/assets/*", echo.WrapHandler(fileServer))

 	e.GET("/", web.EncounterListHandler(s.db))

  	e.GET("/encounters/", web.EncounterListHandler(s.db))
    e.GET("/encounters/:id", web.EncounterShowHandler(s.db))
    e.GET("/encounters/:id/edit", web.EncounterListHandler(s.db))

	e.GET("/health", s.healthHandler)

	return e
}

func (s *Server) healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, s.db.Health())
}
