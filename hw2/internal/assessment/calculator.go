package assessment

import "math"

type Result struct {
	Score             float64
	EvaluatedCriteria int
	TotalCriteria     int
}

// Calculator — доменный сервис расчёта взвешенного итога по шкале 0–10.
type Calculator struct{}

func (Calculator) Calculate(book *Gradebook, studentID StudentID) (Result, error) {
	if book == nil {
		return Result{}, ErrInvalidGradebook
	}
	evaluations, err := book.Evaluations(studentID)
	if err != nil {
		return Result{}, err
	}
	criteria := book.Criteria()
	if len(criteria) == 0 {
		return Result{}, ErrEmptyRubric
	}
	maxWeight := criteria[0].weight.value
	for _, criterion := range criteria {
		maxWeight = math.Max(maxWeight, criterion.weight.value)
	}
	var weightedScore, totalWeight float64
	for _, criterion := range criteria {
		// Масштаб весов не влияет на итог; деление сохраняет и очень малые веса.
		weight := criterion.weight.value / maxWeight
		totalWeight += weight
		for _, evaluation := range evaluations {
			if evaluation.criterionID == criterion.id {
				weightedScore += (evaluation.score.value / criterion.maxScore) * weight
				break
			}
		}
	}
	// Пропущенная оценка даёт ноль, но её критерий остаётся в сумме весов.
	finalScore := 10 * (weightedScore / totalWeight)
	return Result{
		Score:             math.Round(finalScore*100) / 100,
		EvaluatedCriteria: len(evaluations),
		TotalCriteria:     len(criteria),
	}, nil
}
