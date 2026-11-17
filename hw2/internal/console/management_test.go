package console_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dushparra/kpo-2026/hw2/internal/console"
	"github.com/dushparra/kpo-2026/hw2/internal/usecase"
)

func TestConsoleEditsAndDeletesThroughSeparateMenuActions(t *testing.T) {
	service := newService(t)
	path := filepath.Join(t.TempDir(), "updated.json")
	input := strings.Join([]string{
		"1", "1", "Иван", "1", "2", "Анна",
		"3", "1", "Контрольная", "10", "2", "3", "2", "Экзамен", "100", "3",
		"5", "1", "1", "8", "5", "1", "2", "60", "5", "2", "1", "10",
		"11", "1", "1", "Иван Сидоров", "2",
		"11", "2", "1", "Практика", "20", "2", "8", "1",
		"11", "2", "1", "Недопустимое изменение", "7", "2",
		"11", "1", "99", "Не добавлять", "11", "2", "99", "Не добавлять", "10", "1",
		"11", "9", "12", "9", "12", "1", "99", "12", "2", "99",
		"12", "1", "2", "12", "2", "2", "8", "1", "9",
		"10", "json", path, "0", "",
	}, "\n")
	var output bytes.Buffer
	if err := console.Run(service, strings.NewReader(input), &output); err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{
		"11 - изменить студента или критерий", "12 - удалить студента или критерий",
		"Студент изменён", "Критерий изменён", "5.20 / 10", "4.00 / 10",
		"Студент и его оценки удалены", "Критерий и оценки по нему удалены",
		"1 | 1 | Иван Сидоров | 4.00 / 10 | 1/1", "Работа завершена",
	} {
		if !strings.Contains(output.String(), fragment) {
			t.Errorf("expected %q, got:\n%s", fragment, output.String())
		}
	}
	if strings.Count(output.String(), "Ошибка:") != 7 {
		t.Fatalf("invalid edits or deletions were not rejected:\n%s", output.String())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var rows []usecase.ReportRow
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].StudentName != "Иван Сидоров" || rows[0].Score != 4 || rows[0].EvaluatedCriteria != 1 || rows[0].TotalCriteria != 1 {
		t.Fatalf("export contains stale data: %+v", rows)
	}
	criteria, err := service.Criteria()
	if err != nil || len(criteria) != 1 || criteria[0].Name() != "Практика" || criteria[0].MaxScore() != 20 {
		t.Fatalf("failed edit changed the criterion: %+v, %v", criteria, err)
	}
}

func TestConsoleIncompleteEditsAndDeletionsPreserveExistingData(t *testing.T) {
	for _, input := range []string{
		"11\n", "11\n1\n1\n", "11\n2\n1\nНовое название\n",
		"11\n2\n1\nНовое название\n20\n", "12\n1\n", "12\n2\n",
	} {
		t.Run(input, func(t *testing.T) {
			service := newService(t)
			if err := service.AddStudent("1", "Иван"); err != nil {
				t.Fatal(err)
			}
			if err := service.AddCriterion("1", "Контрольная", 10, 2); err != nil {
				t.Fatal(err)
			}
			if err := service.SetEvaluation("1", "1", 8); err != nil {
				t.Fatal(err)
			}
			if err := console.Run(service, strings.NewReader(input), io.Discard); err != nil {
				t.Fatal(err)
			}
			students, err := service.Students()
			if err != nil || len(students) != 1 || students[0].Name() != "Иван" {
				t.Fatalf("incomplete input changed student: %v, %v", students, err)
			}
			criteria, err := service.Criteria()
			if err != nil || len(criteria) != 1 || criteria[0].Name() != "Контрольная" || criteria[0].MaxScore() != 10 {
				t.Fatalf("incomplete input changed criterion: %v, %v", criteria, err)
			}
			result, err := service.Result("1")
			if err != nil || result.Score != 8 || result.EvaluatedCriteria != 1 || result.TotalCriteria != 1 {
				t.Fatalf("incomplete input changed evaluations: %+v, %v", result, err)
			}
		})
	}
}
