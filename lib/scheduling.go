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

func SeedStart(persons []schema.Competitor, initRounds int) error {
	matches, err := createInitialMatches(persons, initRounds)
	if err != nil {
		return err
	}

	// TODO: add status watch/check

	ctx := context.Background()
	_, rerr := schema.RDB.Pipelined(ctx, func(rdb redis.Pipeliner) error {
		for _, match := range matches {
			schema.RDB.HSet(ctx, match.MatchPairID, "competitor1", match.Competitor1.ID)
			schema.RDB.HSet(ctx, match.MatchPairID, "competitor2", match.Competitor2.ID)
			schema.RDB.HSet(ctx, match.MatchPairID, "taken", false) // 0
		}
		return nil
	})
	if rerr != nil {
		return rerr
	}

	return nil
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

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(persons), func(i, j int) { persons[i], persons[j] = persons[j], persons[i] })

	if len(persons)%2 != 0 {
		persons = append(persons, persons[0])
	}

	pairedLunches := make([]schema.MatchPair, len(persons)/2)

	for i, person := range persons[:len(persons)/2] {
		pairedLunches[i] = schema.MatchPair{
			MatchPairID: uuid.New().String(),
			Competitor1: person,
			Competitor2: persons[len(persons)-1-i],
		}
	}

	return pairedLunches
}
