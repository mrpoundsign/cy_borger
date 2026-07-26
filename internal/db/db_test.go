package db

import (
	"os"
	"testing"

	"github.com/mrpoundsign/cy_borger/pkg/chargen"
)

func TestDB(t *testing.T) {
	dbFile := "./test_cy_borger.db"
	defer os.Remove(dbFile)

	database, err := InitDB(dbFile)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// 0. Test User
	u := User{ID: "usr_test123", Username: "Netrunner"}
	if err := database.SaveUser(&u); err != nil {
		t.Fatalf("SaveUser failed: %v", err)
	}

	retrievedUser, err := database.GetUser(u.ID)
	if err != nil || retrievedUser == nil {
		t.Fatalf("GetUser failed: %v", err)
	}
	if retrievedUser.Username != u.Username {
		t.Errorf("Expected username %s, got %s", u.Username, retrievedUser.Username)
	}

	// 1. Test Save & Get Character
	c := chargen.GenerateCharacter()
	if err := database.SaveCharacter(&c, u.ID); err != nil {
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
	g, err := database.CreateGame("Test Campaign", u.ID)
	if err != nil || g == nil {
		t.Fatalf("CreateGame failed: %v", err)
	}

	retrievedGame, err := database.GetGame(g.ID)
	if err != nil || retrievedGame == nil {
		t.Fatalf("GetGame failed: %v", err)
	}

	// 3. Test Join Game
	c.GameID = g.ID
	c.IsSaved = true
	if err := database.SaveCharacter(&c, u.ID); err != nil {
		t.Fatalf("SaveCharacter join game failed: %v", err)
	}

	party, err := database.GetCharactersForGame(g.ID)
	if err != nil {
		t.Fatalf("GetCharactersForGame failed: %v", err)
	}
	if len(party) != 1 {
		t.Errorf("Expected party size 1, got %d", len(party))
	}

	// 4. Test GetGamesByOwner & GetCharactersByOwner
	myGames, err := database.GetGamesByOwner(u.ID)
	if err != nil || len(myGames) != 1 {
		t.Errorf("Expected myGames size 1, got %d", len(myGames))
	}

	myChars, err := database.GetCharactersByOwner(u.ID)
	if err != nil || len(myChars) != 1 {
		t.Errorf("Expected myChars size 1, got %d", len(myChars))
	}
}
