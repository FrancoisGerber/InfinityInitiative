package data

import (
	"InfinityInitiative/utils"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitLocal() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(utils.DBName()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		panic("Failed to connect database")
	}

	// SQLite is a single-writer store: WAL lets readers continue while one writer works,
	// and a busy timeout turns contention into a short wait instead of an error.
	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA busy_timeout=5000")
	db.Exec("PRAGMA foreign_keys=ON")
	db.Exec("PRAGMA synchronous=NORMAL")

	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxOpenConns(1)
	}

	return db
}
