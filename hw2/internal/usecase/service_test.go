package usecase_test

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/dushparra/kpo-2026/hw2/internal/adapters/exporter"
	"github.com/dushparra/kpo-2026/hw2/internal/adapters/memory"
	"github.com/dushparra/kpo-2026/hw2/internal/assessment"
	"github.com/dushparra/kpo-2026/hw2/internal/usecase"
)

func TestServiceRecalculatesResultsAfterCorrectionAndCancellation(t *testing.T) {
	service := newService(t, memory.NewRepository(), map[string]usecase.Exporter{"csv": exporter.CSV{}})
	mustSucceed(t, service.AddStudent("1", "Иван Петров"))
	mustSucceed(t, service.AddStudent("2", "Анна Смирнова"))
	mustSucceed(t, service.AddCriterion("1", "Экзамен", 100, 3))
	mustSucceed(t, service.AddCriterion("2", "Проект", 10, 1))

	mustSucceed(t, service.SetEvaluation(" 1 ", " 1 ", 50))
	assertResult(t, service, "1", 3.75, 1, 2)

	mustSucceed(t, service.SetEvaluation("1", "2", 10))
	assertResult(t, service, "1", 6.25, 2, 2)

	mustSucceed(t, service.SetEvaluation("1", "1", 100))
	assertResult(t, service, "1", 10, 2, 2)
	if err := service.SetEvaluation("1", "1", 101); !errors.Is(err, assessment.ErrInvalidScore) {
		t.Fatalf("expected invalid score, got %v", err)
	}
	if err := service.SetEvaluation("1", "99", 1); !errors.Is(err, assessment.ErrCriterionNotFound) {
		t.Fatalf("expected unknown criterion, got %v", err)
	}
	assertResult(t, service, "1", 10, 2, 2)

	mustSucceed(t, service.CancelEvaluation("1", "2"))
	assertResult(t, service, "1", 7.5, 1, 2)
	evaluations, err := service.Evaluations("1")
	mustSucceed(t, err)
	if len(evaluations) != 1 || evaluations[0].CriterionID() != 1 || evaluations[0].Score().Value() != 100 {
		t.Fatalf("correction or cancellation was not saved: %+v", evaluations)
	}

	report, err := service.Report()
	mustSucceed(t, err)
	if len(report.Rows) != 2 || report.Rows[0].StudentID != 1 || report.Rows[1].StudentID != 2 {
		t.Fatalf("unexpected report order: %+v", report.Rows)
	}
	if report.Rows[0].Score != 7.5 || report.Rows[0].EvaluatedCriteria != 1 || report.Rows[1].Score != 0 {
		t.Fatalf("report does not reflect current grades: %+v", report.Rows)
	}
}

func TestServicePreservesRepositoryErrorsAndDoesNotCommitFailedSave(t *testing.T) {
	storageError := errors.New("storage unavailable")
	repository := &failingRepository{base: memory.NewRepository(), saveError: storageError}
	service := newService(t, repository, map[string]usecase.Exporter{"csv": exporter.CSV{}})
	if err := service.AddStudent("1", "Иван"); !errors.Is(err, storageError) {
		t.Fatalf("expected save error, got %v", err)
	}
	students, err := service.Students()
	mustSucceed(t, err)
	if len(students) != 0 {
		t.Fatal("failed save changed the stored gradebook")
	}

	repository.loadError = storageError
	if _, err := service.Students(); !errors.Is(err, storageError) {
		t.Fatalf("expected load error, got %v", err)
	}
}

