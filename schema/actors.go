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

func (c *Competitor) CreateInDb() error {
	return DB.Create(c).Error
}
