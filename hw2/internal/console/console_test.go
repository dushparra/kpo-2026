package console_test

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dushparra/kpo-2026/hw2/internal/adapters/exporter"
	"github.com/dushparra/kpo-2026/hw2/internal/adapters/memory"
	"github.com/dushparra/kpo-2026/hw2/internal/assessment"
	"github.com/dushparra/kpo-2026/hw2/internal/console"
	"github.com/dushparra/kpo-2026/hw2/internal/usecase"
)

func TestConsoleRunsAssessmentAndBothExports(t *testing.T) {
	service := newService(t)
	directory := t.TempDir()
	csvPath := filepath.Join(directory, "итоги.csv")
	jsonPath := filepath.Join(directory, "итоги.json")
	input := strings.Join([]string{
		"2", "4",
		"1", "1", "Иван Петров",
		"1", "2", "Анна Иванова",
		"2",
		"3", "1", "Контрольная", "20,0", "3,0",
		"3", "2", "Экзамен", "100", "1",
		"4",
		"5", "1", "1", "21", // Ошибка не должна закрыть приложение.
		"5", "1", "1", "10",
		"5", "1", "2", "100",
		"6", "1",
		"8", "1", // (10/20*3 + 100/100*1) / 4 * 10 = 6.25.
		"5", "1", "1", "20",
		"8", "1", // Исправление даёт 10.
		"7", "1", "2",
		"8", "1", // Отменённый экзамен учитывается как 0: итог 7.50.
		"9",
		"10", "csv", csvPath,
		"10", "json", jsonPath,
		"0", "",
	}, "\n")
	var output bytes.Buffer

	if err := console.Run(service, strings.NewReader(input), &output); err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{
		"Студентов пока нет", "Рубрикатор пока пуст", "Иван Петров", "Анна Иванова",
		"Контрольная", "Экзамен", "Ошибка:", "6.25 / 10", "10.00 / 10", "7.50 / 10",
		"Оценка отменена", "Работа завершена",
	} {
		if !strings.Contains(output.String(), fragment) {
			t.Errorf("expected output to contain %q, got:\n%s", fragment, output.String())
		}
	}

	result, err := service.Result("1")
	if err != nil {
		t.Fatal(err)
	}
	if result.Score != 7.5 || result.EvaluatedCriteria != 1 || result.TotalCriteria != 2 {
		t.Fatalf("unexpected final result: %+v", result)
	}
	csvData, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatal(err)
	}
	records, err := csv.NewReader(bytes.NewReader(csvData)).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 3 || records[1][0] != "1" || records[1][2] != "7.50" || records[2][2] != "0.00" {
		t.Fatalf("unexpected exported CSV: %#v", records)
	}
	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		StudentID int64   `json:"student_id"`
		Score     float64 `json:"score"`
	}
	if err := json.Unmarshal(jsonData, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].StudentID != 1 || rows[0].Score != 7.5 || rows[1].Score != 0 {
		t.Fatalf("unexpected exported JSON: %+v", rows)
	}
}

func TestConsoleShowsRankingWithPlacesAndNumericIDTies(t *testing.T) {
	service := newService(t)
	for _, student := range []struct{ id, name string }{
		{"10", "Иван"}, {"2", "Анна"}, {"7", "Пётр"}, {"3", "Мария"},
	} {
		if err := service.AddStudent(student.id, student.name); err != nil {
			t.Fatal(err)
		}
	}
	if err := service.AddCriterion("1", "Экзамен", 10, 1); err != nil {
		t.Fatal(err)
	}
	for _, evaluation := range []struct {
		id    string
		score float64
	}{{"10", 8}, {"2", 8}, {"7", 9}} {
		if err := service.SetEvaluation(evaluation.id, "1", evaluation.score); err != nil {
			t.Fatal(err)
		}
	}
	var output bytes.Buffer
	if err := console.Run(service, strings.NewReader("9\n0\n"), &output); err != nil {
		t.Fatal(err)
	}
	want := "Место | ID | ФИО | Итог / 10 | Оценено критериев\n" +
		"1 | 7 | Пётр | 9.00 / 10 | 1/1\n" +
		"2 | 2 | Анна | 8.00 / 10 | 1/1\n" +
		"3 | 10 | Иван | 8.00 / 10 | 1/1\n" +
		"4 | 3 | Мария | 0.00 / 10 | 0/1\n"
	if !strings.Contains(output.String(), want) {
		t.Fatalf("unexpected ranking:\n%s", output.String())
	}
}

