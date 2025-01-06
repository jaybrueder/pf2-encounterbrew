package models

import (
	"database/sql"
	"errors"
	"fmt"

	"pf2.encounterbrew.com/internal/database"
)

type User struct {
	ID            int
	Name          string
	ActivePartyID int `json:"active_party_id,omitempty"`
}

func GetUserByID(db database.Service, id int) (*User, error) {
	if db == nil {
		return nil, errors.New("database service is nil")
	}

	user := &User{}
	err := db.QueryRow(`
        SELECT id, name, active_party_id
        FROM users
        WHERE id = $1`,
		id).Scan(&user.ID, &user.Name, &user.ActivePartyID)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no user found with ID %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("error getting user: %v", err)
	}

	return user, nil
}

func (u *User) SetActiveParty(db database.Service, partyID int) error {
	if db == nil {
		return errors.New("database service is nil")
	}

	// First verify that the party belongs to this user
	var count int
	err := db.QueryRow(`
        SELECT COUNT(*)
        FROM parties
        WHERE id = $1 AND user_id = $2`,
		partyID, u.ID).Scan(&count)

	if err != nil {
		return fmt.Errorf("error verifying party ownership: %v", err)
	}

	if count == 0 {
		return errors.New("party not found or does not belong to user")
	}

	// Update the active party
	_, err = db.Exec(`
        UPDATE users
        SET active_party_id = $1
        WHERE id = $2`,
		partyID, u.ID)

	if err != nil {
		return fmt.Errorf("error setting active party: %v", err)
	}

	u.ActivePartyID = partyID
	return nil
}

// func (u *User) GetActiveParty(db database.Service) (*Party, error) {
//     if db == nil {
//         return nil, errors.New("database service is nil")
//     }

//     if u.ActivePartyID == nil {
//         return nil, nil // No active party set
//     }

//     party := &Party{}
//     err := db.QueryRow(`
//         SELECT id, name, user_id
//         FROM parties
//         WHERE id = $1 AND user_id = $2`,
//         *u.ActivePartyID, u.ID).Scan(&party.ID, &party.Name, &party.UserID)

//     if err == sql.ErrNoRows {
//         return nil, nil
//     }
//     if err != nil {
//         return nil, fmt.Errorf("error getting active party: %v", err)
//     }

//     // Optionally load the party's players
//     rows, err := db.Query(`
//         SELECT id, name, level, hp, ac, fort, ref, will
//         FROM players
//         WHERE party_id = $1`,
//         party.ID)
//     if err != nil {
//         return nil, fmt.Errorf("error getting party players: %v", err)
//     }
//     defer rows.Close()

//     for rows.Next() {
//         var player Player
//         err := rows.Scan(
//             &player.ID, &player.Name, &player.Level,
//             &player.Hp, &player.Ac, &player.Fort,
//             &player.Ref, &player.Will)
//         if err != nil {
//             return nil, fmt.Errorf("error scanning player: %v", err)
//         }
//         party.Players = append(party.Players, player)
//     }

//     return party, nil
// }
