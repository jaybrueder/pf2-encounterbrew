package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"pf2.encounterbrew.com/internal/database"
)

type Encounter struct {
	ID   		int    		`json:"id"`
	Name 		string 		`json:"name"`
 	UserID 		int			`json:"user_id"`
    User   		*User		`json:"user,omitempty"`
    Monsters 	[]*Monster 	`json:"monsters,omitempty"`
    Combatants 	[]Combatant `json:"combatants,omitempty"`
}

func GetAllEncounters(db database.Service) ([]Encounter, error) {
	if db == nil {
        return nil, errors.New("database service is nil")
    }

	rows, err := db.Query(`
	    SELECT e.id, e.name, e.user_id, u.name AS user_name
	    FROM encounters e
	    JOIN users u ON e.user_id = u.id
	    WHERE e.user_id = $1
	    ORDER BY e.id
    `, 1) // hard-coded User-ID for now
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var encounters []Encounter
    for rows.Next() {
        var e Encounter
        e.User = &User{}

        err := rows.Scan(&e.ID, &e.Name, &e.UserID, &e.User.Name)

        if err != nil {
            return nil, err
        }
        encounters = append(encounters, e)
    }

    if err = rows.Err(); err != nil {
        return nil, err
    }

    // TODO Also fetch monsters for each encounter?

    return encounters, nil
}

func GetEncounter(db database.Service, id string) (Encounter, error) {
	if db == nil {
		return Encounter{}, errors.New("database service is nil")
	}

	encounterID, err := strconv.Atoi(id)
 	if err != nil {
        return Encounter{}, fmt.Errorf("invalid encounter ID: %v", err)
    }

 	var e Encounter
    e.User = &User{}

    err = db.QueryRow(`
        SELECT e.id, e.name, e.user_id, u.name AS user_name
        FROM encounters e
        JOIN users u ON e.user_id = u.id
        WHERE e.user_id = $1 AND e.id = $2
    `, 1, encounterID).Scan(&e.ID, &e.Name, &e.User.ID, &e.User.Name)

    if err != nil {
        if err == sql.ErrNoRows {
            return Encounter{}, fmt.Errorf("no encounter found with ID %d", encounterID)
        }
        return Encounter{}, fmt.Errorf("error scanning encounter row: %v", err)
    }

    // Query for associated monsters
    rows, err := db.Query(`
        SELECT m.id, m.data, em.adjustment, em.count
        FROM monsters m
        JOIN encounter_monsters em ON m.id = em.monster_id
        WHERE em.encounter_id = $1
    `, encounterID)
    if err != nil {
        return e, fmt.Errorf("error querying monsters: %v", err)
    }
    defer rows.Close()

    for rows.Next() {
        var m Monster
        var jsonData []byte
        err := rows.Scan(&m.ID, &jsonData, &m.Adjustment, &m.Count)
        if err != nil {
            return e, fmt.Errorf("error scanning monster row: %v", err)
        }
        err = json.Unmarshal(jsonData, &m.Data)
        if err != nil {
            return e, fmt.Errorf("error unmarshaling monster data: %v", err)
        }
        e.Monsters = append(e.Monsters, &m)
    }

    if err = rows.Err(); err != nil {
        return e, fmt.Errorf("error iterating monster rows: %v", err)
    }

	return e, nil
}

func GetEncounterWithCombatants(db database.Service, id string) (Encounter, error) {
	encounter, err := GetEncounter(db, id)
	if err != nil {
		fmt.Errorf("Error fetching encounter: %v", err)
	}

	// Fetch the active party from the database
	// TODO make this flexible (hard-coded for now to 1)
	party, err := GetParty(db, "1")

	// Get party's players and encounter's monsters
	players := party.Players
	monsters := encounter.Monsters

	combatants := make([]Combatant, 0, len(players)+len(monsters))

    // Add players to combatants
    for i := range players {
        combatants = append(combatants, &players[i])
    }

	// Add monsters to combatants, respecting the count
	for _, monsterPtr := range monsters {
        for i := 0; i < monsterPtr.Count; i++ {
            // Create a non-pointer copy of the monster for each count
            monsterCopy := *monsterPtr

            // Modify the name to differentiate multiple instances
            if monsterPtr.Count > 1 {
                monsterCopy.Data.Name = fmt.Sprintf("%s (%d)", monsterPtr.Data.Name, i+1)
            }

            combatants = append(combatants, &monsterCopy)
        }
    }

    // Set Initiative and sort the combatants
    AssignInitiative(combatants)
    SortCombatantsByInitiative(combatants)

	// Add combatants to the encounter
	encounter.Combatants = combatants

	return encounter, nil
}

