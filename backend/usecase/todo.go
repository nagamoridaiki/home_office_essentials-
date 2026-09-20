package usecase

import (
	"context"

	"backend/domain"
)

type TodoUsecase struct {
	repo domain.TodoRepository
}

func NewTodoUsecase(repo domain.TodoRepository) *TodoUsecase {
	return &TodoUsecase{repo: repo}
}

func (u *TodoUsecase) List(ctx context.Context) ([]domain.Todo, error) {
	return u.repo.FindAll(ctx)
}

func (u *TodoUsecase) Create(ctx context.Context, title string) (domain.Todo, error) {
	todo, err := domain.NewTodo(title)
	if err != nil {
		return domain.Todo{}, err
	}
	return u.repo.Create(ctx, todo)
}

func (u *TodoUsecase) Delete(ctx context.Context, id int64) error {
	return u.repo.Delete(ctx, id)
}
