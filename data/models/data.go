package models

var MaxInventorySize int64 = 24

type Data struct {
	Page         interface{}
	Errors       interface{}
	DropdownData interface{}
	RowsAffected int64
	ActivePage   string
	Id           int64
	Character    FullCharacter
}

var Classes = []string{
	"Fighter",
	"Mage",
	"Rogue",
	"Cleric",
	"Druid",
	"Paladin",
	"Ranger",
	"Bard",
	"Monk",
}

var ItemTypes = []string{
	"Weapon",
	"Armor",
	"Accessory",
	"Resource",
	"Food",
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
