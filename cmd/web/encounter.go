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
		id := c.Param("encounter_id")
		log.Printf("Received ID: %q", id)

		// Fetch the encounter from the database
		encounter, err := models.GetEncounterWithCombatants(db, id)
		if err != nil {
			log.Printf("Error fetching encounter: %v", err)
			return c.String(http.StatusInternalServerError, "Error fetching encounter")
		}

		// Render the template with the encounter
		component := EncounterShow(encounter)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}
}

func EncounterEditHandler(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get the encounter ID from the URL path parameter
		id := c.Param("encounter_id")

		// Fetch the encounter from the database
		encounter, err := models.GetEncounter(db, id)
		if err != nil {
			log.Printf("Error fetching encounter: %v", err)
			return c.String(http.StatusInternalServerError, "Error fetching encounter")
		}

		// Render the template with the encounter
		component := EncounterEdit(encounter)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}
}

// func EncounterUpdateHandler(db database.Service) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		//id := c.Param("id")
// 		//newName := c.FormValue("name")

// 		// Fetch the encounter from the database
// 		// TODO
// 		return nil
// 	}
// }

func EncounterSearchMonster(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		search := c.FormValue("search")
		encounterID := c.Param("encounter_id")

		monsters, err := models.SearchMonsters(db, search)
		if err != nil {
			log.Printf("Error searching for monster: %v", err)
			return c.String(http.StatusInternalServerError, "Error searching for monster")
		}

		component := MonsterSearchResults(encounterID, monsters)
        return component.Render(c.Request().Context(), c.Response().Writer)
	}
}

func EncounterAddMonster(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		encounterID := c.Param("encounter_id")
		monsterID := c.Param("monster_id")

		encounter, err := models.AddMonsterToEncounter(db, encounterID, monsterID)
		if err != nil {
			log.Printf("Error adding monster: %v", err)
			return c.String(http.StatusInternalServerError, "Error adding monster")
		}

		component := MonstersAdded(encounter)
        return component.Render(c.Request().Context(), c.Response().Writer)
	}
}

func EncounterRemoveMonster(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		encounterID := c.Param("encounter_id")
		monsterID := c.Param("monster_id")

		encounter, err := models.RemoveMonsterFromEncounter(db, encounterID, monsterID)
		if err != nil {
			log.Printf("Error removing monster: %v", err)
			return c.String(http.StatusInternalServerError, "Error removing monster")
		}

		component := MonstersAdded(encounter)
        return component.Render(c.Request().Context(), c.Response().Writer)
	}
}
