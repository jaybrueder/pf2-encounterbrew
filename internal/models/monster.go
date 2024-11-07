package models

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"

	"pf2.encounterbrew.com/internal/database"
	"pf2.encounterbrew.com/internal/utils"
)

type Monster struct {
	ID         int `json:"id"`
	AssociationID int `json:"association_id"`
	LevelAdjustment int `json:"level_adjustment"`
	Initiative int `json:"initiative"`
	Conditions []Condition `json:"conditions"`
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

func (m Monster) AdjustMonster() map[string]int {
	currentLevel := m.GetOriginalLevel()
	adjustmentLevel := m.LevelAdjustment

	adjustments := map[string]int{}
	adjustments["level"] = 1 * adjustmentLevel
	adjustments["mod"] = 2 * adjustmentLevel
	adjustments["hp"] = 0

	if (adjustmentLevel + currentLevel) > currentLevel {
		// Elite monster
		switch {
			case currentLevel <= 0:
				adjustments["level"] = adjustments["level"] + 1
				adjustments["hp"] = 10
			case currentLevel == 1:
				adjustments["hp"] = 10
			case currentLevel >= 2 && currentLevel <= 4:
				adjustments["hp"] = 15
			case currentLevel >= 5 && currentLevel <= 19:
				adjustments["hp"] = 20
			case currentLevel >= 20:
				adjustments["hp"] = 30
		}

	} else if (adjustmentLevel + currentLevel) < currentLevel {
		// Weaken monster
		switch {
			case currentLevel <= 1:
				adjustments["level"] = adjustments["level"] - 1
				adjustments["hp"] = -10
			case currentLevel == 2:
				adjustments["hp"] = -10
			case currentLevel >= 3 && currentLevel <= 5:
				adjustments["hp"] = -15
			case currentLevel >= 6 && currentLevel <= 20:
				adjustments["hp"] = -20
			case currentLevel >= 21:
				adjustments["hp"] = -30
		}
	}

	return adjustments
}

// Implement the Combatant interface

func (m Monster) GetName() string {
	if m.LevelAdjustment > 0 {
		return fmt.Sprintf("Elite %s", m.Data.Name)
	} else if m.LevelAdjustment < 0 {
		return fmt.Sprintf("Weak %s", m.Data.Name)
	} else {
		return m.Data.Name
	}
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
	return m.Data.System.Attributes.Hp.Value + m.AdjustMonster()["hp"]
}

func (m *Monster) SetHp(i int) {
	m.Data.System.Attributes.Hp.Value -= i
}

func (m Monster) GetMaxHp() int {
	return m.Data.System.Attributes.Hp.Max + m.AdjustMonster()["hp"]
}

func (m Monster) GetAc() int {
	return m.Data.System.Attributes.Ac.Value + m.AdjustMonster()["mod"]
}

func (m Monster) GetAcDetails() string {
	if m.Data.System.Attributes.Ac.Details == "" {
		return ""
	} else {
		return " " + m.Data.System.Attributes.Ac.Details
	}
}

func (m Monster) GetOriginalLevel() int {
	return m.Data.System.Details.Level.Value
}

func (m Monster) GetLevel() int {
	return m.GetOriginalLevel() + m.AdjustMonster()["level"]
}

func (m Monster) GetSize() string {
	return m.Data.System.Traits.Size.Value
}

func (m Monster) GetTraits() []string {
	return m.Data.System.Traits.Value
}

func (m Monster) GetPerceptionMod() int {
	return m.Data.System.Perception.Mod + m.AdjustMonster()["mod"]
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

	if m.Data.System.Details.Languages.Details != "" {
		languages += m.Data.System.Details.Languages.Details
	}

	return utils.RemoveTrailingComma(languages)
}

func (m Monster) GetSkills() string {
	var skills string

	for key, value := range m.Data.System.Skills {
		skills += fmt.Sprintf("%s +%d, ", utils.CapitalizeFirst(key), value.Base + m.AdjustMonster()["mod"])
	}

	return utils.RemoveTrailingComma(skills)
}

func (m Monster) GetLores() string {
	var lores string

	for _, i := range m.Data.Items {
		if i.Type == "lore"  {
			lores += fmt.Sprintf(", %s +%d", utils.CapitalizeFirst(i.Name), i.System.Mod.Value)
		}
	}

	return lores
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
	return m.Data.System.Saves.Fortitude.Value + m.AdjustMonster()["mod"]
}

func (m Monster) GetRef() int {
	return m.Data.System.Saves.Reflex.Value + m.AdjustMonster()["mod"]
}

func (m Monster) GetWill() int {
	return m.Data.System.Saves.Will.Value + m.AdjustMonster()["mod"]
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
			resistances += fmt.Sprintf("%s %d, ", resistance.Type, resistance.Value)
		}
	}

	return utils.RemoveTrailingComma(resistances)
}

