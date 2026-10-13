package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	"backend/domain"
)

// tokenBytes はトークンの乱数の長さ（32 バイト = 256 ビット）
const tokenBytes = 32

// SessionTokens は usecase.SessionTokenGenerator の実装
type SessionTokens struct{}

func NewSessionTokens() SessionTokens {
	return SessionTokens{}
}

// Generate は Cookie に入れる生のトークンと、DB に保存するハッシュを作る
func (SessionTokens) Generate() (string, domain.SessionTokenHash, error) {
	b := make([]byte, tokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	raw := base64.RawURLEncoding.EncodeToString(b)
	return raw, SessionTokens{}.Hash(raw), nil
}

// Hash は生のトークンを SHA-256 の 16 進文字列（64 文字）にする
func (SessionTokens) Hash(raw string) domain.SessionTokenHash {
	sum := sha256.Sum256([]byte(raw))
	return domain.SessionTokenHash(hex.EncodeToString(sum[:]))
}
