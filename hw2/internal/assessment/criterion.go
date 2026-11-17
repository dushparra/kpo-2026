package assessment

import (
	"math"
	"strconv"
	"strings"
)

type CriterionID int64

func ParseCriterionID(value string) (CriterionID, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil || id <= 0 {
		return 0, ErrInvalidCriterionID
	}
	return CriterionID(id), nil
}

func (id CriterionID) String() string { return strconv.FormatInt(int64(id), 10) }

type Weight struct {
	value float64
}

func NewWeight(value float64) (Weight, error) {
	if !finite(value) || value <= 0 {
		return Weight{}, ErrInvalidWeight
	}
	return Weight{value: value}, nil
}

func (w Weight) Value() float64 { return w.value }

type Criterion struct {
	id       CriterionID
	name     string
	maxScore float64
	weight   Weight
}

func NewCriterion(id CriterionID, name string, maxScore, weight float64) (Criterion, error) {
	if id <= 0 {
		return Criterion{}, ErrInvalidCriterionID
	}
	name = strings.TrimSpace(name)
	if name == "" || !finite(maxScore) || maxScore <= 0 {
		return Criterion{}, ErrInvalidCriterion
	}
	validWeight, err := NewWeight(weight)
	if err != nil {
		return Criterion{}, err
	}
	return Criterion{id: id, name: name, maxScore: maxScore, weight: validWeight}, nil
}

func (c Criterion) ID() CriterionID { return c.id }

func (c Criterion) Name() string { return c.name }

func (c Criterion) MaxScore() float64 { return c.maxScore }

func (c Criterion) Weight() Weight { return c.weight }

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
