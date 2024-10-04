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
}

// Implement the Combatant interface

func (p Player) GetName() string {
	return p.Name
}

func (p Player) GetType() string {
    return "player"
}
