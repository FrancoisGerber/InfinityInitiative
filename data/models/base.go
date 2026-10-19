package models

import (
	"InfinityInitiative/data"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

// updateResult mirrors the two fields the model Update() methods read.
type updateResult struct {
	RowsAffected int64
	Error        error
}

// updateAll writes every column of model, including zero values, and never touches
// loaded associations.
//
// The previous implementation was `db.Select("*").Updates(&model)`: `model` is already a
// pointer, so that passed **T, and with associations preloaded it could upsert the
// related Character/Item/Level rows as a side effect of saving one column.
func updateAll(db *gorm.DB, model any) updateResult {
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(model); err != nil {
		return updateResult{Error: err}
	}

	columns := make([]string, 0, len(stmt.Schema.Fields))
	for _, field := range stmt.Schema.Fields {
		if field.DBName != "" && !field.PrimaryKey {
			columns = append(columns, field.DBName)
		}
	}

	result := db.Model(model).Select(columns).Omit(clause.Associations).Updates(model)

	return updateResult{RowsAffected: result.RowsAffected, Error: result.Error}
}
