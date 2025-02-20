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
