package db

import (
	"os"
	"testing"

	"github.com/mrpoundsign/cy_borger/chargen"
)

func TestDB(t *testing.T) {
	dbFile := "./test_cy_borger.db"
	defer os.Remove(dbFile)

	database, err := InitDB(dbFile)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// 1. Test Save & Get Character
	c := chargen.GenerateCharacter()
	if err := database.SaveCharacter(&c); err != nil {
		t.Fatalf("SaveCharacter failed: %v", err)
	}

	retrieved, err := database.GetCharacter(c.ID)
	if err != nil || retrieved == nil {
		t.Fatalf("GetCharacter failed: %v", err)
	}
	if retrieved.Name != c.Name {
		t.Errorf("Expected name %s, got %s", c.Name, retrieved.Name)
	}

	// 2. Test Create & Get Game
	g, err := database.CreateGame("Test Campaign")
	if err != nil || g == nil {
		t.Fatalf("CreateGame failed: %v", err)
	}

	retrievedGame, err := database.GetGame(g.ID)
	if err != nil || retrievedGame == nil {
		t.Fatalf("GetGame failed: %v", err)
	}

	// 3. Test Join Game
	c.GameID = g.ID
	if err := database.SaveCharacter(&c); err != nil {
		t.Fatalf("SaveCharacter join game failed: %v", err)
	}

	party, err := database.GetCharactersForGame(g.ID)
	if err != nil {
		t.Fatalf("GetCharactersForGame failed: %v", err)
	}
	if len(party) != 1 {
		t.Errorf("Expected party size 1, got %d", len(party))
	}
}
