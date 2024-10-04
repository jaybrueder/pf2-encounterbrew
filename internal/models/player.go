package models

import ()

type Player struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Level   int    `json:"level"`
	Hp	  	int    `json:"hp"`
	Ac     	int    `json:"ac"`
	PartyID int    `json:"party_id"`
	Party   *Party `json:"party,omitempty"`
	Initiative int `json:"initiative"`
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
    return 0
}

func (p *Player) SetHp(i int) {}

func (p Player) GetMaxHp() int {
    return 0
}

func (p Player) GetAc() int {
    return 0
}

func (p Player) GetPerceptionMod() int {
    return 0
}
