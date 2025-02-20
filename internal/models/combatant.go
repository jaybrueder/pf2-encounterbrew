package models

import (
	"sort"

	"pf2.encounterbrew.com/internal/database"
)

type Combatant interface {
	GetName() string
	SetEnumeration(int)
	GetInitiative() int
	SetInitiative(database.Service, int) error
	GenerateInitiative() int
	GetHp() int
	SetHp(int)
	GetMaxHp() int
	GetAc() int
	GetAcDetails() string
	GetType() string
	GetLevel() int
	GetSize() string
	GetTraits() []string
	GetPerceptionMod() int
	GetPerceptionSenses() string
	GetLanguages() string
	GetSkills() string
	GetLores() string
	GetStr() int
	GetDex() int
	GetCon() int
	GetInt() int
	GetWis() int
	GetCha() int
	GetFort() int
	GetRef() int
	GetWill() int
	GetImmunities() string
	GetResistances() string
	GetWeaknesses() string
	GetSpeed() string
	GetOtherSpeeds() string
	GetAttacks() []Item
	GetSpellSchool() Item
	GetSpells() OrderedItemMap
	GetDefensiveActions() []map[string]string
	GetOffensiveActions() []map[string]string
	GetInventory() string
	GetConditions() []Condition
	SetCondition(db database.Service, conditionID int, conditionValue int) []Condition
	RemoveCondition(conditionID int) []Condition
	HasCondition(conditionID int) bool
	GetConditionValue(conditionID int) int
	SetConditionValue(conditionID int, conditionValue int) int
	GetAdjustmentModifier() int
	IsMonster() bool
	IsOffGuard() bool
	AdjustConditions() map[string]int
}

func SortCombatantsByInitiative(combatants []Combatant) {
	sort.Slice(combatants, func(i, j int) bool {
		return combatants[i].GetInitiative() > combatants[j].GetInitiative()
	})
}
