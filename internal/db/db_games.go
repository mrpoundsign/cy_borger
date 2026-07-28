package db

import (
	"database/sql"
	"log"
	"time"

	"github.com/mrpoundsign/cy_borger/pkg/chargen"
)

type Game struct {
	ID         string    `json:"id"`
	GMCode     string    `json:"gm_code"`
	InviteCode string    `json:"invite_code"`
	Name       string    `json:"name"`
	OwnerID    string    `json:"owner_id"`
	SystemID   string    `json:"system_id"`
	IsLocked   bool      `json:"is_locked"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (d *DB) CreateGame(name string, ownerID string) (*Game, error) {
	id := chargen.GenerateRandomID(6)
	gmCode := chargen.GenerateRandomID(6)
	inviteCode := chargen.GenerateRandomID(4)

	query := `INSERT INTO games (id, gm_code, invite_code, name, owner_id, system_id, is_locked, created_at, updated_at) VALUES (?, ?, ?, ?, ?, '0', 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	_, err := d.conn.Exec(query, id, gmCode, inviteCode, name, ownerID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &Game{
		ID:         id,
		GMCode:     gmCode,
		InviteCode: inviteCode,
		Name:       name,
		OwnerID:    ownerID,
		SystemID:   "0",
		IsLocked:   false,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func (d *DB) GetGame(id string) (*Game, error) {
	query := `SELECT id, gm_code, invite_code, name, owner_id, system_id, is_locked, created_at, updated_at FROM games WHERE id = ?`
	row := d.conn.QueryRow(query, id)

	var g Game
	var isLockedInt int
	var createdAt sql.NullTime
	if err := row.Scan(&g.ID, &g.GMCode, &g.InviteCode, &g.Name, &g.OwnerID, &g.SystemID, &isLockedInt, &createdAt, &g.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	g.IsLocked = isLockedInt == 1
	if createdAt.Valid {
		g.CreatedAt = createdAt.Time
	} else {
		g.CreatedAt = g.UpdatedAt
	}
	return &g, nil
}

func (d *DB) GetGameByInviteCode(code string) (*Game, error) {
	query := `SELECT id, gm_code, invite_code, name, owner_id, system_id, is_locked, created_at, updated_at FROM games WHERE invite_code = ?`
	row := d.conn.QueryRow(query, code)

	var g Game
	var isLockedInt int
	var createdAt sql.NullTime
	if err := row.Scan(&g.ID, &g.GMCode, &g.InviteCode, &g.Name, &g.OwnerID, &g.SystemID, &isLockedInt, &createdAt, &g.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	g.IsLocked = isLockedInt == 1
	if createdAt.Valid {
		g.CreatedAt = createdAt.Time
	} else {
		g.CreatedAt = g.UpdatedAt
	}
	return &g, nil
}

func (d *DB) GetGamesForUser(userID string) ([]Game, error) {
	query := `
	SELECT g.id, g.gm_code, g.invite_code, g.name, g.owner_id, g.system_id, g.is_locked, g.created_at, g.updated_at
	FROM games g
	JOIN characters c ON c.game_id = g.id
	WHERE c.owner_id = ?
	GROUP BY g.id
	ORDER BY g.updated_at DESC
	`
	rows, err := d.conn.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []Game
	for rows.Next() {
		var g Game
		var isLockedInt int
		var createdAt sql.NullTime
		if err := rows.Scan(&g.ID, &g.GMCode, &g.InviteCode, &g.Name, &g.OwnerID, &g.SystemID, &isLockedInt, &createdAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		g.IsLocked = isLockedInt == 1
		if createdAt.Valid {
			g.CreatedAt = createdAt.Time
		} else {
			g.CreatedAt = g.UpdatedAt
		}
		games = append(games, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return games, nil
}

func (d *DB) GetGamesByOwner(ownerID string) ([]Game, error) {
	query := `SELECT id, gm_code, invite_code, name, owner_id, system_id, is_locked, created_at, updated_at FROM games WHERE owner_id = ?`
	rows, err := d.conn.Query(query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []Game
	for rows.Next() {
		var g Game
		var isLockedInt int
		var createdAt sql.NullTime
		if err := rows.Scan(&g.ID, &g.GMCode, &g.InviteCode, &g.Name, &g.OwnerID, &g.SystemID, &isLockedInt, &createdAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		g.IsLocked = isLockedInt == 1
		if createdAt.Valid {
			g.CreatedAt = createdAt.Time
		} else {
			g.CreatedAt = g.UpdatedAt
		}
		games = append(games, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return games, nil
}

func (d *DB) AddGameGM(gameID, userID string) error {
	query := `INSERT INTO game_gms (game_id, user_id) VALUES (?, ?) ON CONFLICT DO NOTHING`
	_, err := d.conn.Exec(query, gameID, userID)
	return err
}

func (d *DB) RemoveGameGM(gameID, userID string) error {
	query := `DELETE FROM game_gms WHERE game_id = ? AND user_id = ?`
	_, err := d.conn.Exec(query, gameID, userID)
	return err
}

func (d *DB) GetGameGMs(gameID string) ([]User, error) {
	query := `
	SELECT u.id, u.username, u.handle, u.updated_at
	FROM users u
	JOIN game_gms gm ON gm.user_id = u.id
	WHERE gm.game_id = ?
	`
	rows, err := d.conn.Query(query, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gms []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Handle, &u.UpdatedAt); err != nil {
			return nil, err
		}
		gms = append(gms, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return gms, nil
}

func (d *DB) RenameGame(gameID, newName string) error {
	query := `UPDATE games SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := d.conn.Exec(query, newName, gameID)
	return err
}

func (d *DB) ToggleGameLock(gameID string, isLocked bool) error {
	lockedInt := 0
	if isLocked {
		lockedInt = 1
	}
	query := `UPDATE games SET is_locked = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := d.conn.Exec(query, lockedInt, gameID)
	return err
}

func (d *DB) DeleteGame(id string) error {
	tx, err := d.conn.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("Failed to rollback transaction in DeleteGame: %v", err)
		}
	}()

	if _, err := tx.Exec(`DELETE FROM game_gms WHERE game_id = ?`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM game_logs WHERE game_id = ?`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE characters SET game_id = '' WHERE game_id = ?`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM games WHERE id = ?`, id); err != nil {
		return err
	}

	return tx.Commit()
}
