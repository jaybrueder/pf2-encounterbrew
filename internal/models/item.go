package models

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
		Category string `json:"category"`
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
		Quantity int `json:"quantity"`
		Rules  []any `json:"rules"`
		Slug   any   `json:"slug"`
		Spelldc struct {
			Dc  int `json:"dc"`
			Mod int  `json:"mod"`
			Value int`json:"value"`
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
