package controller_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/controller"
	"backend/usecase"
)

// fakeAuthenticator は決まった結果を返す CustomerAuthenticator
type fakeAuthenticator struct {
	customerID int64
	err        error
}

func (a fakeAuthenticator) AuthenticateRequest(r *http.Request) (int64, error) {
	return a.customerID, a.err
}

func TestRequireCustomer(t *testing.T) {
	t.Run("認証に成功すると次のハンドラでコンテキストから顧客 ID が読める", func(t *testing.T) {
		var gotID int64
		var gotOK bool
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotID, gotOK = controller.CustomerIDFromContext(r.Context())
		})

		rec := httptest.NewRecorder()
		controller.RequireCustomer(fakeAuthenticator{customerID: 42})(next).
			ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/me", nil))

		if !gotOK || gotID != 42 {
			t.Errorf("CustomerIDFromContext() = (%d, %v), want (42, true)", gotID, gotOK)
		}
	})

	t.Run("認証に失敗すると 401 unauthenticated で次のハンドラを呼ばない", func(t *testing.T) {
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })

		rec := httptest.NewRecorder()
		controller.RequireCustomer(fakeAuthenticator{err: usecase.ErrUnauthenticated})(next).
			ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/me", nil))

		if called {
			t.Error("次のハンドラが呼ばれた")
		}
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401", rec.Code)
		}
		var body struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil || body.Error.Code != "unauthenticated" {
			t.Errorf("error.code = %q, err = %v", body.Error.Code, err)
		}
	})

	t.Run("認証の途中で想定外のエラーが起きると 500 で次のハンドラを呼ばない", func(t *testing.T) {
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })

		rec := httptest.NewRecorder()
		controller.RequireCustomer(fakeAuthenticator{err: errors.New("db down")})(next).
			ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/me", nil))

		if called || rec.Code != http.StatusInternalServerError {
			t.Errorf("called = %v, status = %d, want false, 500", called, rec.Code)
		}
	})
}

func TestCookieSessionAuthenticator(t *testing.T) {
	t.Run("Cookie がないとき ErrUnauthenticated", func(t *testing.T) {
		// Cookie がなければユースケースまで行かないので、リポジトリなどは渡さない
		a := controller.NewCookieSessionAuthenticator(usecase.NewCustomerSessionUsecase(nil, nil, time.Now))

		_, err := a.AuthenticateRequest(httptest.NewRequest(http.MethodGet, "/me", nil))

		if !errors.Is(err, usecase.ErrUnauthenticated) {
			t.Errorf("err = %v, want ErrUnauthenticated", err)
		}
	})
}
