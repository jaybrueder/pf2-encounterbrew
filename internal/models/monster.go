package models

import (
	"encoding/json"
	"fmt"
	"log"

	"pf2.encounterbrew.com/internal/database"
)

type Monster struct {
    ID   int `json:"id"`
    Adjustment int `json:"adjustment"`
    Count int `json:"count"`
    Initiative int `json:"initiative"`
    Data struct {
		ID    string `json:"_id"`
		Img   string `json:"img"`
		Items []Item `json:"items"`
		Name   string `json:"name"`
		System struct {
			Abilities struct {
				Cha struct {
					Mod int `json:"mod"`
				} `json:"cha"`
				Con struct {
					Mod int `json:"mod"`
				} `json:"con"`
				Dex struct {
					Mod int `json:"mod"`
				} `json:"dex"`
				Int struct {
					Mod int `json:"mod"`
				} `json:"int"`
				Str struct {
					Mod int `json:"mod"`
				} `json:"str"`
				Wis struct {
					Mod int `json:"mod"`
				} `json:"wis"`
			} `json:"abilities"`
			Attributes struct {
				Ac struct {
					Details string `json:"details"`
					Value   int    `json:"value"`
				} `json:"ac"`
				AllSaves struct {
					Value string `json:"value"`
				} `json:"allSaves"`
				Hp struct {
					Details string `json:"details"`
					Max     int    `json:"max"`
					Temp    int    `json:"temp"`
					Value   int    `json:"value"`
				} `json:"hp"`
				Immunities []struct {
					Type      string `json:"type"`
				} `json:"immunities"`
				Resistances []struct {
					Type       string `json:"type"`
					Value      int    `json:"value"`
				} `json:"resistances"`
				Speed struct {
					OtherSpeeds []struct {
						Type  string `json:"type"`
						Value int    `json:"value"`
					} `json:"otherSpeeds"`
					Value int `json:"value"`
				} `json:"speed"`
				Weaknesses []struct {
					Type       string `json:"type"`
					Value      int    `json:"value"`
				} `json:"weaknesses"`
			} `json:"attributes"`
			Details struct {
				Blurb     string `json:"blurb"`
				Languages struct {
					Details string `json:"details"`
					Value   []any  `json:"value"`
				} `json:"languages"`
				Level struct {
					Value int `json:"value"`
				} `json:"level"`
				PrivateNotes string `json:"privateNotes"`
				PublicNotes  string `json:"publicNotes"`
				Publication  struct {
					License  string `json:"license"`
					Remaster bool   `json:"remaster"`
					Title    string `json:"title"`
				} `json:"publication"`
			} `json:"details"`
			Initiative struct {
				Statistic string `json:"statistic"`
			} `json:"initiative"`
			Perception struct {
				Details string `json:"details"`
				Mod     int    `json:"mod"`
				Senses  []struct {
					Type   string `json:"type"`
					Acuity string `json:"acuity,omitempty"`
					Range  int    `json:"range,omitempty"`
				} `json:"senses"`
			} `json:"perception"`
			Resources struct {
			} `json:"resources"`
			Saves struct {
				Fortitude struct {
					SaveDetail string `json:"saveDetail"`
					Value      int    `json:"value"`
				} `json:"fortitude"`
				Reflex struct {
					SaveDetail string `json:"saveDetail"`
					Value      int    `json:"value"`
				} `json:"reflex"`
				Will struct {
					SaveDetail string `json:"saveDetail"`
					Value      int    `json:"value"`
				} `json:"will"`
			} `json:"saves"`
	      		Skills map[string]struct {
	                Base int `json:"base"`
	            } `json:"skills"`
			Traits struct {
				Rarity string `json:"rarity"`
				Size   struct {
					Value string `json:"value"`
				} `json:"size"`
				Value []string `json:"value"`
			} `json:"traits"`
		} `json:"system"`
		Type string `json:"type"`
		// Initiative int
		// Active bool
		// RelativeXp int
		// Conditions []*Condition
		// Counter int
		// Adjustment int
		// Uuid string
	    }
}

// Implement the Combatant interface

func (m Monster) GetName() string {
	return m.Data.Name
}

func (m Monster) GetType() string {
	return "monster"
}

func (m Monster) GetInitiative() int {
    return m.Initiative
}

func (m *Monster) SetInitiative(i int) {
    m.Initiative = i
}

func (m Monster) GetHp() int {
    return m.Data.System.Attributes.Hp.Value
}

func (m *Monster) SetHp(i int) {
    m.Data.System.Attributes.Hp.Value -= i
}

func (m Monster) GetMaxHp() int {
    return m.Data.System.Attributes.Hp.Max
}

func (m Monster) GetPerceptionMod() int {
    return m.Data.System.Perception.Mod
}

// Databas interactions

func GetAllMonsters(db database.Service) ([]Monster, error) {
    rows, err := db.Query("SELECT id, data FROM monsters")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var monsters []Monster
    for rows.Next() {
        var m Monster
        var jsonData []byte
        err := rows.Scan(&m.ID, &jsonData)
        if err != nil {
            return nil, err
        }
        err = json.Unmarshal(jsonData, &m.Data)
        if err != nil {
            return nil, err
        }
        monsters = append(monsters, m)
    }

    return monsters, nil
}

func SearchMonsters(db database.Service, search string) ([]Monster, error) {
  	query := "SELECT id, data FROM monsters WHERE LOWER(data->>'name') LIKE LOWER($1) LIMIT 10"

	// Search for the monster in the database and return the 10 most relevant results
	rows, err := db.Query(query, "%"+search+"%")
    if err != nil {
        log.Printf("Error executing query: %v", err)
        return nil, fmt.Errorf("database query error: %w", err)
    }
    defer rows.Close()

    var monsters []Monster
    for rows.Next() {
        var m Monster
        var jsonData []byte
        err := rows.Scan(&m.ID, &jsonData)
        if err != nil {
            log.Printf("Error scanning row: %v", err)
            return nil, fmt.Errorf("error scanning row: %w", err)
        }
        err = json.Unmarshal(jsonData, &m.Data)
        if err != nil {
            log.Printf("Error unmarshaling JSON data: %v", err)
            return nil, fmt.Errorf("error unmarshaling JSON: %w", err)
        }

        monsters = append(monsters, m)
    }

    if err = rows.Err(); err != nil {
        log.Printf("Error iterating over rows: %v", err)
        return nil, fmt.Errorf("error iterating over rows: %w", err)
    }

	return monsters, nil
}
