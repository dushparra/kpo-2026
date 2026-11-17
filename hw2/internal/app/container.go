package app

import (
	"fmt"

	"go.uber.org/dig"

	"github.com/dushparra/kpo-2026/hw2/internal/adapters/exporter"
	"github.com/dushparra/kpo-2026/hw2/internal/adapters/memory"
	"github.com/dushparra/kpo-2026/hw2/internal/assessment"
	"github.com/dushparra/kpo-2026/hw2/internal/usecase"
)

// NewContainer регистрирует конструкторы. Dig соберёт граф при вызове Invoke.
func NewContainer() (*dig.Container, error) {
	container := dig.New()
	constructors := []any{
		func() usecase.GradebookRepository { return memory.NewRepository() },
		func() assessment.Calculator { return assessment.Calculator{} },
		func() map[string]usecase.Exporter {
			return map[string]usecase.Exporter{
				"csv":  exporter.CSV{},
				"json": exporter.JSON{},
			}
		},
		usecase.NewService,
	}
	for _, constructor := range constructors {
		if err := container.Provide(constructor); err != nil {
			return nil, fmt.Errorf("регистрация зависимостей: %w", err)
		}
	}
	return container, nil
}
