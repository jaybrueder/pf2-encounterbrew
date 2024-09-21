package models

import (
	"encoding/json"

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
				Value    any  `json:"value"`
			} `json:"value"`
		} `json:"system"`
		Type string `json:"type"`
    }
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
