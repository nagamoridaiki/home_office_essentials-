package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"backend/domain"
	"backend/repository/internal/db"
)

// PostgresTodoRepository は PostgreSQL に Todo を保存する。
// sqlc が生成した型はこのファイルの中だけで扱い、ドメインより内側には出さない。
type PostgresTodoRepository struct {
	queries *db.Queries
}

func NewPostgresTodoRepository(pool *pgxpool.Pool) *PostgresTodoRepository {
	return &PostgresTodoRepository{queries: db.New(pool)}
}

func (r *PostgresTodoRepository) FindAll(ctx context.Context) ([]domain.Todo, error) {
	rows, err := r.queries.ListTodos(ctx)
	if err != nil {
		return nil, err
	}
	todos := make([]domain.Todo, 0, len(rows))
	for _, row := range rows {
		todos = append(todos, toDomain(row))
	}
	return todos, nil
}

func (r *PostgresTodoRepository) Create(ctx context.Context, todo domain.Todo) (domain.Todo, error) {
	row, err := r.queries.CreateTodo(ctx, todo.Title)
	if err != nil {
		return domain.Todo{}, err
	}
	return toDomain(row), nil
}

func (r *PostgresTodoRepository) Delete(ctx context.Context, id int64) error {
	return r.queries.DeleteTodo(ctx, id)
}

// toDomain は生成コードの型をドメインのエンティティに変換する。
// created_at / updated_at はドメインで使っていないため渡さない。
func toDomain(row db.Todo) domain.Todo {
	return domain.Todo{
		ID:        row.ID,
		Title:     row.Title,
		Completed: row.Completed,
	}
}
