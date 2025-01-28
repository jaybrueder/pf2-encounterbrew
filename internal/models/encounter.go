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
	ID                int                        `json:"id"`
	Name              string                     `json:"name"`
	UserID            int                        `json:"user_id"`
	User              *User                      `json:"user,omitempty"`
	Monsters          []*Monster                 `json:"monsters,omitempty"`
	Combatants        []Combatant                `json:"combatants,omitempty"`
	Round             int                        `json:"round"`
	Turn              int                        `json:"turn"`
	GroupedConditions map[string][]ConditionInfo `json:"grouped_conditions"`
	Party             Party                      `json:"party"`
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
        SELECT m.id, m.data, em.level_adjustment, em.id
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
		err := rows.Scan(&m.ID, &jsonData, &m.LevelAdjustment, &m.AssociationID)
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

func GetEncounterWithCombatants(db database.Service, id string, partyId string) (Encounter, error) {
	encounter, err := GetEncounter(db, id)
	if err != nil {
		return Encounter{}, fmt.Errorf("error fetching encounter: %w", err)
	}

	// Fetch the active party from the database
	party, _ := GetParty(db, partyId)

	// Get party's players and encounter's monsters
	players := party.Players
	monsters := encounter.Monsters

	combatants := make([]Combatant, 0, len(players)+len(monsters))

	// Add players to combatants
	for i := range players {
		combatants = append(combatants, &players[i])
	}

	// Add monsters to combatants, respecting the count
	counts := make(map[string]int)

	for _, monster := range monsters {
		exactName := monster.GetName()
		counts[exactName]++

		if counts[exactName] > 1 {
			monster.SetEnumeration(counts[exactName])
			fmt.Printf("%s %d", monster.GetName(), counts[exactName])
		}

		combatants = append(combatants, monster)
	}

	// Set Initiative and sort the combatants
	AssignInitiative(combatants)
	SortCombatantsByInitiative(combatants)

	// Add combatants to the encounter
	encounter.Combatants = combatants

	// Set the party level
	encounter.Party = party

	return encounter, nil
}

func AddMonsterToEncounter(db database.Service, encounterID string, monsterID string, levelAdjustment int) (Encounter, error) {
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
	//nolint:errcheck
	defer tx.Rollback() // Rollback the transaction if it hasn't been committed

	_, err = db.Exec(`
        INSERT INTO encounter_monsters (encounter_id, monster_id, level_adjustment)
        VALUES ($1, $2, $3)
    `, encID, monID, levelAdjustment)

	if err != nil {
		return Encounter{}, fmt.Errorf("Failed to add monster to encounter: %v", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return Encounter{}, fmt.Errorf("Error committing transaction: %v", err)
	}

	encounter, _ := GetEncounter(db, encounterID)

	return encounter, nil
}

func RemoveMonsterFromEncounter(db database.Service, encounterID string, associationID string) (Encounter, error) {
	assID, err := strconv.Atoi(associationID)
	if err != nil {
		return Encounter{}, fmt.Errorf("invalid monster ID: %v", err)
	}

	// Use a transaction to ensure atomicity
	tx, err := db.Begin()
	if err != nil {
		return Encounter{}, fmt.Errorf("error starting transaction: %v", err)
	}
	//nolint:errcheck
	defer tx.Rollback() // Rollback the transaction if it hasn't been committed

	_, err = tx.Exec(`
        DELETE FROM encounter_monsters
        WHERE id = $1
    `, assID)

	if err != nil {
		return Encounter{}, fmt.Errorf("error removing monster from encounter: %v", err)
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
	partyLevel := int(e.Party.GetLevel())
	monsterXpPool := 0.0

	for _, combatant := range e.Combatants {
		if combatant.GetType() == "monster" {
			monsterLevel := combatant.GetLevel()
			difference := monsterLevel - partyLevel

			switch {
			case difference <= -4:
				monsterXpPool += 10.0
			case difference == -3:
				monsterXpPool += 15.0
			case difference == -2:
				monsterXpPool += 20.0
			case difference == -1:
				monsterXpPool += 30.0
			case difference == 0:
				monsterXpPool += 40.0
			case difference == 1:
				monsterXpPool += 60.0
			case difference == 2:
				monsterXpPool += 80.0
			case difference == 3:
				monsterXpPool += 120.0
			case difference >= 4:
				monsterXpPool += 160.0
			}
		}
	}

	threshold := float64(e.Party.GetNumbersOfPlayer()) / 4.0

	switch {
	case monsterXpPool <= 40.0*threshold:
		return 0
	case monsterXpPool <= 60.0*threshold:
		return 1
	case monsterXpPool <= 80.0*threshold:
		return 2
	case monsterXpPool <= 120.0*threshold:
		return 3
	case monsterXpPool > 120.0*threshold:
		return 4
	}

	return 0
}
