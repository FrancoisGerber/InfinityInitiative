package models

import (
	"InfinityInitiative/data"
	"fmt"

	"gorm.io/gorm"
)

type Base struct {
	db     *gorm.DB
	dbName string
}

func (base *Base) initDatabase(dbName string) {
	base.db = data.Init(dbName)
	base.dbName = dbName
}

func (base *Base) Log(message ...interface{}) {
	fmt.Println(message...)
}
