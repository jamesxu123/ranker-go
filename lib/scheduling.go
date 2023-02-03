package lib

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"math/rand"
	"ranker-go/schema"
	"sync"
	"time"
)

const maxRetries = 1000
const settingsKey = "settings:scheduler-state"

func GetAllMatches() ([]schema.MatchPair, error) {
	matches := make([]schema.MatchPair, 0)
	ctx, cancel := context.WithCancel(context.Background()) // implement WithCancelCause when mac gets go1.20
	defer cancel()
	iter := schema.RDB.Scan(ctx, 0, "match:*", 0).Iterator()
	var wg sync.WaitGroup
	var lock sync.Mutex

	for iter.Next(ctx) {
		key := iter.Val()
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			var model schema.RedisMatchPair
			select {
			case <-ctx.Done():
				return
			default:
			}
			err := schema.RDB.HGetAll(ctx, key).Scan(&model)
			if err != nil {
				cancel()
			}
			mp, cErr := model.CreateMatchPair(key)
			if cErr != nil {
				cancel()
			}
			lock.Lock()
			matches = append(matches, mp)
			lock.Unlock()
		}(key)
	}
	wg.Wait()
	return matches, nil
}

func DeleteAllMatches() error {
	ctx := context.Background()

	iter := schema.RDB.Scan(ctx, 0, "match:*", 0).Iterator()
	var wg sync.WaitGroup
	defer wg.Wait()
	for iter.Next(ctx) {
		key := iter.Val()
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			schema.RDB.Del(ctx, key)
		}(key)
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

	return stateTransition(schema.StateNone, schema.StateInit, func(ctx context.Context, pipe redis.Pipeliner) error {
		for _, match := range matches {
			pipe.HSet(ctx, "match:"+match.MatchPairID, "competitor1", match.Competitor1.ID)
			pipe.HSet(ctx, "match:"+match.MatchPairID, "competitor2", match.Competitor2.ID)
			pipe.HSet(ctx, "match:"+match.MatchPairID, "taken", false) // 0
		}
		return nil
	})
}

func createInitialMatches(persons []schema.Competitor, initRounds int) ([]schema.MatchPair, error) {
	if initRounds < 1 {
		return nil, errors.New("initRounds must be at least 1")
	}
	matches := make([]schema.MatchPair, 0)
	for i := 0; i < initRounds; i++ {
		matches = append(matches, genRandomPairs(persons)...)
	}
	return matches, nil
}

func stateTransition(state0 string, state1 string, fn func(ctx context.Context, pipeliner redis.Pipeliner) error) error {
	ctx := context.Background()

	txf := func(tx *redis.Tx) error {
		state, err := tx.Get(ctx, settingsKey).Result()
		if err != nil {
			return err
		}
		if state != state0 {
			return errors.New("settings:scheduler-state has changed")
		}

		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			fnErr := fn(ctx, pipe)
			if fnErr != nil {
				return fnErr
			}
			pipe.Set(ctx, settingsKey, state1, 0)
			return nil
		})
		return err
	}

	// Retry if the key has been changed.
	for i := 0; i < maxRetries; i++ {
		err := schema.RDB.Watch(ctx, txf, settingsKey)
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

func DetermineCurrentState() error {
	// TODO: pass message from TX to main function
	ctx := context.Background()
	txf := func(tx *redis.Tx) error {
		currentRDBState, err := tx.Get(ctx, settingsKey).Result()
		if err != nil {
			//return "", err
		}
		switch currentRDBState {
		case schema.StateNone:
			//return schema.StateNone, err
		case schema.StateInit:

		}
		return err
	}

	// Retry if the key has been changed.
	for i := 0; i < maxRetries; i++ {
		err := schema.RDB.Watch(ctx, txf, settingsKey)
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
