package assessment

import (
	"fmt"
	"slices"
)

// Gradebook — корень агрегата ведомости одного курса.
type Gradebook struct {
	students    []Student
	rubric      Rubric
	evaluations []Evaluation
}

func NewGradebook() *Gradebook { return &Gradebook{} }

func (g *Gradebook) AddStudent(student Student) error {
	if _, err := NewStudent(student.id, student.name); err != nil {
		return err
	}
	for _, current := range g.students {
		if current.id == student.id {
			return fmt.Errorf("%w: %d", ErrDuplicateStudent, student.id)
		}
	}
	g.students = append(g.students, student)
	return nil
}

func (g *Gradebook) UpdateStudent(id StudentID, name string) error {
	updated, err := NewStudent(id, name)
	if err != nil {
		return err
	}
	for i, student := range g.students {
		if student.id == id {
			g.students[i] = updated
			return nil
		}
	}
	return fmt.Errorf("%w: %d", ErrStudentNotFound, id)
}

func (g *Gradebook) DeleteStudent(id StudentID) error {
	if _, err := g.Student(id); err != nil {
		return err
	}
	g.students = slices.DeleteFunc(g.students, func(student Student) bool { return student.id == id })
	g.evaluations = slices.DeleteFunc(g.evaluations, func(evaluation Evaluation) bool {
		return evaluation.studentID == id
	})
	return nil
}

func (g *Gradebook) AddCriterion(criterion Criterion) error {
	return g.rubric.AddCriterion(criterion)
}

func (g *Gradebook) UpdateCriterion(id CriterionID, name string, maxScore, weight float64) error {
	updated, err := NewCriterion(id, name, maxScore, weight)
	if err != nil {
		return err
	}
	if _, err := g.rubric.Criterion(id); err != nil {
		return err
	}
	for _, evaluation := range g.evaluations {
		if evaluation.criterionID == id && evaluation.score.value > maxScore {
			return fmt.Errorf("%w: максимум %g меньше балла %g у студента %d",
				ErrInvalidScore, maxScore, evaluation.score.value, evaluation.studentID)
		}
	}
	return g.rubric.UpdateCriterion(updated)
}

func (g *Gradebook) DeleteCriterion(id CriterionID) error {
	if err := g.rubric.DeleteCriterion(id); err != nil {
		return err
	}
	g.evaluations = slices.DeleteFunc(g.evaluations, func(evaluation Evaluation) bool {
		return evaluation.criterionID == id
	})
	return nil
}

func (g *Gradebook) Student(id StudentID) (Student, error) {
	if id <= 0 {
		return Student{}, ErrInvalidStudentID
	}
	for _, student := range g.students {
		if student.id == id {
			return student, nil
		}
	}
	return Student{}, fmt.Errorf("%w: %d", ErrStudentNotFound, id)
}

func (g *Gradebook) Students() []Student {
	return slices.Clone(g.students)
}

func (g *Gradebook) Criteria() []Criterion { return g.rubric.Criteria() }

func (g *Gradebook) Rubric() Rubric { return g.rubric.clone() }

// SetEvaluation выставляет новую оценку или исправляет существующую.
func (g *Gradebook) SetEvaluation(studentID StudentID, criterionID CriterionID, score float64) error {
	if _, err := g.Student(studentID); err != nil {
		return err
	}
	evaluation, err := NewEvaluation(studentID, g.rubric, criterionID, score)
	if err != nil {
		return err
	}
	for i, current := range g.evaluations {
		if current.studentID == studentID && current.criterionID == criterionID {
			g.evaluations[i] = evaluation
			return nil
		}
	}
	g.evaluations = append(g.evaluations, evaluation)
	return nil
}

func (g *Gradebook) CancelEvaluation(studentID StudentID, criterionID CriterionID) error {
	if _, err := g.Student(studentID); err != nil {
		return err
	}
	if _, err := g.rubric.Criterion(criterionID); err != nil {
		return err
	}
	for i, evaluation := range g.evaluations {
		if evaluation.studentID == studentID && evaluation.criterionID == criterionID {
			g.evaluations = append(g.evaluations[:i], g.evaluations[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("%w: студент %d, критерий %d", ErrEvaluationNotFound, studentID, criterionID)
}

func (g *Gradebook) Evaluations(studentID StudentID) ([]Evaluation, error) {
	if _, err := g.Student(studentID); err != nil {
		return nil, err
	}
	var evaluations []Evaluation
	for _, evaluation := range g.evaluations {
		if evaluation.studentID == studentID {
			evaluations = append(evaluations, evaluation)
		}
	}
	return evaluations, nil
}

// Clone отделяет состояние репозитория от изменений вызывающего кода.
func (g *Gradebook) Clone() *Gradebook {
	return &Gradebook{
		students:    g.Students(),
		rubric:      g.rubric.clone(),
		evaluations: slices.Clone(g.evaluations),
	}
}
