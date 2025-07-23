package party

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

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
		partyID, _ := strconv.Atoi(c.Param("party_id"))

		// Fetch the party from the database
		party, err := models.GetParty(db, partyID)
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
		partyID, _ := strconv.Atoi(c.Param("party_id"))

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
		playerPerception := c.Request().Form["players[]perception"]

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
			perception, _ := strconv.Atoi(playerPerception[i])

			player := models.Player{
				ID:         playerID,
				Name:       playerNames[i],
				Level:      level,
				Ac:         ac,
				Hp:         hp,
				Fort:       fort,
				Ref:        ref,
				Will:       will,
				Perception: perception,
				PartyID:    party.ID,
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

func PlayerNewHandler(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		index, err := strconv.Atoi(c.QueryParam("index"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid index")
		}
		return PlayerForm(index, nil).Render(c.Request().Context(), c.Response().Writer)
	}
}

func PlayerDeleteHandler(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		partyID, _ := strconv.Atoi(c.Param("party_id"))
		playerID, _ := strconv.Atoi(c.Param("player_id"))

		err := models.PlayerDelete(db, playerID)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "Player not deleted")
		}

		// Get the existing party
		party, err := models.GetParty(db, partyID)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "Party not found")
		}

		component := PartyEdit(party)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}
}

func PartyExportHandler(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := 1 // TODO: Get from session

		// Export all parties for the user
		exportData, err := models.ExportAllParties(db, userID)
		if err != nil {
			log.Printf("Error exporting parties: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to export parties")
		}

		// Convert to JSON
		jsonData, err := json.MarshalIndent(exportData, "", "  ")
		if err != nil {
			log.Printf("Error marshaling export data: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to encode export data")
		}

		// Set headers for file download
		filename := fmt.Sprintf("parties_export_%s.json", time.Now().Format("2006-01-02"))
		c.Response().Header().Set("Content-Type", "application/json")
		c.Response().Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

		return c.Blob(http.StatusOK, "application/json", jsonData)
	}
}

func PartyImportHandler(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := 1 // TODO: Get from session
		log.Printf("Party import started for user %d", userID)

		// Get the uploaded file
		file, err := c.FormFile("import_file")
		if err != nil {
			log.Printf("Error getting file from form: %v", err)
			return echo.NewHTTPError(http.StatusBadRequest, "No file uploaded")
		}
		log.Printf("Received file: %s (size: %d bytes)", file.Filename, file.Size)

		// Limit file size to 10MB
		if file.Size > 10*1024*1024 {
			return echo.NewHTTPError(http.StatusBadRequest, "File too large (max 10MB)")
		}

		// Open the file
		src, err := file.Open()
		if err != nil {
			log.Printf("Error opening file: %v", err)
			return echo.NewHTTPError(http.StatusBadRequest, "Failed to open file")
		}
		defer src.Close()

		// Read the file contents
		data, err := io.ReadAll(src)
		if err != nil {
			log.Printf("Error reading file: %v", err)
			return echo.NewHTTPError(http.StatusBadRequest, "Failed to read file")
		}
		log.Printf("Read %d bytes from file", len(data))

		// Parse the JSON
		var importData models.PartiesExportData
		if err := json.Unmarshal(data, &importData); err != nil {
			log.Printf("Error parsing JSON: %v", err)
			log.Printf("JSON content (first 500 chars): %s", string(data[:min(500, len(data))]))
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON format")
		}
		log.Printf("Parsed JSON successfully, found %d parties", len(importData.Parties))

		// Import the parties (this will replace all existing parties)
		if err := models.ImportParties(db, userID, &importData); err != nil {
			log.Printf("Error importing parties: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to import parties")
		}

		log.Printf("Successfully imported %d parties for user %d", len(importData.Parties), userID)
		// Redirect back to the party list
		return c.Redirect(http.StatusSeeOther, "/parties")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
