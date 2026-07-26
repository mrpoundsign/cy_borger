package chargen

import "time"

// Stat represents a character statistic with current and max values.
type Stat struct {
	Current int `json:"current"`
	Max     int `json:"max"`
}

// Weapon represents a weapon with name, damage die, and description.
type Weapon struct {
	Name        string `json:"name"`
	Damage      string `json:"damage"`
	Description string `json:"description,omitempty"`
}

// Armor represents armor with name and damage tier/reduction.
type Armor struct {
	Name      string `json:"name"`
	Tier      string `json:"tier"`
	Reduction string `json:"reduction"`
}

// ClassInfo holds details regarding a character's class.
type ClassInfo struct {
	Name        string `json:"name"`
	Glitch      string `json:"glitch,omitempty"`
	Description string `json:"description"`
	Origin      string `json:"origin,omitempty"`
	Gift        string `json:"gift,omitempty"`
}

// Character represents a complete CY_BORG character.
type Character struct {
	ID        string          `json:"id"`
	EditCode  string          `json:"edit_code"`
	IsSaved   bool            `json:"is_saved"`
	Name      string          `json:"name"`
	Handle    string          `json:"handle"`
	Class     ClassInfo       `json:"class"`
	Style     string          `json:"style"`
	Feature   string          `json:"feature"`
	Quirk     string          `json:"quirk"`
	Obsession string          `json:"obsession"`
	Want      string          `json:"want"`
	Debt      string          `json:"debt"`
	Abilities map[string]Stat `json:"abilities"` // Strength, Agility, Presence, Toughness, Knowledge
	HP        Stat            `json:"hp"`
	Glitches  Stat            `json:"glitches"`
	GlitchDie string          `json:"glitch_die"`
	Gear      []string        `json:"gear"`
	Weapons   []Weapon        `json:"weapons"`
	Armor     []Armor         `json:"armor"`
	Cybertech []string        `json:"cybertech"`
	Apps      []string        `json:"apps"`
	Creds     int             `json:"creds"`
	GameID    string          `json:"game_id,omitempty"`
	OwnerID   string          `json:"owner_id,omitempty"`
	UpdatedAt time.Time       `json:"updated_at,omitempty"`
}
