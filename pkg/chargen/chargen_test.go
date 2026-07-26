package chargen

import (
	"testing"
)

func TestGenerateCharacter(t *testing.T) {
	c := GenerateCharacter()

	if c.ID == "" {
		t.Errorf("Expected non-empty ID")
	}

	if c.EditCode == "" {
		t.Errorf("Expected non-empty EditCode")
	}

	if c.Name == "" {
		t.Errorf("Expected non-empty Name")
	}

	if c.Class.Name == "" {
		t.Errorf("Expected non-empty Class Name")
	}

	if len(c.Abilities) != 5 {
		t.Errorf("Expected 5 abilities, got %d", len(c.Abilities))
	}

	for name, stat := range c.Abilities {
		if stat.Max < -3 || stat.Max > 3 {
			t.Errorf("Stat %s out of expected bounds (-3 to +3): %d", name, stat.Max)
		}
	}

	if c.HP.Max < 1 {
		t.Errorf("HP Max should be at least 1, got %d", c.HP.Max)
	}

	if c.Glitches.Max < 1 {
		t.Errorf("Glitches Max should be at least 1, got %d", c.Glitches.Max)
	}
}
