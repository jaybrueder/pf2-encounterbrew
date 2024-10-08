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
	GetType() string
	GetLevel() int
	GetPerceptionMod() int
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
