package models

import (
	"fmt"
	"sort"

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
		Area struct {
			Type  string `json:"type"`
			Value int    `json:"value"`
		} `json:"area"`
		Bonus struct {
			Value int `json:"value"`
		} `json:"bonus"`
		Category    string `json:"category"`
		DamageRolls map[string]struct {
			Damage     string `json:"damage"`
			DamageType string `json:"damageType"`
		} `json:"damageRolls"`
		Defense struct {
			Save struct {
				Basic     bool   `json:"basic"`
				Statistic string `json:"statistic"`
			} `json:"save"`
		} `json:"defense"`
		Description struct {
			Value string `json:"value"`
		} `json:"description"`
		Duration struct {
			Sustained bool   `json:"sustained"`
			Value     string `json:"value"`
		} `json:"duration"`

		Hardness int `json:"hardness"`
		HP       struct {
			Max int `json:"max"`
		} `json:"hp"`
		Level struct {
			Value int `json:"value"`
		} `json:"level"`
		Location struct {
			HeightenedLevel int `json:"heightenedLevel"`
			Uses            struct {
				Max int `json:"max"`
			} `json:"uses"`
		} `json:"location"`
		Mod struct {
			Value int `json:"value"`
		} `json:"mod"`
		Publication struct {
			License  string `json:"license"`
			Remaster bool   `json:"remaster"`
			Title    string `json:"title"`
		} `json:"publication"`
		Quantity int         `json:"quantity"`
		Range    interface{} `json:"range"`
		Rules    []any       `json:"rules"`
		Runes    struct {
			Potency  int `json:"potency"`
			Striking int `json:"striking"`
		} `json:"runes"`
		Slug    any `json:"slug"`
		Spelldc struct {
			Dc    int `json:"dc"`
			Mod   int `json:"mod"`
			Value int `json:"value"`
		} `json:"spelldc"`
		Target struct {
			Value string `json:"value"`
		} `json:"target"`
		Time struct {
			Value string `json:"value"`
		} `json:"time"`
		Traits struct {
			Value      []string `json:"value"`
			Traditions []string `json:"traditions"`
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

func (i Item) GetDescription() string {
	localizer, _ := utils.GetLocalizer("data/lang/en.json")

	description := localizer.ProcessText(i.System.Description.Value)
	description = utils.NewReplacer().ProcessText(description)

	return description
}

func (i Item) GetLevel() int {
	return i.System.Level.Value
}

func (i Item) GetSpellTraits() []string {
	return i.System.Traits.Value
}

func (i Item) GetSpellTraditions() string {
	traditions := ""

	for _, tradition := range i.System.Traits.Traditions {
		traditions += tradition + ", "
	}

	return traditions
}

func (i Item) GetSpellDefense() string {
	return i.System.Defense.Save.Statistic
}

func (i *Item) GetRange() string {
	switch v := i.System.Range.(type) {
	case map[string]interface{}:
		if val, ok := v["value"].(string); ok {
			return val
		}
	case float64: // JSON numbers are unmarshaled as float64
		return fmt.Sprintf("%d", int(v))
	}
	return ""
}

func (i Item) HasCastTime() bool {
	if i.System.Time.Value != "1" && i.System.Time.Value != "2" && i.System.Time.Value != "3" && i.System.Time.Value != "1 to 3" && i.System.Time.Value != "reaction" {
		return true
	} else {
		return false
	}
}

func (i Item) GetSpellTarget() string {
	return i.System.Target.Value
}

func (i Item) GetSpellDuration() string {
	return i.System.Duration.Value
}

func (i Item) GetSpellTime() string {
	if i.System.Time.Value == "1 to 3" {
		return "1-3"
	} else {
		return i.System.Time.Value
	}
}

func (i Item) GetSpellArea() string {
	area := ""

	if i.System.Area.Value > 0 {
		area = fmt.Sprintf("%d-foot %s", i.System.Area.Value, i.System.Area.Type)
	}

	return area
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
	return i.GetName() + fmt.Sprintf(" (Hardness %d, HP %d, BT %d)", i.System.Hardness, i.System.HP.Max, i.System.HP.Max/2) + ", "
}

func (i Item) GetQuantity() string {
	if i.System.Quantity > 1 {
		return fmt.Sprintf(" (%d)", i.System.Quantity)
	} else {
		return ""
	}
}

// OrderedItemMap is a map of Items with ordered keys
type OrderedItemMap struct {
	Data map[int][]Item
	Keys []int
}

func CreateSortedOrderedItemMap(originalMap map[int][]Item) OrderedItemMap {
	newMap := OrderedItemMap{
		Data: make(map[int][]Item),
		Keys: make([]int, 0, len(originalMap)),
	}

	// Get and sort keys
	for k := range originalMap {
		newMap.Keys = append(newMap.Keys, k)
	}
	sort.Ints(newMap.Keys)

	// Populate map
	for _, k := range newMap.Keys {
		newMap.Data[k] = originalMap[k]
	}

	return newMap
}
