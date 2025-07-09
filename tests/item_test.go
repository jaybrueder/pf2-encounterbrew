package tests

import (
	"strings"
	"testing"

	"pf2.encounterbrew.com/internal/models"
)

const (
	reactionTime    = "reaction"
	testSwordPrefix = "Test sword, "
)

func createSampleItem() models.Item {
	return models.Item{
		ID: "test-item-id",
		Flags: struct {
			Pf2E struct {
				LinkedWeapon string `json:"linkedWeapon"`
			} `json:"pf2e"`
		}{
			Pf2E: struct {
				LinkedWeapon string `json:"linkedWeapon"`
			}{
				LinkedWeapon: "test-weapon",
			},
		},
		Img:  "path/to/item.jpg",
		Name: "test sword",
		Sort: 100,
		System: struct {
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
				Sustained bool                         `json:"sustained"`
				Value     models.FlexibleDurationValue `json:"value"`
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
		}{
			ActionType: struct {
				Value string `json:"value"`
			}{
				Value: "action",
			},
			Actions: struct {
				Value int `json:"value"`
			}{
				Value: 1,
			},
			Attack: struct {
				Value string `json:"value"`
			}{
				Value: "attack",
			},
			AttackEffects: struct {
				Value []any `json:"value"`
			}{
				Value: []any{"bleed"},
			},
			Area: struct {
				Type  string `json:"type"`
				Value int    `json:"value"`
			}{
				Type:  "cone",
				Value: 15,
			},
			Bonus: struct {
				Value int `json:"value"`
			}{
				Value: 5,
			},
			Category: "weapon",
			DamageRolls: map[string]struct {
				Damage     string `json:"damage"`
				DamageType string `json:"damageType"`
			}{
				"damage1": {
					Damage:     "1d8",
					DamageType: "slashing",
				},
			},
			Defense: struct {
				Save struct {
					Basic     bool   `json:"basic"`
					Statistic string `json:"statistic"`
				} `json:"save"`
			}{
				Save: struct {
					Basic     bool   `json:"basic"`
					Statistic string `json:"statistic"`
				}{
					Basic:     true,
					Statistic: "reflex",
				},
			},
			Description: struct {
				Value string `json:"value"`
			}{
				Value: "A sharp sword",
			},
			Duration: struct {
				Sustained bool                         `json:"sustained"`
				Value     models.FlexibleDurationValue `json:"value"`
			}{
				Sustained: false,
				Value: models.FlexibleDurationValue{
					Value: "1 minute",
				},
			},
			Hardness: 8,
			HP: struct {
				Max int `json:"max"`
			}{
				Max: 20,
			},
			Level: struct {
				Value int `json:"value"`
			}{
				Value: 3,
			},
			Location: struct {
				HeightenedLevel int `json:"heightenedLevel"`
				Uses            struct {
					Max int `json:"max"`
				} `json:"uses"`
			}{
				HeightenedLevel: 0,
				Uses: struct {
					Max int `json:"max"`
				}{
					Max: 3,
				},
			},
			Mod: struct {
				Value int `json:"value"`
			}{
				Value: 2,
			},
			Publication: struct {
				License  string `json:"license"`
				Remaster bool   `json:"remaster"`
				Title    string `json:"title"`
			}{
				License:  "OGL",
				Remaster: true,
				Title:    "Test Publication",
			},
			Quantity: 1,
			Range:    30,
			Rules:    []any{},
			Runes: struct {
				Potency  int `json:"potency"`
				Striking int `json:"striking"`
			}{
				Potency:  1,
				Striking: 0,
			},
			Slug: "test-sword",
			Spelldc: struct {
				Dc    int `json:"dc"`
				Mod   int `json:"mod"`
				Value int `json:"value"`
			}{
				Dc:    15,
				Mod:   3,
				Value: 8,
			},
			Target: struct {
				Value string `json:"value"`
			}{
				Value: "1 creature",
			},
			Time: struct {
				Value string `json:"value"`
			}{
				Value: "1",
			},
			Traits: struct {
				Value      []string `json:"value"`
				Traditions []string `json:"traditions"`
			}{
				Value:      []string{"magical", "evocation"},
				Traditions: []string{"arcane", "divine"},
			},
			WeaponType: struct {
				Value string `json:"value"`
			}{
				Value: "sword",
			},
		},
		Type: "weapon",
	}
}

