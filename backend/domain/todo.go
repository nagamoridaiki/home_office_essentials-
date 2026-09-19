package domain

import "errors"

var ErrEmptyTitle = errors.New("title is empty")

type Todo struct {
	ID        int
	Title     string
	Completed bool
}

// NewTodo はIDが未採番のTodoを作る。IDはリポジトリが採番する。
func NewTodo(title string) (Todo, error) {
	if title == "" {
		return Todo{}, ErrEmptyTitle
	}
	return Todo{Title: title}, nil
}

// TodoRepository の実装は外側（repository層）に置く。
type TodoRepository interface {
	FindAll() ([]Todo, error)
	Create(todo Todo) (Todo, error)
	Delete(id int) error
}
