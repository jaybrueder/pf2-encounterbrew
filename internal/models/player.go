package models

import "pf2.encounterbrew.com/internal/database"

type Player struct {
	ID          int         `json:"id"`
	Name        string      `json:"name"`
	Level       int         `json:"level"`
	Hp          int         `json:"hp"`
	Ac          int         `json:"ac"`
	PartyID     int         `json:"party_id"`
	Party       *Party      `json:"party,omitempty"`
	Initiative  int         `json:"initiative"`
	Conditions  []Condition `json:"conditions"`
	Enumeration int         `json:"enumeration"`
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

func (p *Player) SetInitiative(i int) {
	p.Initiative = i
}

func (p Player) GetHp() int {
	return 1
}

func (p *Player) SetHp(i int) {}

func (p Player) GetMaxHp() int {
	return p.Hp
}

func (p Player) GetAc() int {
	return p.Ac
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
	return 0
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
	return 0
}

func (p Player) GetRef() int {
	return 0
}

func (p Player) GetWill() int {
	return 0
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
	// Find the condition in the player's conditions
	for i, c := range p.Conditions {
		if c.ID == conditionID {
			// Remove the condition from the slice
			p.Conditions = append(p.Conditions[:i], p.Conditions[i+1:]...)
			break
		}
	}

	return p.Conditions
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
