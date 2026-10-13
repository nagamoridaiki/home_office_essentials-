package controller

import (
	"context"
	"errors"
	"net/http"

	"backend/usecase"
)

// CustomerAuthenticator はリクエストから顧客 ID を取り出す。
// 特定できなければ usecase.ErrUnauthenticated。認証方式を変えるときはこの実装だけを差し替える
type CustomerAuthenticator interface {
	AuthenticateRequest(r *http.Request) (customerID int64, err error)
}

// CookieSessionAuthenticator は Cookie customer_session のセッションで認証する
type CookieSessionAuthenticator struct {
	sessions *usecase.CustomerSessionUsecase
}

func NewCookieSessionAuthenticator(sessions *usecase.CustomerSessionUsecase) *CookieSessionAuthenticator {
	return &CookieSessionAuthenticator{sessions: sessions}
}

func (a *CookieSessionAuthenticator) AuthenticateRequest(r *http.Request) (int64, error) {
	c, err := r.Cookie(sessionCookieName)
	if err != nil || c.Value == "" {
		return 0, usecase.ErrUnauthenticated
	}
	return a.sessions.Authenticate(r.Context(), c.Value)
}

// customerIDKey はコンテキストに顧客 ID を入れるときのキー。他のパッケージのキーとぶつからないよう専用の型にする
type customerIDKey struct{}

// RequireCustomer はログイン中の顧客だけを次のハンドラへ通す。通すときは顧客 ID をコンテキストに入れる
func RequireCustomer(a CustomerAuthenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, err := a.AuthenticateRequest(r)
			if errors.Is(err, usecase.ErrUnauthenticated) {
				writeUnauthenticated(w)
				return
			}
			if err != nil {
				writeInternalError(w)
				return
			}
			ctx := context.WithValue(r.Context(), customerIDKey{}, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// CustomerIDFromContext は RequireCustomer が入れた顧客 ID を読む
func CustomerIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(customerIDKey{}).(int64)
	return id, ok
}
