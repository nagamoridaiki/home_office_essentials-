package domain

import (
	"context"
	"errors"
)

var ErrEmptyTitle = errors.New("title is empty")

type Todo struct {
	// DB の id は bigint のため int64 で受ける
	ID        int64
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
// ctx はリクエストが中断されたときに問い合わせも止めるために渡す。
type TodoRepository interface {
	FindAll(ctx context.Context) ([]Todo, error)
	Create(ctx context.Context, todo Todo) (Todo, error)
	Delete(ctx context.Context, id int64) error
}
