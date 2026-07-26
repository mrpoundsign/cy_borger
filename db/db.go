package db

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/mrpoundsign/cy_borger/chargen"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Game struct {
	ID         string    `json:"id"`
	GMCode     string    `json:"gm_code"`
	InviteCode string    `json:"invite_code"`
	Name       string    `json:"name"`
	OwnerID    string    `json:"owner_id"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type DB struct {
	conn *sql.DB
}

func InitDB(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Run migrations using golang-migrate
	if err := runMigrations(conn); err != nil {
		return nil, fmt.Errorf("migration failure: %w", err)
	}

	return &DB{conn: conn}, nil
}

func runMigrations(conn *sql.DB) error {
	driver, err := sqlite.WithInstance(conn, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	subFS, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to get migrations subfs: %w", err)
	}

	sourceDriver, err := iofs.New(subFS, ".")
	if err != nil {
		return fmt.Errorf("failed to create iofs source driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "sqlite", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func (d *DB) SaveUser(u *User) error {
	existing, err := d.GetUser(u.ID)
	if err != nil {
		return err
	}
	if existing != nil {
		return fmt.Errorf("user handle has already been customized and cannot be edited again")
	}

	query := `INSERT INTO users (id, username, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)`
	_, err = d.conn.Exec(query, u.ID, u.Username)
	return err
}

func (d *DB) GetUser(id string) (*User, error) {
	query := `SELECT id, username, updated_at FROM users WHERE id = ?`
	row := d.conn.QueryRow(query, id)

	var u User
	if err := row.Scan(&u.ID, &u.Username, &u.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (d *DB) SaveCharacter(c *chargen.Character, ownerID string) error {
	data, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal character: %w", err)
	}

	isSavedInt := 0
	if c.IsSaved {
		isSavedInt = 1
	}

	query := `
	INSERT INTO characters (id, edit_code, game_id, owner_id, is_saved, data_json, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(id) DO UPDATE SET
		game_id = excluded.game_id,
		is_saved = excluded.is_saved,
		data_json = excluded.data_json,
		updated_at = CURRENT_TIMESTAMP;
	`
	_, err = d.conn.Exec(query, c.ID, c.EditCode, c.GameID, ownerID, isSavedInt, string(data))
	return err
}

func (d *DB) GetCharacter(id string) (*chargen.Character, error) {
	query := `SELECT game_id, is_saved, data_json, updated_at FROM characters WHERE id = ?`
	row := d.conn.QueryRow(query, id)

	var gameID, dataStr string
	var isSavedInt int
	var updatedAt time.Time
	if err := row.Scan(&gameID, &isSavedInt, &dataStr, &updatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	var c chargen.Character
	if err := json.Unmarshal([]byte(dataStr), &c); err != nil {
		return nil, err
	}
	c.GameID = gameID
	c.IsSaved = isSavedInt == 1 || c.IsSaved
	c.UpdatedAt = updatedAt

	return &c, nil
}

func (d *DB) CreateGame(name string, ownerID string) (*Game, error) {
	id := chargen.GenerateRandomID(6)
	gmCode := chargen.GenerateRandomID(6)
	inviteCode := chargen.GenerateRandomID(4)

	query := `INSERT INTO games (id, gm_code, invite_code, name, owner_id, updated_at) VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`
	_, err := d.conn.Exec(query, id, gmCode, inviteCode, name, ownerID)
	if err != nil {
		return nil, err
	}

	return &Game{
		ID:         id,
		GMCode:     gmCode,
		InviteCode: inviteCode,
		Name:       name,
		OwnerID:    ownerID,
		UpdatedAt:  time.Now(),
	}, nil
}

func (d *DB) GetGame(id string) (*Game, error) {
	query := `SELECT id, gm_code, invite_code, name, owner_id, updated_at FROM games WHERE id = ?`
	row := d.conn.QueryRow(query, id)

	var g Game
	if err := row.Scan(&g.ID, &g.GMCode, &g.InviteCode, &g.Name, &g.OwnerID, &g.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &g, nil
}

func (d *DB) GetGameByInviteCode(code string) (*Game, error) {
	query := `SELECT id, gm_code, invite_code, name, owner_id, updated_at FROM games WHERE invite_code = ?`
	row := d.conn.QueryRow(query, code)

	var g Game
	if err := row.Scan(&g.ID, &g.GMCode, &g.InviteCode, &g.Name, &g.OwnerID, &g.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &g, nil
}

func (d *DB) GetCharactersForGame(gameID string) ([]chargen.Character, error) {
	query := `SELECT game_id, is_saved, data_json, updated_at FROM characters WHERE game_id = ?`
	rows, err := d.conn.Query(query, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chars []chargen.Character
	for rows.Next() {
		var gameID, dataStr string
		var isSavedInt int
		var updatedAt time.Time
		if err := rows.Scan(&gameID, &isSavedInt, &dataStr, &updatedAt); err != nil {
			return nil, err
		}
		var c chargen.Character
		if err := json.Unmarshal([]byte(dataStr), &c); err == nil {
			c.GameID = gameID
			c.IsSaved = isSavedInt == 1 || c.IsSaved
			c.UpdatedAt = updatedAt
			chars = append(chars, c)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return chars, nil
}

func (d *DB) GetGamesByOwner(ownerID string) ([]Game, error) {
	query := `SELECT id, gm_code, invite_code, name, owner_id, updated_at FROM games WHERE owner_id = ?`
	rows, err := d.conn.Query(query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []Game
	for rows.Next() {
		var g Game
		if err := rows.Scan(&g.ID, &g.GMCode, &g.InviteCode, &g.Name, &g.OwnerID, &g.UpdatedAt); err != nil {
			return nil, err
		}
		games = append(games, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return games, nil
}

func (d *DB) GetCharactersByOwner(ownerID string) ([]chargen.Character, error) {
	query := `SELECT game_id, is_saved, data_json, updated_at FROM characters WHERE owner_id = ?`
	rows, err := d.conn.Query(query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chars []chargen.Character
	for rows.Next() {
		var gameID, dataStr string
		var isSavedInt int
		var updatedAt time.Time
		if err := rows.Scan(&gameID, &isSavedInt, &dataStr, &updatedAt); err != nil {
			return nil, err
		}
		var c chargen.Character
		if err := json.Unmarshal([]byte(dataStr), &c); err == nil {
			c.GameID = gameID
			c.IsSaved = isSavedInt == 1 || c.IsSaved
			c.UpdatedAt = updatedAt
			chars = append(chars, c)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return chars, nil
}
