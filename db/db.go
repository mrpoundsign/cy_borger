package db

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/mrpoundsign/cy_borger/chargen"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

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
	query := `
	INSERT INTO users (id, username)
	VALUES (?, ?)
	ON CONFLICT(id) DO UPDATE SET
		username = excluded.username;
	`
	_, err := d.conn.Exec(query, u.ID, u.Username)
	return err
}

func (d *DB) GetUser(id string) (*User, error) {
	query := `SELECT id, username FROM users WHERE id = ?`
	row := d.conn.QueryRow(query, id)

	var u User
	if err := row.Scan(&u.ID, &u.Username); err != nil {
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

	query := `
	INSERT INTO characters (id, edit_code, game_id, owner_id, data_json)
	VALUES (?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		game_id = excluded.game_id,
		owner_id = excluded.owner_id,
		data_json = excluded.data_json;
	`
	_, err = d.conn.Exec(query, c.ID, c.EditCode, c.GameID, ownerID, string(data))
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
