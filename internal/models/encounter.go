package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"pf2.encounterbrew.com/internal/database"
)

type Encounter struct {
	ID                int                        `json:"id"`
	Name              string                     `json:"name"`
	UserID            int                        `json:"user_id"`
	User              *User                      `json:"user,omitempty"`
	PartyID           int                        `json:"party_id"`
	Party             *Party                     `json:"party,omitempty"`
	Monsters          []*Monster                 `json:"monsters,omitempty"`
	Players           []*Player                  `json:"players,omitempty"`
	Combatants        []Combatant                `json:"combatants,omitempty"`
	Round             int                        `json:"round"`
	Turn              int                        `json:"turn"`
	GroupedConditions map[string][]ConditionInfo `json:"grouped_conditions"`
}

func CreateEncounter(db database.Service, name string, partyId int) (Encounter, error) {
	if db == nil {
		return Encounter{}, errors.New("database service is nil")
	}

	var e Encounter
	e.Name = name
	e.PartyID = partyId

	err := db.QueryRow(`
	    INSERT INTO encounters (name, party_id, user_id)
	    VALUES ($1, $2, $3)
	    RETURNING id
    `, name, partyId, 1).Scan(&e.ID)

	if err != nil {
		return Encounter{}, err
	}

	// Get all players from the new party
	rows, err := db.Query(`
		SELECT id, hp FROM players
		WHERE party_id = $1
	`, partyId)
	if err != nil {
		return Encounter{}, fmt.Errorf("error getting party players: %v", err)
	}
	defer rows.Close()

	// Collect all player IDs and their HP
	var playerIDs []int
	var playerHPs []int
	for rows.Next() {
		var playerID, playerHP int
		if err := rows.Scan(&playerID, &playerHP); err != nil {
			return Encounter{}, fmt.Errorf("error scanning player ID and HP: %v", err)
		}
		playerIDs = append(playerIDs, playerID)
		playerHPs = append(playerHPs, playerHP)
	}
	if err = rows.Err(); err != nil {
		return Encounter{}, fmt.Errorf("error iterating over players: %v", err)
	}

	// Add each player from the new party to the encounter with their initial HP
	for i, playerID := range playerIDs {
		_, err = db.Exec(`
			INSERT INTO encounter_players (encounter_id, player_id, initiative, hp)
			VALUES ($1, $2, $3, $4)
		`, e.ID, playerID, 0, playerHPs[i])

		if err != nil {
			return Encounter{}, fmt.Errorf("error adding player to encounter: %v", err)
		}
	}

	return e, nil
}

func UpdateEncounter(db database.Service, encounterId int, name string, partyId int) error {
	if db == nil {
		return errors.New("database service is nil")
	}

	// First, verify the new party exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM parties WHERE id = $1)", partyId).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error checking party existence: %v", err)
	}
	if !exists {
		return fmt.Errorf("party with ID %d does not exist", partyId)
	}

	// First, get the current party_id
	var currentPartyID int
	err = db.QueryRow(`
		SELECT party_id
		FROM encounters
		WHERE id = $1
	`, encounterId).Scan(&currentPartyID)

	if err != nil {
		return fmt.Errorf("failed to get current encounter: %w", err)
	}

	// Check if party_id has changed
	if currentPartyID != partyId {
		// Start a transaction for the party change
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("error starting transaction: %v", err)
		}
		defer func() {
			if tx != nil {
				_ = tx.Rollback()
			}
		}()

		// First, remove all existing players from the encounter
		_, err = tx.Exec(`
			DELETE FROM encounter_players
			WHERE encounter_id = $1
		`, encounterId)
		if err != nil {
			return fmt.Errorf("error removing existing players: %v", err)
		}

		// Get all players from the new party
		rows, err := tx.Query(`
			SELECT id, hp FROM players
			WHERE party_id = $1
		`, partyId)
		if err != nil {
			return fmt.Errorf("error getting party players: %v", err)
		}
		defer rows.Close()

		// Collect all player IDs and their HP
		var playerIDs []int
		var playerHPs []int
		for rows.Next() {
			var playerID, playerHP int
			if err := rows.Scan(&playerID, &playerHP); err != nil {
				return fmt.Errorf("error scanning player ID and HP: %v", err)
			}
			playerIDs = append(playerIDs, playerID)
			playerHPs = append(playerHPs, playerHP)
		}
		if err = rows.Err(); err != nil {
			return fmt.Errorf("error iterating over players: %v", err)
		}

		// Add each player from the new party to the encounter with their initial HP
		for i, playerID := range playerIDs {
			_, err = tx.Exec(`
				INSERT INTO encounter_players (encounter_id, player_id, initiative, hp)
				VALUES ($1, $2, $3, $4)
			`, encounterId, playerID, 0, playerHPs[i])

			if err != nil {
				return fmt.Errorf("error adding player to encounter: %v", err)
			}
		}

		// Update the encounter itself
		_, err = tx.Exec(`
			UPDATE encounters
			SET name = $1, party_id = $2
			WHERE id = $3
		`, name, partyId, encounterId)

		if err != nil {
			return fmt.Errorf("failed to update encounter: %w", err)
		}

		// Commit the transaction
		if err = tx.Commit(); err != nil {
			return fmt.Errorf("error committing transaction: %v", err)
		}
		tx = nil // Prevent rollback after successful commit
	} else {
		// If party hasn't changed, just update the name
		_, err = db.Exec(`
			UPDATE encounters
			SET name = $1
			WHERE id = $2
		`, name, encounterId)

		if err != nil {
			return fmt.Errorf("failed to update encounter: %w", err)
		}
	}

	return nil
}

