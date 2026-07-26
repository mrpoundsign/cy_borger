package db

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/hex"
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
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Handle       string    `json:"handle"`
	PasswordHash string    `json:"-"`
	Salt         string    `json:"-"`
	UpdatedAt    time.Time `json:"updated_at"`
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

func GenerateSalt() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func HashPassword(password, salt string) string {
	hash := sha256.Sum256([]byte(salt + ":" + password))
	return hex.EncodeToString(hash[:])
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
	INSERT INTO users (id, username, handle, password_hash, salt, updated_at)
	VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(id) DO UPDATE SET
		username = excluded.username,
		handle = excluded.handle,
		updated_at = CURRENT_TIMESTAMP;
	`
	_, err := d.conn.Exec(query, u.ID, u.Username, u.Handle, u.PasswordHash, u.Salt)
	return err
}

func (d *DB) CreateUserAccount(username, password string) (*User, error) {
	existing, err := d.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("Username already registered")
	}

	salt, err := GenerateSalt()
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := HashPassword(password, salt)
	userID := "usr_" + chargen.GenerateRandomID(6)

	num := (time.Now().UnixNano() % 9000) + 1000
	handle := fmt.Sprintf("%s#%04d", username, num)

	query := `INSERT INTO users (id, username, handle, password_hash, salt, updated_at) VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`
	_, err = d.conn.Exec(query, userID, username, handle, hash, salt)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:           userID,
		Username:     username,
		Handle:       handle,
		PasswordHash: hash,
		Salt:         salt,
		UpdatedAt:    time.Now(),
	}, nil
}

func (d *DB) AuthenticateUser(username, password string) (*User, error) {
	u, err := d.GetUserByUsername(username)
	if err != nil || u == nil {
		return nil, fmt.Errorf("Invalid username or password")
	}

	expectedHash := HashPassword(password, u.Salt)
	if u.PasswordHash != expectedHash {
		return nil, fmt.Errorf("Invalid username or password")
	}

	return u, nil
}

func (d *DB) GetUserByUsername(username string) (*User, error) {
	query := `SELECT id, username, handle, password_hash, salt, updated_at FROM users WHERE LOWER(username) = LOWER(?)`
	row := d.conn.QueryRow(query, username)

	var u User
	if err := row.Scan(&u.ID, &u.Username, &u.Handle, &u.PasswordHash, &u.Salt, &u.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (d *DB) GetUser(id string) (*User, error) {
	query := `SELECT id, username, handle, password_hash, salt, updated_at FROM users WHERE id = ?`
	row := d.conn.QueryRow(query, id)

	var u User
	if err := row.Scan(&u.ID, &u.Username, &u.Handle, &u.PasswordHash, &u.Salt, &u.UpdatedAt); err != nil {
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
	isDeadInt := 0
	if c.IsDead {
		isDeadInt = 1
	}

	var diedAtVal interface{}
	if !c.DiedAt.IsZero() {
		diedAtVal = c.DiedAt
	}

	query := `
	INSERT INTO characters (id, edit_code, game_id, owner_id, is_saved, is_dead, death_note, died_at, data_json, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(id) DO UPDATE SET
		game_id = excluded.game_id,
		is_saved = excluded.is_saved,
		is_dead = excluded.is_dead,
		death_note = excluded.death_note,
		died_at = excluded.died_at,
		data_json = excluded.data_json,
		updated_at = CURRENT_TIMESTAMP;
	`
	_, err = d.conn.Exec(query, c.ID, c.EditCode, c.GameID, ownerID, isSavedInt, isDeadInt, c.DeathNote, diedAtVal, string(data))
	return err
}

func (d *DB) GetCharacter(id string) (*chargen.Character, error) {
	query := `
	SELECT c.owner_id, c.game_id, c.is_saved, c.is_dead, c.death_note, c.died_at, c.data_json, c.updated_at, COALESCE(NULLIF(u.handle, ''), u.username, '') AS owner_username
	FROM characters c
	LEFT JOIN users u ON c.owner_id = u.id
	WHERE c.id = ?
	`
	row := d.conn.QueryRow(query, id)

	var ownerID, gameID, deathNote, dataStr, ownerUsername string
	var isSavedInt, isDeadInt int
	var diedAt sql.NullTime
	var updatedAt time.Time
	if err := row.Scan(&ownerID, &gameID, &isSavedInt, &isDeadInt, &deathNote, &diedAt, &dataStr, &updatedAt, &ownerUsername); err != nil {
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
	query := `
	SELECT c.owner_id, c.game_id, c.is_saved, c.is_dead, c.death_note, c.died_at, c.data_json, c.updated_at, COALESCE(NULLIF(u.handle, ''), u.username, '') AS owner_username
	FROM characters c
	LEFT JOIN users u ON c.owner_id = u.id
	WHERE c.game_id = ?
	ORDER BY c.is_dead ASC, c.died_at DESC, c.updated_at DESC
	`
	rows, err := d.conn.Query(query, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chars []chargen.Character
	for rows.Next() {
		var ownerID, gameID, deathNote, dataStr, ownerUsername string
		var isSavedInt, isDeadInt int
		var diedAt sql.NullTime
		var updatedAt time.Time
		if err := rows.Scan(&ownerID, &gameID, &isSavedInt, &isDeadInt, &deathNote, &diedAt, &dataStr, &updatedAt, &ownerUsername); err != nil {
			return nil, err
		}
		var c chargen.Character
		if err := json.Unmarshal([]byte(dataStr), &c); err == nil {
			c.OwnerID = ownerID
			c.OwnerUsername = ownerUsername
			c.GameID = gameID
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
	query := `
	SELECT c.owner_id, c.game_id, c.is_saved, c.is_dead, c.death_note, c.died_at, c.data_json, c.updated_at, COALESCE(u.username, '') AS owner_username
	FROM characters c
	LEFT JOIN users u ON c.owner_id = u.id
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
		var ownerIDVal, gameID, deathNote, dataStr, ownerUsername string
		var isSavedInt, isDeadInt int
		var diedAt sql.NullTime
		var updatedAt time.Time
		if err := rows.Scan(&ownerIDVal, &gameID, &isSavedInt, &isDeadInt, &deathNote, &diedAt, &dataStr, &updatedAt, &ownerUsername); err != nil {
			return nil, err
		}
		var c chargen.Character
		if err := json.Unmarshal([]byte(dataStr), &c); err == nil {
			c.OwnerID = ownerIDVal
			c.OwnerUsername = ownerUsername
			c.GameID = gameID
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
