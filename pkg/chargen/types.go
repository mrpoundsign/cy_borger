package chargen

import (
	"strings"
	"time"
)

// Stat represents a character statistic with current and max values.
type Stat struct {
	Current int `json:"current"`
	Max     int `json:"max"`
}

// Stats represents the core character statistics in CY_BORG as integer modifiers.
type Stats struct {
	Strength  int `json:"strength"`
	Agility   int `json:"agility"`
	Presence  int `json:"presence"`
	Toughness int `json:"toughness"`
	Knowledge int `json:"knowledge"`
}

// NamedStatInt pairs a stat's display name with its integer modifier value.
type NamedStatInt struct {
	Name  string
	Value int
}

// List returns the character stats in fixed, canonical display order.
func (s Stats) List() []NamedStatInt {
	return []NamedStatInt{
		{Name: "Strength", Value: s.Strength},
		{Name: "Agility", Value: s.Agility},
		{Name: "Presence", Value: s.Presence},
		{Name: "Toughness", Value: s.Toughness},
		{Name: "Knowledge", Value: s.Knowledge},
	}
}

// Get fetches a stat integer value by name case-insensitively.
func (s *Stats) Get(name string) (int, bool) {
	switch strings.ToLower(name) {
	case "strength":
		return s.Strength, true
	case "agility":
		return s.Agility, true
	case "presence":
		return s.Presence, true
	case "toughness":
		return s.Toughness, true
	case "knowledge":
		return s.Knowledge, true
	default:
		return 0, false
	}
}

// Set updates a stat integer value by name case-insensitively.
func (s *Stats) Set(name string, val int) bool {
	switch strings.ToLower(name) {
	case "strength":
		s.Strength = val
		return true
	case "agility":
		s.Agility = val
		return true
	case "presence":
		s.Presence = val
		return true
	case "toughness":
		s.Toughness = val
		return true
	case "knowledge":
		s.Knowledge = val
		return true
	default:
		return false
	}
}

// Weapon represents a weapon with name, damage die, and description.
type Weapon struct {
	Name        string `json:"name"`
	Damage      string `json:"damage"`
	Hands       string `json:"hands,omitempty"`
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
	ID            string    `json:"id"`
	EditCode      string    `json:"edit_code"`
	IsSaved       bool      `json:"is_saved"`
	IsDead        bool      `json:"is_dead"`
	DeathNote     string    `json:"death_note"`
	DiedAt        time.Time `json:"died_at,omitempty"`
	OwnerID       string    `json:"owner_id,omitempty"`
	OwnerUsername string    `json:"owner_username,omitempty"`
	Name          string    `json:"name"`
	Handle        string    `json:"handle"`
	Class         ClassInfo `json:"class"`
	Style         string    `json:"style"`
	Feature       string    `json:"feature"`
	Quirk         string    `json:"quirk"`
	Obsession     string    `json:"obsession"`
	Want          string    `json:"want"`
	Debt          string    `json:"debt"`
	Stats         Stats     `json:"stats"` // Strength, Agility, Presence, Toughness, Knowledge
	HP            Stat      `json:"hp"`
	Glitches      Stat      `json:"glitches"`
	GlitchDie     string    `json:"glitch_die"`
	Gear          []string  `json:"gear"`
	Weapons       []Weapon  `json:"weapons"`
	Armor         []Armor   `json:"armor"`
	Cybertech     []string  `json:"cybertech"`
	Apps          []string  `json:"apps"`
	Creds         int       `json:"creds"`
	GameID        string    `json:"game_id,omitempty"`
	GameName      string    `json:"game_name,omitempty"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
}
