package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"pf2.encounterbrew.com/internal/database"
)

type Party struct {
	ID      int      `json:"id"`
	Name    string   `json:"name"`
	UserID  int      `json:"user_id"`
	User    *User    `json:"user,omitempty"`
	Players []Player `json:"players,omitempty"`
}

func (p *Party) GetLevel() float64 {
	if len(p.Players) == 0 {
		return 0
	}

	totalLevel := 0
	for _, player := range p.Players {
		totalLevel += player.Level
	}

	average := float64(totalLevel) / float64(len(p.Players))

	return average
}

func (p Party) GetNumbersOfPlayer() float64 {
	return float64(len(p.Players))
}

// Database interaction
func (p *Party) Create(db database.Service) (int, error) {
	if db == nil {
		return 0, errors.New("database service is nil")
	}

	id, err := db.InsertReturningID(
		"parties",
		[]string{"name", "user_id"},
		p.Name, p.UserID,
	)
	if err != nil {
		return 0, fmt.Errorf("error creating party: %v", err)
	}

	return id, nil
}

func GetAllParties(db database.Service) ([]Party, error) {
	if db == nil {
		return nil, errors.New("database service is nil")
	}

	// First get all parties
	rows, err := db.Query(`
        SELECT p.id, p.name, p.user_id, u.name AS user_name
        FROM parties p
        JOIN users u ON p.user_id = u.id
        WHERE p.user_id = $1
        ORDER BY p.id
    `, 1)
	if err != nil {
		return nil, fmt.Errorf("error querying parties: %v", err)
	}
	defer rows.Close()

	var parties []Party
	for rows.Next() {
		var p Party
		p.User = &User{}

		err := rows.Scan(&p.ID, &p.Name, &p.UserID, &p.User.Name)
		if err != nil {
			return nil, fmt.Errorf("error scanning party row: %v", err)
		}

		// Get players for this party
		playerRows, err := db.Query(`
            SELECT id, name, level, hp, ac, fort, ref, will
            FROM players
            WHERE party_id = $1
        `, p.ID)
		if err != nil {
			return nil, fmt.Errorf("error querying players: %v", err)
		}
		defer playerRows.Close()

		var players []Player
		for playerRows.Next() {
			var player Player
			err := playerRows.Scan(&player.ID, &player.Name, &player.Level, &player.Hp, &player.Ac, &player.Fort, &player.Ref, &player.Will)
			if err != nil {
				return nil, fmt.Errorf("error scanning player row: %v", err)
			}
			player.PartyID = p.ID
			players = append(players, player)
		}

		if err = playerRows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating player rows: %v", err)
		}

		p.Players = players
		parties = append(parties, p)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating party rows: %v", err)
	}

	return parties, nil
}

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
        SELECT id, name, level, hp, ac, fort, ref, will
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
		err := rows.Scan(&player.ID, &player.Name, &player.Level, &player.Hp, &player.Ac, &player.Fort, &player.Ref, &player.Will)
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

func PartyExists(db database.Service, partyID int) (bool, error) {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM parties WHERE id = $1)
	`, partyID).Scan(&exists)

	return exists, err
}

// Update updates the party's name in the database
func (p *Party) Update(db database.Service) error {
	if db == nil {
		return errors.New("database service is nil")
	}

	result, err := db.Exec(`
        UPDATE parties
        SET name = $1
        WHERE id = $2 AND user_id = $3`,
		p.Name, p.ID, p.UserID)

	if err != nil {
		return fmt.Errorf("error updating party: %v", err)
	}

	// Check if any row was affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("party not found or user not authorized")
	}

	return nil
}

// UpdateWithPlayers updates the party and all its players in a single transaction
func (p *Party) UpdateWithPlayers(db database.Service, playersToDelete []int) error {
	if db == nil {
		return errors.New("database service is nil")
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}

	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("error rolling back transaction: %v", err)
		}
	}()

	// Update party
	_, err = tx.Exec(`
        UPDATE parties
        SET name = $1
        WHERE id = $2 AND user_id = $3`,
		p.Name, p.ID, p.UserID)

	if err != nil {
		return fmt.Errorf("error updating party: %v", err)
	}

	// Delete removed players
	if len(playersToDelete) > 0 {
		placeholders := make([]string, len(playersToDelete))
		args := make([]interface{}, len(playersToDelete))
		for i, id := range playersToDelete {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			args[i] = id
		}
		//nolint:gosec
		query := fmt.Sprintf("DELETE FROM players WHERE id IN (%s)", strings.Join(placeholders, ","))
		_, err := tx.Exec(query, args...)
		if err != nil {
			return fmt.Errorf("error deleting players: %v", err)
		}
	}

	// Update or insert players
	for _, player := range p.Players {
		if player.ID == 0 {
			// Insert new player
			err := tx.QueryRow(`
                INSERT INTO players (name, level, ac, hp, fort, ref, will, party_id)
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
                RETURNING id`,
				player.Name, player.Level, player.Ac, player.Hp, player.Fort, player.Ref, player.Will, p.ID).Scan(&player.ID)
			if err != nil {
				return fmt.Errorf("error inserting player: %v", err)
			}
		} else {
			// Update existing player
			result, err := tx.Exec(`
                UPDATE players
                SET name = $1, level = $2, ac = $3, hp = $4, fort = $5, ref = $6, will = $7
                WHERE id = $8 AND party_id = $9`,
				player.Name, player.Level, player.Ac, player.Hp, player.Fort, player.Ref, player.Will, player.ID, p.ID)

			if err != nil {
				return fmt.Errorf("error updating player: %v", err)
			}

			rowsAffected, err := result.RowsAffected()
			if err != nil {
				return fmt.Errorf("error getting rows affected: %v", err)
			}
			if rowsAffected == 0 {
				return fmt.Errorf("player %d not found or not associated with party", player.ID)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}

// Removes party and associated players
func (p *Party) Delete(db database.Service) error {
	if db == nil {
		return errors.New("database service is nil")
	}

	result, err := db.Exec(`
        DELETE FROM parties
        WHERE id = $1 AND user_id = $2`,
		p.ID, p.UserID)
	if err != nil {
		return fmt.Errorf("error deleting party: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("party not found or user not authorized")
	}

	return nil
}
