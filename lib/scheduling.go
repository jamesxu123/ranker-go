package lib

import (
	"github.com/google/uuid"
	"math/rand"
	"ranker-go/schema"
	"time"
)

func genRandomPairs(persons []schema.Competitor) ([]schema.MatchPair, error) {

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

	return pairedLunches, nil
}