func (m Monster) GetWeaknesses() string {
	weaknesses := ""

	if len(m.Data.System.Attributes.Weaknesses) > 0 {
		for _, weakness := range m.Data.System.Attributes.Weaknesses {
			weaknesses += fmt.Sprintf("%s %d, ", weakness.Type, weakness.Value)
		}
	}

	return utils.RemoveTrailingComma(weaknesses)
}

func (m Monster) GetSpeed() string {
	return fmt.Sprintf("%d feet", m.Data.System.Attributes.Speed.Value)
}

func (m Monster) GetOtherSpeeds() string {
	if len(m.Data.System.Attributes.Speed.OtherSpeeds) > 0 {
		var otherSpeeds string

		for _, speed := range m.Data.System.Attributes.Speed.OtherSpeeds {
			otherSpeeds += fmt.Sprintf(", %s %d feet", speed.Type, speed.Value)
		}

		return otherSpeeds
	} else {
		return ""
	}
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

func (m Monster) GetSpells() map[string]string {
	spellsByLevel := make(map[int][]map[string]string)

 	for _, spell := range m.Data.Items {
  	 	if spell.Type == "spell" {
	        level := spell.System.Level.Value
	        if utils.Contains(spell.System.Traits.Value, "cantrip") {
	            level = 0
	        }

			if level < spell.System.Location.HeightenedLevel {
				level = spell.System.Location.HeightenedLevel
			}

	        spellInfo := map[string]string{
	            "name":        spell.Name,
	            "description": spell.System.Description.Value,
	            "level":       strconv.Itoa(level),
	            "uses":        strconv.Itoa(spell.System.Location.Uses.Max),
	            "type":        spell.Type,
	        }

			spellsByLevel[level] = append(spellsByLevel[level], spellInfo)
     	}
    }

    var levels []int
    for level := range spellsByLevel {
        levels = append(levels, level)
    }
    sort.Sort(sort.Reverse(sort.IntSlice(levels)))

    var sortedSpells []map[string]string
    for _, level := range levels {
        sortedSpells = append(sortedSpells, spellsByLevel[level]...)
    }

    return utils.FormatSortedSpells(sortedSpells, utils.DivideAndRoundUp(m.GetLevel()))
}

func (m Monster) GetActions(category string) []map[string]string {
	actions := []map[string]string{}

	for _, i := range m.Data.Items {
		if i.Type == "action" && i.System.Category == category {
			action := map[string]string{}

			action["name"] = i.Name
			action["actionType"] = i.System.ActionType.Value
			action["actionCost"] = strconv.Itoa(i.System.Actions.Value)

			var traits string
			for _, trait := range i.System.Traits.Value {
				traits += trait + ", "
			}

			action["traits"] = utils.RemoveTrailingComma(traits)

			description := i.System.Description.Value
			description = utils.RemoveHTML(description)
			action["description"] = description

			actions = append(actions, action)
		}
	}

	return actions
}

func (m Monster) GetDefensiveActions() []map[string]string {
	return m.GetActions("defensive")
}

func (m Monster) GetOffensiveActions() []map[string]string {
	return m.GetActions("offensive")
}

func (m Monster) GetInventory() string {
	var inventory string

	for _, i := range m.Data.Items {
		switch i.Type {
			case "equipment":
				inventory += i.FormatEquipmentName()
			case "weapon":
				inventory += i.FormatWeaponName()
			case "armor":
				inventory += i.FormatWeaponName()
			case "consumable":
				inventory += i.FormatConsumableName()
			case "shield":
				inventory += i.FormatShieldName()
		}
	}

	return utils.RemoveTrailingComma(inventory)
}

func (m Monster) GetConditions() []Condition {
	return m.Conditions
}

func (m *Monster) SetCondition(db database.Service, conditionID int, conditionValue int) []Condition {
	// Get condition from the database
    condition, _ := GetCondition(db, conditionID)

    // Set the condition's value
    condition.Data.System.Value.Value = conditionValue

    // Initialize the Conditions slice if it's nil
    if m.Conditions == nil {
        m.Conditions = make([]Condition, 0)
    }

    // Add the condition to the monsters's conditions
    m.Conditions = append(m.Conditions, condition)

    return m.Conditions
}

func (m *Monster) RemoveCondition(conditionID int) []Condition {
	// Find the condition in the player's conditions
	for i, c := range m.Conditions {
		if c.ID == conditionID {
			// Remove the condition from the slice
			m.Conditions = append(m.Conditions[:i], m.Conditions[i+1:]...)
			break
		}
	}

	return m.Conditions
}

func (m Monster) GetAdjustmentModifier() int {
	return m.AdjustMonster()["mod"]
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
