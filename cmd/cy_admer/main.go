package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
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
			fmt.Println("Usage: cy_admer set-password <username> <password>")
			return
		}
		username := os.Args[2]
		password := os.Args[3]

		db, err := sql.Open("sqlite", "./cy_borger.db")
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}
		defer db.Close()

		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Failed to hash password: %v", err)
		}

		res, err := db.Exec("UPDATE users SET password_hash = ? WHERE username = ?", string(hash), username)
		if err != nil {
			log.Fatalf("Failed to update password: %v", err)
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			log.Fatalf("Failed to check rows affected: %v", err)
		}

		if rowsAffected == 0 {
			fmt.Printf("User '%s' not found.\n", username)
		} else {
			fmt.Printf("Password updated successfully for user '%s'.\n", username)
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
	}
}

func printUsage() {
	fmt.Println("CY_BORGER Admin Tool (cy_admer)")
	fmt.Println("Commands:")
	fmt.Println("  set-password <username> <password>   Set a user's password")
}
