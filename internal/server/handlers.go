package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (s *Server) HomeHandler(c echo.Context) error {
	resp := map[string]string{
		"message": "Hello World",
	}

	return c.JSON(http.StatusOK, resp)
}
