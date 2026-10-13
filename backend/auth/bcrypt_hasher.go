package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"

	"backend/usecase"
)

// bcryptCost は計算の重さ（research R1）
const bcryptCost = 12

// BcryptHasher は usecase.PasswordHasher の bcrypt による実装
type BcryptHasher struct {
	// dummyHash は DummyCompare で照合するためのハッシュ。起動時に 1 度だけ作る
	dummyHash []byte
}

func NewBcryptHasher() (*BcryptHasher, error) {
	dummy, err := bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), bcryptCost)
	if err != nil {
		return nil, err
	}
	return &BcryptHasher{dummyHash: dummy}, nil
}

func (h *BcryptHasher) Hash(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (h *BcryptHasher) Compare(hash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return usecase.ErrPasswordMismatch
	}
	return err
}

// DummyCompare は照合の時間だけかけて、結果は捨てる
func (h *BcryptHasher) DummyCompare(password string) {
	_ = bcrypt.CompareHashAndPassword(h.dummyHash, []byte(password))
}
