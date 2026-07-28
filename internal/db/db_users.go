package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/mrpoundsign/cy_borger/pkg/chargen"
)

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Handle       string    `json:"handle"`
	PasswordHash string    `json:"-"`
	Salt         string    `json:"-"`
	UpdatedAt    time.Time `json:"updated_at"`
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

func (d *DB) BanUser(ownerID, bannedUserID string) error {
	query := `INSERT INTO user_bans (owner_id, banned_user_id) VALUES (?, ?) ON CONFLICT DO NOTHING`
	_, err := d.conn.Exec(query, ownerID, bannedUserID)
	return err
}

func (d *DB) IsUserBanned(ownerID, userID string) (bool, error) {
	query := `SELECT 1 FROM user_bans WHERE owner_id = ? AND banned_user_id = ?`
	row := d.conn.QueryRow(query, ownerID, userID)
	var dummy int
	err := row.Scan(&dummy)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
