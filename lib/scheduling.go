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
const settingsKey = "settings:scheduler-state" // for future, this should be stored elsewhere

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
}

func SeedStart(persons []schema.Competitor, initRounds int) error {
	matches, err := createInitialMatches(persons, initRounds)
	if err != nil {
		return err
	}

	return stateTransition(schema.StateNone, schema.StateInit, func(ctx context.Context, pipe redis.Pipeliner) error {
		for i := range persons {
			pipe.ZAdd(ctx, "scheduler:competitors", redis.Z{Score: 0, Member: persons[i].ID})
		}

		for _, match := range matches {
			pipe.HSet(ctx, "match:"+match.MatchPairID, "competitor1", match.Competitor1.ID)
			pipe.HSet(ctx, "match:"+match.MatchPairID, "competitor2", match.Competitor2.ID)
			pipe.HSet(ctx, "match:"+match.MatchPairID, "taken", false) // 0
			pipe.ZAdd(ctx, "scheduler:assignment_status", redis.Z{
				Score:  0,
				Member: match.MatchPairID,
			})
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

	return schema.RedisRetryWithWatch(ctx, txf, settingsKey, maxRetries)
}

func SetSchedulerState(newState string) {
	schema.RDB.Set(context.Background(), settingsKey, newState, 0)
}

func DetermineCurrentState() (string, error) {
	// TODO: pass message from TX to main function
	ctx := context.Background()
	s := make(chan string)
	txf := func(tx *redis.Tx) error {
		currentRDBState, err := tx.Get(ctx, settingsKey).Result()
		if err != nil {
			//return "", err
			return err
		}
		switch currentRDBState {
		case schema.StateNone:
			s <- schema.StateNone
		case schema.StateInit:
			iter := tx.Scan(ctx, 0, "match:*", 0).Iterator()
			if iter.Next(ctx) == false {
				s <- schema.StateContinuous
			} else {
				s <- schema.StateInit
			}
		case schema.StateContinuous:
			s <- schema.StateContinuous
		}
		return nil
	}
	err := make(chan error)
	go func() {
		err <- schema.RedisRetryWithWatch(ctx, txf, settingsKey, maxRetries)
	}()
	return <-s, <-err
}

func GetMatchForJudge() error {
	// TODO: in progress
	ctx := context.Background()

	state, err := schema.RDB.Get(ctx, settingsKey).Result()
	if err != nil {
		return err
	}

	switch state {
	case schema.StateInit:
		// logic as follows
		// 1. rank using scheduler:assignment_status the matches that have had the least hits
		// 2. if not marked, mark the match as taken in match:* and increment assignment_status. Do this atomically
		// 3. increment scheduler:competitors for each person in the match
		// 4. profit
	case schema.StateContinuous:
		// logic as follows
		// 1. need some logic about cleaning up the matches that have been abandoned? maybe at a later date
		// 2. find the competitor with the highest variation or lowest confidence
		// 3. pair with the closest score
		// 4. add this to the list of matches and assignment_status. do this atomically
	case schema.StateFinishing:
		// clean up first by ranking in assignment_status by least and completing all the 0 ones
		// terminate
	}
	return nil
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

	for i, person := range clonedPersons[:len(clonedPersons)/2] {
		pairedLunches[i] = schema.MatchPair{
			MatchPairID: uuid.New().String(),
			Competitor1: person,
			Competitor2: clonedPersons[len(persons)-1-i],
		}
	}

	return pairedLunches
}
