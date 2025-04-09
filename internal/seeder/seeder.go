package seeder

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

type partiesData struct {
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

func Run(dbService database.Service) error {
	if dbService == nil {
		return fmt.Errorf("database service cannot be nil")
	}

	log.Println("Starting data seeding...")
	var finalErr error // Variable to collect errors without stopping immediately

	// --- Seed conditions ---
	log.Println("Seeding conditions...")
	conditionsChanged := 0
	// Use relative paths assuming execution from project root
	conditionsPath := "data/conditions"
	err := filepath.Walk(conditionsPath, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			log.Printf("Error accessing path %s: %v\n", path, walkErr)
			return nil // Continue walking other parts if possible
		}
		if !info.IsDir() && filepath.Ext(path) == ".json" {
			changed, seedErr := upsertSeedFile(dbService, path, "conditions")
			if seedErr != nil {
				log.Printf("ERROR seeding file %s: %v\n", path, seedErr)
				// Collect the error but continue seeding other files
				if finalErr == nil {
					finalErr = fmt.Errorf("error seeding conditions: %w", seedErr)
				} else {
					finalErr = fmt.Errorf("%w; error seeding conditions: %w", finalErr, seedErr)
				}
			} else if changed {
				conditionsChanged++
			}
		}
		return nil // Continue walking
	})
	// Check for error during the walk itself (e.g., directory not found)
	if err != nil {
		walkErrorMsg := fmt.Sprintf("error walking conditions directory '%s': %v", conditionsPath, err)
		log.Printf("FATAL during seeding: %s\n", walkErrorMsg)
		if finalErr == nil {
			finalErr = errors.New(walkErrorMsg)
		} else {
			finalErr = fmt.Errorf("%w; %s", finalErr, walkErrorMsg)
		}
		// Return immediately if the walk failed catastrophically
		return finalErr
	}

	if finalErr == nil { // Only log success summary if no file errors occurred
		if conditionsChanged == 0 {
			log.Println("Conditions data already up-to-date.")
		} else {
			log.Printf("Finished seeding conditions. %d files resulted in changes.\n", conditionsChanged)
		}
	}

	// --- Seed parties ---
	log.Println("Seeding parties...")
	partiesFilePath := filepath.Join("data", "parties.json")
	err = upsertSeedParties(dbService, partiesFilePath)
	if err != nil {
		// Check if the error is specifically because the file doesn't exist
		if errors.Is(err, os.ErrNotExist) {
			log.Printf("Optional parties file '%s' not found, skipping party seeding.", partiesFilePath)
		} else {
			// For any other error, record it as potentially fatal
			partyErrorMsg := fmt.Sprintf("error seeding parties from '%s': %v", partiesFilePath, err)
			log.Printf("ERROR during seeding: %s\n", partyErrorMsg)
			if finalErr == nil {
				finalErr = errors.New(partyErrorMsg)
			} else {
				finalErr = fmt.Errorf("%w; %s", finalErr, partyErrorMsg)
			}
		}
	}

	// --- Seed monsters ---
	log.Println("Seeding monsters...")
	monstersChanged := 0
	monstersPath := "data/bestiaries"
	err = filepath.Walk(monstersPath, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			log.Printf("Error accessing path %s: %v\n", path, walkErr)
			return nil // Continue walking
		}
		if !info.IsDir() && filepath.Ext(path) == ".json" && filepath.Base(path) != "_folders.json" {
			changed, seedErr := upsertSeedFile(dbService, path, "monsters")
			if seedErr != nil {
				log.Printf("ERROR seeding file %s: %v\n", path, seedErr)
				if finalErr == nil {
					finalErr = fmt.Errorf("error seeding monsters: %w", seedErr)
				} else {
					finalErr = fmt.Errorf("%w; error seeding monsters: %w", finalErr, seedErr)
				}
			} else if changed {
				monstersChanged++
			}
		}
		return nil // Continue walking
	})
	// Check for error during the walk itself
	if err != nil {
		walkErrorMsg := fmt.Sprintf("error walking bestiaries directory '%s': %v", monstersPath, err)
		log.Printf("FATAL during seeding: %s\n", walkErrorMsg)
		if finalErr == nil {
			finalErr = errors.New(walkErrorMsg)
		} else {
			finalErr = fmt.Errorf("%w; %s", finalErr, walkErrorMsg)
		}
		// Return immediately if the walk failed catastrophically
		return finalErr
	}

	if finalErr == nil { // Only log success summary if no file errors occurred
		if monstersChanged == 0 {
			log.Println("Monsters data already up-to-date.")
		} else {
			log.Printf("Finished seeding monsters. %d files resulted in changes.\n", monstersChanged)
		}
	}

	// --- Final Log ---
	if finalErr != nil {
		log.Printf("Data seeding process completed with errors: %v", finalErr)
	} else {
		log.Println("Data seeding process completed successfully.")
	}

	return finalErr
}

