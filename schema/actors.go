package schema

import "gorm.io/gorm"

type Competitor struct {
	gorm.Model
	Name        string `json:"name" gorm:"uniqueIndex;type:text;not null"`
	Location    string `json:"location"`
	Description string `json:"description"`
}

type MatchPair struct {
	MatchPairID string     `json:"id"`
	Competitor1 Competitor `json:"competitor1"`
	Competitor2 Competitor `json:"competitor2"`
}

type RedisMatchPair struct {
	Taken         bool `json:"taken" redis:"taken"`
	Competitor1ID uint `json:"competitor_1_id" redis:"competitor1"`
	Competitor2ID uint `json:"competitor_2_id" redis:"competitor2"`
}

func (c *Competitor) CreateInDb() error {
	return DB.Create(c).Error
}

func findCompetitorByID(id uint) (Competitor, error) {
	var c Competitor
	DB.First(&c, id) // only get the first
	err := DB.Error
	return c, err
}

func (rmp RedisMatchPair) CreateMatchPair(id string) (MatchPair, error) {
	c1, e1 := findCompetitorByID(rmp.Competitor1ID)
	c2, e2 := findCompetitorByID(rmp.Competitor2ID)
	if e1 != nil {
		return MatchPair{}, e1
	} else if e2 != nil {
		return MatchPair{}, e2
	}
	return MatchPair{
		MatchPairID: id,
		Competitor1: c1,
		Competitor2: c2,
	}, nil
}
