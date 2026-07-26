package chargen

import (
	crand "crypto/rand"
	"encoding/hex"
	"math/big"
	mrand "math/rand"
	"strconv"
)

// GenerateRandomID returns a unique random string ID.
func GenerateRandomID(bytesLen int) string {
	b := make([]byte, bytesLen)
	_, _ = crand.Read(b)
	return hex.EncodeToString(b)
}

func pickRandom(slice []string) string {
	if len(slice) == 0 {
		return ""
	}
	return slice[mrand.Intn(len(slice))]
}

func roll3d6Modifier() int {
	sum := mrand.Intn(6) + 1 + mrand.Intn(6) + 1 + mrand.Intn(6) + 1
	switch {
	case sum <= 4:
		return -3
	case sum <= 6:
		return -2
	case sum <= 8:
		return -1
	case sum <= 12:
		return 0
	case sum <= 14:
		return 1
	case sum <= 16:
		return 2
	default:
		return 3
	}
}

// GenerateCharacter rolls a new CY_BORG character.
func GenerateCharacter() Character {
	id := GenerateRandomID(8)
	editCode := GenerateRandomID(6)

	strVal := roll3d6Modifier()
	agiVal := roll3d6Modifier()
	preVal := roll3d6Modifier()
	touVal := roll3d6Modifier()
	knoVal := roll3d6Modifier()

	hpMax := mrand.Intn(6) + 1 + touVal
	if hpMax < 1 {
		hpMax = 1
	}

	glitchesMax := mrand.Intn(3) + 1

	cls := Classes[mrand.Intn(len(Classes))]
	origin := pickRandom(cls.Origins)
	gift := pickRandom(cls.Gifts)

	wpn1 := WeaponsList[mrand.Intn(len(WeaponsList))]
	wpn2 := WeaponsList[mrand.Intn(len(WeaponsList))]
	armor := ArmorList[mrand.Intn(len(ArmorList))]

	creds := (mrand.Intn(6) + 1 + mrand.Intn(6) + 1) * 10

	return Character{
		ID:        id,
		EditCode:  editCode,
		IsSaved:   false,
		Name:      pickRandom(FirstNames),
		Handle:    pickRandom(Handles),
		Style:     pickRandom(Styles),
		Feature:   pickRandom(Features),
		Quirk:     pickRandom(Quirks),
		Obsession: pickRandom(Obsessions),
		Want:      pickRandom(Wants),
		Debt:      pickRandom(Debts),
		Class: ClassInfo{
			Name:        cls.Name,
			Glitch:      cls.Glitch,
			Description: cls.Description,
			Origin:      origin,
			Gift:        gift,
		},
		Abilities: map[string]Stat{
			"Strength":  {Current: strVal, Max: strVal},
			"Agility":   {Current: agiVal, Max: agiVal},
			"Presence":  {Current: preVal, Max: preVal},
			"Toughness": {Current: touVal, Max: touVal},
			"Knowledge": {Current: knoVal, Max: knoVal},
		},
		HP:        Stat{Current: hpMax, Max: hpMax},
		Glitches:  Stat{Current: glitchesMax, Max: glitchesMax},
		GlitchDie: "d3",
		Gear: []string{
			strconv.Itoa(creds) + "¤",
			pickRandom(GearList),
			pickRandom(GearList),
		},
		Weapons:   []Weapon{wpn1, wpn2},
		Armor:     []Armor{armor},
		Cybertech: []string{pickRandom(CybertechList)},
		Apps:      []string{pickRandom(AppsList)},
		Creds:     creds,
	}
}

// CreateBlankCharacter initializes a blank CY_BORG character ready for custom editing.
func CreateBlankCharacter() Character {
	id := GenerateRandomID(8)
	editCode := GenerateRandomID(6)

	return Character{
		ID:        id,
		EditCode:  editCode,
		IsSaved:   true,
		Name:      "UNNAMED OPERATOR",
		Handle:    "operator",
		Style:     "Chipped",
		Feature:   "Tattooed with glowing neon ink",
		Quirk:     "Talks to synthetic pets",
		Obsession: "Hacking corp servers",
		Want:      "Wants a clean slate",
		Debt:      "Owes 5,000¤ to the Syndicate",
		Class: ClassInfo{
			Name:        "CYBER-PUNK",
			Glitch:      "Feedback Loop",
			Description: "A street runner making their own path in Cy.",
			Origin:      "Gutter punk",
			Gift:        "Custom Cyberware",
		},
		Abilities: map[string]Stat{
			"Strength":  {Current: 0, Max: 0},
			"Agility":   {Current: 0, Max: 0},
			"Presence":  {Current: 0, Max: 0},
			"Toughness": {Current: 0, Max: 0},
			"Knowledge": {Current: 0, Max: 0},
		},
		HP:        Stat{Current: 4, Max: 4},
		Glitches:  Stat{Current: 2, Max: 2},
		GlitchDie: "d3",
		Gear:      []string{"Default Deck", "Flashlight"},
		Weapons: []Weapon{
			{Name: "Light Pistol", Damage: "d6", Description: "Standard sidearm"},
		},
		Armor:     []Armor{},
		Cybertech: []string{},
		Apps:      []string{},
		Creds:     100,
	}
}

func CryptoRandomCode(length int) string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	result := make([]byte, length)
	for i := range result {
		num, _ := crand.Int(crand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[num.Int64()]
	}
	return string(result)
}
