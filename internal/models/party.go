package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"pf2.encounterbrew.com/internal/database"
)

type Party struct {
	ID      int      `json:"id"`
	Name    string   `json:"name"`
	UserID  int      `json:"user_id"`
	User    *User    `json:"user,omitempty"`
	Players []Player `json:"players,omitempty"`
}

// Database interaction

func GetParty(db database.Service, id string) (Party, error) {
	if db == nil {
		return Party{}, errors.New("database service is nil")
	}

	partyID, err := strconv.Atoi(id)
	if err != nil {
		return Party{}, fmt.Errorf("invalid party ID: %v", err)
	}

	var p Party
	p.User = &User{}

	err = db.QueryRow(`
        SELECT p.id, p.name, p.user_id, u.name AS user_name
        FROM parties p
        JOIN users u ON p.user_id = u.id
        WHERE p.user_id = $1 AND p.id = $2
    `, 1, partyID).Scan(&p.ID, &p.Name, &p.User.ID, &p.User.Name)

	if err != nil {
		if err == sql.ErrNoRows {
			return Party{}, fmt.Errorf("no party found with ID %d", partyID)
		}
		return Party{}, fmt.Errorf("error scanning party row: %v", err)
	}

	// Query for associated players
	rows, err := db.Query(`
        SELECT id, name, level, hp, ac
        FROM players
        WHERE party_id = $1
    `, partyID)
	if err != nil {
		return Party{}, fmt.Errorf("error querying players: %v", err)
	}
	defer rows.Close()

	var players []Player
	for rows.Next() {
		var player Player
		err := rows.Scan(&player.ID, &player.Name, &player.Level, &player.Hp, &player.Ac)
		if err != nil {
			return Party{}, fmt.Errorf("error scanning player row: %v", err)
		}
		player.PartyID = partyID
		players = append(players, player)
	}

	if err = rows.Err(); err != nil {
		return Party{}, fmt.Errorf("error iterating player rows: %v", err)
	}

	p.Players = players

	return p, nil
}
