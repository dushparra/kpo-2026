package assessment_test

import (
	"errors"
	"math"
	"testing"

	"github.com/dushparra/kpo-2026/hw2/internal/assessment"
)

func TestConstructorsRejectInvalidStudentAndCriterion(t *testing.T) {
	for _, input := range []struct {
		id   assessment.StudentID
		name string
		want error
	}{{0, "Иван", assessment.ErrInvalidStudentID}, {1, "  ", assessment.ErrInvalidStudent}} {
		if _, err := assessment.NewStudent(input.id, input.name); !errors.Is(err, input.want) {
			t.Fatalf("NewStudent(%d, %q): %v", input.id, input.name, err)
		}
	}

	tests := []struct {
		name   string
		id     assessment.CriterionID
		title  string
		max    float64
		weight float64
		want   error
	}{
		{"zero id", 0, "Экзамен", 10, 1, assessment.ErrInvalidCriterionID},
		{"empty name", 1, " ", 10, 1, assessment.ErrInvalidCriterion},
		{"zero maximum", 1, "Экзамен", 0, 1, assessment.ErrInvalidCriterion},
		{"negative maximum", 1, "Экзамен", -10, 1, assessment.ErrInvalidCriterion},
		{"NaN maximum", 1, "Экзамен", math.NaN(), 1, assessment.ErrInvalidCriterion},
		{"infinite maximum", 1, "Экзамен", math.Inf(1), 1, assessment.ErrInvalidCriterion},
		{"zero weight", 1, "Экзамен", 10, 0, assessment.ErrInvalidWeight},
		{"negative weight", 1, "Экзамен", 10, -1, assessment.ErrInvalidWeight},
		{"NaN weight", 1, "Экзамен", 10, math.NaN(), assessment.ErrInvalidWeight},
		{"infinite weight", 1, "Экзамен", 10, math.Inf(1), assessment.ErrInvalidWeight},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := assessment.NewCriterion(tt.id, tt.title, tt.max, tt.weight)
			if !errors.Is(err, tt.want) {
				t.Fatalf("want %v, got %v", tt.want, err)
			}
		})
	}
}

func TestGradebookRejectsDuplicatesAndZeroValues(t *testing.T) {
	book := newTestBook(t)
	student := mustStudent(t, 1, "  Другое имя  ")
	if err := book.AddStudent(student); !errors.Is(err, assessment.ErrDuplicateStudent) {
		t.Fatalf("duplicate student: %v", err)
	}
	criterion := mustCriterion(t, 1, 10, 1)
	if err := book.AddCriterion(criterion); !errors.Is(err, assessment.ErrDuplicateCriterion) {
		t.Fatalf("duplicate criterion: %v", err)
	}
	if err := book.AddStudent(assessment.Student{}); !errors.Is(err, assessment.ErrInvalidStudentID) {
		t.Fatalf("zero student: %v", err)
	}
	if err := book.AddCriterion(assessment.Criterion{}); !errors.Is(err, assessment.ErrInvalidCriterionID) {
		t.Fatalf("zero criterion: %v", err)
	}
	if len(book.Students()) != 1 || len(book.Criteria()) != 1 {
		t.Fatal("rejected additions changed the book")
	}
	if student.ID() != 1 || student.Name() != "Другое имя" {
		t.Fatal("constructor did not trim student fields")
	}
	// Одинаковые имена допустимы: идентичность задаётся ID.
	if err := book.AddStudent(mustStudent(t, 2, "Иван")); err != nil {
		t.Fatal(err)
	}
}

func TestEvaluationFactoryChecksRubricAndScoreBoundaries(t *testing.T) {
	book := newTestBook(t)
	tests := []struct {
		name  string
		score float64
		want  error
	}{
		{"zero", 0, nil},
		{"maximum", 10, nil},
		{"fraction", 7.5, nil},
		{"negative", -0.01, assessment.ErrInvalidScore},
		{"above maximum", 10.01, assessment.ErrInvalidScore},
		{"NaN", math.NaN(), assessment.ErrInvalidScore},
		{"positive infinity", math.Inf(1), assessment.ErrInvalidScore},
		{"negative infinity", math.Inf(-1), assessment.ErrInvalidScore},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluation, err := assessment.NewEvaluation(1, book.Rubric(), 1, tt.score)
			if !errors.Is(err, tt.want) {
				t.Fatalf("want %v, got %v", tt.want, err)
			}
			if tt.want == nil && (evaluation.Score().Value() != tt.score || evaluation.StudentID() != 1 || evaluation.CriterionID() != 1) {
				t.Fatalf("wrong evaluation: %+v", evaluation)
			}
		})
	}
	if _, err := assessment.NewEvaluation(1, book.Rubric(), 99, 5); !errors.Is(err, assessment.ErrCriterionNotFound) {
		t.Fatalf("criterion outside rubric: %v", err)
	}
	if _, err := assessment.NewEvaluation(0, book.Rubric(), 1, 5); !errors.Is(err, assessment.ErrInvalidStudentID) {
		t.Fatalf("empty student id: %v", err)
	}
}

