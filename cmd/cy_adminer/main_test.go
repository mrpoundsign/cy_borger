package main

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/mrpoundsign/cy_borger/internal/db"
	_ "modernc.org/sqlite"
)

func TestSetPassword(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_adminer.db")

	database, err := db.InitDB(dbPath)
	if err != nil {
		t.Fatalf("failed to initialize test db: %v", err)
	}

	_, err = database.CreateUserAccount("mrp", "oldpass")
	if err != nil {
		t.Fatalf("CreateUserAccount error: %v", err)
	}

	// Verify old password works initially
	authedUser, err := database.AuthenticateUser("mrp", "oldpass")
	if err != nil || authedUser == nil {
		t.Fatalf("AuthenticateUser failed with old pass: %v", err)
	}

	// Open raw sql connection to test setPassword
	sqlConn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sql db: %v", err)
	}
	defer sqlConn.Close()

	if err := setPassword(sqlConn, "mrp", "newpass"); err != nil {
		t.Fatalf("setPassword failed: %v", err)
	}

	// Verify new password authenticates successfully
	authedUserNew, err := database.AuthenticateUser("mrp", "newpass")
	if err != nil || authedUserNew == nil {
		t.Fatalf("AuthenticateUser failed with new pass: %v", err)
	}

	// Verify old password no longer works
	_, err = database.AuthenticateUser("mrp", "oldpass")
	if err == nil {
		t.Fatalf("expected authentication to fail with old pass, but it succeeded")
	}
}
