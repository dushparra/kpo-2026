package assessment

import "fmt"

type Score struct {
	value float64
}

func NewScore(value float64) (Score, error) {
	if !finite(value) || value < 0 {
		return Score{}, ErrInvalidScore
	}
	return Score{value: value}, nil
}

func (s Score) Value() float64 { return s.value }

type Evaluation struct {
	studentID   StudentID
	criterionID CriterionID
	score       Score
}

// NewEvaluation создаёт оценку по реальному критерию рубрикатора.
// Наличие студента в ведомости проверяет корень агрегата Gradebook.
func NewEvaluation(studentID StudentID, rubric Rubric, criterionID CriterionID, value float64) (Evaluation, error) {
	if studentID <= 0 {
		return Evaluation{}, ErrInvalidStudentID
	}
	criterion, err := rubric.Criterion(criterionID)
	if err != nil {
		return Evaluation{}, err
	}
	score, err := NewScore(value)
	if err != nil {
		return Evaluation{}, err
	}
	if score.value > criterion.maxScore {
		return Evaluation{}, fmt.Errorf("%w: допустимо от 0 до %g", ErrInvalidScore, criterion.maxScore)
	}
	return Evaluation{studentID: studentID, criterionID: criterionID, score: score}, nil
}

func (e Evaluation) StudentID() StudentID { return e.studentID }

func (e Evaluation) CriterionID() CriterionID { return e.criterionID }

func (e Evaluation) Score() Score { return e.score }
