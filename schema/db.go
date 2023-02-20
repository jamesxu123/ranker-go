package schema

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	StateInit       string = "state_init"
	StateNone       string = "state_none"
	StateContinuous string = "state_continuous"
	StateFinishing  string = "state_finishing"
	StateEnd        string = "state_end"
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
	_, err = RDB.Ping(context.Background()).Result()
	if err != nil {
		return err
	}
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

func RedisRetryWithWatch(ctx context.Context, txf func(tx *redis.Tx) error, key string, maxRetries int) error {
	// Retry if the key has been changed.
	for i := 0; i < maxRetries; i++ {
		err := RDB.Watch(ctx, txf, key)
		if err == nil {
			// Success.
			return nil
		}
		if err == redis.TxFailedErr {
			// Optimistic lock lost. Retry.
			continue
		}
		// Return any other error.
		return err
	}
	return errors.New("increment reached maximum number of retries")
}
