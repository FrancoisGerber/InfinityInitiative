package data

import (
	"InfinityInitiative/utils"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitRemote() *gorm.DB {
	db, err := gorm.Open(sqlite.New(sqlite.Config{
		DSN:        utils.RemoteConn(),
		DriverName: "libsql",
	}), &gorm.Config{})

	if err != nil {
		panic("Failed to connect to database")
	}

	DB = db
	return db
}
