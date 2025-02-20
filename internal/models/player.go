package models

import (
	"errors"
	"fmt"
	"math/rand"

	"pf2.encounterbrew.com/internal/database"
)

type Player struct {
	ID            int         `json:"id"`
	AssociationID int         `json:"association_id"`
	Name          string      `json:"name"`
	Level         int         `json:"level"`
	Hp            int         `json:"hp"`
	Ac            int         `json:"ac"`
	Fort          int         `json:"for"`
	Ref           int         `json:"ref"`
	Will          int         `json:"wil"`
	Perception    int         `json:"perception"`
	PartyID       int         `json:"party_id"`
	Party         *Party      `json:"party,omitempty"`
	Initiative    int         `json:"initiative"`
	Conditions    []Condition `json:"conditions"`
	Enumeration   int         `json:"enumeration"`
}

// Implement the Combatant interface

func (p Player) GetName() string {
	return p.Name
}

func (p Player) GetType() string {
	return "player"
}

func (p Player) GetInitiative() int {
	return p.Initiative
}

func (p *Player) SetInitiative(db database.Service, i int) error {
	// Update the local struct
	p.Initiative = i

	// Update the database
	_, err := db.Exec(`
        UPDATE encounter_players
        SET initiative = $1
        WHERE id = $2
    `, i, p.AssociationID)

	if err != nil {
		return fmt.Errorf("error updating player initiative in database: %v", err)
	}

	return nil
}

func (p Player) GetHp() int {
	return p.Hp
}

func (p *Player) SetHp(i int) {
	p.Hp -= i
}

func (p Player) GetMaxHp() int {
	return p.Hp
}

func (p Player) GetAc() int {
	return p.Ac + p.AdjustConditions()["ac"]
}

func (p Player) GetAcDetails() string {
	return ""
}

func (p Player) GetLevel() int {
	return p.Level
}

func (p Player) GetSize() string {
	return ""
}

func (p Player) GetTraits() []string {
	return []string{}
}

func (p Player) GetPerceptionMod() int {
	return p.Perception
}

func (p Player) GetPerceptionSenses() string {
	return ""
}

func (p Player) GetLanguages() string {
	return ""
}

func (p Player) GetSkills() string {
	return ""
}

func (p Player) GetLores() string {
	return ""
}

func (p Player) GetStr() int {
	return 0
}

func (p Player) GetDex() int {
	return 0
}

func (p Player) GetCon() int {
	return 0
}

func (p Player) GetInt() int {
	return 0
}

func (p Player) GetWis() int {
	return 0
}

func (p Player) GetCha() int {
	return 0
}

func (p Player) GetFort() int {
	return p.Fort
}

func (p Player) GetRef() int {
	return p.Ref
}

func (p Player) GetWill() int {
	return p.Will
}

func (p Player) GetImmunities() string {
	return ""
}

func (p Player) GetResistances() string {
	return ""
}

func (p Player) GetWeaknesses() string {
	return ""
}

func (p Player) GetSpeed() string {
	return ""
}

func (p Player) GetOtherSpeeds() string {
	return ""
}

func (p Player) GetAttacks() []Item {
	return []Item{}
}

func (p Player) GetSpellSchool() Item {
	return Item{}
}

func (p Player) GetSpells() OrderedItemMap {
	return CreateSortedOrderedItemMap(map[int][]Item{})
}

func (p Player) GetDefensiveActions() []map[string]string {
	return []map[string]string{}
}

func (p Player) GetOffensiveActions() []map[string]string {
	return []map[string]string{}
}

func (p Player) GetInventory() string {
	return ""
}

func (p Player) GetConditions() []Condition {
	return p.Conditions
}

func (p *Player) SetCondition(db database.Service, conditionID int, conditionValue int) []Condition {
	// Get condition from the database
	condition, _ := GetCondition(db, conditionID)

	// Set the condition's value
	condition.Data.System.Value.Value = conditionValue

	// Initialize the Conditions slice if it's nil
	if p.Conditions == nil {
		p.Conditions = make([]Condition, 0)
	}

	// Add the condition to the player's conditions
	p.Conditions = append(p.Conditions, condition)

	return p.Conditions
}

func (p *Player) RemoveCondition(conditionID int) []Condition {
	for i, c := range p.Conditions {
		if c.ID == conditionID {
			// Remove the condition from the slice
			p.Conditions = append(p.Conditions[:i], p.Conditions[i+1:]...)
			break
		}
	}

	return p.Conditions
}

func (p *Player) HasCondition(conditionID int) bool {
	for _, c := range p.Conditions {
		if c.ID == conditionID {
			return true
		}
	}

	return false
}

func (p *Player) GetConditionValue(conditionID int) int {
	for _, c := range p.Conditions {
		if c.ID == conditionID {
			return c.GetValue()
		}
	}

	return 0
}

func (p *Player) SetConditionValue(conditionID int, conditionValue int) int {
	for _, c := range p.Conditions {
		if c.ID == conditionID {
			c.SetValue(conditionValue)
			return c.GetValue()
		}
	}

	return 0
}

func (p Player) GetAdjustmentModifier() int {
	return 0
}

func (p *Player) SetEnumeration(value int) {
	p.Enumeration = value
}

func (p Player) IsMonster() bool {
	return false
}

func (p Player) IsOffGuard() bool {
	for _, c := range p.Conditions {
		if c.GetName() == "Off-Guard" {
			return true
		}
	}

	return false
}

func (p Player) AdjustConditions() map[string]int {
	conditions := map[string]int{}
	conditions["ac"] = 0

	if p.IsOffGuard() {
		conditions["ac"] = -2
	}

	return conditions
}

func (p Player) GenerateInitiative() int {
	//nolint:gosec
	return rand.Intn(20) + 1 + p.GetPerceptionMod()
}

// Database interactions
func PlayerDelete(db database.Service, id int) error {
	if db == nil {
		return errors.New("database service is nil")
	}

	_, err := db.Exec(`
        DELETE FROM players
        WHERE id = $1
    `, id)

	return err
}

// Update updates the player's details in the database
func (p *Player) Update(db database.Service) error {
	if db == nil {
		return errors.New("database service is nil")
	}

	err := db.QueryRow(`
        UPDATE players
        SET name = $1, level = $2, ac = $3, hp = $4, fort = $5, ref = $6, will = $7
        WHERE id = $8 AND party_id = $9
        RETURNING id`,
		p.Name, p.Level, p.Ac, p.Hp, p.Fort, p.Ref, p.Will, p.ID, p.PartyID).Scan(&p.ID)

	if err != nil {
		return fmt.Errorf("error updating player: %v", err)
	}

	return nil
}
