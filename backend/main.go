package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"backend/controller"
	"backend/repository"
	"backend/usecase"
)

func main() {
	pool, err := newDBPool(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	todoRepo := repository.NewMemoryTodoRepository()
	todoUsecase := usecase.NewTodoUsecase(todoRepo)
	todoController := controller.NewTodoController(todoUsecase)

	mux := http.NewServeMux()
	todoController.Register(mux)

	fmt.Println("Backend running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", controller.EnableCORS(mux)))
}

// newDBPool は DATABASE_URL から接続プールを作る。
// pgxpool.New はこの時点ではまだ接続しないため、Ping で起動時に疎通を確かめ、
// つながらなければ起動を失敗させる（あとのリクエストまで異常に気づけないのを防ぐ）。
func newDBPool(ctx context.Context) (*pgxpool.Pool, error) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return nil, errors.New("環境変数 DATABASE_URL が設定されていません")
	}

	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("DB プールの作成に失敗しました: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("DB への接続に失敗しました: %w", err)
	}

	return pool, nil
}