func TestServiceSelectsExporterAndPreservesItsError(t *testing.T) {
	writeError := errors.New("write failed")
	selected := &exporterSpy{err: writeError}
	unused := &exporterSpy{}
	formats := map[string]usecase.Exporter{"csv": unused, "json": selected}
	service := newService(t, memory.NewRepository(), formats)
	delete(formats, "json") // The caller must not change the service's configuration afterwards.
	mustSucceed(t, service.AddStudent("1", "Иван"))
	mustSucceed(t, service.AddCriterion("1", "Экзамен", 10, 1))
	mustSucceed(t, service.SetEvaluation("1", "1", 8))

	if err := service.Export(" JSON ", " result.json "); !errors.Is(err, writeError) {
		t.Fatalf("expected export error, got %v", err)
	}
	if selected.calls != 1 || unused.calls != 0 || selected.path != "result.json" {
		t.Fatalf("wrong exporter selected: selected=%+v, unused=%+v", selected, unused)
	}
	if len(selected.report.Rows) != 1 || selected.report.Rows[0].Score != 8 {
		t.Fatalf("export received incorrect data: %+v", selected.report)
	}
}

func TestServiceDoesNotOverwriteExportWhenRequestIsInvalid(t *testing.T) {
	path := filepath.Join(t.TempDir(), "results.csv")
	const original = "previous results"
	mustSucceed(t, os.WriteFile(path, []byte(original), 0o644))
	service := newService(t, memory.NewRepository(), map[string]usecase.Exporter{"csv": exporter.CSV{}})

	if err := service.Export("xml", path); !errors.Is(err, usecase.ErrUnsupportedFormat) {
		t.Fatalf("expected unsupported format, got %v", err)
	}
	if err := service.Export("csv", path); !errors.Is(err, assessment.ErrEmptyRubric) {
		t.Fatalf("expected empty rubric, got %v", err)
	}
	if err := service.Export("csv", "   "); !errors.Is(err, usecase.ErrEmptyPath) {
		t.Fatalf("expected empty path, got %v", err)
	}
	data, err := os.ReadFile(path)
	mustSucceed(t, err)
	if string(data) != original {
		t.Fatalf("invalid export changed existing file: %q", data)
	}
}

func TestServiceRejectsInvalidIDsWithoutChangingGradebook(t *testing.T) {
	service := newService(t, memory.NewRepository(), map[string]usecase.Exporter{"csv": exporter.CSV{}})
	mustSucceed(t, service.AddStudent("1", "Иван"))
	mustSucceed(t, service.AddCriterion("1", "Экзамен", 10, 1))
	mustSucceed(t, service.SetEvaluation("1", "1", 8))

	for _, id := range []string{"", " ", "0", "-1", "1.5", "1e3", "abc", "NaN", "Inf", "9223372036854775808"} {
		t.Run(id, func(t *testing.T) {
			if err := service.AddStudent(id, "Анна"); !errors.Is(err, assessment.ErrInvalidStudentID) {
				t.Fatalf("AddStudent: expected invalid student ID, got %v", err)
			}
			if err := service.AddCriterion(id, "Проект", 10, 1); !errors.Is(err, assessment.ErrInvalidCriterionID) {
				t.Fatalf("AddCriterion: expected invalid criterion ID, got %v", err)
			}
			if err := service.UpdateStudent(id, "Анна"); !errors.Is(err, assessment.ErrInvalidStudentID) {
				t.Fatalf("UpdateStudent: %v", err)
			}
			if err := service.DeleteStudent(id); !errors.Is(err, assessment.ErrInvalidStudentID) {
				t.Fatalf("DeleteStudent: %v", err)
			}
			if err := service.UpdateCriterion(id, "Проект", 10, 1); !errors.Is(err, assessment.ErrInvalidCriterionID) {
				t.Fatalf("UpdateCriterion: %v", err)
			}
			if err := service.DeleteCriterion(id); !errors.Is(err, assessment.ErrInvalidCriterionID) {
				t.Fatalf("DeleteCriterion: %v", err)
			}
			if err := service.SetEvaluation(id, "1", 2); !errors.Is(err, assessment.ErrInvalidStudentID) {
				t.Fatalf("SetEvaluation student ID: %v", err)
			}
			if err := service.SetEvaluation("1", id, 2); !errors.Is(err, assessment.ErrInvalidCriterionID) {
				t.Fatalf("SetEvaluation criterion ID: %v", err)
			}
			if err := service.CancelEvaluation(id, "1"); !errors.Is(err, assessment.ErrInvalidStudentID) {
				t.Fatalf("CancelEvaluation student ID: %v", err)
			}
			if err := service.CancelEvaluation("1", id); !errors.Is(err, assessment.ErrInvalidCriterionID) {
				t.Fatalf("CancelEvaluation criterion ID: %v", err)
			}
			if _, err := service.Evaluations(id); !errors.Is(err, assessment.ErrInvalidStudentID) {
				t.Fatalf("Evaluations student ID: %v", err)
			}
			if _, err := service.Result(id); !errors.Is(err, assessment.ErrInvalidStudentID) {
				t.Fatalf("Result student ID: %v", err)
			}
			assertResult(t, service, "1", 8, 1, 1)
		})
	}
	students, err := service.Students()
	mustSucceed(t, err)
	criteria, err := service.Criteria()
	mustSucceed(t, err)
	if len(students) != 1 || len(criteria) != 1 {
		t.Fatalf("invalid IDs changed gradebook: students=%v criteria=%v", students, criteria)
	}
}

