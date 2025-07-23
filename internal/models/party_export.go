package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"pf2.encounterbrew.com/internal/database"
)

// PartiesExportData represents the structure for exporting/importing parties
type PartiesExportData struct {
	Parties []PartyExportData `json:"parties"`
}

// PartyExportData represents a single party in the export format
type PartyExportData struct {
	Name    string             `json:"name"`
	Players []PlayerExportData `json:"players"`
}

// PlayerExportData represents a player in the export format
type PlayerExportData struct {
	Name       string `json:"name"`
	Level      int    `json:"level"`
	Hp         int    `json:"hp"`
	Ac         int    `json:"ac"`
	Fort       int    `json:"for"`
	Ref        int    `json:"ref"`
	Will       int    `json:"wil"`
	Perception int    `json:"perception"`
}

// ExportAllParties exports all parties for a user in the seeder format
func ExportAllParties(db database.Service, userID int) (*PartiesExportData, error) {
	if db == nil {
		return nil, errors.New("database service is nil")
	}

	// Get all parties with their players
	rows, err := db.Query(`
		SELECT p.id, p.name
		FROM parties p
		WHERE p.user_id = $1
		ORDER BY p.id
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("error querying parties: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Printf("error closing rows: %v\n", err)
		}
	}()

	exportData := &PartiesExportData{
		Parties: make([]PartyExportData, 0),
	}

	for rows.Next() {
		var partyID int
		var partyName string
		err := rows.Scan(&partyID, &partyName)
		if err != nil {
			return nil, fmt.Errorf("error scanning party row: %v", err)
		}

		// Get players for this party
		playerRows, err := db.Query(`
			SELECT name, level, hp, ac, fort, ref, will, perception
			FROM players
			WHERE party_id = $1
			ORDER BY id
		`, partyID)
		if err != nil {
			return nil, fmt.Errorf("error querying players: %v", err)
		}
		defer func() {
			if err := playerRows.Close(); err != nil {
				fmt.Printf("error closing playerRows: %v\n", err)
			}
		}()

		partyExport := PartyExportData{
			Name:    partyName,
			Players: make([]PlayerExportData, 0),
		}

		for playerRows.Next() {
			var player PlayerExportData
			err := playerRows.Scan(
				&player.Name, &player.Level, &player.Hp, &player.Ac,
				&player.Fort, &player.Ref, &player.Will, &player.Perception,
			)
			if err != nil {
				return nil, fmt.Errorf("error scanning player row: %v", err)
			}
			partyExport.Players = append(partyExport.Players, player)
		}

		if err = playerRows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating player rows: %v", err)
		}

		exportData.Parties = append(exportData.Parties, partyExport)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating party rows: %v", err)
	}

	return exportData, nil
}

// DeleteAllUserParties deletes all parties for a specific user
func DeleteAllUserParties(db database.Service, userID int) error {
	if db == nil {
		return errors.New("database service is nil")
	}

	_, err := db.Exec(`
		DELETE FROM parties
		WHERE user_id = $1
	`, userID)

	if err != nil {
		return fmt.Errorf("error deleting parties: %v", err)
	}

	return nil
}

// ImportParties imports parties from the export format, replacing all existing parties
func ImportParties(db database.Service, userID int, importData *PartiesExportData) error {
	if db == nil {
		return errors.New("database service is nil")
	}

	if importData == nil {
		return errors.New("import data is nil")
	}

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}

	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("error rolling back transaction: %v", err)
		}
	}()

	// Delete all existing parties for the user
	_, err = tx.Exec(`
		DELETE FROM parties
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("error deleting existing parties: %v", err)
	}

	// Import new parties
	for _, partyData := range importData.Parties {
		trimmedPartyName := strings.TrimSpace(partyData.Name)
		if trimmedPartyName == "" {
			continue // Skip parties with empty names
		}

		// Insert party
		var partyID int
		err := tx.QueryRow(`
			INSERT INTO parties (name, user_id)
			VALUES ($1, $2)
			RETURNING id
		`, trimmedPartyName, userID).Scan(&partyID)
		if err != nil {
			return fmt.Errorf("error inserting party '%s': %v", trimmedPartyName, err)
		}

		// Insert players
		for _, playerData := range partyData.Players {
			trimmedPlayerName := strings.TrimSpace(playerData.Name)
			if trimmedPlayerName == "" {
				continue // Skip players with empty names
			}

			_, err := tx.Exec(`
				INSERT INTO players (
					name, level, hp, ac, fort, ref, will, perception, party_id
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			`, trimmedPlayerName, playerData.Level, playerData.Hp, playerData.Ac,
				playerData.Fort, playerData.Ref, playerData.Will, playerData.Perception,
				partyID)
			if err != nil {
				return fmt.Errorf("error inserting player '%s': %v", trimmedPlayerName, err)
			}
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}
