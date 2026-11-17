package console

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/dushparra/kpo-2026/hw2/internal/assessment"
	"github.com/dushparra/kpo-2026/hw2/internal/usecase"
)

type terminal struct {
	service     *usecase.Service
	scanner     *bufio.Scanner
	output      io.Writer
	ioError     error
	inputClosed bool
}

func Run(service *usecase.Service, input io.Reader, output io.Writer) error {
	if service == nil || input == nil || output == nil {
		return errors.New("сервис, ввод и вывод обязательны")
	}

	c := &terminal{
		service: service,
		scanner: bufio.NewScanner(input),
		output:  output,
	}
	c.write("=== ПОС: преподавательская оценка студентов ===\n")
	c.write("Данные хранятся в памяти до выхода. Результаты можно экспортировать в CSV/JSON.\n")
	c.write("Итог считается по шкале 0–10; невыставленные оценки учитываются как 0.\n")
	c.printMenu()

	for {
		command, err := c.readLine("\nВыберите действие: ")
		if c.inputClosed {
			return nil
		}
		if err != nil {
			return err
		}

		var actionError error
		switch strings.ToLower(command) {
		case "1":
			actionError = c.addStudent()
		case "2":
			actionError = c.listStudents()
		case "3":
			actionError = c.addCriterion()
		case "4":
			actionError = c.listCriteria()
		case "5":
			actionError = c.setEvaluation()
		case "6":
			actionError = c.listEvaluations()
		case "7":
			actionError = c.cancelEvaluation()
		case "8":
			actionError = c.showResult()
		case "9":
			actionError = c.showReport()
		case "10":
			actionError = c.exportReport()
		case "11":
			actionError = c.editEntity()
		case "12":
			actionError = c.deleteEntity()
		case "help":
			c.printMenu()
		case "0":
			c.write("Работа завершена.\n")
			return c.ioError
		default:
			c.write("Неизвестная команда. Введите help, чтобы увидеть меню.\n")
		}

		if c.inputClosed {
			return nil
		}
		if c.ioError != nil {
			return c.ioError
		}
		if actionError != nil {
			c.write("Ошибка: %v\n", actionError)
		}
		if c.ioError != nil {
			return c.ioError
		}
	}
}

func (c *terminal) printMenu() {
	c.write("\n1 - добавить студента\n")
	c.write("2 - показать студентов\n")
	c.write("3 - добавить критерий\n")
	c.write("4 - показать рубрикатор\n")
	c.write("5 - выставить или исправить оценку\n")
	c.write("6 - показать оценки студента\n")
	c.write("7 - отменить оценку\n")
	c.write("8 - посчитать итог студента\n")
	c.write("9 - показать рейтинг студентов\n")
	c.write("10 - экспортировать итоги\n")
	c.write("11 - изменить студента или критерий\n")
	c.write("12 - удалить студента или критерий\n")
	c.write("help - повторить меню\n")
	c.write("0 - выйти\n")
}

func (c *terminal) addStudent() error {
	id, err := c.readLine("ID студента: ")
	if err != nil {
		return err
	}
	name, err := c.readLine("ФИО студента: ")
	if err != nil {
		return err
	}
	if err := c.service.AddStudent(id, name); err != nil {
		return err
	}
	c.write("Студент добавлен.\n")
	return nil
}

func (c *terminal) listStudents() error {
	students, err := c.service.Students()
	if err != nil {
		return err
	}
	if len(students) == 0 {
		c.write("Студентов пока нет.\n")
		return nil
	}
	c.write("ID | ФИО\n")
	for _, student := range students {
		c.write("%d | %s\n", student.ID(), student.Name())
	}
	return nil
}

func (c *terminal) addCriterion() error {
	id, err := c.readLine("ID критерия: ")
	if err != nil {
		return err
	}
	name, err := c.readLine("Название критерия: ")
	if err != nil {
		return err
	}
	maximum, err := c.readNumber("Максимальный балл: ")
	if err != nil {
		return err
	}
	weight, err := c.readNumber("Вес критерия: ")
	if err != nil {
		return err
	}
	if err := c.service.AddCriterion(id, name, maximum, weight); err != nil {
		return err
	}
	c.write("Критерий добавлен.\n")
	return nil
}