func UpdateTurnAndRound(db database.Service, turn int, round int, id int) error {
	if id == 0 {
		return errors.New("invalid encounter ID")
	}

	if turn < 0 || round < 0 {
		return errors.New("invalid turn or round")
	}

	fmt.Printf("Turn and round updates: %d %d", turn, round)

	_, err := db.Exec(`
		UPDATE encounters
		SET turn = $1, round = $2
		WHERE id = $3
	`, turn, round, id)

	if err != nil {
		return fmt.Errorf("failed to update encounter: %w", err)
	}

	return nil
}

func DeleteEncounter(db database.Service, id int) error {
	if db == nil {
		return errors.New("database service is nil")
	}

	_, err := db.Exec(`
		DELETE FROM encounters
		WHERE id = $1
	`, id)

	if err != nil {
		return fmt.Errorf("error deleting encounter: %v", err)
	}

	return nil
}

func GetEncounter(db database.Service, encounterId int) (Encounter, error) {
	if db == nil {
		return Encounter{}, errors.New("database service is nil")
	}

	var e Encounter
	e.User = &User{}
	e.Party = &Party{}

	err := db.QueryRow(`
       SELECT e.id, e.name, e.user_id, e.party_id, e.turn, e.round, u.name AS user_name, p.name AS party_name
       FROM encounters e
       JOIN users u ON e.user_id = u.id
       JOIN parties p ON e.party_id = p.id
       WHERE e.user_id = $1 AND e.id = $2
   `, 1, encounterId).Scan(&e.ID, &e.Name, &e.User.ID, &e.Party.ID, &e.Turn, &e.Round, &e.User.Name, &e.Party.Name)

	if err != nil {
		if err == sql.ErrNoRows {
			return Encounter{}, fmt.Errorf("no encounter found with ID %d", encounterId)
		}
		return Encounter{}, fmt.Errorf("error scanning encounter row: %v", err)
	}

	rows, err := db.Query(`
        SELECT m.id, m.data, em.level_adjustment, em.id, em.initiative, em.hp as current_hp, em.enumeration
        FROM monsters m
        JOIN encounter_monsters em ON m.id = em.monster_id
        WHERE em.encounter_id = $1
    `, encounterId)
	if err != nil {
		return e, fmt.Errorf("error querying monsters: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var m Monster
		var jsonData []byte
		var currentHp int
		err := rows.Scan(
			&m.ID,
			&jsonData,
			&m.LevelAdjustment,
			&m.AssociationID,
			&m.Initiative,
			&currentHp,
			&m.Enumeration,
		)
		if err != nil {
			return e, fmt.Errorf("error scanning monster row: %v", err)
		}
		err = json.Unmarshal(jsonData, &m.Data)
		if err != nil {
			return e, fmt.Errorf("error unmarshaling monster data: %v", err)
		}
		m.Data.System.Attributes.Hp.Value = currentHp // Set the current HP
		e.Monsters = append(e.Monsters, &m)
	}

	if err = rows.Err(); err != nil {
		return e, fmt.Errorf("error iterating monster rows: %v", err)
	}

	// Query for associated players
	playerRows, err := db.Query(`
    SELECT
        p.id,
        p.name,
        p.level,
        p.hp,
        p.ac,
        p.fort,
        p.ref,
        p.will,
        ep.initiative,
        ep.id as association_id,
        ep.hp as current_hp
    FROM players p
    JOIN encounter_players ep ON p.id = ep.player_id
    WHERE ep.encounter_id = $1
`, encounterId)
	if err != nil {
		return e, fmt.Errorf("error querying players: %v", err)
	}
	defer playerRows.Close()

	for playerRows.Next() {
		var player Player
		var currentHp int
		err := playerRows.Scan(
			&player.ID,
			&player.Name,
			&player.Level,
			&player.Hp,
			&player.Ac,
			&player.Fort,
			&player.Ref,
			&player.Will,
			&player.Initiative,
			&player.AssociationID,
			&currentHp,
		)
		if err != nil {
			return e, fmt.Errorf("error scanning player row: %v", err)
		}
		player.Hp = currentHp // Set the current HP
		e.Players = append(e.Players, &player)
	}

	if err = playerRows.Err(); err != nil {
		return e, fmt.Errorf("error iterating player rows: %v", err)
	}

	return e, nil
}

func GetEncounterWithCombatants(db database.Service, encounterId int) (Encounter, error) {
	encounter, err := GetEncounter(db, encounterId)
	if err != nil {
		return Encounter{}, fmt.Errorf("error fetching encounter: %w", err)
	}

	// Get party's players and encounter's monsters
	players := encounter.Players
	monsters := encounter.Monsters

	combatants := make([]Combatant, 0, len(players)+len(monsters))

	// Add players to combatants
	for i := range players {
		combatants = append(combatants, players[i])
	}

	// Add monsters to combatants
	for _, monster := range monsters {
		combatants = append(combatants, monster)
	}

	// Add combatants to the encounter
	encounter.Combatants = combatants

	// Fetch conditions for each combatant
	for i := range encounter.Combatants {
		isMonster := encounter.Combatants[i].IsMonster()
		conditions, err := GetCombatantConditions(db, encounterId, encounter.Combatants[i].GetAssociationID(), isMonster)
		if err != nil {
			return Encounter{}, fmt.Errorf("error fetching conditions for combatant: %w", err)
		}
		encounter.Combatants[i].SetConditions(conditions)
	}

	return encounter, nil
}

func GetAllEncounters(db database.Service) ([]Encounter, error) {
	if db == nil {
		return nil, errors.New("database service is nil")
	}

	rows, err := db.Query(`
        SELECT
            e.id,
            e.name,
            e.user_id,
            e.party_id,
            u.name AS user_name,
            p.name AS party_name
        FROM encounters e
        JOIN users u ON e.user_id = u.id
        JOIN parties p ON e.party_id = p.id
        WHERE e.user_id = $1
        ORDER BY e.id
    `, 1) // hard-coded User-ID for now
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var encounters []Encounter
	for rows.Next() {
		var e Encounter
		e.User = &User{}
		e.Party = &Party{}

		err := rows.Scan(
			&e.ID,
			&e.Name,
			&e.UserID,
			&e.PartyID,
			&e.User.Name,
			&e.Party.Name,
		)

		if err != nil {
			return nil, err
		}
		encounters = append(encounters, e)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return encounters, nil
}

func AddMonsterToEncounter(db database.Service, encounterId int, monsterID int, levelAdjustment int, initiative int) (Encounter, error) {
	// Use a transaction to ensure atomicity
	tx, err := db.Begin()
	if err != nil {
		return Encounter{}, fmt.Errorf("Error starting transaction: %v", err)
	}
	//nolint:errcheck
	defer tx.Rollback() // Rollback the transaction if it hasn't been committed

	// Get the monster's initial HP and name from the data JSON field
	var monsterData []byte
	err = db.QueryRow(`
        SELECT data FROM monsters WHERE id = $1
    `, monsterID).Scan(&monsterData)
	if err != nil {
		return Encounter{}, fmt.Errorf("Failed to get monster's data: %v", err)
	}

	var monster Monster
	err = json.Unmarshal(monsterData, &monster.Data)
	if err != nil {
		return Encounter{}, fmt.Errorf("Failed to unmarshal monster data: %v", err)
	}

	monsterHP := monster.Data.System.Attributes.Hp.Value
	monsterName := monster.Data.Name

	// Find the highest enumeration value for monsters of the same type in this encounter
	var maxEnumeration int
	err = db.QueryRow(`
		SELECT COALESCE(MAX(em.enumeration), 0)
		FROM encounter_monsters em
		JOIN monsters m ON em.monster_id = m.id
		WHERE em.encounter_id = $1
		AND m.data->>'name' = $2
	`, encounterId, monsterName).Scan(&maxEnumeration)
	if err != nil {
		return Encounter{}, fmt.Errorf("Failed to get max enumeration: %v", err)
	}

	// Set the new enumeration to be one more than the highest existing enumeration
	newEnumeration := maxEnumeration + 1

	_, err = db.Exec(`
        INSERT INTO encounter_monsters (encounter_id, monster_id, level_adjustment, initiative, hp, enumeration)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, encounterId, monsterID, levelAdjustment, initiative, monsterHP, newEnumeration)

	if err != nil {
		return Encounter{}, fmt.Errorf("Failed to add monster to encounter: %v", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return Encounter{}, fmt.Errorf("Error committing transaction: %v", err)
	}

	encounter, _ := GetEncounter(db, encounterId)

	return encounter, nil
}

func RemoveMonsterFromEncounter(db database.Service, encounterId int, associationID int) error {
	// Use a transaction to ensure atomicity
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	//nolint:errcheck
	defer tx.Rollback() // Rollback the transaction if it hasn't been committed

	_, err = tx.Exec(`
        DELETE FROM encounter_monsters
        WHERE id = $1
    `, associationID)

	if err != nil {
		return fmt.Errorf("error removing monster from encounter: %v", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}

func RemovePlayerFromEncounter(db database.Service, encounterId int, associationID int) error {
	// Use a transaction to ensure atomicity
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	//nolint:errcheck
	defer tx.Rollback() // Rollback the transaction if it hasn't been committed

	_, err = tx.Exec(`
        DELETE FROM encounter_players
        WHERE id = $1
    `, associationID)

	if err != nil {
		return fmt.Errorf("error removing player from encounter: %v", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}

func (e Encounter) GetPartyName() string {
	return e.Party.Name
}

func (e Encounter) GetDifficulty() int {
	levels := 0
	for _, player := range e.Players {
		levels += player.Level
	}

	partyLevel := int(float64(levels) / float64(len(e.Players)))
	monsterXpPool := 0.0

	for _, combatant := range e.Combatants {
		if combatant.GetType() == "monster" {
			monsterLevel := combatant.GetLevel()
			difference := monsterLevel - partyLevel

			switch {
			case difference <= -4:
				monsterXpPool += 10.0
			case difference == -3:
				monsterXpPool += 15.0
			case difference == -2:
				monsterXpPool += 20.0
			case difference == -1:
				monsterXpPool += 30.0
			case difference == 0:
				monsterXpPool += 40.0
			case difference == 1:
				monsterXpPool += 60.0
			case difference == 2:
				monsterXpPool += 80.0
			case difference == 3:
				monsterXpPool += 120.0
			case difference >= 4:
				monsterXpPool += 160.0
			}
		}
	}

	threshold := float64(len(e.Players)) / 4.0

	switch {
	case monsterXpPool <= 40.0*threshold:
		return 0
	case monsterXpPool <= 60.0*threshold:
		return 1
	case monsterXpPool <= 80.0*threshold:
		return 2
	case monsterXpPool <= 120.0*threshold:
		return 3
	case monsterXpPool > 120.0*threshold:
		return 4
	}

	return 0
}

func GetCombatantConditions(db database.Service, encounterID int, associationID int, isMonster bool) ([]Condition, error) {
	var query string
	if isMonster {
		query = `
            SELECT c.id, c.data, cc.condition_value
            FROM combatant_conditions cc
            JOIN conditions c ON cc.condition_id = c.id
            WHERE cc.encounter_id = $1 AND cc.encounter_monster_id = $2
        `
	} else {
		query = `
            SELECT c.id, c.data, cc.condition_value
            FROM combatant_conditions cc
            JOIN conditions c ON cc.condition_id = c.id
            WHERE cc.encounter_id = $1 AND cc.encounter_player_id = $2
        `
	}

	rows, err := db.Query(query, encounterID, associationID)
	if err != nil {
		return nil, fmt.Errorf("error querying combatant conditions: %v", err)
	}
	defer rows.Close()

	var conditions []Condition
	for rows.Next() {
		var c Condition
		var jsonData []byte
		var conditionValue int
		err := rows.Scan(&c.ID, &jsonData, &conditionValue)
		if err != nil {
			return nil, fmt.Errorf("error scanning condition row: %v", err)
		}
		err = json.Unmarshal(jsonData, &c.Data)
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling condition data: %v", err)
		}
		c.Data.System.Value.Value = conditionValue
		conditions = append(conditions, c)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating condition rows: %v", err)
	}

	return conditions, nil
}