func upsertSeedFile(db database.Service, filePath string, table string) (bool, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("unable to read file %s: %w", filePath, err)
	}

	// Validate JSON structure *before* trying to upsert
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		if errors.As(err, &syntaxError) {
			return false, fmt.Errorf("unable to parse JSON from %s (syntax error at byte %d): %w", filePath, syntaxError.Offset, err)
		} else if errors.As(err, &unmarshalTypeError) {
			return false, fmt.Errorf("unable to parse JSON from %s (type error at byte %d, field '%s', expected type '%s'): %w", filePath, unmarshalTypeError.Offset, unmarshalTypeError.Field, unmarshalTypeError.Value, err)
		}
		return false, fmt.Errorf("unable to parse JSON from %s: %w", filePath, err)
	}

	// Extract and validate name
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
	        name = EXCLUDED.name -- Keep name update in case casing changed etc.
	    WHERE %s.data::jsonb IS DISTINCT FROM EXCLUDED.data::jsonb;
	`, table, table)

	res, err := db.Exec(query, trimmedName, data)
	if err != nil {
		return false, fmt.Errorf("unable to upsert data into table '%s' from %s (name: '%s'): %w", table, filePath, trimmedName, err)
	}

	rowsAffected, raErr := res.RowsAffected()
	if raErr != nil {
		log.Printf("Warning: Could not get RowsAffected for %s (name: %s): %v\n", filePath, trimmedName, raErr)
		return false, nil
	}

	if rowsAffected > 0 {
		return true, nil // Indicate change occurred
	}

	// No rows affected means data was identical or conflict occurred with no update needed
	return false, nil
}

func upsertSeedParties(db database.Service, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("unable to read parties file '%s': %w", filePath, err)
	}

	var partiesData partiesData

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

	// Use a transaction for party/player upserts
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction for %s: %w", filePath, err)
	}
	// Defer rollback/commit handling
	var txErr error
	defer func() {
		if txErr != nil {
			log.Printf("Rolling back transaction for %s due to error: %v", filePath, txErr)
			if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
				log.Printf("Error rolling back transaction for %s: %v", filePath, rbErr)
			}
			// Ensure the function returns the original error that caused the rollback
			err = txErr
		} else {
			// No error during processing, attempt to commit
			if cmtErr := tx.Commit(); cmtErr != nil {
				log.Printf("Error committing transaction for %s: %v", filePath, cmtErr)
				err = fmt.Errorf("error committing transaction for %s: %w", filePath, cmtErr)
			}
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
		userID := 1 // Assuming a default user ID for seeded parties

		// --- Party Upsert Logic ---
		// Try to insert, if conflict on (name, user_id), do nothing. Then fetch the ID.
		// This handles existing parties without updating them unnecessarily.
		partyInsertQuery := `
            INSERT INTO parties (name, user_id)
            VALUES ($1, $2)
            ON CONFLICT (name, user_id) DO NOTHING;
        `
		_, txErr = tx.Exec(partyInsertQuery, trimmedPartyName, userID)
		if txErr != nil {
			txErr = fmt.Errorf("error inserting party '%s' in %s: %w", trimmedPartyName, filePath, txErr)
			return txErr
		}

		// Always fetch the ID after attempting insert/on conflict
		partyFetchQuery := `SELECT id FROM parties WHERE name = $1 AND user_id = $2;`
		txErr = tx.QueryRow(partyFetchQuery, trimmedPartyName, userID).Scan(&partyID)
		if txErr != nil {
			txErr = fmt.Errorf("error fetching ID for party '%s' in %s after upsert: %w", trimmedPartyName, filePath, txErr)
			return txErr
		}
		// --- End Party Upsert Logic ---


		log.Printf("Processing party '%s' (ID: %d)\n", trimmedPartyName, partyID) // Log consistently

		// --- Player Upserts within Transaction ---
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
			res, playerErr := tx.Exec(playerQuery,
				trimmedPlayerName, playerData.Level, playerData.Hp, playerData.Ac,
				playerData.Fort, playerData.Ref, playerData.Will, playerData.Perception,
				partyID,
			)
			if playerErr != nil {
				// If a player fails, the transaction will be rolled back by defer
				txErr = fmt.Errorf("error upserting player '%s' for party '%s' in %s: %w",
					trimmedPlayerName, trimmedPartyName, filePath, playerErr)
				return txErr
			}

			rowsAffected, raErr := res.RowsAffected()
			if raErr != nil {
				log.Printf("Warning: Could not get RowsAffected for player '%s' in party '%s': %v\n", trimmedPlayerName, trimmedPartyName, raErr)
			} else {
				partyPlayersAffected += rowsAffected
			}
			playersProcessedInParty++
		} // End player loop

		// Log summary for the party's players after processing all players for that party
		if partyPlayersAffected > 0 {
			log.Printf("-> Upserted/Updated %d players for party '%s'.\n", partyPlayersAffected, trimmedPartyName)
		} else if playersProcessedInParty > 0 {
			log.Printf("-> All %d players for party '%s' were already up-to-date.\n", playersProcessedInParty, trimmedPartyName)
		} else if len(partyData.Players) > 0 {
			log.Printf("-> No valid players found to process for party '%s'.\n", trimmedPartyName)
		}

		totalPartiesProcessed++
		totalPlayersAffected += partyPlayersAffected

	} // End party loop

	// Final log message based on transaction outcome (set by defer)
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
