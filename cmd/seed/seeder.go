package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"pf2.encounterbrew.com/internal/database"
	"pf2.encounterbrew.com/internal/models"
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

	// Seed conditions
	err := filepath.Walk("data/conditions", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".json" {
			if err := seedFile(dbService, path, "conditions"); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking through files: %v", err)
	}

	// Add this new call to seed parties
	if err := seedParties(dbService, "data/parties.json"); err != nil {
		fmt.Fprintf(os.Stderr, "Error seeding parties: %v\n", err)
		os.Exit(1)
	}

	// Set active party for default user
	// TODO: Static user id of 1
	if err := setDefaultUserActiveParty(dbService, 1); err != nil {
		fmt.Fprintf(os.Stderr, "Error setting default user's active party: %v\n", err)
		os.Exit(1)
	}

	// Seed a default encounter
	if err := seedEncounter(dbService, 1); err != nil {
		fmt.Fprintf(os.Stderr, "Error seeding encounter: %v\n", err)
		os.Exit(1)
	}

	// Seed monsters
	err = filepath.Walk("data/bestiaries", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".json" {
			if err := seedFile(dbService, path, "monsters"); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error seeding data: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Data seeded successfully")
}

func seedFile(db database.Service, filePath string, table string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("unable to read file %s: %v", filePath, err)
	}

	var jsonData map[string]interface{}
	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		return fmt.Errorf("unable to parse JSON from %s: %v", filePath, err)
	}

	_, err = db.Insert(table, []string{"data"}, jsonData)
	if err != nil {
		return fmt.Errorf("unable to insert data from %s: %v", filePath, err)
	}

	fmt.Printf("Seeded data from %s\n", filePath)
	return nil
}

func seedParties(db database.Service, filePath string) error {
	// Read the JSON file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("unable to read parties file: %v", err)
	}

	// Parse the JSON data
	var partiesData PartiesData
	if err := json.Unmarshal(data, &partiesData); err != nil {
		return fmt.Errorf("unable to parse parties JSON: %v", err)
	}

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}

	defer func() {
		if err := tx.Rollback(); err != nil {
			log.Printf("error rolling back transaction: %v", err)
		}
	}()

	// Insert each party and its players
	for _, partyData := range partiesData.Parties {
		// Create party
		party := &models.Party{
			Name:   partyData.Name,
			UserID: 1, // Assuming default user ID of 1
		}

		// Insert party
		var partyID int
		err := tx.QueryRow(`
            INSERT INTO parties (name, user_id)
            VALUES ($1, $2)
            RETURNING id
        `, party.Name, party.UserID).Scan(&partyID)
		if err != nil {
			return fmt.Errorf("error inserting party %s: %v", party.Name, err)
		}

		// Insert players for this party
		for _, playerData := range partyData.Players {
			_, err := tx.Exec(`
                INSERT INTO players (
                    name, level, hp, ac, fort, ref, will, perception, party_id
                ) VALUES (
                    $1, $2, $3, $4, $5, $6, $7, $8, $9
                )
            `,
				playerData.Name,
				playerData.Level,
				playerData.Hp,
				playerData.Ac,
				playerData.Fort,
				playerData.Ref,
				playerData.Will,
				playerData.Perception,
				partyID,
			)

			if err != nil {
				return fmt.Errorf("error inserting player %s: %v", playerData.Name, err)
			}
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	fmt.Printf("Successfully seeded %d parties\n", len(partiesData.Parties))
	return nil
}

func seedEncounter(db database.Service, partyID int) error {
	// First, verify the party exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM parties WHERE id = $1)", partyID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error checking party existence: %v", err)
	}
	if !exists {
		return fmt.Errorf("party with ID %d does not exist", partyID)
	}

	// Create a new encounter
	var encounterID int
	err = db.QueryRow(`
        INSERT INTO encounters (name, user_id)
        VALUES ($1, $2)
        RETURNING id
    `, "Default Encounter", 1).Scan(&encounterID)

	if err != nil {
		return fmt.Errorf("error creating encounter: %v", err)
	}

	// Get all players from the party
	rows, err := db.Query(`
        SELECT id FROM players
        WHERE party_id = $1
    `, partyID)
	if err != nil {
		return fmt.Errorf("error getting party players: %v", err)
	}
	defer rows.Close()

	// Collect all player IDs first
	var playerIDs []int
	for rows.Next() {
		var playerID int
		if err := rows.Scan(&playerID); err != nil {
			return fmt.Errorf("error scanning player ID: %v", err)
		}
		playerIDs = append(playerIDs, playerID)
	}
	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating over players: %v", err)
	}

	// Start a new transaction for inserting encounter players
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	// Add each player to the encounter_players table
	for _, playerID := range playerIDs {
		_, err = tx.Exec(`
            INSERT INTO encounter_players (encounter_id, player_id, initiative)
            VALUES ($1, $2, $3)
        `, encounterID, playerID, 0)

		if err != nil {
			return fmt.Errorf("error adding player to encounter: %v", err)
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}
	tx = nil // Prevent rollback after successful commit

	fmt.Printf("Successfully created encounter %d with players from party %d\n", encounterID, partyID)
	return nil
}

func setDefaultUserActiveParty(db database.Service, userID int) error {
	// Set the first party (ID 1) as the active party for the default user
	_, err := db.Exec(`
        UPDATE users
        SET active_party_id = 1
        WHERE id = $1
    `, userID)

	if err != nil {
		return fmt.Errorf("error setting default user's active party: %v", err)
	}

	fmt.Println("Set default user's active party")
	return nil
}
