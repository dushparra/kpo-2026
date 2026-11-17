package assessment

import (
	"strconv"
	"strings"
)

type StudentID int64

func ParseStudentID(value string) (StudentID, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil || id <= 0 {
		return 0, ErrInvalidStudentID
	}
	return StudentID(id), nil
}

func (id StudentID) String() string { return strconv.FormatInt(int64(id), 10) }

type Student struct {
	id   StudentID
	name string
}

func NewStudent(id StudentID, name string) (Student, error) {
	if id <= 0 {
		return Student{}, ErrInvalidStudentID
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return Student{}, ErrInvalidStudent
	}
	return Student{id: id, name: name}, nil
}

func (s Student) ID() StudentID { return s.id }

func (s Student) Name() string { return s.name }