func TestGradebookCorrectsAndCancelsWithoutPartialChanges(t *testing.T) {
	book := newTestBook(t)
	if err := book.SetEvaluation(1, 1, 8); err != nil {
		t.Fatal(err)
	}
	for _, input := range []struct {
		student   assessment.StudentID
		criterion assessment.CriterionID
		score     float64
		want      error
	}{
		{99, 1, 5, assessment.ErrStudentNotFound},
		{1, 99, 5, assessment.ErrCriterionNotFound},
		{1, 1, 11, assessment.ErrInvalidScore},
	} {
		if err := book.SetEvaluation(input.student, input.criterion, input.score); !errors.Is(err, input.want) {
			t.Fatalf("rejected evaluation: %v", err)
		}
		assertSingleScore(t, book, 8)
	}
	if err := book.SetEvaluation(1, 1, 9); err != nil {
		t.Fatal(err)
	}
	assertSingleScore(t, book, 9)
	if err := book.CancelEvaluation(1, 99); !errors.Is(err, assessment.ErrCriterionNotFound) {
		t.Fatalf("cancel unknown criterion: %v", err)
	}
	if err := book.CancelEvaluation(99, 1); !errors.Is(err, assessment.ErrStudentNotFound) {
		t.Fatalf("cancel unknown student: %v", err)
	}
	assertSingleScore(t, book, 9)
	if err := book.CancelEvaluation(1, 1); err != nil {
		t.Fatal(err)
	}
	evaluations, err := book.Evaluations(1)
	if err != nil || len(evaluations) != 0 {
		t.Fatalf("cancelled evaluation remains: %v, %v", evaluations, err)
	}
	if err := book.CancelEvaluation(1, 1); !errors.Is(err, assessment.ErrEvaluationNotFound) {
		t.Fatalf("cancel absent evaluation: %v", err)
	}
}

func TestGradebookDoesNotExposeMutableCollections(t *testing.T) {
	book := newTestBook(t)
	if err := book.SetEvaluation(1, 1, 8); err != nil {
		t.Fatal(err)
	}
	book.Students()[0] = assessment.Student{}
	book.Criteria()[0] = assessment.Criterion{}
	evaluations, err := book.Evaluations(1)
	if err != nil {
		t.Fatal(err)
	}
	evaluations[0] = assessment.Evaluation{}
	rubric := book.Rubric()
	if err := rubric.AddCriterion(mustCriterion(t, 3, 100, 2)); err != nil {
		t.Fatal(err)
	}
	clone := book.Clone()
	if err := clone.SetEvaluation(1, 1, 1); err != nil {
		t.Fatal(err)
	}
	if err := clone.AddStudent(mustStudent(t, 2, "Анна")); err != nil {
		t.Fatal(err)
	}
	if err := clone.AddCriterion(mustCriterion(t, 2, 100, 2)); err != nil {
		t.Fatal(err)
	}
	assertSingleScore(t, book, 8)
	if len(book.Students()) != 1 || len(book.Criteria()) != 1 || book.Criteria()[0].ID() != 1 {
		t.Fatal("changing a returned collection or clone changed original data")
	}
}

func TestRubricRejectsWeightOverflowWithoutAddingCriterion(t *testing.T) {
	book := assessment.NewGradebook()
	if err := book.AddCriterion(mustCriterion(t, 1, 10, math.MaxFloat64)); err != nil {
		t.Fatal(err)
	}
	if err := book.AddCriterion(mustCriterion(t, 2, 10, math.MaxFloat64)); !errors.Is(err, assessment.ErrInvalidWeight) {
		t.Fatalf("overflowing weight sum: %v", err)
	}
	if len(book.Criteria()) != 1 {
		t.Fatal("invalid criterion was added")
	}
}

