package usecase_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"backend/domain"
	"backend/usecase"
)

// fakeHasher は "hashed:" を前に付けるだけのハッシュ
type fakeHasher struct{}

func (fakeHasher) Hash(password string) (string, error) { return "hashed:" + password, nil }

func (fakeHasher) Compare(hash, password string) error {
	if hash != "hashed:"+password {
		return usecase.ErrPasswordMismatch
	}
	return nil
}

func (fakeHasher) DummyCompare(string) {}

// fakeCredentialStore は渡された値を覚えておくストア
type fakeCredentialStore struct {
	called    bool
	gotEmail  domain.Email
	gotHash   string
	createErr error
}

func (s *fakeCredentialStore) CreateCustomerWithPassword(_ context.Context, email domain.Email, passwordHash string) (domain.Customer, error) {
	s.called = true
	s.gotEmail = email
	s.gotHash = passwordHash
	if s.createErr != nil {
		return domain.Customer{}, s.createErr
	}
	return domain.Customer{ID: 1, Email: email}, nil
}

type fakeNotifier struct {
	err    error
	called bool
}

func (n *fakeNotifier) NotifyRegistered(context.Context, domain.Customer) error {
	n.called = true
	return n.err
}

func TestPasswordAuthUsecase_SignUp(t *testing.T) {
	ctx := context.Background()

	t.Run("正しい入力で顧客が作られ、メールアドレスが小文字にそろう", func(t *testing.T) {
		store, notifier := &fakeCredentialStore{}, &fakeNotifier{}
		u := usecase.NewPasswordAuthUsecase(store, fakeHasher{}, notifier)

		c, err := u.SignUp(ctx, " Taro@Example.com ", "correct-horse")

		if err != nil {
			t.Fatalf("SignUp() error = %v", err)
		}
		if c.Email.String() != "taro@example.com" {
			t.Errorf("email = %q", c.Email.String())
		}
		if !notifier.called {
			t.Error("通知の口が呼ばれていない")
		}
	})

	t.Run("ストアに渡るのは生のパスワードではなくハッシュ", func(t *testing.T) {
		store := &fakeCredentialStore{}
		u := usecase.NewPasswordAuthUsecase(store, fakeHasher{}, &fakeNotifier{})

		if _, err := u.SignUp(ctx, "taro@example.com", "correct-horse"); err != nil {
			t.Fatalf("SignUp() error = %v", err)
		}
		if store.gotHash != "hashed:correct-horse" {
			t.Errorf("passwordHash = %q", store.gotHash)
		}
	})

	tests := []struct {
		name     string
		email    string
		password string
		want     []usecase.FieldError
	}{
		{
			name:     "メールアドレスの形式不正とパスワード 7 文字は両方の理由が返る",
			email:    "x",
			password: "1234567",
			want: []usecase.FieldError{
				{Field: "email", Code: usecase.ReasonInvalidEmail},
				{Field: "password", Code: usecase.ReasonTooShort},
			},
		},
		{
			name:     "空のメールアドレスとパスワードは required",
			email:    "  ",
			password: "",
			want: []usecase.FieldError{
				{Field: "email", Code: usecase.ReasonRequired},
				{Field: "password", Code: usecase.ReasonRequired},
			},
		},
		{
			name:     "パスワード 73 バイトは too_long",
			email:    "taro@example.com",
			password: strings.Repeat("a", 73),
			want:     []usecase.FieldError{{Field: "password", Code: usecase.ReasonTooLong}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeCredentialStore{}
			u := usecase.NewPasswordAuthUsecase(store, fakeHasher{}, &fakeNotifier{})

			_, err := u.SignUp(ctx, tt.email, tt.password)

			var ve *usecase.ValidationError
			if !errors.As(err, &ve) {
				t.Fatalf("err = %v, want ValidationError", err)
			}
			if !reflect.DeepEqual(ve.Fields, tt.want) {
				t.Errorf("Fields = %v, want %v", ve.Fields, tt.want)
			}
			if store.called {
				t.Error("検証エラーなのにストアが呼ばれた")
			}
		})
	}

	t.Run("ストアが ErrEmailAlreadyUsed を返すとそのまま伝わる", func(t *testing.T) {
		store := &fakeCredentialStore{createErr: domain.ErrEmailAlreadyUsed}
		u := usecase.NewPasswordAuthUsecase(store, fakeHasher{}, &fakeNotifier{})

		_, err := u.SignUp(ctx, "taro@example.com", "correct-horse")

		if !errors.Is(err, domain.ErrEmailAlreadyUsed) {
			t.Errorf("err = %v, want ErrEmailAlreadyUsed", err)
		}
	})

	t.Run("通知が失敗しても SignUp は成功する", func(t *testing.T) {
		u := usecase.NewPasswordAuthUsecase(&fakeCredentialStore{}, fakeHasher{}, &fakeNotifier{err: errors.New("smtp down")})

		if _, err := u.SignUp(ctx, "taro@example.com", "correct-horse"); err != nil {
			t.Errorf("SignUp() error = %v, want nil", err)
		}
	})
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     error
	}{
		{"8 文字ちょうどは通る", "12345678", nil},
		{"全角 8 文字（24 バイト）は通る", "あいうえおかきく", nil},
		{"7 文字は短すぎる", "1234567", usecase.ErrPasswordTooShort},
		{"72 バイトちょうどは通る", strings.Repeat("a", 72), nil},
		{"73 バイトは長すぎる", strings.Repeat("a", 73), usecase.ErrPasswordTooLong},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := usecase.ValidatePassword(tt.password); !errors.Is(err, tt.want) {
				t.Errorf("ValidatePassword() = %v, want %v", err, tt.want)
			}
		})
	}
}
