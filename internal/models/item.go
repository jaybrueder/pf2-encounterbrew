package models

import "pf2.encounterbrew.com/internal/utils"

type Item struct {
	ID    string `json:"_id"`
	Flags struct {
		Pf2E struct {
			LinkedWeapon string `json:"linkedWeapon"`
		} `json:"pf2e"`
	} `json:"flags"`
	Img    string `json:"img"`
	Name   string `json:"name"`
	Sort   int    `json:"sort"`
	System struct {
		ActionType struct {
			Value string `json:"value"`
		} `json:"actionType"`
		Actions struct {
			Value int `json:"value"`
		} `json:"actions"`
		Attack struct {
			Value string `json:"value"`
		} `json:"attack"`
		AttackEffects struct {
			Value []any `json:"value"`
		} `json:"attackEffects"`
		Bonus struct {
			Value int `json:"value"`
		} `json:"bonus"`
		Category    string `json:"category"`
		DamageRolls map[string]struct {
			Damage     string `json:"damage"`
			DamageType string `json:"damageType"`
		} `json:"damageRolls"`
		Description struct {
			Value string `json:"value"`
		} `json:"description"`
		Level struct {
			Value int `json:"value"`
		} `json:"level"`
		Publication struct {
			License  string `json:"license"`
			Remaster bool   `json:"remaster"`
			Title    string `json:"title"`
		} `json:"publication"`
		Quantity int   `json:"quantity"`
		Rules    []any `json:"rules"`
		Slug     any   `json:"slug"`
		Spelldc  struct {
			Dc    int `json:"dc"`
			Mod   int `json:"mod"`
			Value int `json:"value"`
		} `json:"spelldc"`
		Traits struct {
			Value []string `json:"value"`
		} `json:"traits"`
		WeaponType struct {
			Value string `json:"value"`
		} `json:"weaponType"`
	} `json:"system"`
	Type string `json:"type"`
}

func (i Item) GetWeaponType() string {
	return utils.CapitalizeFirst(i.System.WeaponType.Value)
}

func (i Item) GetName() string {
	return utils.CapitalizeFirst(i.Name)
}

func (i Item) GetAttackValue() int {
	return i.System.Bonus.Value
}

func (i Item) GetTraits() string {
	var traits string

	for _, trait := range i.System.Traits.Value {
		traits += trait + ", "
	}

	if len(traits) > 0 {
		return " (" + utils.RemoveTrailingComma(traits) + ")"
	} else {
		return ""
	}
}

func (i Item) GetDamageValue() string {
	var damage string

	for _, damageRoll := range i.System.DamageRolls {
		damage += damageRoll.Damage + " " + damageRoll.DamageType + ", "
	}

	return utils.RemoveTrailingComma(damage)
}

func (i Item) GetSpellDC() int {
	return i.System.Spelldc.Dc
}

func (i Item) GetSpellAttackValue() int {
	return i.System.Spelldc.Value
}
