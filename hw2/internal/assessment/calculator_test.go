package assessment_test

import (
	"errors"
	"math"
	"testing"

	"github.com/dushparra/kpo-2026/hw2/internal/assessment"
)

func TestCalculatorUsesWeightsAndNormalizesDifferentMaxima(t *testing.T) {
	book := newTestBook(t)
	if err := book.AddCriterion(mustCriterion(t, 2, 100, 3)); err != nil {
		t.Fatal(err)
	}
	calculator := assessment.Calculator{}

	assertResult(t, calculator, book, 0, 0, 2)
	if err := book.SetEvaluation(1, 1, 8); err != nil {
		t.Fatal(err)
	}
	assertResult(t, calculator, book, 3.2, 1, 2)
	if err := book.SetEvaluation(1, 2, 60); err != nil {
		t.Fatal(err)
	}
	assertResult(t, calculator, book, 6.8, 2, 2)
	if err := book.SetEvaluation(1, 1, 10); err != nil {
		t.Fatal(err)
	}
	assertResult(t, calculator, book, 7.6, 2, 2)
	if err := book.CancelEvaluation(1, 1); err != nil {
		t.Fatal(err)
	}
	assertResult(t, calculator, book, 3.6, 1, 2)
	// Новый критерий тоже сразу участвует в текущем итоге.
	if err := book.AddCriterion(mustCriterion(t, 3, 20, 5)); err != nil {
		t.Fatal(err)
	}
	assertResult(t, calculator, book, 1.8, 1, 3)
}

func TestCalculatorRoundsOnlyFinalResult(t *testing.T) {
	book := assessment.NewGradebook()
	if err := book.AddStudent(mustStudent(t, 1, "Иван")); err != nil {
		t.Fatal(err)
	}
	for _, id := range []assessment.CriterionID{1, 2, 3} {
		if err := book.AddCriterion(mustCriterion(t, id, 3, 1)); err != nil {
			t.Fatal(err)
		}
		if err := book.SetEvaluation(1, id, 1); err != nil {
			t.Fatal(err)
		}
	}
	assertResult(t, assessment.Calculator{}, book, 3.33, 3, 3)
}

func TestCalculatorSupportsRelativeAndLargeFiniteWeights(t *testing.T) {
	for _, scale := range []float64{math.SmallestNonzeroFloat64, 1e-307, 0.1, 1, 1e307} {
		book := assessment.NewGradebook()
		if err := book.AddStudent(mustStudent(t, 1, "Иван")); err != nil {
			t.Fatal(err)
		}
		if err := book.AddCriterion(mustCriterion(t, 1, 10, 2*scale)); err != nil {
			t.Fatal(err)
		}
		if err := book.AddCriterion(mustCriterion(t, 2, 100, 3*scale)); err != nil {
			t.Fatal(err)
		}
		if err := book.SetEvaluation(1, 1, 8); err != nil {
			t.Fatal(err)
		}
		if err := book.SetEvaluation(1, 2, 60); err != nil {
			t.Fatal(err)
		}
		assertResult(t, assessment.Calculator{}, book, 6.8, 2, 2)
	}
}

func TestCalculatorRejectsUnavailableInputs(t *testing.T) {
	calculator := assessment.Calculator{}
	book := assessment.NewGradebook()
	if err := book.AddStudent(mustStudent(t, 1, "Иван")); err != nil {
		t.Fatal(err)
	}
	if _, err := calculator.Calculate(book, 1); !errors.Is(err, assessment.ErrEmptyRubric) {
		t.Fatalf("empty rubric: %v", err)
	}
	if _, err := calculator.Calculate(book, 99); !errors.Is(err, assessment.ErrStudentNotFound) {
		t.Fatalf("unknown student: %v", err)
	}
	if _, err := calculator.Calculate(nil, 1); !errors.Is(err, assessment.ErrInvalidGradebook) {
		t.Fatalf("nil book: %v", err)
	}
}

func assertResult(t *testing.T, calculator assessment.Calculator, book *assessment.Gradebook, score float64, graded, total int) {
	t.Helper()
	result, err := calculator.Calculate(book, 1)
	if err != nil {
		t.Fatal(err)
	}
	if result.Score != score || result.EvaluatedCriteria != graded || result.TotalCriteria != total {
		t.Fatalf("want score %g and %d/%d criteria, got %+v", score, graded, total, result)
	}
}
