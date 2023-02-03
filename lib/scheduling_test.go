package lib

import (
	"fmt"
	"ranker-go/schema"
	"testing"
)

func TestGenerateRandomPairs(t *testing.T) {
	p1 := schema.Competitor{
		Name:        "P1",
		Location:    "l1",
		Description: "d1",
	}
	p2 := schema.Competitor{
		Name:        "P2",
		Location:    "l2",
		Description: "d2",
	}
	//p3 := schema.Competitor{
	//	Name:        "P3",
	//	Location:    "l3",
	//	Description: "d3",
	//}

	c := []schema.Competitor{p1, p2}

	matches := genRandomPairs(c)

	if len(matches) != 1 {
		t.Errorf("got length %d, wanted %d", len(matches), 1)
	}

	if len(c) != 2 {
		t.Errorf("unexpected side-effect on backing array, got length %d, wanted %d", len(c), 1)
	}
}

func TestGenerateRandomPairsOdd(t *testing.T) {
	p1 := schema.Competitor{
		Name:        "P1",
		Location:    "l1",
		Description: "d1",
	}
	p2 := schema.Competitor{
		Name:        "P2",
		Location:    "l2",
		Description: "d2",
	}
	p3 := schema.Competitor{
		Name:        "P3",
		Location:    "l3",
		Description: "d3",
	}

	c := []schema.Competitor{p1, p2, p3}

	matches := genRandomPairs(c)

	if len(matches) != 2 {
		t.Errorf("got length %d, wanted %d", len(matches), 2)
	}

	if len(c) != 3 {
		t.Errorf("unexpected side-effect on backing array, got length %d, wanted %d", len(c), 3)
	}
}

func TestCreateInitialMatches(t *testing.T) {
	p1 := schema.Competitor{
		Name:        "P1",
		Location:    "l1",
		Description: "d1",
	}
	p2 := schema.Competitor{
		Name:        "P2",
		Location:    "l2",
		Description: "d2",
	}
	//p3 := schema.Competitor{
	//	Name:        "P3",
	//	Location:    "l3",
	//	Description: "d3",
	//}

	c := []schema.Competitor{p1, p2}

	matches, err := createInitialMatches(c, 3)
	if err != nil {
		t.Error(err)
	}

	if len(matches) != 3 {
		fmt.Printf("%v", matches)
		t.Errorf("got length %d, wanted %d", len(matches), 3)
	}
}