func TestDomainRejectsNonpositiveIDsWithoutConsole(t *testing.T) {
	book := newTestBook(t)
	if err := book.SetEvaluation(1, 1, 8); err != nil {
		t.Fatal(err)
	}
	for _, id := range []int64{0, -1} {
		studentID := assessment.StudentID(id)
		criterionID := assessment.CriterionID(id)
		if _, err := assessment.NewStudent(studentID, "Иван"); !errors.Is(err, assessment.ErrInvalidStudentID) {
			t.Fatalf("student ID %d: %v", id, err)
		}
		if _, err := assessment.NewCriterion(criterionID, "Экзамен", 10, 1); !errors.Is(err, assessment.ErrInvalidCriterionID) {
			t.Fatalf("criterion ID %d: %v", id, err)
		}
		if _, err := assessment.NewEvaluation(studentID, book.Rubric(), 1, 5); !errors.Is(err, assessment.ErrInvalidStudentID) {
			t.Fatalf("factory accepted student ID %d: %v", id, err)
		}
		if _, err := assessment.NewEvaluation(1, book.Rubric(), criterionID, 5); !errors.Is(err, assessment.ErrInvalidCriterionID) {
			t.Fatalf("factory accepted criterion ID %d: %v", id, err)
		}
		if err := book.SetEvaluation(studentID, 1, 5); !errors.Is(err, assessment.ErrInvalidStudentID) {
			t.Fatalf("gradebook accepted student ID %d: %v", id, err)
		}
		if err := book.SetEvaluation(1, criterionID, 5); !errors.Is(err, assessment.ErrInvalidCriterionID) {
			t.Fatalf("gradebook accepted criterion ID %d: %v", id, err)
		}
		assertSingleScore(t, book, 8)
	}
}

func TestGradebookUpdatesStudentWithoutChangingIdentityOrGrades(t *testing.T) {
	book := newTestBook(t)
	if err := book.SetEvaluation(1, 1, 8); err != nil {
		t.Fatal(err)
	}
	if err := book.UpdateStudent(1, "  Иван Сидоров  "); err != nil {
		t.Fatal(err)
	}
	for _, input := range []struct {
		id   assessment.StudentID
		name string
		want error
	}{
		{0, "Иван", assessment.ErrInvalidStudentID},
		{-1, "Иван", assessment.ErrInvalidStudentID},
		{99, "Иван", assessment.ErrStudentNotFound},
		{1, " ", assessment.ErrInvalidStudent},
	} {
		if err := book.UpdateStudent(input.id, input.name); !errors.Is(err, input.want) {
			t.Fatalf("invalid student update: %v", err)
		}
	}
	student, err := book.Student(1)
	if err != nil || student.ID() != 1 || student.Name() != "Иван Сидоров" || len(book.Students()) != 1 {
		t.Fatalf("update changed identity or failed update changed data: %+v, %v", student, err)
	}
	assertSingleScore(t, book, 8)
}

func TestGradebookUpdatesCriterionAndRejectsInvalidChangesAtomically(t *testing.T) {
	book := newTestBook(t)
	if err := book.SetEvaluation(1, 1, 8); err != nil {
		t.Fatal(err)
	}
	if err := book.UpdateCriterion(1, "  Практика  ", 20, 4); err != nil {
		t.Fatal(err)
	}
	assertResult(t, assessment.Calculator{}, book, 4, 1, 1)
	for _, input := range []struct {
		id          assessment.CriterionID
		name        string
		max, weight float64
		want        error
	}{
		{0, "Критерий", 20, 4, assessment.ErrInvalidCriterionID},
		{-1, "Критерий", 20, 4, assessment.ErrInvalidCriterionID},
		{99, "Критерий", 20, 4, assessment.ErrCriterionNotFound},
		{1, " ", 20, 4, assessment.ErrInvalidCriterion},
		{1, "Плохой максимум", 7, 4, assessment.ErrInvalidScore},
		{1, "NaN", math.NaN(), 4, assessment.ErrInvalidCriterion},
		{1, "Бесконечный вес", 20, math.Inf(1), assessment.ErrInvalidWeight},
	} {
		if err := book.UpdateCriterion(input.id, input.name, input.max, input.weight); !errors.Is(err, input.want) {
			t.Fatalf("invalid criterion update: %v", err)
		}
	}
	criterion := book.Criteria()[0]
	if criterion.ID() != 1 || criterion.Name() != "Практика" || criterion.MaxScore() != 20 || criterion.Weight().Value() != 4 {
		t.Fatalf("failed update changed criterion: %+v", criterion)
	}
	assertSingleScore(t, book, 8)
	assertResult(t, assessment.Calculator{}, book, 4, 1, 1)

	if err := book.AddCriterion(mustCriterion(t, 2, 100, math.MaxFloat64)); err != nil {
		t.Fatal(err)
	}
	// При замене старый вес исключается из суммы, а не прибавляется повторно.
	if err := book.UpdateCriterion(2, "Экзамен", 100, math.MaxFloat64); err != nil {
		t.Fatal(err)
	}
	if err := book.UpdateCriterion(1, "Переполнение", 20, math.MaxFloat64); !errors.Is(err, assessment.ErrInvalidWeight) {
		t.Fatalf("overflowing replacement weight: %v", err)
	}
	if book.Criteria()[0].Name() != "Практика" || book.Criteria()[0].Weight().Value() != 4 {
		t.Fatal("overflowing weight changed the criterion")
	}
}

