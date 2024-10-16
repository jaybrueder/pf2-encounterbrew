package models

import (
	"math/rand"
	"sort"
)

type Combatant interface {
	GetName() string
	GetInitiative() int
	SetInitiative(int)
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
	GetSpells() []map[string]string
	GetDefensiveActions() []map[string]string
	GetOffensiveActions() []map[string]string
	GetInventory() string
}

func AssignInitiative(combatants []Combatant) {
	for _, c := range combatants {
		c.SetInitiative(rand.Intn(20) + 1 + c.GetPerceptionMod())
	}
}

func SortCombatantsByInitiative(combatants []Combatant) {
	sort.Slice(combatants, func(i, j int) bool {
		return combatants[i].GetInitiative() > combatants[j].GetInitiative()
	})
}