func TestServiceNormalizesNumericIDsInEveryScenario(t *testing.T) {
	service := newService(t, memory.NewRepository(), map[string]usecase.Exporter{"json": exporter.JSON{}})
	mustSucceed(t, service.AddStudent(" 001 ", "Иван"))
	mustSucceed(t, service.AddCriterion("0002", "Экзамен", 10, 1))
	if err := service.AddStudent("1", "Анна"); !errors.Is(err, assessment.ErrDuplicateStudent) {
		t.Fatalf("leading-zero student ID bypassed uniqueness: %v", err)
	}
	if err := service.AddCriterion("2", "Проект", 10, 1); !errors.Is(err, assessment.ErrDuplicateCriterion) {
		t.Fatalf("leading-zero criterion ID bypassed uniqueness: %v", err)
	}
	mustSucceed(t, service.SetEvaluation("+1", " 02 ", 8))
	assertResult(t, service, "01", 8, 1, 1)
	evaluations, err := service.Evaluations("001")
	mustSucceed(t, err)
	if len(evaluations) != 1 || evaluations[0].StudentID() != 1 || evaluations[0].CriterionID() != 2 {
		t.Fatalf("IDs not canonical: %+v", evaluations)
	}
	report, err := service.Report()
	mustSucceed(t, err)
	if len(report.Rows) != 1 || report.Rows[0].StudentID != 1 {
		t.Fatalf("report ID not numeric: %+v", report.Rows)
	}
	mustSucceed(t, service.CancelEvaluation("0001", "+2"))
	assertResult(t, service, "1", 0, 0, 1)
}

func TestServiceAcceptsLargestInt64ID(t *testing.T) {
	const id = "9223372036854775807"
	service := newService(t, memory.NewRepository(), map[string]usecase.Exporter{"json": exporter.JSON{}})
	mustSucceed(t, service.AddStudent(id, "Иван"))
	mustSucceed(t, service.AddCriterion(id, "Экзамен", 10, 1))
	mustSucceed(t, service.SetEvaluation(id, id, 10))
	assertResult(t, service, id, 10, 1, 1)
}

func TestServiceRanksReportByScoreThenNumericID(t *testing.T) {
	service := newRankingService(t)
	assertReportOrder(t, service, 7, 2, 10, 3)

	mustSucceed(t, service.SetEvaluation("3", "1", 10))
	assertReportOrder(t, service, 3, 7, 2, 10)
	mustSucceed(t, service.CancelEvaluation("7", "1"))
	assertReportOrder(t, service, 3, 2, 10, 7)

	students, err := service.Students()
	mustSucceed(t, err)
	wantInsertionOrder := []int64{10, 2, 7, 3}
	if len(students) != len(wantInsertionOrder) {
		t.Fatalf("ranking changed student count: %v", students)
	}
	for i, student := range students {
		if int64(student.ID()) != wantInsertionOrder[i] {
			t.Fatalf("ranking changed the stored student order: %v", students)
		}
	}
}

