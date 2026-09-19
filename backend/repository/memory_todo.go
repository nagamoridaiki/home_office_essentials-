package repository

import (
	"sync"

	"backend/domain"
)

// MemoryTodoRepository はメモリ上にTodoを保持する。再起動でデータは消える。
type MemoryTodoRepository struct {
	mu     sync.Mutex
	todos  []domain.Todo
	nextID int
}

func NewMemoryTodoRepository() *MemoryTodoRepository {
	return &MemoryTodoRepository{nextID: 1}
}

func (r *MemoryTodoRepository) FindAll() ([]domain.Todo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	// 呼び出し側が内部のスライスを書き換えないようコピーを返す
	todos := make([]domain.Todo, len(r.todos))
	copy(todos, r.todos)
	return todos, nil
}

func (r *MemoryTodoRepository) Create(todo domain.Todo) (domain.Todo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	todo.ID = r.nextID
	r.nextID++
	r.todos = append(r.todos, todo)
	return todo, nil
}

func (r *MemoryTodoRepository) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	remaining := make([]domain.Todo, 0, len(r.todos))
	for _, t := range r.todos {
		if t.ID != id {
			remaining = append(remaining, t)
		}
	}
	r.todos = remaining
	return nil
}
