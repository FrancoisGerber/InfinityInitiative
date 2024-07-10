package models

import (
	"InfinityInitiative/data"
	"fmt"

	"gorm.io/gorm"
)

type Base struct {
	db *gorm.DB
}

func (base *Base) initDatabase() {
	base.db = data.Init()
}

func (base *Base) Log(message ...interface{}) {
	fmt.Println(message...)
}
