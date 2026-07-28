package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/mrpoundsign/cy_borger/pkg/chargen"
)

func (d *DB) SaveCharacter(c *chargen.Character, ownerID string) error {
	data, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal character: %w", err)
	}

	isSavedInt := 0
	if c.IsSaved {
		isSavedInt = 1
	}
	isDeadInt := 0
	if c.IsDead {
		isDeadInt = 1
	}

	var diedAtVal interface{}
	if !c.DiedAt.IsZero() {
		diedAtVal = c.DiedAt
	}

	query := `
	INSERT INTO characters (id, edit_code, game_id, owner_id, system_id, is_saved, is_dead, death_note, died_at, data_json, updated_at)
	VALUES (?, ?, ?, ?, COALESCE(NULLIF(?, ''), '0'), ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(id) DO UPDATE SET
		game_id = excluded.game_id,
		system_id = excluded.system_id,
		is_saved = excluded.is_saved,
		is_dead = excluded.is_dead,
		death_note = excluded.death_note,
		died_at = excluded.died_at,
		data_json = excluded.data_json,
		updated_at = CURRENT_TIMESTAMP;
	`
	_, err = d.conn.Exec(query, c.ID, c.EditCode, c.GameID, ownerID, c.SystemID, isSavedInt, isDeadInt, c.DeathNote, diedAtVal, string(data))
	return err
}

func (d *DB) GetCharacter(id string) (*chargen.Character, error) {
	query := `
	SELECT c.owner_id, c.game_id, c.system_id, c.is_saved, c.is_dead, c.death_note, c.died_at, c.data_json, c.updated_at, COALESCE(NULLIF(u.handle, ''), u.username, '') AS owner_username
	FROM characters c
	LEFT JOIN users u ON c.owner_id = u.id
	WHERE c.id = ?
	`
	row := d.conn.QueryRow(query, id)

	var ownerID, gameID, systemID, deathNote, dataStr, ownerUsername string
	var isSavedInt, isDeadInt int
	var diedAt sql.NullTime
	var updatedAt time.Time
	if err := row.Scan(&ownerID, &gameID, &systemID, &isSavedInt, &isDeadInt, &deathNote, &diedAt, &dataStr, &updatedAt, &ownerUsername); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	var c chargen.Character
	if err := json.Unmarshal([]byte(dataStr), &c); err != nil {
		return nil, err
	}
	c.OwnerID = ownerID
	c.OwnerUsername = ownerUsername
	c.GameID = gameID
	c.SystemID = systemID
	c.IsSaved = isSavedInt == 1 || c.IsSaved
	c.IsDead = isDeadInt == 1 || c.IsDead
	if deathNote != "" {
		c.DeathNote = deathNote
	}
	if diedAt.Valid {
		c.DiedAt = diedAt.Time
	}
	c.UpdatedAt = updatedAt

	return &c, nil
}

func (d *DB) GetCharactersForGame(gameID string) ([]chargen.Character, error) {
	query := `
	SELECT c.owner_id, c.game_id, c.system_id, c.is_saved, c.is_dead, c.death_note, c.died_at, c.data_json, c.updated_at, COALESCE(NULLIF(u.handle, ''), u.username, '') AS owner_username
	FROM characters c
	LEFT JOIN users u ON c.owner_id = u.id
	WHERE c.game_id = ? AND c.is_saved = 1
	ORDER BY c.is_dead ASC, c.died_at DESC, c.updated_at DESC
	`
	rows, err := d.conn.Query(query, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chars []chargen.Character
	for rows.Next() {
		var ownerID, gameID, systemID, deathNote, dataStr, ownerUsername string
		var isSavedInt, isDeadInt int
		var diedAt sql.NullTime
		var updatedAt time.Time
		if err := rows.Scan(&ownerID, &gameID, &systemID, &isSavedInt, &isDeadInt, &deathNote, &diedAt, &dataStr, &updatedAt, &ownerUsername); err != nil {
			log.Printf("GetCharactersForGame Scan error: %v", err)
			return nil, err
		}
		var c chargen.Character
		if err := json.Unmarshal([]byte(dataStr), &c); err == nil {
			c.OwnerID = ownerID
			c.OwnerUsername = ownerUsername
			c.GameID = gameID
			c.SystemID = systemID
			c.IsSaved = isSavedInt == 1 || c.IsSaved
			c.IsDead = isDeadInt == 1 || c.IsDead
			if deathNote != "" {
				c.DeathNote = deathNote
			}
			if diedAt.Valid {
				c.DiedAt = diedAt.Time
			}
			c.UpdatedAt = updatedAt
			chars = append(chars, c)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return chars, nil
}

func (d *DB) GetCharactersByOwner(ownerID string) ([]chargen.Character, error) {
	query := `
	SELECT c.owner_id, c.game_id, c.system_id, c.is_saved, c.is_dead, c.death_note, c.died_at, c.data_json, c.updated_at, COALESCE(u.username, '') AS owner_username, COALESCE(g.name, '') AS game_name
	FROM characters c
	LEFT JOIN users u ON c.owner_id = u.id
	LEFT JOIN games g ON c.game_id = g.id
	WHERE c.owner_id = ?
	ORDER BY c.is_dead ASC, c.died_at DESC, c.updated_at DESC
	`
	rows, err := d.conn.Query(query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chars []chargen.Character
	for rows.Next() {
		var ownerIDVal, gameID, systemID, deathNote, dataStr, ownerUsername, gameName string
		var isSavedInt, isDeadInt int
		var diedAt sql.NullTime
		var updatedAt time.Time
		if err := rows.Scan(&ownerIDVal, &gameID, &systemID, &isSavedInt, &isDeadInt, &deathNote, &diedAt, &dataStr, &updatedAt, &ownerUsername, &gameName); err != nil {
			return nil, err
		}
		var c chargen.Character
		if err := json.Unmarshal([]byte(dataStr), &c); err == nil {
			c.OwnerID = ownerIDVal
			c.OwnerUsername = ownerUsername
			c.GameID = gameID
			c.SystemID = systemID
			c.GameName = gameName
			c.IsSaved = isSavedInt == 1 || c.IsSaved
			c.IsDead = isDeadInt == 1 || c.IsDead
			if deathNote != "" {
				c.DeathNote = deathNote
			}
			if diedAt.Valid {
				c.DiedAt = diedAt.Time
			}
			c.UpdatedAt = updatedAt
			chars = append(chars, c)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return chars, nil
}

func (d *DB) DeleteCharacter(id string) error {
	query := `DELETE FROM characters WHERE id = ?`
	_, err := d.conn.Exec(query, id)
	return err
}

func (d *DB) KickUserFromGame(gameID, userID string) error {
	query := `UPDATE characters SET game_id = '', updated_at = CURRENT_TIMESTAMP WHERE game_id = ? AND owner_id = ?`
	_, err := d.conn.Exec(query, gameID, userID)
	return err
}

func (d *DB) KickUserFromAllOwnerGames(ownerID, userID string) error {
	query := `
	UPDATE characters
	SET game_id = '', updated_at = CURRENT_TIMESTAMP
	WHERE owner_id = ? AND game_id IN (
		SELECT id FROM games WHERE owner_id = ?
	)
	`
	_, err := d.conn.Exec(query, userID, ownerID)
	return err
}
