package data

import (
	"InfinityInitiative/data/interfaces"
	"sync"

	"gorm.io/gorm"
)

var (
	once   sync.Once
	handle *gorm.DB
)

// Init returns the process-wide database handle.
//
// Every model constructor calls this, and it used to open a *new* connection pool each
// time — a single game page could hold 5-10 pools to the same SQLite file, which is how
// "database is locked" happens.
func Init() *gorm.DB {
	once.Do(func() {
		handle = InitLocal()
	})

	return handle
}

func AutoMigration(model interfaces.Model) error {
	err := model.Migrate()
	return err
}
