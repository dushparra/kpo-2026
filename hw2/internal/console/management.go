package console

import "errors"

func (c *terminal) editEntity() error {
	choice, err := c.readLine("Что изменить (1 - студент, 2 - критерий): ")
	if err != nil {
		return err
	}
	switch choice {
	case "1":
		id, err := c.readLine("ID студента: ")
		if err != nil {
			return err
		}
		name, err := c.readLine("Новое ФИО студента: ")
		if err != nil {
			return err
		}
		if err := c.service.UpdateStudent(id, name); err != nil {
			return err
		}
		c.write("Студент изменён.\n")
	case "2":
		id, err := c.readLine("ID критерия: ")
		if err != nil {
			return err
		}
		name, err := c.readLine("Новое название критерия: ")
		if err != nil {
			return err
		}
		maximum, err := c.readNumber("Новый максимальный балл: ")
		if err != nil {
			return err
		}
		weight, err := c.readNumber("Новый вес критерия: ")
		if err != nil {
			return err
		}
		if err := c.service.UpdateCriterion(id, name, maximum, weight); err != nil {
			return err
		}
		c.write("Критерий изменён. Итоги будут рассчитаны по новым данным.\n")
	default:
		return errors.New("выберите 1 для студента или 2 для критерия")
	}
	return nil
}

func (c *terminal) deleteEntity() error {
	choice, err := c.readLine("Что удалить (1 - студент, 2 - критерий): ")
	if err != nil {
		return err
	}
	switch choice {
	case "1":
		id, err := c.readLine("ID студента: ")
		if err != nil {
			return err
		}
		if err := c.service.DeleteStudent(id); err != nil {
			return err
		}
		c.write("Студент и его оценки удалены.\n")
	case "2":
		id, err := c.readLine("ID критерия: ")
		if err != nil {
			return err
		}
		if err := c.service.DeleteCriterion(id); err != nil {
			return err
		}
		c.write("Критерий и оценки по нему удалены.\n")
	default:
		return errors.New("выберите 1 для студента или 2 для критерия")
	}
	return nil
}
