package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"pf2.encounterbrew.com/internal/database"
)

type PartiesData struct {
	Parties []struct {
		Name    string `json:"name"`
		Players []struct {
			Name       string `json:"name"`
			Level      int    `json:"level"`
			Hp         int    `json:"hp"`
			Ac         int    `json:"ac"`
			Fort       int    `json:"for"`
			Ref        int    `json:"ref"`
			Will       int    `json:"wil"`
			Perception int    `json:"perception"`
		} `json:"players"`
	} `json:"parties"`
}

func main() {
	dbService := database.New()

	log.Println("Starting data seeding...")

	// Seed conditions
	log.Println("Seeding conditions...")
	conditionsChanged := 0
	err := filepath.Walk("data/conditions", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error accessing path %s: %v\n", path, err)
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".json" {
			changed, err := upsertSeedFile(dbService, path, "conditions")
			if err != nil {
				log.Printf("ERROR seeding file %s: %v\n", path, err)
				// return err // Optional: stop on first file error
			}
			if changed {
				conditionsChanged++
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("FATAL: Error walking conditions directory: %v\n", err)
		// os.Exit(1)
	}
	if conditionsChanged == 0 {
		log.Println("Conditions data already up-to-date.")
	} else {
		log.Printf("Finished seeding conditions. %d files resulted in changes.\n", conditionsChanged)
	}

	// Seed parties - This function already logs changes internally
	log.Println("Seeding parties...")
	if err := upsertSeedParties(dbService, "data/parties.json"); err != nil {
		log.Fatalf("FATAL: Error seeding parties: %v\n", err)
	}

	// Seed monsters
	log.Println("Seeding monsters...") // Log section start
	monstersChanged := 0              // Counter for changes
	err = filepath.Walk("data/bestiaries", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error accessing path %s: %v\n", path, err)
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".json" && filepath.Base(path) != "_folders.json" {
			changed, err := upsertSeedFile(dbService, path, "monsters") // Capture changed status
			if err != nil {
				log.Printf("ERROR seeding file %s: %v\n", path, err)
				// return err // Optional: stop on first file error
			}
			if changed {
				monstersChanged++
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("FATAL: Error walking bestiaries directory: %v\n", err)
		// os.Exit(1)
	}
	if monstersChanged == 0 {
		log.Println("Monsters data already up-to-date.")
	} else {
		log.Printf("Finished seeding monsters. %d files resulted in changes.\n", monstersChanged)
	}

	log.Println("Data seeding process completed.")
}

func upsertSeedFile(db database.Service, filePath string, table string) (bool, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("unable to read file %s: %w", filePath, err)
	}

	var jsonData map[string]interface{}
	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		if errors.As(err, &syntaxError) {
			return false, fmt.Errorf("unable to parse JSON from %s (syntax error at byte %d): %w", filePath, syntaxError.Offset, err)
		} else if errors.As(err, &unmarshalTypeError) {
			return false, fmt.Errorf("unable to parse JSON from %s (type error at byte %d, field '%s', expected type '%s'): %w", filePath, unmarshalTypeError.Offset, unmarshalTypeError.Field, unmarshalTypeError.Value, err)
		}
		return false, fmt.Errorf("unable to parse JSON from %s: %w", filePath, err)
	}

	nameVal, ok := jsonData["name"]
	if !ok {
		return false, fmt.Errorf("missing 'name' field in JSON file %s", filePath)
	}
	nameStr, ok := nameVal.(string)
	if !ok {
		return false, fmt.Errorf("'name' field in JSON file %s is not a string (type: %T)", filePath, nameVal)
	}
	trimmedName := strings.TrimSpace(nameStr)
	if trimmedName == "" {
		return false, fmt.Errorf("'name' field in JSON file %s cannot be empty or only whitespace", filePath)
	}

	query := fmt.Sprintf(`
	    INSERT INTO %s (name, data)
	    VALUES ($1, $2)
	    ON CONFLICT (name) DO UPDATE SET
	        data = EXCLUDED.data,
	        name = EXCLUDED.name
	    WHERE %s.data::jsonb IS DISTINCT FROM EXCLUDED.data::jsonb;
	`, table, table)

	res, err := db.Exec(query, trimmedName, data)
	if err != nil {
		return false, fmt.Errorf("unable to upsert data into table '%s' from %s (name: '%s'): %w", table, filePath, trimmedName, err)
	}

	rowsAffected, raErr := res.RowsAffected()
	if raErr != nil {
		// Log the warning about RowsAffected error, but don't log general success
		log.Printf("Warning: Could not get RowsAffected for %s (name: %s): %v\n", filePath, trimmedName, raErr)
		return false, nil // Return false as we couldn't confirm change, but Exec was successful
	}

	if rowsAffected > 0 {
		log.Printf("Upserted/Updated data from %s (Name: %s)\n", filePath, trimmedName)
		return true, nil // Indicate change occurred
	}

	return false, nil // Indicate no change occurred
}

func upsertSeedParties(db database.Service, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("unable to read parties file '%s': %w", filePath, err)
	}

	var partiesData PartiesData

	if err := json.Unmarshal(data, &partiesData); err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		if errors.As(err, &syntaxError) {
			return fmt.Errorf("unable to parse JSON from %s (syntax error at byte %d): %w", filePath, syntaxError.Offset, err)
		} else if errors.As(err, &unmarshalTypeError) {
			return fmt.Errorf("unable to parse JSON from %s (type error at byte %d, field '%s', expected type '%s'): %w", filePath, unmarshalTypeError.Offset, unmarshalTypeError.Field, unmarshalTypeError.Value, err)
		}
		return fmt.Errorf("unable to parse parties JSON from '%s': %w", filePath, err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction for %s: %w", filePath, err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("Error rolling back transaction for %s: %v\n", filePath, err)
		}
	}()

	totalPartiesProcessed := 0
	totalPlayersAffected := int64(0)

	for _, partyData := range partiesData.Parties {
		trimmedPartyName := strings.TrimSpace(partyData.Name)
		if trimmedPartyName == "" {
			log.Printf("Skipping party with empty name in %s\n", filePath)
			continue
		}

		var partyID int
		userID := 1 // Default user ID

		// Party Upsert - Still logs processing
		partyQuery := `
		    WITH party_data AS (
		        SELECT id FROM parties WHERE name = $1 AND user_id = $2
		    ),
		    upsert AS (
		        INSERT INTO parties (name, user_id)
		        VALUES ($1, $2)
		        ON CONFLICT (name, user_id) DO UPDATE SET
		            name = EXCLUDED.name
		        WHERE parties.name IS DISTINCT FROM EXCLUDED.name
		        RETURNING id
		    )
		    SELECT id FROM upsert
		    UNION ALL
		    SELECT id FROM party_data
		    WHERE NOT EXISTS (SELECT 1 FROM upsert)
		    LIMIT 1;
		`
		err := tx.QueryRow(partyQuery, trimmedPartyName, userID).Scan(&partyID)
		if err != nil {
			return fmt.Errorf("error upserting party '%s' in %s: %w", trimmedPartyName, filePath, err)
		}

		// Keep party processing log for context within the transaction
		log.Printf("Processed party '%s' (ID: %d)\n", trimmedPartyName, partyID)

		// Player Upserts
		partyPlayersAffected := int64(0)
		playersProcessedInParty := 0
		for _, playerData := range partyData.Players {
			trimmedPlayerName := strings.TrimSpace(playerData.Name)
			if trimmedPlayerName == "" {
				log.Printf("Skipping player with empty name in party '%s' (%s)\n", trimmedPartyName, filePath)
				continue
			}

			playerQuery := `
			    INSERT INTO players (
			        name, level, hp, ac, fort, ref, will, perception, party_id
			    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			    ON CONFLICT (name, party_id) DO UPDATE SET
			        level = EXCLUDED.level, hp = EXCLUDED.hp, ac = EXCLUDED.ac,
			        fort = EXCLUDED.fort, ref = EXCLUDED.ref, will = EXCLUDED.will,
			        perception = EXCLUDED.perception
			    WHERE
			        players.level IS DISTINCT FROM EXCLUDED.level OR
			        players.hp IS DISTINCT FROM EXCLUDED.hp OR
			        players.ac IS DISTINCT FROM EXCLUDED.ac OR
			        players.fort IS DISTINCT FROM EXCLUDED.fort OR
			        players.ref IS DISTINCT FROM EXCLUDED.ref OR
			        players.will IS DISTINCT FROM EXCLUDED.will OR
			        players.perception IS DISTINCT FROM EXCLUDED.perception;
			`
			res, err := tx.Exec(playerQuery,
				trimmedPlayerName, playerData.Level, playerData.Hp, playerData.Ac,
				playerData.Fort, playerData.Ref, playerData.Will, playerData.Perception,
				partyID,
			)
			if err != nil {
				return fmt.Errorf("error upserting player '%s' for party '%s' in %s: %w",
					trimmedPlayerName, trimmedPartyName, filePath, err)
			}

			rowsAffected, raErr := res.RowsAffected()
			if raErr != nil {
				log.Printf("Warning: Could not get RowsAffected for player '%s' in party '%s': %v\n", trimmedPlayerName, trimmedPartyName, raErr)
			} else {
				partyPlayersAffected += rowsAffected // Sum changes within the party
			}
			playersProcessedInParty++
		}
		// Log summary for the party's players only if changes occurred or if it's informative
		if partyPlayersAffected > 0 {
			log.Printf("-> Upserted/Updated %d players for party '%s'.\n", partyPlayersAffected, trimmedPartyName)
		} else if playersProcessedInParty > 0 {
			log.Printf("-> All %d players for party '%s' were already up-to-date.\n", playersProcessedInParty, trimmedPartyName)
		} else {
			log.Printf("-> No players processed for party '%s'.\n", trimmedPartyName)
		}

		totalPartiesProcessed++
		totalPlayersAffected += partyPlayersAffected
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction for %s: %w", filePath, err)
	}

	// Final summary for parties file
	if totalPlayersAffected > 0 {
		log.Printf("Successfully processed %d parties from %s. Total players inserted/updated: %d\n", totalPartiesProcessed, filePath, totalPlayersAffected)
	} else if totalPartiesProcessed > 0 {
		log.Printf("Successfully processed %d parties from %s. All player data was already up-to-date.\n", totalPartiesProcessed, filePath)
	} else {
		log.Printf("No parties processed from %s.\n", filePath)
	}
	return nil
}
