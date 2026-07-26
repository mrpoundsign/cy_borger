package db

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/mrpoundsign/cy_borger/chargen"
	_ "modernc.org/sqlite"
)

type Game struct {
	ID         string `json:"id"`
	GMCode     string `json:"gm_code"`
	InviteCode string `json:"invite_code"`
	Name       string `json:"name"`
}

type DB struct {
	conn *sql.DB
}

func InitDB(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	createTablesSQL := `
	CREATE TABLE IF NOT EXISTS characters (
		id TEXT PRIMARY KEY,
		edit_code TEXT NOT NULL,
		game_id TEXT DEFAULT '',
		data_json TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS games (
		id TEXT PRIMARY KEY,
		gm_code TEXT NOT NULL,
		invite_code TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL
	);
	`

	_, err = conn.Exec(createTablesSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &DB{conn: conn}, nil
}

func (d *DB) SaveCharacter(c *chargen.Character) error {
	data, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal character: %w", err)
	}

	query := `
	INSERT INTO characters (id, edit_code, game_id, data_json)
	VALUES (?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		game_id = excluded.game_id,
		data_json = excluded.data_json;
	`
	_, err = d.conn.Exec(query, c.ID, c.EditCode, c.GameID, string(data))
	return err
}

func (d *DB) GetCharacter(id string) (*chargen.Character, error) {
	query := `SELECT data_json FROM characters WHERE id = ?`
	row := d.conn.QueryRow(query, id)

	var dataStr string
	if err := row.Scan(&dataStr); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	var c chargen.Character
	if err := json.Unmarshal([]byte(dataStr), &c); err != nil {
		return nil, err
	}

	return &c, nil
}

func (d *DB) CreateGame(name string) (*Game, error) {
	id := chargen.GenerateRandomID(6)
	gmCode := chargen.GenerateRandomID(6)
	inviteCode := chargen.GenerateRandomID(4)

	query := `INSERT INTO games (id, gm_code, invite_code, name) VALUES (?, ?, ?, ?)`
	_, err := d.conn.Exec(query, id, gmCode, inviteCode, name)
	if err != nil {
		return nil, err
	}

	return &Game{
		ID:         id,
		GMCode:     gmCode,
		InviteCode: inviteCode,
		Name:       name,
	}, nil
}

func (d *DB) GetGame(id string) (*Game, error) {
	query := `SELECT id, gm_code, invite_code, name FROM games WHERE id = ?`
	row := d.conn.QueryRow(query, id)

	var g Game
	if err := row.Scan(&g.ID, &g.GMCode, &g.InviteCode, &g.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &g, nil
}

func (d *DB) GetGameByInviteCode(code string) (*Game, error) {
	query := `SELECT id, gm_code, invite_code, name FROM games WHERE invite_code = ?`
	row := d.conn.QueryRow(query, code)

	var g Game
	if err := row.Scan(&g.ID, &g.GMCode, &g.InviteCode, &g.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &g, nil
}

func (d *DB) GetCharactersForGame(gameID string) ([]chargen.Character, error) {
	query := `SELECT data_json FROM characters WHERE game_id = ?`
	rows, err := d.conn.Query(query, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chars []chargen.Character
	for rows.Next() {
		var dataStr string
		if err := rows.Scan(&dataStr); err != nil {
			return nil, err
		}
		var c chargen.Character
		if err := json.Unmarshal([]byte(dataStr), &c); err == nil {
			chars = append(chars, c)
		}
	}
	return chars, nil
}
