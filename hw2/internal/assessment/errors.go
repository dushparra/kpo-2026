package assessment

import "errors"

var (
	ErrInvalidStudentID   = errors.New("ID студента должен быть положительным целым числом")
	ErrInvalidCriterionID = errors.New("ID критерия должен быть положительным целым числом")
	ErrInvalidStudent     = errors.New("имя студента не может быть пустым")
	ErrInvalidCriterion   = errors.New("у критерия должны быть название и положительный конечный максимум")
	ErrInvalidWeight      = errors.New("вес и сумма весов должны быть положительными конечными числами")
	ErrInvalidScore       = errors.New("балл должен быть конечным числом от нуля до максимума критерия")
	ErrDuplicateStudent   = errors.New("студент с таким идентификатором уже существует")
	ErrDuplicateCriterion = errors.New("критерий с таким идентификатором уже существует")
	ErrStudentNotFound    = errors.New("студент не найден")
	ErrCriterionNotFound  = errors.New("критерий отсутствует в рубрикаторе")
	ErrEvaluationNotFound = errors.New("оценка не найдена")
	ErrEmptyRubric        = errors.New("для расчёта итога нужно добавить хотя бы один критерий")
	ErrInvalidGradebook   = errors.New("ведомость не задана")
)