func TestServiceExportsStudentsInRankingOrder(t *testing.T) {
	service := newRankingService(t)
	directory := t.TempDir()
	csvPath := filepath.Join(directory, "rating.csv")
	jsonPath := filepath.Join(directory, "rating.json")
	mustSucceed(t, service.Export("csv", csvPath))
	mustSucceed(t, service.Export("json", jsonPath))

	csvData, err := os.ReadFile(csvPath)
	mustSucceed(t, err)
	records, err := csv.NewReader(strings.NewReader(string(csvData))).ReadAll()
	mustSucceed(t, err)
	jsonData, err := os.ReadFile(jsonPath)
	mustSucceed(t, err)
	var rows []struct {
		StudentID int64 `json:"student_id"`
	}
	mustSucceed(t, json.Unmarshal(jsonData, &rows))
	want := []int64{7, 2, 10, 3}
	if len(records) != len(want)+1 || len(rows) != len(want) {
		t.Fatalf("exports lost students: CSV=%v JSON=%v", records, rows)
	}
	for i, id := range want {
		if records[i+1][0] != strconv.FormatInt(id, 10) || rows[i].StudentID != id {
			t.Fatalf("exports differ from ranking at place %d: CSV=%v JSON=%v", i+1, records, rows)
		}
	}
}

func TestServiceEditsAndDeletesStudentsAndCriteriaAndRebuildsRanking(t *testing.T) {
	service := newRankingService(t)
	mustSucceed(t, service.AddCriterion("2", "Проект", 100, 1))
	mustSucceed(t, service.SetEvaluation("3", "2", 100))
	assertReportOrder(t, service, 3, 7, 2, 10)

	mustSucceed(t, service.UpdateStudent("02", "Анна Новая"))
	mustSucceed(t, service.UpdateCriterion("01", "Экзамен с новым весом", 10, 3))
	assertReportOrder(t, service, 7, 2, 10, 3)
	assertResult(t, service, "7", 6.75, 1, 2)
	if err := service.UpdateCriterion("1", "Недопустимый максимум", 7, 1); !errors.Is(err, assessment.ErrInvalidScore) {
		t.Fatalf("expected existing grade protection, got %v", err)
	}
	if err := service.UpdateStudent("99", "Не должен появиться"); !errors.Is(err, assessment.ErrStudentNotFound) {
		t.Fatalf("editing created an unknown student: %v", err)
	}
	if err := service.UpdateCriterion("99", "Не должен появиться", 10, 1); !errors.Is(err, assessment.ErrCriterionNotFound) {
		t.Fatalf("editing created an unknown criterion: %v", err)
	}
	assertResult(t, service, "7", 6.75, 1, 2)

	mustSucceed(t, service.DeleteStudent("007"))
	mustSucceed(t, service.DeleteCriterion("002"))
	assertReportOrder(t, service, 2, 10, 3)
	assertResult(t, service, "2", 8, 1, 1)
	assertResult(t, service, "3", 0, 0, 1)
	path := filepath.Join(t.TempDir(), "updated.json")
	mustSucceed(t, service.Export("json", path))
	data, err := os.ReadFile(path)
	mustSucceed(t, err)
	var rows []usecase.ReportRow
	mustSucceed(t, json.Unmarshal(data, &rows))
	if len(rows) != 3 || rows[0].StudentID != 2 || rows[0].StudentName != "Анна Новая" || rows[0].Score != 8 || rows[0].TotalCriteria != 1 {
		t.Fatalf("export did not reflect edits and deletions: %+v", rows)
	}
}

