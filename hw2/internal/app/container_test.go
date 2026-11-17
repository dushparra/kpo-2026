package app_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/dig"

	"github.com/dushparra/kpo-2026/hw2/internal/app"
	"github.com/dushparra/kpo-2026/hw2/internal/usecase"
)

func TestDigContainerRunsAssessmentAndExport(t *testing.T) {
	container, err := app.NewContainer()
	if err != nil {
		t.Fatal(err)
	}
	service := resolveService(t, container)
	if err := service.AddStudent("1", "Иван"); err != nil {
		t.Fatal(err)
	}
	if err := service.AddCriterion("1", "Экзамен", 20, 1); err != nil {
		t.Fatal(err)
	}
	if err := service.SetEvaluation("1", "1", 15); err != nil {
		t.Fatal(err)
	}
	result, err := service.Result("1")
	if err != nil || result.Score != 7.5 {
		t.Fatalf("result: %+v, %v", result, err)
	}
	for _, format := range []string{"csv", "json"} {
		path := filepath.Join(t.TempDir(), "results."+format)
		if err := service.Export(format, path); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(path)
		if err != nil || len(data) == 0 {
			t.Fatalf("export %s: %q, %v", format, data, err)
		}
		if format == "json" && !json.Valid(data) {
			t.Fatalf("invalid JSON: %s", data)
		}
	}
	// Повторное разрешение зависимостей в одном контейнере сохраняет сеанс.
	students, err := resolveService(t, container).Students()
	if err != nil || len(students) != 1 {
		t.Fatal("the container did not reuse the session's repository")
	}
	other, err := app.NewContainer()
	if err != nil {
		t.Fatal(err)
	}
	students, err = resolveService(t, other).Students()
	if err != nil || len(students) != 0 {
		t.Fatal("a new container reused another session's state")
	}
}

func resolveService(t *testing.T, container *dig.Container) *usecase.Service {
	t.Helper()
	var service *usecase.Service
	if err := container.Invoke(func(resolved *usecase.Service) {
		service = resolved
	}); err != nil {
		t.Fatal(err)
	}
	return service
}
