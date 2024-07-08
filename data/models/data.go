package models

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
	"Enemy",
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
