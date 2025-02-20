package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"pf2.encounterbrew.com/internal/database"
	"pf2.encounterbrew.com/internal/utils"
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

type ConditionInfo struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

func (c Condition) GetValue() int {
	return c.Data.System.Value.Value
}

func (c *Condition) SetValue(value int) {
	c.Data.System.Value.Value = value
}

func (c Condition) GetName() string {
	return utils.RemoveHTML(c.Data.Name)
}

func (c Condition) IsValued() bool {
	return c.Data.System.Value.IsValued
}

func (c Condition) GetDescription() string {
	return utils.RemoveHTML(utils.NewReplacer().ProcessText(c.Data.System.Description.Value))
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

func GetGroupedConditions(db database.Service) (map[string][]ConditionInfo, error) {
	if db == nil {
		return nil, errors.New("database service is nil")
	}

	// Query all conditions
	rows, err := db.Query("SELECT id, data FROM conditions")
	if err != nil {
		return nil, fmt.Errorf("error querying conditions: %w", err)
	}
	defer rows.Close()

	// Initialize the result map
	groupedConditions := make(map[string][]ConditionInfo)

	// Process each row
	for rows.Next() {
		var c Condition
		var jsonData []byte
		err := rows.Scan(&c.ID, &jsonData)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		err = json.Unmarshal(jsonData, &c.Data)
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling JSON: %w", err)
		}

		// Determine the group
		group := "other"
		if c.Data.System.Group != nil {
			if groupStr, ok := c.Data.System.Group.(string); ok && groupStr != "" {
				group = groupStr
			}
		}

		// Create the condition info
		condInfo := ConditionInfo{
			Name: c.GetName(),
			ID:   c.ID,
		}

		// Add to the appropriate group
		groupedConditions[group] = append(groupedConditions[group], condInfo)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return groupedConditions, nil
}
