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
        SELECT m.id, m.data, em.adjustment
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
        err := rows.Scan(&m.ID, &jsonData, &m.Adjustment)
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
