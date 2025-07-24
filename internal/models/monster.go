package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"

	"github.com/lib/pq"
	"pf2.encounterbrew.com/internal/database"
	"pf2.encounterbrew.com/internal/utils"
)

type Monster struct {
	ID              int         `json:"id"`
	AssociationID   int         `json:"association_id"`
	LevelAdjustment int         `json:"level_adjustment"`
	Enumeration     int         `json:"enumeration"`
	Initiative      int         `json:"initiative"`
	Conditions      []Condition `json:"conditions"`
	Data            struct {
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
					Details string   `json:"details"`
					Value   []string `json:"value"`
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
				Details string  `json:"details"`
				Mod     int     `json:"mod"`
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

func (m Monster) AdjustConditions() map[string]int {
	conditions := map[string]int{}
	conditions["ac"] = 0

	if m.IsOffGuard() {
		conditions["ac"] = -2
	}

	return conditions
}

// Implement the Combatant interface

func (m Monster) GetName() string {
	var name string

	if m.LevelAdjustment > 0 {
		name = fmt.Sprintf("Elite %s", m.Data.Name)
	} else if m.LevelAdjustment < 0 {
		name = fmt.Sprintf("Weak %s", m.Data.Name)
	} else {
		name = m.Data.Name
	}

	if m.Enumeration > 0 {
		name = fmt.Sprintf("%s %d", name, m.Enumeration)
	}

	return name
}

func (m Monster) GetType() string {
	return "monster"
}

func (m Monster) GetInitiative() int {
	return m.Initiative
}

func (m *Monster) SetInitiative(db database.Service, i int) error {
	// Update the local struct
	m.Initiative = i

	// Update the database
	_, err := db.Exec(`
        UPDATE encounter_monsters
        SET initiative = $1
        WHERE id = $2
    `, i, m.AssociationID)

	if err != nil {
		return fmt.Errorf("error updating monster initiative in database: %v", err)
	}

	return nil
}

func (m Monster) GetHp() int {
	return m.Data.System.Attributes.Hp.Value + m.AdjustMonster()["hp"]
}

func (m *Monster) SetHp(db database.Service, i int) error {
	m.Data.System.Attributes.Hp.Value -= i

	// Update the hp in the encounter_monsters table
	_, err := db.Exec(`
        UPDATE encounter_monsters
        SET hp = $1
        WHERE id = $2
    `, m.Data.System.Attributes.Hp.Value, m.AssociationID)

	if err != nil {
		return fmt.Errorf("error updating monster hp in database: %v", err)
	}

	return nil
}

func (m Monster) GetMaxHp() int {
	return m.Data.System.Attributes.Hp.Max + m.AdjustMonster()["hp"]
}

func (m Monster) GetAc() int {
	return m.Data.System.Attributes.Ac.Value + m.AdjustMonster()["mod"] + m.AdjustConditions()["ac"]
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
		skills += fmt.Sprintf("%s +%d, ", utils.CapitalizeFirst(key), value.Base+m.AdjustMonster()["mod"])
	}

	return utils.RemoveTrailingComma(skills)
}

func (m Monster) GetLores() string {
	var lores string

	for _, i := range m.Data.Items {
		if i.Type == "lore" {
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

func (m Monster) GetCon() int {
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

func (m Monster) GetSpells() OrderedItemMap {
	spellsByLevel := make(map[int][]Item)

	for _, spell := range m.Data.Items {
		if spell.Type == "spell" {
			level := spell.System.Level.Value
			if utils.Contains(spell.System.Traits.Value, "cantrip") {
				level = 0
			}

			if level < spell.System.Location.HeightenedLevel {
				level = spell.System.Location.HeightenedLevel
			}

			spellsByLevel[level] = append(spellsByLevel[level], spell)
		}
	}

	return CreateSortedOrderedItemMap(spellsByLevel)
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
			action["description"] = utils.RemoveHTML(i.GetDescription())

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

func (m Monster) GetInteractions() []map[string]string {
	return m.GetActions("interaction")
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

func (m *Monster) SetConditions(conditions []Condition) {
	m.Conditions = conditions
}

func (m *Monster) SetCondition(db database.Service, encounterID int, conditionID int, conditionValue int) error {
	// Initialize the Conditions slice if it's nil
	if m.Conditions == nil {
		m.Conditions = make([]Condition, 0)
	}

	// Check if the condition already exists
	for i, c := range m.Conditions {
		if c.ID == conditionID {
			// Increment the existing condition's value
			newValue := c.GetValue() + conditionValue
			c.SetValue(newValue)
			m.Conditions[i] = c

			// Update the condition in the database
			_, err := db.Exec(`
				UPDATE combatant_conditions
				SET condition_value = $1
				WHERE encounter_id = $2 AND encounter_monster_id = $3 AND condition_id = $4
			`, newValue, encounterID, m.AssociationID, conditionID)

			if err != nil {
				return fmt.Errorf("error updating condition in combatant_conditions: %v", err)
			}

			return nil
		}
	}

	// Condition doesn't exist, so add it
	// Get condition from the database
	condition, err := GetCondition(db, conditionID)
	if err != nil {
		return err
	}

	// Set the condition's value
	condition.Data.System.Value.Value = conditionValue

	// Add the condition to the monster's conditions
	m.Conditions = append(m.Conditions, condition)

	// Insert the condition into the combatant_conditions table
	_, err = db.Exec(`
        INSERT INTO combatant_conditions (encounter_id, encounter_monster_id, condition_id, condition_value)
        VALUES ($1, $2, $3, $4)
    `, encounterID, m.AssociationID, conditionID, conditionValue)

	if err != nil {
		return fmt.Errorf("error inserting condition into combatant_conditions: %v", err)
	}

	return nil
}

func (m *Monster) RemoveCondition(db database.Service, encounterID int, conditionID int) error {
	// Find the condition in the monster's conditions
	for i, c := range m.Conditions {
		if c.ID == conditionID {
			// Remove the condition from the slice
			m.Conditions = append(m.Conditions[:i], m.Conditions[i+1:]...)
			break
		}
	}

	// Remove the condition from the combatant_conditions table
	_, err := db.Exec(`
        DELETE FROM combatant_conditions
        WHERE encounter_id = $1 AND encounter_monster_id = $2 AND condition_id = $3
    `, encounterID, m.AssociationID, conditionID)

	if err != nil {
		return fmt.Errorf("error removing condition from combatant_conditions: %v", err)
	}

	return nil
}

func (m *Monster) HasCondition(conditionID int) bool {
	for _, c := range m.Conditions {
		if c.ID == conditionID {
			return true
		}
	}

	return false
}

func (m *Monster) GetConditionValue(conditionID int) int {
	for _, c := range m.Conditions {
		if c.ID == conditionID {
			return c.GetValue()
		}
	}

	return 0
}

func (m *Monster) SetConditionValue(conditionID int, conditionValue int) int {
	for _, c := range m.Conditions {
		if c.ID == conditionID {
			c.SetValue(conditionValue)
			return c.GetValue()
		}
	}

	return 0
}

func (m Monster) GetAdjustmentModifier() int {
	return m.AdjustMonster()["mod"]
}

func (m *Monster) SetEnumeration(value int) {
	m.Enumeration = value
}

func (m Monster) IsMonster() bool {
	return true
}

func (m Monster) IsOffGuard() bool {
	for _, c := range m.Conditions {
		if c.GetName() == "Off-Guard" {
			return true
		}
	}

	return false
}

func (m Monster) GenerateInitiative() int {
	//nolint:gosec
	return rand.Intn(20) + 1 + m.GetPerceptionMod()
}

func (m *Monster) GetAssociationID() int {
	return m.AssociationID
}

// Database interactions

func GetAllMonsters(db database.Service) ([]Monster, error) {
	rows, err := db.Query("SELECT id, data FROM monsters")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Printf("error closing rows: %v\n", err)
		}
	}()

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

type MonsterSearchFilters struct {
	MinLevel        *int     `json:"min_level"`
	MaxLevel        *int     `json:"max_level"`
	ExcludedSources []string `json:"excluded_sources"`
	ExcludedSizes   []string `json:"excluded_sizes"`
}

func SearchMonsters(db database.Service, search string) ([]Monster, error) {
	return SearchMonstersWithFilters(db, search, MonsterSearchFilters{})
}

func SearchMonstersWithFilters(db database.Service, search string, filters MonsterSearchFilters) ([]Monster, error) {
	// Build dynamic query with filters
	queryArgs := []interface{}{search}
	argCounter := 1

	whereConditions := []string{}

	// Level filter conditions
	if filters.MinLevel != nil {
		argCounter++
		whereConditions = append(whereConditions, fmt.Sprintf("(data->'system'->'details'->'level'->>'value')::int >= $%d", argCounter))
		queryArgs = append(queryArgs, *filters.MinLevel)
	}

	if filters.MaxLevel != nil {
		argCounter++
		whereConditions = append(whereConditions, fmt.Sprintf("(data->'system'->'details'->'level'->>'value')::int <= $%d", argCounter))
		queryArgs = append(queryArgs, *filters.MaxLevel)
	}

	// Check if "Other" is in excluded sources
	hasOther := false
	otherIndex := -1
	for i, source := range filters.ExcludedSources {
		if source == "Other" {
			hasOther = true
			otherIndex = i
			break
		}
	}

	// If "Other" is checked, handle it specially
	if hasOther {
		// Remove "Other" from the excluded sources
		filters.ExcludedSources = append(filters.ExcludedSources[:otherIndex], filters.ExcludedSources[otherIndex+1:]...)
		
		// Include ONLY the core books (excluding all other sources)
		coreBooks := []string{
			"Pathfinder Bestiary",
			"Pathfinder Bestiary 2",
			"Pathfinder Bestiary 3",
			"Pathfinder Monster Core",
		}
		argCounter++
		whereConditions = append(whereConditions, fmt.Sprintf("(data->'system'->'details'->'publication'->>'title' = ANY($%d))", argCounter))
		queryArgs = append(queryArgs, pq.Array(coreBooks))
	}

	// Source exclusion filter (for remaining sources)
	if len(filters.ExcludedSources) > 0 {
		argCounter++
		whereConditions = append(whereConditions, fmt.Sprintf("NOT (data->'system'->'details'->'publication'->>'title' = ANY($%d))", argCounter))
		queryArgs = append(queryArgs, pq.Array(filters.ExcludedSources))
	}

	// Size exclusion filter
	if len(filters.ExcludedSizes) > 0 {
		argCounter++
		whereConditions = append(whereConditions, fmt.Sprintf("NOT (data->'system'->'traits'->'size'->>'value' = ANY($%d))", argCounter))
		queryArgs = append(queryArgs, pq.Array(filters.ExcludedSizes))
	}

	// Build WHERE clause for filters
	filterClause := ""
	if len(whereConditions) > 0 {
		filterClause = " WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Use a more sophisticated search that prioritizes exact matches, then prefix matches, then partial matches
	query := fmt.Sprintf(`
		WITH filtered_monsters AS (
			SELECT id, data, LOWER(data->>'name') as name_lower
			FROM monsters
			%s
		),
		search_results AS (
			-- Exact match (case-insensitive)
			SELECT id, data, 1 as priority, name_lower
			FROM filtered_monsters
			WHERE name_lower = LOWER($1)

			UNION

			-- Prefix match (case-insensitive)
			SELECT id, data, 2 as priority, name_lower
			FROM filtered_monsters
			WHERE name_lower LIKE LOWER($1 || '%%')
			AND name_lower != LOWER($1)

			UNION

			-- Contains match (case-insensitive)
			SELECT id, data, 3 as priority, name_lower
			FROM filtered_monsters
			WHERE name_lower LIKE LOWER('%%' || $1 || '%%')
			AND name_lower NOT LIKE LOWER($1 || '%%')
		)
		SELECT id, data, priority, name_lower
		FROM search_results
		ORDER BY priority, name_lower
		LIMIT 10
	`, filterClause)

	rows, err := db.Query(query, queryArgs...)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return nil, fmt.Errorf("database query error: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Printf("error closing rows: %v\n", err)
		}
	}()

	var monsters []Monster
	for rows.Next() {
		var m Monster
		var jsonData []byte
		var priority int
		var nameLower string
		err := rows.Scan(&m.ID, &jsonData, &priority, &nameLower)
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

func GetMonster(db database.Service, id int) (Monster, error) {
	if db == nil {
		return Monster{}, errors.New("database service is nil")
	}

	var m Monster
	var jsonData []byte
	err := db.QueryRow("SELECT id, data FROM monsters WHERE id = $1", id).Scan(&m.ID, &jsonData)

	if err != nil {
		if err == sql.ErrNoRows {
			return Monster{}, fmt.Errorf("no monster found with ID %d", id)
		}
		return Monster{}, fmt.Errorf("error scanning monster row: %v", err)
	}

	err = json.Unmarshal(jsonData, &m.Data)
	if err != nil {
		return Monster{}, fmt.Errorf("error unmarshaling monster data: %v", err)
	}

	return m, nil
}
