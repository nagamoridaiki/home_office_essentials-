package usecase

import "backend/domain"

type TodoUsecase struct {
	repo domain.TodoRepository
}

func NewTodoUsecase(repo domain.TodoRepository) *TodoUsecase {
	return &TodoUsecase{repo: repo}
}

func (u *TodoUsecase) List() ([]domain.Todo, error) {
	return u.repo.FindAll()
}

func (u *TodoUsecase) Create(title string) (domain.Todo, error) {
	todo, err := domain.NewTodo(title)
	if err != nil {
		return domain.Todo{}, err
	}
	return u.repo.Create(todo)
}

func (u *TodoUsecase) Delete(id int) error {
	return u.repo.Delete(id)
}
