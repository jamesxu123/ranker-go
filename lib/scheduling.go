package lib

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"math/rand"
	"ranker-go/schema"
	"time"
)

const maxRetries = 1000

func GetAllMatches() ([]schema.MatchPair, error) {
	matches := make([]schema.MatchPair, 0)
	ctx := context.Background()
	iter := schema.RDB.Scan(ctx, 0, "match:*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		var model schema.RedisMatchPair
		err := schema.RDB.HGetAll(ctx, key).Scan(&model)
		if err != nil {
			return nil, err
		}
		mp, cErr := model.CreateMatchPair(key)
		if cErr != nil {
			return nil, cErr
		}
		matches = append(matches, mp)
	}
	return matches, nil
}

func DeleteAllMatches() error {
	ctx := context.Background()

	iter := schema.RDB.Scan(ctx, 0, "match:*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		schema.RDB.Del(ctx, key)
	}
	return nil

	//cmds, err := schema.RDB.Pipelined(ctx, func(pipe redis.Pipeliner) error {
	//	iter := pipe.Scan(ctx, 0, "match:*", 0).Iterator()
	//	for iter.Next(ctx) {
	//		key := iter.Val()
	//		pipe.Del(ctx, key)
	//	}
	//	return nil
	//})
	//
	//for _, cmd := range cmds {
	//	fmt.Println(cmd.(*redis.ScanCmd).Val())
	//}

	//return err
}

func SeedStart(persons []schema.Competitor, initRounds int) error {
	matches, err := createInitialMatches(persons, initRounds)
	if err != nil {
		return err
	}

	ctx := context.Background()

	txf := func(tx *redis.Tx) error {
		state, err := tx.Get(ctx, "settings:scheduler-state").Result()
		if err != nil {
			return err
		}
		if state != schema.StateNone {
			return errors.New("settings:scheduler-state has changed")
		}

		state = schema.StateInit

		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			for _, match := range matches {
				pipe.HSet(ctx, "match:"+match.MatchPairID, "competitor1", match.Competitor1.ID)
				pipe.HSet(ctx, "match:"+match.MatchPairID, "competitor2", match.Competitor2.ID)
				pipe.HSet(ctx, "match:"+match.MatchPairID, "taken", false) // 0
			}
			pipe.Set(ctx, "settings:scheduler-state", state, 0)
			return nil
		})
		return err
	}

	// Retry if the key has been changed.
	for i := 0; i < maxRetries; i++ {
		err := schema.RDB.Watch(ctx, txf, "settings:scheduler-state")
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

func createInitialMatches(persons []schema.Competitor, initRounds int) ([]schema.MatchPair, error) {
	if initRounds < 1 {
		return nil, errors.New("initRounds must be at least 1")
	}
	matches := make([]schema.MatchPair, len(persons)*initRounds)
	for i := 0; i < initRounds; i++ {
		matches = append(matches, genRandomPairs(persons)...)
	}
	return matches, nil
}

func genRandomPairs(persons []schema.Competitor) []schema.MatchPair {

	clonedPersons := make([]schema.Competitor, len(persons))
	copy(clonedPersons, persons)

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(clonedPersons), func(i, j int) { clonedPersons[i], clonedPersons[j] = clonedPersons[j], clonedPersons[i] })

	if len(persons)%2 != 0 {
		clonedPersons = append(clonedPersons, clonedPersons[0])
	}

	pairedLunches := make([]schema.MatchPair, len(clonedPersons)/2)

	for i, person := range clonedPersons[:len(clonedPersons)/2] { // TODO: Bug over here, empty elements
		pairedLunches[i] = schema.MatchPair{
			MatchPairID: uuid.New().String(),
			Competitor1: person,
			Competitor2: clonedPersons[len(persons)-1-i],
		}
	}

	return pairedLunches
}
