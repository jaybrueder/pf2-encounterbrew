package party

import (
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"pf2.encounterbrew.com/internal/database"
	"pf2.encounterbrew.com/internal/models"
)

func PartyListHandler(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Fetch all parties for a given user
		parties, err := models.GetAllParties(db)
		if err != nil {
			log.Printf("Error fetching parties: %v", err)
			return c.String(http.StatusInternalServerError, "Error fetching parties")
		}

		// Render the template with the encounters
		component := PartyList(parties)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}
}

func PartyEditHandler(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get the party ID from the URL path parameter
		id := c.Param("party_id")

		// Fetch the party from the database
		party, err := models.GetParty(db, id)
		if err != nil {
			log.Printf("Error fetching party: %v", err)
			return c.String(http.StatusInternalServerError, "Error fetching party")
		}

		// Render the template with the party
		component := PartyEdit(party)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}
}

func PartyUpdateHandler(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get the party ID from the URL path parameter
		partyID := c.Param("party_id")

		// Parse the form
		if err := c.Request().ParseForm(); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Failed to parse form")
		}

		// Get the existing party
		party, err := models.GetParty(db, partyID)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "Party not found")
		}

		// Update party name
		party.Name = c.FormValue("party_name")
		party.UserID = 1 // Set the user ID

		// Get player data from form
		playerIDs := c.Request().Form["players[]id"]
		playerNames := c.Request().Form["players[]name"]
		playerLevels := c.Request().Form["players[]level"]
		playerACs := c.Request().Form["players[]ac"]
		playerHPs := c.Request().Form["players[]hp"]

		// Update players
		for i := range playerIDs {
			if i < len(party.Players) {
				id, _ := strconv.Atoi(playerIDs[i])
				level, _ := strconv.Atoi(playerLevels[i])
				ac, _ := strconv.Atoi(playerACs[i])
				hp, _ := strconv.Atoi(playerHPs[i])

				party.Players[i].ID = id
				party.Players[i].Name = playerNames[i]
				party.Players[i].Level = level
				party.Players[i].Ac = ac
				party.Players[i].Hp = hp
				party.Players[i].PartyID = party.ID
			}
		}

		// Save all changes
		if err := party.UpdateWithPlayers(db); err != nil {
			log.Printf("Failed to update party: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update party")
		}

		component := PartyEdit(party)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}
}
