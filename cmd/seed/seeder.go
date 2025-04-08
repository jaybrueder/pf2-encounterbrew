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
	partiesFilePath := filepath.Join("data", "parties.json") // Use filepath.Join
	err = upsertSeedParties(dbService, partiesFilePath)      // Assign error directly

	if err != nil {
		// Check if the error is specifically because the file doesn't exist
		// os.ErrNotExist is generally fine, io/fs.ErrNotExist is slightly more modern
		if errors.Is(err, os.ErrNotExist) {
			log.Printf("Optional parties file '%s' not found, skipping party seeding.", partiesFilePath)
			// No fatal error here, just log and continue
		} else {
			// For any other error (permissions, JSON parse error inside the file, DB error during upsert), treat it as fatal
			log.Fatalf("FATAL: Error seeding parties from '%s': %v\n", partiesFilePath, err)
		}
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
		// This error will be checked in main using errors.Is(err, os.ErrNotExist)
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
	// Use named return for easier error handling in defer
	var txErr error
	defer func() {
		if txErr != nil {
			// If an error occurred during the transaction, rollback
			if rbErr := tx.Rollback(); rbErr != nil && rbErr != sql.ErrTxDone {
				log.Printf("Error rolling back transaction for %s (original error: %v): %v\n", filePath, txErr, rbErr)
			}
		} else {
			// If no error occurred, try to commit
			if cmtErr := tx.Commit(); cmtErr != nil {
				// If commit fails, assign it to txErr so the function returns it
				txErr = fmt.Errorf("error committing transaction for %s: %w", filePath, cmtErr)
				log.Printf("Error during commit: %v", txErr) // Log commit error specifically
			}
		}
	}() // Pass txErr by reference (implicitly via closure)

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

		// Party Upsert
		partyQuery := `
		    WITH existing_party AS (
		        SELECT id FROM parties WHERE name = $1 AND user_id = $2
		    ), inserted_party AS (
		        INSERT INTO parties (name, user_id)
		        VALUES ($1, $2)
		        ON CONFLICT (name, user_id) DO NOTHING -- Avoid unnecessary updates if name/user_id match
		        RETURNING id
		    )
		    SELECT id FROM inserted_party
		    UNION ALL
		    SELECT id FROM existing_party WHERE NOT EXISTS (SELECT 1 FROM inserted_party)
			LIMIT 1; -- Ensure only one ID is returned
		`
		// Scan can potentially fail if the party already exists and ON CONFLICT DO NOTHING returns no rows.
		// We need to handle sql.ErrNoRows specifically after the upsert attempt.
		err = tx.QueryRow(partyQuery, trimmedPartyName, userID).Scan(&partyID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// This means the party existed, and DO NOTHING was triggered. We need to get the existing ID.
				fetchQuery := `SELECT id FROM parties WHERE name = $1 AND user_id = $2`
				err = tx.QueryRow(fetchQuery, trimmedPartyName, userID).Scan(&partyID)
				if err != nil {
					txErr = fmt.Errorf("error fetching existing party ID for '%s' in %s after failed upsert scan: %w", trimmedPartyName, filePath, err)
					return txErr // Assign to txErr and return
				}
			} else {
				// Genuine error during insert/query
				txErr = fmt.Errorf("error upserting/finding party '%s' in %s: %w", trimmedPartyName, filePath, err)
				return txErr // Assign to txErr and return
			}
		}


		// Log party processing regardless of player changes
		log.Printf("Processing party '%s' (ID: %d)\n", trimmedPartyName, partyID)

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
			    WHERE -- Only update if any relevant field has actually changed
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
				// If a player fails, rollback the whole transaction for this file
				txErr = fmt.Errorf("error upserting player '%s' for party '%s' in %s: %w",
					trimmedPlayerName, trimmedPartyName, filePath, err)
				return txErr // Assign to txErr and return
			}

			rowsAffected, raErr := res.RowsAffected()
			if raErr != nil {
				log.Printf("Warning: Could not get RowsAffected for player '%s' in party '%s': %v\n", trimmedPlayerName, trimmedPartyName, raErr)
				// Continue processing other players even if RowsAffected fails
			} else {
				partyPlayersAffected += rowsAffected // Sum changes within the party
			}
			playersProcessedInParty++
		}
		// Log summary for the party's players
		if partyPlayersAffected > 0 {
			log.Printf("-> Upserted/Updated %d players for party '%s'.\n", partyPlayersAffected, trimmedPartyName)
		} else if playersProcessedInParty > 0 {
			log.Printf("-> All %d players for party '%s' were already up-to-date.\n", playersProcessedInParty, trimmedPartyName)
		} else if len(partyData.Players) > 0 {
            log.Printf("-> No valid players found to process for party '%s'.\n", trimmedPartyName)
        } // No need to log if partyData.Players was empty


		totalPartiesProcessed++
		totalPlayersAffected += partyPlayersAffected
	}

	if txErr == nil {
		if totalPlayersAffected > 0 {
			log.Printf("Successfully processed %d parties from %s. Total players inserted/updated: %d\n", totalPartiesProcessed, filePath, totalPlayersAffected)
		} else if totalPartiesProcessed > 0 {
			log.Printf("Successfully processed %d parties from %s. All player data was already up-to-date.\n", totalPartiesProcessed, filePath)
		} else {
			log.Printf("No valid parties processed from %s.\n", filePath)
		}
	}

	return txErr
}
