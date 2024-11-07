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
		Hardness int `json:"hardness"`
		HP struct {
			Max int `json:"max"`
		} `json:"hp"`
		Level struct {
			Value int `json:"value"`
		} `json:"level"`
		Location struct {
			HeightenedLevel int `json:"heightenedLevel"`
			Uses struct {
				Max   int `json:"max"`
			} `json:"uses"`
		}  `json:"location"`
		Mod struct {
			Value int `json:"value"`
		} `json:"mod"`
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

func (i Item) GetAttackValue(modifier int) int {
	return i.System.Bonus.Value + modifier
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

func (i Item) GetDamageValue(modifier int) string {
	var damageString string

	for _, damageRoll := range i.System.DamageRolls {
		damageString += utils.ModifyDamage(damageRoll.Damage, modifier) + " " + damageRoll.DamageType + ", "
	}

	return utils.RemoveTrailingComma(damageString)
}

func (i Item) GetDamageEffect() string {
	if len(i.System.AttackEffects.Value) > 0 {
		return i.System.AttackEffects.Value[0].(string)
	} else {
		return ""
	}
}

func (i Item) GetSpellDC(modifier int) int {
	return i.System.Spelldc.Dc + modifier
}

func (i Item) GetSpellAttackValue(modifier int) int {
	return i.System.Spelldc.Value + modifier
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

func (i Item) FormatConsumableName() string {
	return i.GetName() + i.GetQuantity() + ", "
}

func (i Item) FormatShieldName() string {
	return i.GetName() + fmt.Sprintf(" (Hardness %d, HP %d, BT %d)", i.System.Hardness, i.System.HP.Max, i.System.HP.Max / 2) +  ", "
}

func (i Item) GetQuantity() string {
	if i.System.Quantity > 1 {
		return fmt.Sprintf(" (%d)", i.System.Quantity)
	} else {
		return ""
	}
}
