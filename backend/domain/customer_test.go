package domain_test

import (
	"errors"
	"strings"
	"testing"

	"backend/domain"
)

func TestNewEmail(t *testing.T) {
	// 254 文字ちょうど（64 + 1 + 186 + 3）
	max := strings.Repeat("a", 64) + "@" + strings.Repeat("b", 186) + ".jp"

	ok := []struct {
		name string
		raw  string
		want string
	}{
		{"前後の空白を除き小文字にする", " Taro@Example.com ", "taro@example.com"},
		{"254 文字ちょうどは通る", max, max},
	}
	for _, tt := range ok {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.NewEmail(tt.raw)
			if err != nil {
				t.Fatalf("エラーになった: %v", err)
			}
			if got.String() != tt.want {
				t.Errorf("got %q, want %q", got.String(), tt.want)
			}
		})
	}

	ng := []struct {
		name string
		raw  string
	}{
		{"空文字", ""},
		{"@ がない", "x"},
		{"ドメインがない", "a@"},
		{"表示名付き", `"Taro" <a@b.jp>`},
		{"255 文字以上", "a" + max},
	}
	for _, tt := range ng {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := domain.NewEmail(tt.raw); !errors.Is(err, domain.ErrInvalidEmail) {
				t.Errorf("err = %v, want ErrInvalidEmail", err)
			}
		})
	}
}
