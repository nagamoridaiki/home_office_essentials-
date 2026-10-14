package usecase

import (
	"context"
	"errors"
	"log"
	"strings"
	"unicode/utf8"

	"backend/domain"
)

var (
	ErrPasswordTooShort = errors.New("password too short")
	ErrPasswordTooLong  = errors.New("password too long")
	// ErrPasswordMismatch はパスワードがハッシュと一致しないとき
	ErrPasswordMismatch = errors.New("password mismatch")
	// ErrInvalidCredentials はログインできないとき。未登録かパスワード違いかは区別しない
	ErrInvalidCredentials = errors.New("invalid credentials")
)

const (
	passwordMinChars = 8
	// bcrypt が扱えるのは 72 バイトまで
	passwordMaxBytes = 72
)

// ValidatePassword はパスワードが 8 文字（rune）以上、72 バイト以下かを確かめる
func ValidatePassword(raw string) error {
	if utf8.RuneCountInString(raw) < passwordMinChars {
		return ErrPasswordTooShort
	}
	if len(raw) > passwordMaxBytes {
		return ErrPasswordTooLong
	}
	return nil
}

// 入力の検証に失敗した理由
const (
	ReasonRequired     = "required"
	ReasonInvalidEmail = "invalid_email"
	ReasonTooShort     = "too_short"
	ReasonTooLong      = "too_long"
)

// FieldError は項目名と理由の組
type FieldError struct {
	Field string
	Code  string
}

// ValidationError は入力の検証エラー。問題のある項目をまとめて持つ
type ValidationError struct {
	Fields []FieldError
}

func (e *ValidationError) Error() string {
	return "validation failed"
}

// PasswordHasher はパスワードのハッシュ化と照合。実装は外側（backend/auth）に置く
type PasswordHasher interface {
	Hash(password string) (string, error)
	// Compare は一致しなければ ErrPasswordMismatch
	Compare(hash, password string) error
	// DummyCompare は未登録のときも照合の時間をかけ、応答時間で登録の有無を悟らせない
	DummyCompare(password string)
}

// PasswordCredentialStore は customers と customer_credentials をまたいで保存する
type PasswordCredentialStore interface {
	// CreateCustomerWithPassword は 1 トランザクションで作る。メールアドレスの重複は domain.ErrEmailAlreadyUsed
	CreateCustomerWithPassword(ctx context.Context, email domain.Email, passwordHash string) (domain.Customer, error)
	// FindPasswordHashByEmail は顧客 ID とハッシュを返す。なければ domain.ErrCustomerNotFound
	FindPasswordHashByEmail(ctx context.Context, email domain.Email) (customerID int64, passwordHash string, err error)
}

// PasswordAuthUsecase はメールアドレスとパスワードで顧客を確かめる
type PasswordAuthUsecase struct {
	store    PasswordCredentialStore
	hasher   PasswordHasher
	notifier CustomerNotifier
}

func NewPasswordAuthUsecase(store PasswordCredentialStore, hasher PasswordHasher, notifier CustomerNotifier) *PasswordAuthUsecase {
	return &PasswordAuthUsecase{store: store, hasher: hasher, notifier: notifier}
}

// SignUp はアカウントを作る。セッションの発行は呼び出し側（ハンドラ）が続けて行う
func (u *PasswordAuthUsecase) SignUp(ctx context.Context, rawEmail, password string) (domain.Customer, error) {
	email, err := validateSignUp(rawEmail, password)
	if err != nil {
		return domain.Customer{}, err
	}

	hash, err := u.hasher.Hash(password)
	if err != nil {
		return domain.Customer{}, err
	}
	c, err := u.store.CreateCustomerWithPassword(ctx, email, hash)
	if err != nil {
		return domain.Customer{}, err
	}

	// 通知に失敗してもアカウントはできているので、登録は成功にする
	if err := u.notifier.NotifyRegistered(ctx, c); err != nil {
		log.Printf("登録の通知に失敗しました（customer_id=%d）: %v", c.ID, err)
	}
	return c, nil
}

// Login はメールアドレスとパスワードを確かめ、顧客 ID を返す。セッションの発行は呼び出し側（ハンドラ）が行う
func (u *PasswordAuthUsecase) Login(ctx context.Context, rawEmail, password string) (int64, error) {
	email, err := domain.NewEmail(rawEmail)
	if err != nil {
		// 形式不正でも照合の時間をかけ、応答時間で理由を悟らせない
		u.hasher.DummyCompare(password)
		return 0, ErrInvalidCredentials
	}

	customerID, hash, err := u.store.FindPasswordHashByEmail(ctx, email)
	if errors.Is(err, domain.ErrCustomerNotFound) {
		u.hasher.DummyCompare(password)
		return 0, ErrInvalidCredentials
	}
	if err != nil {
		return 0, err
	}

	if err := u.hasher.Compare(hash, password); err != nil {
		if errors.Is(err, ErrPasswordMismatch) {
			return 0, ErrInvalidCredentials
		}
		return 0, err
	}
	return customerID, nil
}

// validateSignUp はメールアドレスとパスワードを両方確かめ、問題をまとめて返す
func validateSignUp(rawEmail, password string) (domain.Email, error) {
	var fields []FieldError

	email, err := domain.NewEmail(rawEmail)
	switch {
	case strings.TrimSpace(rawEmail) == "":
		fields = append(fields, FieldError{Field: "email", Code: ReasonRequired})
	case err != nil:
		fields = append(fields, FieldError{Field: "email", Code: ReasonInvalidEmail})
	}

	switch err := ValidatePassword(password); {
	case password == "":
		fields = append(fields, FieldError{Field: "password", Code: ReasonRequired})
	case errors.Is(err, ErrPasswordTooShort):
		fields = append(fields, FieldError{Field: "password", Code: ReasonTooShort})
	case errors.Is(err, ErrPasswordTooLong):
		fields = append(fields, FieldError{Field: "password", Code: ReasonTooLong})
	}

	if len(fields) > 0 {
		return domain.Email{}, &ValidationError{Fields: fields}
	}
	return email, nil
}
