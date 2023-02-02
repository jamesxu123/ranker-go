package schema

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB
var RDB *redis.Client

func RedisOpen(connectionString string) error {
	opt, err := redis.ParseURL(connectionString)
	if err != nil {
		return err
	}
	RDB = redis.NewClient(opt)
	return nil
}

func Open() error {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	err = db.AutoMigrate(Competitor{})
	if err != nil {
		return err
	}
	DB = db
	return nil
}
