package assessment

import (
	"fmt"
	"slices"
)

type Rubric struct {
	criteria []Criterion
}

func (r *Rubric) AddCriterion(criterion Criterion) error {
	// Проверяем и нулевое значение Criterion{}, которое в Go можно создать без New.
	if _, err := NewCriterion(criterion.id, criterion.name, criterion.maxScore, criterion.weight.value); err != nil {
		return err
	}
	var totalWeight float64
	for _, current := range r.criteria {
		if current.id == criterion.id {
			return fmt.Errorf("%w: %d", ErrDuplicateCriterion, criterion.id)
		}
		totalWeight += current.weight.value
	}
	if !finite(totalWeight + criterion.weight.value) {
		return ErrInvalidWeight
	}
	r.criteria = append(r.criteria, criterion)
	return nil
}

func (r *Rubric) UpdateCriterion(criterion Criterion) error {
	if _, err := NewCriterion(criterion.id, criterion.name, criterion.maxScore, criterion.weight.value); err != nil {
		return err
	}
	index := -1
	totalWeight := criterion.weight.value
	for i, current := range r.criteria {
		if current.id == criterion.id {
			index = i
		} else {
			totalWeight += current.weight.value
		}
	}
	if index == -1 {
		return fmt.Errorf("%w: %d", ErrCriterionNotFound, criterion.id)
	}
	if !finite(totalWeight) {
		return ErrInvalidWeight
	}
	r.criteria[index] = criterion
	return nil
}

func (r *Rubric) DeleteCriterion(id CriterionID) error {
	if id <= 0 {
		return ErrInvalidCriterionID
	}
	for i, criterion := range r.criteria {
		if criterion.id == id {
			r.criteria = slices.Delete(r.criteria, i, i+1)
			return nil
		}
	}
	return fmt.Errorf("%w: %d", ErrCriterionNotFound, id)
}

func (r Rubric) Criterion(id CriterionID) (Criterion, error) {
	if id <= 0 {
		return Criterion{}, ErrInvalidCriterionID
	}
	for _, criterion := range r.criteria {
		if criterion.id == id {
			return criterion, nil
		}
	}
	return Criterion{}, fmt.Errorf("%w: %d", ErrCriterionNotFound, id)
}

func (r Rubric) Criteria() []Criterion {
	return slices.Clone(r.criteria)
}

func (r Rubric) clone() Rubric {
	return Rubric{criteria: r.Criteria()}
}