func AddMonsterToEncounter(db database.Service, encounterID string, monsterID string) (Encounter, error) {
 	// Convert string IDs to integers
  	encID, err := strconv.Atoi(encounterID)
  	if err != nil {
        return Encounter{}, fmt.Errorf("invalid encounter ID: %v", err)
     }

    monID, err := strconv.Atoi(monsterID)
    if err != nil {
       return Encounter{}, fmt.Errorf("invalid monster ID: %v", err)
    }

    // Use a transaction to ensure atomicity
    tx, err := db.Begin()
    if err != nil {
    	return Encounter{}, fmt.Errorf("Error starting transaction: %v", err)
    }
    defer tx.Rollback() // Rollback the transaction if it hasn't been committed

    // Try to update the count if the monster already exists in the encounter
    result, err := tx.Exec(`
        UPDATE encounter_monsters
        SET count = count + 1
        WHERE encounter_id = $1 AND monster_id = $2
    `, encID, monID)
    if err != nil {
    	return Encounter{}, fmt.Errorf("Error updating monster count: %v", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
    	return Encounter{}, fmt.Errorf("Error getting rows affected: %v", err)
    }

    if rowsAffected == 0 {
	    // Insert the new relationship into the encounter_monsters table
	    _, err = db.Exec(`
	        INSERT INTO encounter_monsters (encounter_id, monster_id)
	        VALUES ($1, $2)
	        ON CONFLICT (encounter_id, monster_id) DO NOTHING
	    `, encID, monID)

	    if err != nil {
	        return Encounter{}, fmt.Errorf("Failed to add monster to encounter: %v", err)
	    }
	}

 	// Commit the transaction
    if err = tx.Commit(); err != nil {
    	return Encounter{}, fmt.Errorf("Error committing transaction: %v", err)
    }

    encounter, err := GetEncounter(db, encounterID)

    return encounter, nil
}

func RemoveMonsterFromEncounter(db database.Service, encounterID string, monsterID string) (Encounter, error) {
    // Convert string IDs to integers
    encID, err := strconv.Atoi(encounterID)
    if err != nil {
        return Encounter{}, fmt.Errorf("invalid encounter ID: %v", err)
    }

    monID, err := strconv.Atoi(monsterID)
    if err != nil {
        return Encounter{}, fmt.Errorf("invalid monster ID: %v", err)
    }

    // Use a transaction to ensure atomicity
    tx, err := db.Begin()
    if err != nil {
        return Encounter{}, fmt.Errorf("error starting transaction: %v", err)
    }
    defer tx.Rollback() // Rollback the transaction if it hasn't been committed

    // Get the current count of the monster in the encounter
    var count int
    err = tx.QueryRow(`
        SELECT count FROM encounter_monsters
        WHERE encounter_id = $1 AND monster_id = $2
    `, encID, monID).Scan(&count)
    if err != nil {
        if err == sql.ErrNoRows {
            return Encounter{}, fmt.Errorf("monster not found in encounter")
        }
        return Encounter{}, fmt.Errorf("error getting monster count: %v", err)
    }

    if count > 1 {
        // If count is greater than 1, decrement the count
        _, err = tx.Exec(`
            UPDATE encounter_monsters
            SET count = count - 1
            WHERE encounter_id = $1 AND monster_id = $2
        `, encID, monID)
        if err != nil {
            return Encounter{}, fmt.Errorf("error updating monster count: %v", err)
        }
    } else {
        // If count is 1, remove the monster from the encounter
        _, err = tx.Exec(`
            DELETE FROM encounter_monsters
            WHERE encounter_id = $1 AND monster_id = $2
        `, encID, monID)
        if err != nil {
            return Encounter{}, fmt.Errorf("error removing monster from encounter: %v", err)
        }
    }

    // Commit the transaction
    if err = tx.Commit(); err != nil {
        return Encounter{}, fmt.Errorf("error committing transaction: %v", err)
    }

    // Fetch the updated encounter
    encounter, err := GetEncounter(db, encounterID)
    if err != nil {
        return Encounter{}, fmt.Errorf("error fetching updated encounter: %v", err)
    }

    return encounter, nil
}

func (e Encounter) GetDifficulty() int {
	partyLevel := 1
	monsterXpPool := 0

	for _, combatant := range e.Combatants {
		if combatant.GetType() == "monster" {
			monsterLevel := combatant.GetLevel()
			difference := monsterLevel - partyLevel

			switch {
				case difference <= -4:
					monsterXpPool += 10
				case difference == -3:
					monsterXpPool += 15
				case difference == -2:
					monsterXpPool += 20
				case difference == -1:
					monsterXpPool += 30
				case difference == 0:
					monsterXpPool += 40
				case difference == 1:
					monsterXpPool += 60
				case difference == 2:
					monsterXpPool += 80
				case difference == 3:
					monsterXpPool += 120
				case difference >= 4:
					monsterXpPool += 160
			}
		}
	}

	switch {
		case monsterXpPool <= 40:
			return 0
		case monsterXpPool <= 60:
			return 1
		case monsterXpPool <= 80:
			return 2
		case monsterXpPool <= 120:
			return 3
		case monsterXpPool > 120:
			return 4
	}

	return 0
}
