package domain

import (
	"context"
	"errors"
	"net/mail"
	"strings"
)

var (
	ErrInvalidEmail     = errors.New("invalid email")
	ErrEmailAlreadyUsed = errors.New("email already used")
	ErrCustomerNotFound = errors.New("customer not found")
)

// メールアドレスの最大長（RFC 5321 の上限）
const maxEmailLength = 254

// Email は小文字にそろえたメールアドレス。NewEmail を通した値だけを持つ
type Email struct {
	value string
}

// NewEmail は前後の空白を取り除いて小文字にし、形式を確かめる
func NewEmail(raw string) (Email, error) {
	v := strings.ToLower(strings.TrimSpace(raw))
	if v == "" || len(v) > maxEmailLength {
		return Email{}, ErrInvalidEmail
	}
	// 表示名付き（"Taro" <a@b>）や余分な文字を含む入力は、解釈結果が入力と一致しないので弾ける
	addr, err := mail.ParseAddress(v)
	if err != nil || addr.Name != "" || addr.Address != v {
		return Email{}, ErrInvalidEmail
	}
	return Email{value: v}, nil
}

func (e Email) String() string {
	return e.value
}

// Customer は顧客。氏名・フリガナ・電話番号は未登録なら nil。
// 作成はパスワードの登録と一体で行うため、生成関数は持たない
type Customer struct {
	ID             int64
	Email          Email
	FamilyName     *string
	GivenName      *string
	FamilyNameKana *string
	GivenNameKana  *string
	PhoneNumber    *string
}

type CustomerRepository interface {
	// なければ ErrCustomerNotFound
	FindByID(ctx context.Context, id int64) (Customer, error)
}