func TestFlexibleDurationValue_UnmarshalJSON_String(t *testing.T) {
	fdv := models.FlexibleDurationValue{}
	jsonData := []byte(`"1 minute"`)

	err := fdv.UnmarshalJSON(jsonData)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if fdv.Value != "1 minute" {
		t.Errorf("expected '1 minute', got '%s'", fdv.Value)
	}
}

func TestFlexibleDurationValue_UnmarshalJSON_Number(t *testing.T) {
	fdv := models.FlexibleDurationValue{}
	jsonData := []byte(`60`)

	err := fdv.UnmarshalJSON(jsonData)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if fdv.Value != "60" {
		t.Errorf("expected '60', got '%s'", fdv.Value)
	}
}

func TestFlexibleDurationValue_UnmarshalJSON_Float(t *testing.T) {
	fdv := models.FlexibleDurationValue{}
	jsonData := []byte(`60.5`)

	err := fdv.UnmarshalJSON(jsonData)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if fdv.Value != "60.5" {
		t.Errorf("expected '60.5', got '%s'", fdv.Value)
	}
}

func TestFlexibleDurationValue_UnmarshalJSON_Invalid(t *testing.T) {
	fdv := models.FlexibleDurationValue{}
	jsonData := []byte(`{"invalid": "object"}`)

	err := fdv.UnmarshalJSON(jsonData)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}

	expectedError := "value must be either a string or a number"
	if err.Error() != expectedError {
		t.Errorf("expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestItem_GetWeaponType(t *testing.T) {
	item := createSampleItem()
	weaponType := item.GetWeaponType()

	expected := "Sword"
	if weaponType != expected {
		t.Errorf("expected '%s', got '%s'", expected, weaponType)
	}
}

func TestItem_GetName(t *testing.T) {
	item := createSampleItem()
	name := item.GetName()

	expected := "Test sword"
	if name != expected {
		t.Errorf("expected '%s', got '%s'", expected, name)
	}
}

func TestItem_GetLevel(t *testing.T) {
	item := createSampleItem()
	level := item.GetLevel()

	expected := 3
	if level != expected {
		t.Errorf("expected %d, got %d", expected, level)
	}
}

func TestItem_GetSpellTraits(t *testing.T) {
	item := createSampleItem()
	traits := item.GetSpellTraits()

	expected := []string{"magical", "evocation"}
	if len(traits) != len(expected) {
		t.Errorf("expected %d traits, got %d", len(expected), len(traits))
	}

	for i, trait := range expected {
		if traits[i] != trait {
			t.Errorf("expected trait '%s', got '%s'", trait, traits[i])
		}
	}
}

func TestItem_GetSpellTraditions(t *testing.T) {
	item := createSampleItem()
	traditions := item.GetSpellTraditions()

	expected := "arcane, divine, "
	if traditions != expected {
		t.Errorf("expected '%s', got '%s'", expected, traditions)
	}
}

func TestItem_GetSpellDefense(t *testing.T) {
	item := createSampleItem()
	defense := item.GetSpellDefense()

	expected := "reflex"
	if defense != expected {
		t.Errorf("expected '%s', got '%s'", expected, defense)
	}
}

func TestItem_GetRange_Float64(t *testing.T) {
	item := createSampleItem()
	item.System.Range = float64(30)

	rangeValue := (&item).GetRange()

	expected := "30"
	if rangeValue != expected {
		t.Errorf("expected '%s', got '%s'", expected, rangeValue)
	}
}

func TestItem_GetRange_Int(t *testing.T) {
	item := createSampleItem()
	item.System.Range = 30

	rangeValue := (&item).GetRange()

	// Int type is not handled by GetRange, should return empty string
	expected := ""
	if rangeValue != expected {
		t.Errorf("expected '%s', got '%s'", expected, rangeValue)
	}
}

func TestItem_GetRange_FloatValue(t *testing.T) {
	item := createSampleItem()
	item.System.Range = float64(30.5)

	rangeValue := (&item).GetRange()

	expected := "30"
	if rangeValue != expected {
		t.Errorf("expected '%s', got '%s'", expected, rangeValue)
	}
}

func TestItem_GetRange_MapWithValue(t *testing.T) {
	item := createSampleItem()
	item.System.Range = map[string]interface{}{
		"value": "touch",
	}
	rangeValue := (&item).GetRange()

	expected := "touch"
	if rangeValue != expected {
		t.Errorf("expected '%s', got '%s'", expected, rangeValue)
	}
}

func TestItem_GetRange_MapWithoutValue(t *testing.T) {
	item := createSampleItem()
	item.System.Range = map[string]interface{}{
		"other": "data",
	}
	rangeValue := (&item).GetRange()

	expected := ""
	if rangeValue != expected {
		t.Errorf("expected '%s', got '%s'", expected, rangeValue)
	}
}

func TestItem_GetRange_Nil(t *testing.T) {
	item := createSampleItem()
	item.System.Range = nil
	rangeValue := (&item).GetRange()

	expected := ""
	if rangeValue != expected {
		t.Errorf("expected '%s', got '%s'", expected, rangeValue)
	}
}

func TestItem_HasCastTime_True(t *testing.T) {
	item := createSampleItem()
	item.System.Time.Value = "10 minutes"

	if !item.HasCastTime() {
		t.Error("expected HasCastTime to return true for '10 minutes'")
	}
}

func TestItem_HasCastTime_False_One(t *testing.T) {
	item := createSampleItem()
	item.System.Time.Value = "1"

	if item.HasCastTime() {
		t.Error("expected HasCastTime to return false for '1'")
	}
}

func TestItem_HasCastTime_False_Two(t *testing.T) {
	item := createSampleItem()
	item.System.Time.Value = "2"

	if item.HasCastTime() {
		t.Error("expected HasCastTime to return false for '2'")
	}
}

func TestItem_HasCastTime_False_Three(t *testing.T) {
	item := createSampleItem()
	item.System.Time.Value = "3"

	if item.HasCastTime() {
		t.Error("expected HasCastTime to return false for '3'")
	}
}

func TestItem_HasCastTime_False_OneToThree(t *testing.T) {
	item := createSampleItem()
	item.System.Time.Value = "1 to 3"

	if item.HasCastTime() {
		t.Error("expected HasCastTime to return false for '1 to 3'")
	}
}

func TestItem_HasCastTime_False_Reaction(t *testing.T) {
	item := createSampleItem()
	item.System.Time.Value = reactionTime

	if item.HasCastTime() {
		t.Error("expected HasCastTime to return false for 'reaction'")
	}
}

func TestItem_GetSpellTarget(t *testing.T) {
	item := createSampleItem()
	target := item.GetSpellTarget()

	expected := "1 creature"
	if target != expected {
		t.Errorf("expected '%s', got '%s'", expected, target)
	}
}

func TestItem_GetSpellDuration(t *testing.T) {
	item := createSampleItem()
	duration := item.GetSpellDuration()

	expected := "1 minute"
	if duration != expected {
		t.Errorf("expected '%s', got '%s'", expected, duration)
	}
}

func TestItem_GetSpellTime_OneToThree(t *testing.T) {
	item := createSampleItem()
	item.System.Time.Value = "1 to 3"
	time := item.GetSpellTime()

	expected := "1-3"
	if time != expected {
		t.Errorf("expected '%s', got '%s'", expected, time)
	}
}

func TestItem_GetSpellTime_Other(t *testing.T) {
	item := createSampleItem()
	item.System.Time.Value = reactionTime
	time := item.GetSpellTime()

	expected := reactionTime
	if time != expected {
		t.Errorf("expected '%s', got '%s'", expected, time)
	}
}

func TestItem_GetSpellArea(t *testing.T) {
	item := createSampleItem()
	area := item.GetSpellArea()

	expected := "15-foot cone"
	if area != expected {
		t.Errorf("expected '%s', got '%s'", expected, area)
	}
}

func TestItem_GetSpellArea_ZeroValue(t *testing.T) {
	item := createSampleItem()
	item.System.Area.Value = 0
	area := item.GetSpellArea()

	expected := ""
	if area != expected {
		t.Errorf("expected '%s', got '%s'", expected, area)
	}
}

func TestItem_GetAttackValue(t *testing.T) {
	item := createSampleItem()
	attackValue := item.GetAttackValue(3)

	expected := 8 // 5 + 3
	if attackValue != expected {
		t.Errorf("expected %d, got %d", expected, attackValue)
	}
}

func TestItem_GetActionCost_Zero(t *testing.T) {
	item := createSampleItem()
	item.System.Actions.Value = 0
	actionCost := item.GetActionCost()

	expected := 1
	if actionCost != expected {
		t.Errorf("expected %d, got %d", expected, actionCost)
	}
}

func TestItem_GetActionCost_NonZero(t *testing.T) {
	item := createSampleItem()
	item.System.Actions.Value = 2
	actionCost := item.GetActionCost()

	expected := 2
	if actionCost != expected {
		t.Errorf("expected %d, got %d", expected, actionCost)
	}
}

func TestItem_GetTraits_WithTraits(t *testing.T) {
	item := createSampleItem()
	traits := item.GetTraits()

	expected := " (magical, evocation)"
	if traits != expected {
		t.Errorf("expected '%s', got '%s'", expected, traits)
	}
}

func TestItem_GetTraits_Empty(t *testing.T) {
	item := createSampleItem()
	item.System.Traits.Value = []string{}
	traits := item.GetTraits()

	expected := ""
	if traits != expected {
		t.Errorf("expected '%s', got '%s'", expected, traits)
	}
}

func TestItem_GetDamageValue(t *testing.T) {
	item := createSampleItem()
	damageValue := item.GetDamageValue(2)

	expected := "1d8+2 slashing"
	if damageValue != expected {
		t.Errorf("expected '%s', got '%s'", expected, damageValue)
	}
}

func TestItem_GetDamageValue_MultipleDamage(t *testing.T) {
	item := createSampleItem()
	item.System.DamageRolls["damage2"] = struct {
		Damage     string `json:"damage"`
		DamageType string `json:"damageType"`
	}{
		Damage:     "1d4",
		DamageType: "fire",
	}
	damageValue := item.GetDamageValue(1)

	// Result should contain both damage types (order may vary due to map iteration)
	if !strings.Contains(damageValue, "slashing") || !strings.Contains(damageValue, "fire") {
		t.Errorf("expected damage string to contain both 'slashing' and 'fire', got '%s'", damageValue)
	}
}

func TestItem_GetDamageEffect_WithEffect(t *testing.T) {
	item := createSampleItem()
	effect := item.GetDamageEffect()

	expected := "bleed"
	if effect != expected {
		t.Errorf("expected '%s', got '%s'", expected, effect)
	}
}

func TestItem_GetDamageEffect_Empty(t *testing.T) {
	item := createSampleItem()
	item.System.AttackEffects.Value = []any{}
	effect := item.GetDamageEffect()

	expected := ""
	if effect != expected {
		t.Errorf("expected '%s', got '%s'", expected, effect)
	}
}

func TestItem_GetSpellDC(t *testing.T) {
	item := createSampleItem()
	spellDC := item.GetSpellDC(2)

	expected := 17 // 15 + 2
	if spellDC != expected {
		t.Errorf("expected %d, got %d", expected, spellDC)
	}
}

func TestItem_GetSpellAttackValue(t *testing.T) {
	item := createSampleItem()
	attackValue := item.GetSpellAttackValue(3)

	expected := 11 // 8 + 3
	if attackValue != expected {
		t.Errorf("expected %d, got %d", expected, attackValue)
	}
}

func TestItem_FormatEquipmentName(t *testing.T) {
	item := createSampleItem()
	formatted := item.FormatEquipmentName()

	expected := testSwordPrefix
	if formatted != expected {
		t.Errorf("expected '%s', got '%s'", expected, formatted)
	}
}

func TestItem_FormatEquipmentName_WithQuantity(t *testing.T) {
	item := createSampleItem()
	item.System.Quantity = 3
	formatted := item.FormatEquipmentName()

	expected := "Test sword (3), "
	if formatted != expected {
		t.Errorf("expected '%s', got '%s'", expected, formatted)
	}
}

func TestItem_FormatWeaponName_WithPotency(t *testing.T) {
	item := createSampleItem()
	item.System.Runes.Potency = 2
	formatted := item.FormatWeaponName()

	expected := "+2 Test sword, "
	if formatted != expected {
		t.Errorf("expected '%s', got '%s'", expected, formatted)
	}
}

func TestItem_FormatWeaponName_WithStriking(t *testing.T) {
	item := createSampleItem()
	item.System.Runes.Potency = 1
	item.System.Runes.Striking = 2
	formatted := item.FormatWeaponName()

	expected := "+2 striking Test sword, "
	if formatted != expected {
		t.Errorf("expected '%s', got '%s'", expected, formatted)
	}
}

func TestItem_FormatWeaponName_NoRunes(t *testing.T) {
	item := createSampleItem()
	item.System.Runes.Potency = 0
	item.System.Runes.Striking = 0
	formatted := item.FormatWeaponName()

	expected := testSwordPrefix
	if formatted != expected {
		t.Errorf("expected '%s', got '%s'", expected, formatted)
	}
}

func TestItem_FormatConsumableName(t *testing.T) {
	item := createSampleItem()
	formatted := item.FormatConsumableName()

	expected := testSwordPrefix
	if formatted != expected {
		t.Errorf("expected '%s', got '%s'", expected, formatted)
	}
}

func TestItem_FormatShieldName(t *testing.T) {
	item := createSampleItem()
	formatted := item.FormatShieldName()

	expected := "Test sword (Hardness 8, HP 20, BT 10), "
	if formatted != expected {
		t.Errorf("expected '%s', got '%s'", expected, formatted)
	}
}

func TestItem_GetQuantity_Single(t *testing.T) {
	item := createSampleItem()
	quantity := item.GetQuantity()

	expected := ""
	if quantity != expected {
		t.Errorf("expected '%s', got '%s'", expected, quantity)
	}
}

func TestItem_GetQuantity_Multiple(t *testing.T) {
	item := createSampleItem()
	item.System.Quantity = 5
	quantity := item.GetQuantity()

	expected := " (5)"
	if quantity != expected {
		t.Errorf("expected '%s', got '%s'", expected, quantity)
	}
}

func TestCreateSortedOrderedItemMap(t *testing.T) {
	originalMap := map[int][]models.Item{
		3: {createSampleItem()},
		1: {createSampleItem()},
		2: {createSampleItem()},
	}

	orderedMap := models.CreateSortedOrderedItemMap(originalMap)

	expectedKeys := []int{1, 2, 3}
	if len(orderedMap.Keys) != len(expectedKeys) {
		t.Errorf("expected %d keys, got %d", len(expectedKeys), len(orderedMap.Keys))
	}

	for i, key := range expectedKeys {
		if orderedMap.Keys[i] != key {
			t.Errorf("expected key %d at position %d, got %d", key, i, orderedMap.Keys[i])
		}
	}

	// Check that all data is preserved
	for key, items := range originalMap {
		if len(orderedMap.Data[key]) != len(items) {
			t.Errorf("expected %d items for key %d, got %d", len(items), key, len(orderedMap.Data[key]))
		}
	}
}

func TestCreateSortedOrderedItemMap_EmptyMap(t *testing.T) {
	originalMap := map[int][]models.Item{}

	orderedMap := models.CreateSortedOrderedItemMap(originalMap)

	if len(orderedMap.Keys) != 0 {
		t.Errorf("expected 0 keys, got %d", len(orderedMap.Keys))
	}

	if len(orderedMap.Data) != 0 {
		t.Errorf("expected 0 data entries, got %d", len(orderedMap.Data))
	}
}