func (c *terminal) listCriteria() error {
	criteria, err := c.service.Criteria()
	if err != nil {
		return err
	}
	if len(criteria) == 0 {
		c.write("Рубрикатор пока пуст.\n")
		return nil
	}
	c.write("ID | Критерий | Максимум | Вес\n")
	for _, criterion := range criteria {
		c.write("%d | %s | %g | %g\n",
			criterion.ID(), criterion.Name(), criterion.MaxScore(), criterion.Weight().Value())
	}
	return nil
}

func (c *terminal) setEvaluation() error {
	studentID, err := c.readLine("ID студента: ")
	if err != nil {
		return err
	}
	criterionID, err := c.readLine("ID критерия: ")
	if err != nil {
		return err
	}
	score, err := c.readNumber("Балл: ")
	if err != nil {
		return err
	}
	if err := c.service.SetEvaluation(studentID, criterionID, score); err != nil {
		return err
	}
	c.write("Оценка сохранена. Повторный ввод по тому же критерию исправляет оценку.\n")
	return nil
}

func (c *terminal) listEvaluations() error {
	studentID, err := c.readLine("ID студента: ")
	if err != nil {
		return err
	}
	evaluations, err := c.service.Evaluations(studentID)
	if err != nil {
		return err
	}
	if len(evaluations) == 0 {
		c.write("У студента пока нет оценок.\n")
		return nil
	}
	criteria, err := c.service.Criteria()
	if err != nil {
		return err
	}
	byID := make(map[assessment.CriterionID]assessment.Criterion, len(criteria))
	for _, criterion := range criteria {
		byID[criterion.ID()] = criterion
	}
	c.write("Критерий | Название | Балл / Максимум\n")
	for _, evaluation := range evaluations {
		criterion := byID[evaluation.CriterionID()]
		c.write("%d | %s | %g / %g\n",
			criterion.ID(), criterion.Name(), evaluation.Score().Value(), criterion.MaxScore())
	}
	return nil
}

func (c *terminal) cancelEvaluation() error {
	studentID, err := c.readLine("ID студента: ")
	if err != nil {
		return err
	}
	criterionID, err := c.readLine("ID критерия: ")
	if err != nil {
		return err
	}
	if err := c.service.CancelEvaluation(studentID, criterionID); err != nil {
		return err
	}
	c.write("Оценка отменена.\n")
	return nil
}

func (c *terminal) showResult() error {
	studentID, err := c.readLine("ID студента: ")
	if err != nil {
		return err
	}
	result, err := c.service.Result(studentID)
	if err != nil {
		return err
	}
	c.write("Итог студента %s: %.2f / 10 (оценено %d/%d критериев).\n",
		studentID, result.Score, result.EvaluatedCriteria, result.TotalCriteria)
	return nil
}

func (c *terminal) showReport() error {
	report, err := c.service.Report()
	if err != nil {
		return err
	}
	if len(report.Rows) == 0 {
		c.write("Ведомость пуста: студентов пока нет.\n")
		return nil
	}
	c.write("Место | ID | ФИО | Итог / 10 | Оценено критериев\n")
	for i, row := range report.Rows {
		c.write("%d | %d | %s | %.2f / 10 | %d/%d\n",
			i+1, row.StudentID, row.StudentName, row.Score, row.EvaluatedCriteria, row.TotalCriteria)
	}
	return nil
}

func (c *terminal) exportReport() error {
	format, err := c.readLine("Формат (csv/json): ")
	if err != nil {
		return err
	}
	path, err := c.readLine("Путь к файлу: ")
	if err != nil {
		return err
	}
	if err := c.service.Export(format, path); err != nil {
		return err
	}
	c.write("Итоги экспортированы: %s\n", path)
	return nil
}

func (c *terminal) write(format string, args ...any) {
	if c.ioError == nil {
		_, c.ioError = fmt.Fprintf(c.output, format, args...)
	}
}