func TestConsoleRejectsInvalidIDsAndContinues(t *testing.T) {
	service := newService(t)
	var commands []string
	for _, id := range []string{"", "0", "-1", "1.5", "S1", "9223372036854775808"} {
		commands = append(commands,
			"1", id, "Недопустимый студент",
			"3", id, "Недопустимый критерий", "10", "1",
		)
	}
	path := filepath.Join(t.TempDir(), "results.json")
	commands = append(commands,
		"1", "001", "Иван",
		"1", "1", "Повторный студент",
		"3", "001", "Домашка", "10", "1",
		"5", "01", "+1", "8",
		"5", "1", "-1", "9", // Ошибка ID не меняет существующую оценку.
		"8", "S1",
		"7", "0", "1",
		"6", "1", "8", "1", "9",
		"10", "json", path,
		"0", "",
	)
	var output bytes.Buffer
	if err := console.Run(service, strings.NewReader(strings.Join(commands, "\n")), &output); err != nil {
		t.Fatal(err)
	}
	students, err := service.Students()
	if err != nil || len(students) != 1 || students[0].ID() != 1 || students[0].Name() != "Иван" {
		t.Fatalf("invalid or duplicate ID added a student: %v, %v", students, err)
	}
	criteria, err := service.Criteria()
	if err != nil || len(criteria) != 1 || criteria[0].ID() != 1 {
		t.Fatalf("invalid ID added a criterion: %v, %v", criteria, err)
	}
	result, err := service.Result("1")
	if err != nil || result.Score != 8 || result.EvaluatedCriteria != 1 {
		t.Fatalf("numeric ID scenario produced wrong result: %+v, %v", result, err)
	}
	for _, fragment := range []string{"Ошибка:", "Итог студента 1: 8.00 / 10", "1 | Иван | 8.00", "Работа завершена"} {
		if !strings.Contains(output.String(), fragment) {
			t.Errorf("expected %q, got:\n%s", fragment, output.String())
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		StudentID int64   `json:"student_id"`
		Score     float64 `json:"score"`
	}
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].StudentID != 1 || rows[0].Score != 8 {
		t.Fatalf("export did not retain canonical numeric ID: %+v", rows)
	}
}

func TestConsoleRejectsInvalidNumbersAndReturnsToMenu(t *testing.T) {
	for _, number := range []string{"не число", "NaN", "+Inf", "-Inf", "1e9999"} {
		t.Run(number, func(t *testing.T) {
			service := newService(t)
			var output bytes.Buffer
			input := "3\n1\nКритерий\n" + number + "\n0\n"

			if err := console.Run(service, strings.NewReader(input), &output); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(output.String(), "нужно ввести конечное число") ||
				!strings.Contains(output.String(), "Работа завершена") {
				t.Fatalf("expected a readable error and clean exit, got:\n%s", output.String())
			}
			criteria, err := service.Criteria()
			if err != nil || len(criteria) != 0 {
				t.Fatalf("invalid input added a criterion: %v, %v", criteria, err)
			}
		})
	}
}

func TestConsoleEndsCleanlyOnEOFWithoutPartialChanges(t *testing.T) {
	for _, input := range []string{
		"", "1\n1\n", "3\n1\nКонтрольная\n10\n", "5\n1\n1\n",
	} {
		t.Run(input, func(t *testing.T) {
			service := newService(t)
			if err := console.Run(service, strings.NewReader(input), io.Discard); err != nil {
				t.Fatalf("EOF should end normally, got %v", err)
			}
			students, err := service.Students()
			if err != nil || len(students) != 0 {
				t.Fatalf("partial input added a student: %v, %v", students, err)
			}
			criteria, err := service.Criteria()
			if err != nil || len(criteria) != 0 {
				t.Fatalf("partial input added a criterion: %v, %v", criteria, err)
			}
		})
	}
}

func TestConsoleReturnsReaderFailureDuringAction(t *testing.T) {
	want := errors.New("reader failed")
	input := io.MultiReader(strings.NewReader("1\n1\n"), failingReader{err: want})
	service := newService(t)

	if err := console.Run(service, input, io.Discard); !errors.Is(err, want) {
		t.Fatalf("expected reader failure, got %v", err)
	}
	students, err := service.Students()
	if err != nil || len(students) != 0 {
		t.Fatalf("reader failure left partial changes: %v, %v", students, err)
	}
}

func TestConsoleReturnsWriterFailures(t *testing.T) {
	for _, want := range []error{io.ErrClosedPipe, io.EOF} {
		if err := console.Run(newService(t), strings.NewReader("0\n"), failingWriter{err: want}); !errors.Is(err, want) {
			t.Fatalf("expected writer failure %v, got %v", want, err)
		}
	}
}

func TestConsoleRecoversFromUnknownCommandAndExportError(t *testing.T) {
	service := newService(t)
	if err := service.AddStudent("1", "Иван Петров"); err != nil {
		t.Fatal(err)
	}
	if err := service.AddCriterion("1", "Работа", 10, 1); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "missing", "result.csv")
	input := strings.Join([]string{"unknown", "help", "10", "csv", path, "0", ""}, "\n")
	var output bytes.Buffer

	if err := console.Run(service, strings.NewReader(input), &output); err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{"Неизвестная команда", "Ошибка:", "Работа завершена"} {
		if !strings.Contains(output.String(), fragment) {
			t.Fatalf("expected %q, got:\n%s", fragment, output.String())
		}
	}
}

func newService(t *testing.T) *usecase.Service {
	t.Helper()
	return usecase.NewService(memory.NewRepository(), assessment.Calculator{}, map[string]usecase.Exporter{
		"csv":  exporter.CSV{},
		"json": exporter.JSON{},
	})
}

type failingReader struct {
	err error
}

func (r failingReader) Read([]byte) (int, error) {
	return 0, r.err
}

type failingWriter struct {
	err error
}

func (w failingWriter) Write([]byte) (int, error) {
	return 0, w.err
}
