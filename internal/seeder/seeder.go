package seeder

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"pf2.encounterbrew.com/internal/database"
)

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
			changed, seedErr := UpsertSeedFile(dbService, path, "conditions")
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
			changed, seedErr := UpsertSeedFile(dbService, path, "monsters")
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

func UpsertSeedFile(db database.Service, filePath string, table string) (bool, error) {
	data, err := os.ReadFile(filePath) // #nosec G304 - file path is controlled and validated
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
