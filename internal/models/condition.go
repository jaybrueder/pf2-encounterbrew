package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"pf2.encounterbrew.com/internal/database"
)

type Condition struct {
	ID   int `json:"id"`
	Data struct {
		ID     string `json:"_id"`
		Img    string `json:"img"`
		Name   string `json:"name"`
		System struct {
			Description struct {
				Value string `json:"value"`
			} `json:"description"`
			Duration struct {
				Expiry any    `json:"expiry"`
				Unit   string `json:"unit"`
				Value  int    `json:"value"`
			} `json:"duration"`
			Group       any   `json:"group"`
			Overrides   []any `json:"overrides"`
			Publication struct {
				License  string `json:"license"`
				Remaster bool   `json:"remaster"`
				Title    string `json:"title"`
			} `json:"publication"`
			References struct {
				Children     []any `json:"children"`
				ImmunityFrom []any `json:"immunityFrom"`
				OverriddenBy []any `json:"overriddenBy"`
				Overrides    []any `json:"overrides"`
			} `json:"references"`
			Rules  []any `json:"rules"`
			Traits struct {
				Value []any `json:"value"`
			} `json:"traits"`
			Value struct {
				IsValued bool `json:"isValued"`
				Value    int  `json:"value"`
			} `json:"value"`
		} `json:"system"`
		Type string `json:"type"`
	}
}

func (c Condition) GetValue() int {
	return c.Data.System.Value.Value
}

func GetCondition(db database.Service, conditionID int) (Condition, error) {
	if db == nil {
		return Condition{}, errors.New("database service is nil")
	}

	// Query the condition from the database
    var c Condition
    var jsonData []byte

   	err := db.QueryRow(`
        SELECT id, data
        FROM conditions p
        WHERE id = $1
    `, conditionID).Scan(&c.ID, &jsonData)

    if err != nil {
		if err == sql.ErrNoRows {
			return Condition{}, fmt.Errorf("no condition found with ID %d", conditionID)
		}
		return Condition{}, fmt.Errorf("error scanning condition row: %v", err)
	}

	// Unmarshal the JSON data
    err = json.Unmarshal(jsonData, &c.Data)
    if err != nil {
        return Condition{}, fmt.Errorf("error unmarshaling condition data: %w", err)
    }

    return c, nil
}

func GetAllConditions(db database.Service) ([]Condition, error) {
	rows, err := db.Query("SELECT id, data FROM conditions")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conditions []Condition
	for rows.Next() {
		var c Condition
		var jsonData []byte
		err := rows.Scan(&c.ID, &jsonData)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(jsonData, &c.Data)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, c)
	}

	return conditions, nil
}

func SearchConditions(db database.Service, search string) ([]Condition, error) {
	query := "SELECT id, data FROM conditions WHERE LOWER(data->>'name') LIKE LOWER($1) LIMIT 5"

	// Search for the monster in the database and return the 10 most relevant results
	rows, err := db.Query(query, "%"+search+"%")
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return nil, fmt.Errorf("database query error: %w", err)
	}
	defer rows.Close()

	var conditions []Condition
	for rows.Next() {
		var c Condition
		var jsonData []byte
		err := rows.Scan(&c.ID, &jsonData)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		err = json.Unmarshal(jsonData, &c.Data)
		if err != nil {
			log.Printf("Error unmarshaling JSON data: %v", err)
			return nil, fmt.Errorf("error unmarshaling JSON: %w", err)
		}

		conditions = append(conditions, c)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating over rows: %v", err)
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return conditions, nil
}
