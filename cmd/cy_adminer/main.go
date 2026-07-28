package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/mrpoundsign/cy_borger/internal/db"
	_ "modernc.org/sqlite"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "set-password":
		if len(os.Args) < 4 {
			fmt.Println("Usage: cy_adminer set-password <username> <password>")
			return
		}
		username := os.Args[2]
		password := os.Args[3]

		dbConn, err := sql.Open("sqlite", "./cy_borger.db")
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}
		defer dbConn.Close()

		if err := setPassword(dbConn, username, password); err != nil {
			log.Fatalf("%v", err)
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
	}
}

func setPassword(dbConn *sql.DB, username, password string) error {
	var count int
	err := dbConn.QueryRow("SELECT COUNT(*) FROM users WHERE LOWER(username) = LOWER(?)", username).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to query user: %w", err)
	}
	if count == 0 {
		fmt.Printf("User '%s' not found.\n", username)
		return nil
	}

	salt, err := db.GenerateSalt()
	if err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := db.HashPassword(password, salt)

	res, err := dbConn.Exec("UPDATE users SET password_hash = ?, salt = ?, updated_at = CURRENT_TIMESTAMP WHERE LOWER(username) = LOWER(?)", hash, salt, username)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		fmt.Printf("User '%s' not found.\n", username)
	} else {
		fmt.Printf("Password updated successfully for user '%s'.\n", username)
	}
	return nil
}

func printUsage() {
	fmt.Println("CY_BORGER Admin Tool (cy_adminer)")
	fmt.Println("Commands:")
	fmt.Println("  set-password <username> <password>   Set a user's password")
}
