package memory

import (
	"errors"

	"github.com/dushparra/kpo-2026/hw2/internal/assessment"
)

type Repository struct {
	book *assessment.Gradebook
}

func NewRepository() *Repository {
	return &Repository{book: assessment.NewGradebook()}
}

func (r *Repository) Load() (*assessment.Gradebook, error) {
	if r.book == nil {
		return assessment.NewGradebook(), nil
	}
	return r.book.Clone(), nil
}

func (r *Repository) Save(book *assessment.Gradebook) error {
	if book == nil {
		return errors.New("нельзя сохранить пустую ведомость")
	}
	r.book = book.Clone()
	return nil
}
