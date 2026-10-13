package auth_test

import (
	"encoding/base64"
	"testing"

	"backend/auth"
)

func TestSessionTokens_Generate(t *testing.T) {
	tokens := auth.NewSessionTokens()

	raw, hash, err := tokens.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	t.Run("生のトークンは 32 バイトの base64url", func(t *testing.T) {
		b, err := base64.RawURLEncoding.DecodeString(raw)
		if err != nil || len(b) != 32 {
			t.Errorf("raw = %q, len = %d, err = %v", raw, len(b), err)
		}
	})
	t.Run("ハッシュは 64 文字で Hash の結果と同じ", func(t *testing.T) {
		if len(hash) != 64 || hash != tokens.Hash(raw) {
			t.Errorf("hash = %q", hash)
		}
	})
	t.Run("呼ぶたびに違うトークンになる", func(t *testing.T) {
		raw2, _, _ := tokens.Generate()
		if raw == raw2 {
			t.Error("同じトークンが作られた")
		}
	})
}
