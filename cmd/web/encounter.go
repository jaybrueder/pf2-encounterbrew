package web

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	"pf2.encounterbrew.com/internal/database"
	"pf2.encounterbrew.com/internal/models"
)

func EncounterListHandler(db database.Service) echo.HandlerFunc {
    return func(c echo.Context) error {
        // Fetch all encounters for given user
        encounters, err := models.GetAllEncounters(db)
        if err != nil {
			log.Printf("Error fetching encounters: %v", err)
			return c.String(http.StatusInternalServerError, "Error fetching encounters")
		}

        // Render the template with the encounters
        component := EncounterList(encounters)
        return component.Render(c.Request().Context(), c.Response().Writer)
    }
}

func EncounterShowHandler(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get the encounter ID from the URL path parameter
		id := c.Param("id")
		log.Printf("Received ID: %q", id)

		// Fetch the encounter from the database
		encounter, err := models.GetEncounter(db, id)
		if err != nil {
			log.Printf("Error fetching encounter: %v", err)
			return c.String(http.StatusInternalServerError, "Error fetching encounter")
		}

		// Render the template with the encounter
		component := EncounterShow(encounter)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}
}