func TestGradebookDeletesStudentsAndCriteriaWithTheirEvaluations(t *testing.T) {
	book := newTestBook(t)
	if err := book.AddStudent(mustStudent(t, 2, "Анна")); err != nil {
		t.Fatal(err)
	}
	if err := book.AddCriterion(mustCriterion(t, 2, 100, 3)); err != nil {
		t.Fatal(err)
	}
	for _, grade := range []struct {
		student   assessment.StudentID
		criterion assessment.CriterionID
		score     float64
	}{{1, 1, 8}, {1, 2, 60}, {2, 1, 5}, {2, 2, 100}} {
		if err := book.SetEvaluation(grade.student, grade.criterion, grade.score); err != nil {
			t.Fatal(err)
		}
	}
	for _, input := range []struct {
		id                           int64
		studentError, criterionError error
	}{{0, assessment.ErrInvalidStudentID, assessment.ErrInvalidCriterionID},
		{-1, assessment.ErrInvalidStudentID, assessment.ErrInvalidCriterionID},
		{99, assessment.ErrStudentNotFound, assessment.ErrCriterionNotFound}} {
		if err := book.DeleteStudent(assessment.StudentID(input.id)); !errors.Is(err, input.studentError) {
			t.Fatalf("invalid student deletion: %v", err)
		}
		if err := book.DeleteCriterion(assessment.CriterionID(input.id)); !errors.Is(err, input.criterionError) {
			t.Fatalf("invalid criterion deletion: %v", err)
		}
	}
	if err := book.DeleteStudent(1); err != nil {
		t.Fatal(err)
	}
	if _, err := book.Student(1); !errors.Is(err, assessment.ErrStudentNotFound) || len(book.Students()) != 1 {
		t.Fatalf("student was not deleted: %v", err)
	}
	if err := book.AddStudent(mustStudent(t, 1, "Другой студент")); err != nil {
		t.Fatal(err)
	}
	if evaluations, err := book.Evaluations(1); err != nil || len(evaluations) != 0 {
		t.Fatalf("deleted student's grades reappeared: %v, %v", evaluations, err)
	}
	if err := book.DeleteCriterion(1); err != nil {
		t.Fatal(err)
	}
	evaluations, err := book.Evaluations(2)
	if err != nil || len(evaluations) != 1 || evaluations[0].CriterionID() != 2 || evaluations[0].Score().Value() != 100 {
		t.Fatalf("criterion deletion damaged unrelated grades: %v, %v", evaluations, err)
	}
	if result, err := (assessment.Calculator{}).Calculate(book, 2); err != nil || result.Score != 10 || result.TotalCriteria != 1 {
		t.Fatalf("result did not reflect criterion deletion: %+v, %v", result, err)
	}
	if err := book.AddCriterion(mustCriterion(t, 1, 10, 2)); err != nil {
		t.Fatal(err)
	}
	evaluations, err = book.Evaluations(2)
	if err != nil || len(evaluations) != 1 {
		t.Fatalf("deleted criterion's grades reappeared: %v, %v", evaluations, err)
	}
	for _, id := range []assessment.CriterionID{2, 1} {
		if err := book.DeleteCriterion(id); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := (assessment.Calculator{}).Calculate(book, 2); !errors.Is(err, assessment.ErrEmptyRubric) {
		t.Fatalf("last criterion deletion: %v", err)
	}
}

func newTestBook(t *testing.T) *assessment.Gradebook {
	t.Helper()
	book := assessment.NewGradebook()
	if err := book.AddStudent(mustStudent(t, 1, "Иван")); err != nil {
		t.Fatal(err)
	}
	if err := book.AddCriterion(mustCriterion(t, 1, 10, 2)); err != nil {
		t.Fatal(err)
	}
	return book
}

func mustStudent(t *testing.T, id assessment.StudentID, name string) assessment.Student {
	t.Helper()
	student, err := assessment.NewStudent(id, name)
	if err != nil {
		t.Fatal(err)
	}
	return student
}

func mustCriterion(t *testing.T, id assessment.CriterionID, maxScore, weight float64) assessment.Criterion {
	t.Helper()
	criterion, err := assessment.NewCriterion(id, "Критерий "+id.String(), maxScore, weight)
	if err != nil {
		t.Fatal(err)
	}
	return criterion
}

func assertSingleScore(t *testing.T, book *assessment.Gradebook, score float64) {
	t.Helper()
	evaluations, err := book.Evaluations(1)
	if err != nil || len(evaluations) != 1 || evaluations[0].Score().Value() != score {
		t.Fatalf("want single score %g, got %v, %v", score, evaluations, err)
	}
}
