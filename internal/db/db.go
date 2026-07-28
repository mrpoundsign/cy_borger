package db

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

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
	conn, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
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
