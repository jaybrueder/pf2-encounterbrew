package models

import (
	"fmt"

	"pf2.encounterbrew.com/internal/utils"
)

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
		Location struct {
			Uses struct {
				Max   int `json:"max"`
			} `json:"uses"`
		}  `json:"location"`
		Publication struct {
			License  string `json:"license"`
			Remaster bool   `json:"remaster"`
			Title    string `json:"title"`
		} `json:"publication"`
		Quantity int   `json:"quantity"`
		Rules    []any `json:"rules"`
		Runes	struct {
			Potency int `json:"potency"`
			Striking int `json:"striking"`
		} `json:"runes"`
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

func (i Item) GetActionCost() int {
	if i.System.Actions.Value == 0 {
		return 1
	} else {
		return i.System.Actions.Value
	}
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

func (i Item) FormatEquipmentName() string {
	return i.GetName() + i.GetQuantity() + ", "
}

func (i Item) FormatWeaponName() string {
	var potency string

	if i.System.Runes.Potency > 0 {
		if i.System.Runes.Striking > 0 {
			potency = fmt.Sprintf("+%d striking ", i.System.Runes.Striking)
		} else {
			potency = fmt.Sprintf("+%d ", i.System.Runes.Potency)
		}
	}

	return potency + i.GetName() + i.GetQuantity() + ", "
}

func (i Item) GetQuantity() string {
	if i.System.Quantity > 1 {
		return fmt.Sprintf(" (%d)", i.System.Quantity)
	} else {
		return ""
	}
}