func TestServiceDoesNotCommitFailedEditsOrDeletions(t *testing.T) {
	repository := &failingRepository{base: memory.NewRepository()}
	service := newService(t, repository, map[string]usecase.Exporter{"csv": exporter.CSV{}})
	mustSucceed(t, service.AddStudent("1", "Иван"))
	mustSucceed(t, service.AddCriterion("1", "Экзамен", 10, 1))
	mustSucceed(t, service.SetEvaluation("1", "1", 8))
	want := errors.New("save failed")
	repository.saveError = want
	for _, change := range []func() error{
		func() error { return service.UpdateStudent("1", "Другое имя") },
		func() error { return service.UpdateCriterion("1", "Другое название", 20, 3) },
		func() error { return service.DeleteStudent("1") },
		func() error { return service.DeleteCriterion("1") },
	} {
		if err := change(); !errors.Is(err, want) {
			t.Fatalf("expected save failure, got %v", err)
		}
		students, err := service.Students()
		mustSucceed(t, err)
		criteria, err := service.Criteria()
		mustSucceed(t, err)
		if len(students) != 1 || students[0].Name() != "Иван" || len(criteria) != 1 || criteria[0].Name() != "Экзамен" || criteria[0].MaxScore() != 10 || criteria[0].Weight().Value() != 1 {
			t.Fatalf("failed save changed stored data: %v, %v", students, criteria)
		}
		assertResult(t, service, "1", 8, 1, 1)
	}
}

func newRankingService(t *testing.T) *usecase.Service {
	t.Helper()
	service := newService(t, memory.NewRepository(), map[string]usecase.Exporter{
		"csv": exporter.CSV{}, "json": exporter.JSON{},
	})
	for _, id := range []string{"10", "2", "7", "3"} {
		mustSucceed(t, service.AddStudent(id, "Студент "+id))
	}
	mustSucceed(t, service.AddCriterion("1", "Экзамен", 10, 1))
	mustSucceed(t, service.SetEvaluation("7", "1", 9))
	mustSucceed(t, service.SetEvaluation("2", "1", 8))
	mustSucceed(t, service.SetEvaluation("10", "1", 8))
	return service
}

func assertReportOrder(t *testing.T, service *usecase.Service, want ...int64) {
	t.Helper()
	report, err := service.Report()
	mustSucceed(t, err)
	if len(report.Rows) != len(want) {
		t.Fatalf("unexpected ranking size: %+v", report.Rows)
	}
	for i, row := range report.Rows {
		if row.StudentID != want[i] {
			t.Fatalf("unexpected student at place %d: %+v, want IDs %v", i+1, report.Rows, want)
		}
	}
}

type failingRepository struct {
	base      *memory.Repository
	loadError error
	saveError error
}

func (r *failingRepository) Load() (*assessment.Gradebook, error) {
	if r.loadError != nil {
		return nil, r.loadError
	}
	return r.base.Load()
}

func (r *failingRepository) Save(book *assessment.Gradebook) error {
	if r.saveError != nil {
		return r.saveError
	}
	return r.base.Save(book)
}

type exporterSpy struct {
	path   string
	report usecase.Report
	calls  int
	err    error
}

func (s *exporterSpy) Export(report usecase.Report, path string) error {
	s.path = path
	s.report = report
	s.calls++
	return s.err
}

func newService(t *testing.T, repository usecase.GradebookRepository, exporters map[string]usecase.Exporter) *usecase.Service {
	t.Helper()
	return usecase.NewService(repository, assessment.Calculator{}, exporters)
}

func mustSucceed(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func assertResult(t *testing.T, service *usecase.Service, studentID string, score float64, graded, total int) {
	t.Helper()
	result, err := service.Result(studentID)
	mustSucceed(t, err)
	if result.Score != score || result.EvaluatedCriteria != graded || result.TotalCriteria != total {
		t.Fatalf("unexpected result for %s: %+v, want score=%v graded=%d total=%d", studentID, result, score, graded, total)
	}
}
