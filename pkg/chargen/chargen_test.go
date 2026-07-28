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

	statsList := c.Stats.List()
	if len(statsList) != 5 {
		t.Errorf("Expected 5 stats, got %d", len(statsList))
	}

	for _, item := range statsList {
		if item.Value < -3 || item.Value > 3 {
			t.Errorf("Stat %s out of expected bounds (-3 to +3): %d", item.Name, item.Value)
		}
	}

	if c.HP.Max < 1 {
		t.Errorf("HP Max should be at least 1, got %d", c.HP.Max)
	}

	if c.Glitches.Max < 1 {
		t.Errorf("Glitches Max should be at least 1, got %d", c.Glitches.Max)
	}
}
