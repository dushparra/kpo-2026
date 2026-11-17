package memory_test

import (
	"testing"

	"github.com/dushparra/kpo-2026/hw2/internal/adapters/memory"
	"github.com/dushparra/kpo-2026/hw2/internal/assessment"
)

func TestRepositoryDoesNotShareMutableGradebookWithCaller(t *testing.T) {
	repository := memory.NewRepository()
	book, err := repository.Load()
	if err != nil {
		t.Fatal(err)
	}
	student, err := assessment.NewStudent(1, "Иван")
	if err != nil {
		t.Fatal(err)
	}
	if err := book.AddStudent(student); err != nil {
		t.Fatal(err)
	}
	beforeSave, err := repository.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(beforeSave.Students()) != 0 {
		t.Fatal("changing loaded copy changed repository before Save")
	}
	if err := repository.Save(book); err != nil {
		t.Fatal(err)
	}
	second, err := assessment.NewStudent(2, "Анна")
	if err != nil {
		t.Fatal(err)
	}
	if err := book.AddStudent(second); err != nil {
		t.Fatal(err)
	}
	stored, err := repository.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(stored.Students()) != 1 || stored.Students()[0].ID() != 1 {
		t.Fatalf("saved snapshot changed through old pointer: %+v", stored.Students())
	}
	if err := repository.Save(nil); err == nil {
		t.Fatal("nil gradebook was accepted")
	}
}
