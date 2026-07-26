package chargen

var FirstNames = []string{
	"SPICE", "CHROME", "RAZOR", "HEX", "VOIX", "KILOWATT", "NULL", "CYPHER",
	"VEX", "ASH", "CINDER", "NEXUS", "RIPPED", "GLITCH", "STATIC", "ZERO",
	"SHADOW", "BLADE", "VIPER", "VENOM", "HAWK", "ECHO", "GHOST", "SPECTRE",
}

var Handles = []string{
	"spicexhaxxor-wemutx172770",
	"ghost_in_the_wire_99",
	"cyber_pawn_x",
	"void_walker_01",
	"neon_blade_runner",
	"null_pointer_ex",
	"daemon_slayer_66",
	"acid_burn_v2",
}

var Styles = []string{
	"heavy makeup, hat and shades",
	"leather trenchcoat and LED glowing tattoos",
	"spiked chrome hair and torn corporate suit",
	"full tactical gear with neon graffiti",
	"hooded tech-wear with digital visor",
	"visor, combat boots, patched denim vest",
}

var Features = []string{
	"always minding your three favorite plants",
	"constantly twitching chrome eye",
	"hums retro synth tunes under breath",
	"smokes electronic cigarettes non-stop",
	"obsessively cleans weapons",
	"speaks in low monotone drone",
}

var Quirks = []string{
	"You try to start each day with meditation.",
	"You never look anyone directly in the eyes.",
	"You refuse to eat synth-food.",
	"You record all your conversations.",
	"You talk to your cyberware when stressed.",
	"You trust no one who doesn't have cybertech.",
}

var Obsessions = []string{
	"Overly interested in printed shirts.",
	"Collecting obsolete floppy disks.",
	"Tracking down rare pre-collapse vinyl records.",
	"Fixated on corporate conspiracy theories.",
	"Obsessed with cybernetic upgrading.",
}

var Wants = []string{
	"You want revenge.",
	"You want to escape G0 city.",
	"You want to destroy the Megacorp that ruined your life.",
	"You want to pay off your massive debt.",
	"You want to become a legendary cyberpunk mercenary.",
}

var Debts = []string{
	"You owe 14,000¤ to a gambling den with layers upon layers of owners.",
	"You owe 25,000¤ to a shady street doc for black-market cyberware.",
	"You owe 50,000¤ to a ruthless loan shark syndicate.",
	"You owe 8,000¤ to a corrupt corporate enforcer.",
}

type ClassPreset struct {
	Name        string
	Glitch      string
	Description string
	Origins     []string
	Gifts       []string
}

var Classes = []ClassPreset{
	{
		Name:        "Renegade Cyberslasher",
		Glitch:      "YOU DIDN'T ASK FOR THIS",
		Description: "You are DEATH incarnate—a frenzied flurry of chrome, murder and blood-stained steel. But yours is no mindless rage. You match your trained and cybernetically enhanced body with an equally disciplined mind. You used to kill for a cause, for an ideal. Now? You kill for money",
		Origins: []string{
			"Discharged dishonorably from corporate security.",
			"Former pit fighter in the illegal underground ring.",
			"Ex-mercenary left for dead in the ruins of G0.",
		},
		Gifts: []string{
			"Steelcutter chainsaw (d8 damage): Absolutely not made for combat. When hitting for maximum damage, it gets stuck for d3 rounds, dealing damage automatically.",
			"Dual Logans (d8 damage): Retractable chrome claws embedded in fists.",
		},
	},
	{
		Name:        "Orphaned Cyberdeck Hacker",
		Glitch:      "SYSTEM ERROR 404",
		Description: "You grew up plugged into the Net, deciphering corporate encryption before you could read printed text.",
		Origins: []string{
			"Raised in a subterranean illegal data server farm.",
			"Escaped from a corporate AI research program.",
		},
		Gifts: []string{
			"Deck-Jack: Direct neural connection to military-grade decks.",
			"Custom Exploit Suite: Allows hacking secured locks at DR10.",
		},
	},
	{
		Name:        "Forsaken Gang Goon",
		Glitch:      "LEFT FOR DEAD",
		Description: "Your gang was your family until they betrayed you or got wiped out. Now you wander the streets alone.",
		Origins: []string{
			"Your gang was taken out by a rival gang. They think you are dead too.",
			"Your gang broke the rules and you left on bad terms.",
		},
		Gifts: []string{
			"Heavy Chrome Brass Knuckles (d6 damage).",
			"Street Reputation: Intimidate street punks at DR10.",
		},
	},
}

var WeaponsList = []Weapon{
	{Name: "Vibro-knife", Damage: "d4", Hands: "1h", Description: "Cuts through cheap locks easily."},
	{Name: "Telescopic baton", Damage: "d6", Hands: "1h", Description: "Easily concealed."},
	{Name: "Heavy Pistol", Damage: "d6", Hands: "1h", Description: "Standard sidearm."},
	{Name: "Shotgun", Damage: "d8", Hands: "2h", Description: "Devastating at close range."},
	{Name: "Assault Rifle", Damage: "d8", Hands: "2h", Description: "Military grade."},
	{Name: "SMG", Damage: "d6a", Hands: "1h", Description: "Autofire capability."},
	{Name: "Sniper Rifle", Damage: "d10", Hands: "2h", Description: "Includes targeting scope."},
	{Name: "Steelcutter chainsaw", Damage: "d8", Hands: "2h", Description: "Absolutely not made for combat. Gets stuck for d3 rounds on max damage."},
	{Name: "Monofilament Whip", Damage: "d8", Hands: "1h", Description: "Ignores 2 points of armor."},
	{Name: "Heavy Machine Gun", Damage: "d10a", Hands: "2h", Description: "Requires high Strength or tripod to fire accurately."},
}

var ArmorList = []Armor{
	{Name: "StyleGuard", Tier: "-d2", Reduction: "-d2 damage (Looks just like clothes!)"},
	{Name: "Subdermal Plating", Tier: "-d4", Reduction: "-d4 damage"},
	{Name: "Heavy Tactical Armor", Tier: "-d6", Reduction: "-d6 damage"},
}

var GearList = []string{
	"Breathing mask (provides oxygen in gas or underwater)",
	"Mirrorshades",
	"Fake ID (Good enough to pass random checks)",
	"2 mags ammo",
	"Cyberdeck with standard OS",
	"Grappling hook & high-tensile wire",
	"Medical stim-injector (heals d6 HP)",
	"Signal jammer",
}

var CybertechList = []string{
	"Autocamo: Subdermal projection of ever-changing anti-facial-recognition patterns.",
	"Optical Camouflage: Render invisible for 1d4 rounds once per day.",
	"Neural Speed Booster: +1 to Agility defense rolls.",
	"Cyber-Arm: Integrated heavy weapon mount.",
	"Subdermal Armor: +1 Armor Tier.",
}

var AppsList = []string{
	"Blink: Teleport 30 feet in line of sight.",
	"Overclock: Double action speed for 1 round, take d4 damage after.",
	"Siphon Creds: Drain credits from electronic terminals.",
}
