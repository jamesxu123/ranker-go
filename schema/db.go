package schema

import (
	"context"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	StateInit string = "state_init"
	StateNone string = "state_none"
)

var DB *gorm.DB
var RDB *redis.Client

func initRedis(rdb *redis.Client) {
	rdb.SetNX(context.Background(), "settings:scheduler-state", StateNone, 0)
}

func RedisOpen(connectionString string) error {
	opt, err := redis.ParseURL(connectionString)
	if err != nil {
		return err
	}
	RDB = redis.NewClient(opt)
	initRedis(RDB)
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
