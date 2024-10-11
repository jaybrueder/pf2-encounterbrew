package models

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"pf2.encounterbrew.com/internal/database"
	"pf2.encounterbrew.com/internal/utils"
)

type Monster struct {
	ID         int `json:"id"`
	Adjustment int `json:"adjustment"`
	Count      int `json:"count"`
	Initiative int `json:"initiative"`
	Data       struct {
		ID     string `json:"_id"`
		Img    string `json:"img"`
		Items  []Item `json:"items"`
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
					Type string `json:"type"`
				} `json:"immunities"`
				Resistances []struct {
					Type  string `json:"type"`
					Value int    `json:"value"`
				} `json:"resistances"`
				Speed struct {
					OtherSpeeds []struct {
						Type  string `json:"type"`
						Value int    `json:"value"`
					} `json:"otherSpeeds"`
					Value int `json:"value"`
				} `json:"speed"`
				Weaknesses []struct {
					Type  string `json:"type"`
					Value int    `json:"value"`
				} `json:"weaknesses"`
			} `json:"attributes"`
			Details struct {
				Blurb     string `json:"blurb"`
				Languages struct {
					Details string `json:"details"`
					Value   []string  `json:"value"`
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
				Senses  []Sense `json:"senses"`
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
	}
}

type Sense struct {
	Type   string `json:"type"`
	Acuity string `json:"acuity,omitempty"`
	Range  int    `json:"range,omitempty"`
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

func (m Monster) GetAc() int {
	return m.Data.System.Attributes.Ac.Value
}

func (m Monster) GetLevel() int {
	return m.Data.System.Details.Level.Value
}

func (m Monster) GetSize() string {
	return m.Data.System.Traits.Size.Value
}

func (m Monster) GetTraits() []string {
	return m.Data.System.Traits.Value
}

func (m Monster) GetPerceptionMod() int {
	return m.Data.System.Perception.Mod
}

func (m Monster) GetPerceptionSenses() string {
	var senses string

	for _, sense := range m.Data.System.Perception.Senses {
		if sense.Range == 0 {
			senses += fmt.Sprintf("%s, ", sense.Type)
		} else {
			senses += fmt.Sprintf("%s (%s) %dft, ", sense.Type, sense.Acuity, sense.Range)
		}
	}

	return utils.RemoveTrailingComma(senses)
}

func (m Monster) GetLanguages() string {
	var languages string

	for _, language := range m.Data.System.Details.Languages.Value {
		languages += fmt.Sprintf("%s, ", utils.CapitalizeFirst(language))
	}

	return utils.RemoveTrailingComma(languages)
}

func (m Monster) GetSkills() string {
	var skills string

	for key, value := range m.Data.System.Skills {
		skills += fmt.Sprintf("%s +%d, ", utils.CapitalizeFirst(key), value.Base)
	}

	return utils.RemoveTrailingComma(skills)
}

func (m Monster) GetStr() int {
	return m.Data.System.Abilities.Str.Mod
}

func (m Monster) GetDex() int {
	return m.Data.System.Abilities.Dex.Mod
}

func (m Monster) GetCon() int{
	return m.Data.System.Abilities.Con.Mod
}

func (m Monster) GetInt() int {
	return m.Data.System.Abilities.Int.Mod
}

func (m Monster) GetWis() int {
	return m.Data.System.Abilities.Wis.Mod
}

func (m Monster) GetCha() int {
	return m.Data.System.Abilities.Cha.Mod
}

func (m Monster) GetFort() int {
	return m.Data.System.Saves.Fortitude.Value
}

func (m Monster) GetRef() int {
	return m.Data.System.Saves.Reflex.Value
}

func (m Monster) GetWill() int {
	return m.Data.System.Saves.Will.Value
}

func (m Monster) GetImmunities() string {
	immunities := ""

	if len(m.Data.System.Attributes.Immunities) > 0 {
		for _, immunity := range m.Data.System.Attributes.Immunities {
			immunities += fmt.Sprintf("%s, ", immunity.Type)
		}
	}

	return utils.RemoveTrailingComma(immunities)
}

func (m Monster) GetResistances() string {
	resistances := ""

	if len(m.Data.System.Attributes.Resistances) > 0 {
		for _, resistance := range m.Data.System.Attributes.Resistances {
			resistances += fmt.Sprintf("%s, ", resistance.Type)
		}
	}

	return utils.RemoveTrailingComma(resistances)
}

func (m Monster) GetWeaknesses() string {
	weaknesses := ""

	if len(m.Data.System.Attributes.Weaknesses) > 0 {
		for _, weakness := range m.Data.System.Attributes.Weaknesses {
			weaknesses += fmt.Sprintf("%s, ", weakness.Type)
		}
	}

	return utils.RemoveTrailingComma(weaknesses)
}

func (m Monster) GetSpeed() string {
	return fmt.Sprintf("%d feet", m.Data.System.Attributes.Speed.Value)
}

func (m Monster) GetAttacks() []Item {
	attacks := []Item{}

	for _, i := range m.Data.Items {
		if i.Type == "melee" {
			attacks = append(attacks, i)
		}
	}

	return attacks
}

func (m Monster) GetSpellSchool() Item {
	spellSchool := Item{}

	for _, s := range m.Data.Items {
		if s.Type == "spellcastingEntry" {
			spellSchool = s
		}
	}

	return spellSchool
}

func (m Monster) GetSpells() []map[string]string {
	spells := []map[string]string{}

	for _, i := range m.Data.Items {
		if i.Type == "spell" {
			spell := map[string]string{}

			spell["name"] = i.Name

			description := i.System.Description.Value
			description = strings.ReplaceAll(description, "<p>", "")
			description = strings.ReplaceAll(description, "</p>", "")
			spell["description"] = description

			spell["level"] = strconv.Itoa(i.System.Level.Value)
			spell["type"] = i.Type

			spells = append(spells, spell)
		}
	}
	return spells
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
