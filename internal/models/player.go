package models

type Player struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Level      int    `json:"level"`
	Hp         int    `json:"hp"`
	Ac         int    `json:"ac"`
	PartyID    int    `json:"party_id"`
	Party      *Party `json:"party,omitempty"`
	Initiative int    `json:"initiative"`
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

func (p Player) GetStr() int {
	return 0
}

func (p Player) GetDex() int {
	return 0
}

func (p Player) GetCon() int{
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

func (p Player) GetAttacks() []Item {
	return []Item{}
}

func (p Player) GetSpellSchool() Item {
	return Item{}
}

func (p Player) GetSpells() []map[string]string  {
	return []map[string]string{}
}

func (p Player) GetDefensiveActions() []map[string]string {
	return []map[string]string{}
}

func (p Player) GetOffensiveActions()[]map[string]string {
	return []map[string]string{}
}

func (p Player) GetInventory() string {
	return ""
}
