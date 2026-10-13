package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"backend/auth"
	"backend/controller"
	"backend/notification"
	"backend/repository"
	"backend/usecase"
)

func main() {
	pool, err := newDBPool(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	todoRepo := repository.NewPostgresTodoRepository(pool)
	todoUsecase := usecase.NewTodoUsecase(todoRepo)
	todoController := controller.NewTodoController(todoUsecase)

	hasher, err := auth.NewBcryptHasher()
	if err != nil {
		log.Fatal(err)
	}
	customerRepo := repository.NewPostgresCustomerRepository(pool)
	passwordAuthUsecase := usecase.NewPasswordAuthUsecase(customerRepo, hasher, notification.NoopCustomerNotifier{})
	sessionUsecase := usecase.NewCustomerSessionUsecase(
		repository.NewPostgresCustomerSessionRepository(pool), auth.NewSessionTokens(), time.Now)
	customerUsecase := usecase.NewCustomerUsecase(customerRepo)
	requireCustomer := controller.RequireCustomer(controller.NewCookieSessionAuthenticator(sessionUsecase))
	meController := controller.NewMeController(customerUsecase)
	// SESSION_COOKIE_SECURE が false のときだけ Secure を外す（http の localhost で動かすため）
	sessionCookie := controller.NewSessionCookie(os.Getenv("SESSION_COOKIE_SECURE") != "false")
	customerAuthController := controller.NewCustomerAuthController(passwordAuthUsecase, sessionUsecase, sessionCookie)

	mux := http.NewServeMux()
	todoController.Register(mux)
	meController.Register(mux, requireCustomer)
	customerAuthController.Register(mux)

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
