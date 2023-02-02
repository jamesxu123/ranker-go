package schema

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Open() error {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	err = db.AutoMigrate(Competitor{})
	if err != nil {
		panic("failed to migrate Competitor")
	}
	DB = db
	return err
}
