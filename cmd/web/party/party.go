package party

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"pf2.encounterbrew.com/internal/database"
	"pf2.encounterbrew.com/internal/models"
)

func PartyNewHandler(c echo.Context) error {
	return PartyCreate().Render(c.Request().Context(), c.Response().Writer)
}

func PartyCreateHandler(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		partyName := c.FormValue("party_name")
		if partyName == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "Party name is required")
		}

		party := models.Party{
			Name:   partyName,
			UserID: 1, // Replace with actual user ID from session
		}

		id, err := party.Create(db)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create party")
		}

		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/parties/%d/edit", id))
	}
}

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
		playerFort := c.Request().Form["players[]fort"]
		playerRef := c.Request().Form["players[]ref"]
		playerWill := c.Request().Form["players[]will"]

		// Create a map of existing player IDs for tracking deletions
		existingPlayers := make(map[int]bool)
		for _, player := range party.Players {
			existingPlayers[player.ID] = true
		}

		// Update or create players
		var updatedPlayers []models.Player
		for i := range playerIDs {
			playerID, _ := strconv.Atoi(playerIDs[i])
			level, _ := strconv.Atoi(playerLevels[i])
			ac, _ := strconv.Atoi(playerACs[i])
			hp, _ := strconv.Atoi(playerHPs[i])
			fort, _ := strconv.Atoi(playerFort[i])
			ref, _ := strconv.Atoi(playerRef[i])
			will, _ := strconv.Atoi(playerWill[i])

			player := models.Player{
				ID:      playerID,
				Name:    playerNames[i],
				Level:   level,
				Ac:      ac,
				Hp:      hp,
				Fort:    fort,
				Ref:     ref,
				Will:    will,
				PartyID: party.ID,
			}

			if playerID != 0 {
				delete(existingPlayers, playerID) // Remove from tracking map
			}
			updatedPlayers = append(updatedPlayers, player)
		}

		// Any remaining IDs in existingPlayers map represent players to be deleted
		playersToDelete := make([]int, 0)
		for id := range existingPlayers {
			playersToDelete = append(playersToDelete, id)
		}

		party.Players = updatedPlayers

		// Save all changes
		if err := party.UpdateWithPlayers(db, playersToDelete); err != nil {
			log.Printf("Failed to update party: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update party")
		}

		component := PartyEdit(party)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}
}

func DeletePartyHandler(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		partyID := c.Param("party_id")
		id, err := strconv.Atoi(partyID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid party ID")
		}

		party := models.Party{
			ID:     id,
			UserID: 1, // Replace with actual user ID from session
		}

		if err := party.Delete(db); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete party")
		}

		return c.Redirect(http.StatusSeeOther, "/parties")
	}
}

func NewPlayerFormHandler(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		index, err := strconv.Atoi(c.QueryParam("index"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid index")
		}
		return PlayerForm(index, nil).Render(c.Request().Context(), c.Response().Writer)
	}
}

func PartyActiveHandler(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		partyID, err := strconv.Atoi(c.Param("party_id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid party ID")
		}

		// Set the active party for the user
		user := models.User{ID: 1} // Replace with actual user ID from session
		if err := user.SetActiveParty(db, partyID); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to set active party")
		}

		return c.Redirect(http.StatusSeeOther, "/parties")
	}
}
