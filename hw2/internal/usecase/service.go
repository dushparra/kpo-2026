package usecase

import (
	"errors"
	"fmt"
	"maps"
	"sort"
	"strings"

	"github.com/dushparra/kpo-2026/hw2/internal/assessment"
)

var (
	ErrUnsupportedFormat = errors.New("неизвестный формат экспорта")
	ErrEmptyPath         = errors.New("путь к файлу не может быть пустым")
)

// Service is a facade for console scenarios. Domain rules remain in assessment.
type Service struct {
	repository GradebookRepository
	calculator assessment.Calculator
	exporters  map[string]Exporter
}

func NewService(
	repository GradebookRepository,
	calculator assessment.Calculator,
	exporters map[string]Exporter,
) *Service {
	return &Service{
		repository: repository,
		calculator: calculator,
		exporters:  maps.Clone(exporters),
	}
}

func (s *Service) AddStudent(id, name string) error {
	studentID, err := assessment.ParseStudentID(id)
	if err != nil {
		return err
	}
	student, err := assessment.NewStudent(studentID, name)
	if err != nil {
		return err
	}
	return s.update(func(book *assessment.Gradebook) error {
		return book.AddStudent(student)
	})
}

func (s *Service) AddCriterion(id, name string, maxScore, weight float64) error {
	criterionID, err := assessment.ParseCriterionID(id)
	if err != nil {
		return err
	}
	criterion, err := assessment.NewCriterion(criterionID, name, maxScore, weight)
	if err != nil {
		return err
	}
	return s.update(func(book *assessment.Gradebook) error {
		return book.AddCriterion(criterion)
	})
}

func (s *Service) UpdateStudent(id, name string) error {
	studentID, err := assessment.ParseStudentID(id)
	if err != nil {
		return err
	}
	return s.update(func(book *assessment.Gradebook) error {
		return book.UpdateStudent(studentID, name)
	})
}

func (s *Service) DeleteStudent(id string) error {
	studentID, err := assessment.ParseStudentID(id)
	if err != nil {
		return err
	}
	return s.update(func(book *assessment.Gradebook) error {
		return book.DeleteStudent(studentID)
	})
}

func (s *Service) UpdateCriterion(id, name string, maxScore, weight float64) error {
	criterionID, err := assessment.ParseCriterionID(id)
	if err != nil {
		return err
	}
	return s.update(func(book *assessment.Gradebook) error {
		return book.UpdateCriterion(criterionID, name, maxScore, weight)
	})
}

func (s *Service) DeleteCriterion(id string) error {
	criterionID, err := assessment.ParseCriterionID(id)
	if err != nil {
		return err
	}
	return s.update(func(book *assessment.Gradebook) error {
		return book.DeleteCriterion(criterionID)
	})
}

func (s *Service) Students() ([]assessment.Student, error) {
	book, err := s.load()
	if err != nil {
		return nil, err
	}
	return book.Students(), nil
}

func (s *Service) Criteria() ([]assessment.Criterion, error) {
	book, err := s.load()
	if err != nil {
		return nil, err
	}
	return book.Criteria(), nil
}

func (s *Service) SetEvaluation(studentID, criterionID string, score float64) error {
	student, err := assessment.ParseStudentID(studentID)
	if err != nil {
		return err
	}
	criterion, err := assessment.ParseCriterionID(criterionID)
	if err != nil {
		return err
	}
	return s.update(func(book *assessment.Gradebook) error {
		return book.SetEvaluation(student, criterion, score)
	})
}

func (s *Service) CancelEvaluation(studentID, criterionID string) error {
	student, err := assessment.ParseStudentID(studentID)
	if err != nil {
		return err
	}
	criterion, err := assessment.ParseCriterionID(criterionID)
	if err != nil {
		return err
	}
	return s.update(func(book *assessment.Gradebook) error {
		return book.CancelEvaluation(student, criterion)
	})
}

func (s *Service) Evaluations(studentID string) ([]assessment.Evaluation, error) {
	id, err := assessment.ParseStudentID(studentID)
	if err != nil {
		return nil, err
	}
	book, err := s.load()
	if err != nil {
		return nil, err
	}
	return book.Evaluations(id)
}

func (s *Service) Result(studentID string) (assessment.Result, error) {
	id, err := assessment.ParseStudentID(studentID)
	if err != nil {
		return assessment.Result{}, err
	}
	book, err := s.load()
	if err != nil {
		return assessment.Result{}, err
	}
	return s.calculator.Calculate(book, id)
}

func (s *Service) Report() (Report, error) {
	book, err := s.load()
	if err != nil {
		return Report{}, err
	}
	if len(book.Criteria()) == 0 {
		return Report{}, assessment.ErrEmptyRubric
	}

	students := book.Students()
	report := Report{Rows: make([]ReportRow, 0, len(students))}
	for _, student := range students {
		result, err := s.calculator.Calculate(book, student.ID())
		if err != nil {
			return Report{}, fmt.Errorf("итог студента %d: %w", student.ID(), err)
		}
		report.Rows = append(report.Rows, ReportRow{
			StudentID:         int64(student.ID()),
			StudentName:       student.Name(),
			Score:             result.Score,
			EvaluatedCriteria: result.EvaluatedCriteria,
			TotalCriteria:     result.TotalCriteria,
		})
	}
	sort.Slice(report.Rows, func(i, j int) bool {
		if report.Rows[i].Score != report.Rows[j].Score {
			return report.Rows[i].Score > report.Rows[j].Score
		}
		return report.Rows[i].StudentID < report.Rows[j].StudentID
	})
	return report, nil
}

func (s *Service) Export(format, path string) error {
	format = strings.ToLower(strings.TrimSpace(format))
	exporter, exists := s.exporters[format]
	if !exists {
		return fmt.Errorf("%w: %q", ErrUnsupportedFormat, format)
	}
	path = strings.TrimSpace(path)
	if path == "" {
		return ErrEmptyPath
	}

	report, err := s.Report()
	if err != nil {
		return err
	}
	if err := exporter.Export(report, path); err != nil {
		return fmt.Errorf("экспорт %s: %w", format, err)
	}
	return nil
}

func (s *Service) load() (*assessment.Gradebook, error) {
	book, err := s.repository.Load()
	if err != nil {
		return nil, fmt.Errorf("загрузка ведомости: %w", err)
	}
	if book == nil {
		return nil, errors.New("репозиторий вернул пустую ведомость")
	}
	return book, nil
}

func (s *Service) update(change func(*assessment.Gradebook) error) error {
	book, err := s.load()
	if err != nil {
		return err
	}
	if err := change(book); err != nil {
		return err
	}
	if err := s.repository.Save(book); err != nil {
		return fmt.Errorf("сохранение ведомости: %w", err)
	}
	return nil
}
