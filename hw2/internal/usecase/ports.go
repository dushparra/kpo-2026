package usecase

import "github.com/dushparra/kpo-2026/hw2/internal/assessment"

type GradebookRepository interface {
	Load() (*assessment.Gradebook, error)
	Save(book *assessment.Gradebook) error
}

type Exporter interface {
	Export(report Report, path string) error
}
