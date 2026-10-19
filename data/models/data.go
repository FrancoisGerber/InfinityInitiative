package models

// MaxInventorySize is how many distinct stacks a character can hold.
const MaxInventorySize int64 = 36

// MaxCharacters caps how many characters one account may own.
const MaxCharacters int64 = 5

type Data struct {
	Page         interface{}
	Errors       interface{}
	DropdownData interface{}
	RowsAffected int64
	ActivePage   string
	Id           int64
	Character    FullCharacter
}

// Classes are the playable classes. Keep this list in step with the art in
// views/assets/img/classes — Rogue, Druid, Paladin, Bard and Monk had no portrait, so a
// character created with them rendered a broken image (B25).
var Classes = []string{
	"Fighter",
	"Mage",
	"Cleric",
	"Ranger",
}

var ItemTypes = []string{
	"Weapon",
	"Armor",
	"Accessory",
	"Resource",
	"Food",
}

var EquippableItemTypes = []string{
	"Weapon",
	"Armor",
	"Accessory",
}

var Skills = []string{
	"Mining",
	"Woodcutting",
	"Fishing",
	"Smithing",
	"Crafting",
	"Cooking",
	"Farming",
}
