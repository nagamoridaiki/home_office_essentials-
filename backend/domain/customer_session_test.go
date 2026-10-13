package domain_test

import (
	"errors"
	"testing"
	"time"

	"backend/domain"
)

var baseTime = time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)

func newSession() domain.CustomerSession {
	return domain.NewCustomerSession(1, domain.SessionTokenHash("hash"), baseTime)
}

func TestNewCustomerSession(t *testing.T) {
	s := newSession()
	if want := baseTime.Add(30 * 24 * time.Hour); !s.ExpiresAt.Equal(want) {
		t.Errorf("ExpiresAt = %v, want %v", s.ExpiresAt, want)
	}
	if s.RevokedAt != nil {
		t.Error("作った直後なのに RevokedAt が入っている")
	}
}

func TestCustomerSession_IsActive(t *testing.T) {
	expires := baseTime.Add(domain.SessionLifetime)
	tests := []struct {
		name string
		now  time.Time
		want bool
	}{
		{"期限の直前は有効", expires.Add(-time.Nanosecond), true},
		{"期限ちょうどは無効", expires, false},
		{"期限を過ぎたら無効", expires.Add(time.Second), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newSession()
			if got := s.IsActive(tt.now); got != tt.want {
				t.Errorf("IsActive = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("ログアウト済みは無効", func(t *testing.T) {
		s := newSession()
		if err := s.Revoke(baseTime.Add(time.Hour)); err != nil {
			t.Fatal(err)
		}
		if s.IsActive(baseTime.Add(2 * time.Hour)) {
			t.Error("ログアウト済みなのに有効")
		}
	})

	t.Run("ログアウト済みは時刻を戻しても無効のまま", func(t *testing.T) {
		s := newSession()
		if err := s.Revoke(baseTime.Add(time.Hour)); err != nil {
			t.Fatal(err)
		}
		if s.IsActive(baseTime) {
			t.Error("ログアウト前の時刻で有効に戻った")
		}
	})
}

func TestCustomerSession_Revoke(t *testing.T) {
	t.Run("有効なセッションはログアウトできる", func(t *testing.T) {
		s := newSession()
		at := baseTime.Add(time.Hour)
		if err := s.Revoke(at); err != nil {
			t.Fatal(err)
		}
		if s.RevokedAt == nil || !s.RevokedAt.Equal(at) {
			t.Errorf("RevokedAt = %v, want %v", s.RevokedAt, at)
		}
	})

	t.Run("ログアウト済みはもう一度ログアウトできない", func(t *testing.T) {
		s := newSession()
		first := baseTime.Add(time.Hour)
		if err := s.Revoke(first); err != nil {
			t.Fatal(err)
		}
		if err := s.Revoke(baseTime.Add(2 * time.Hour)); !errors.Is(err, domain.ErrSessionNotActive) {
			t.Errorf("err = %v, want ErrSessionNotActive", err)
		}
		if !s.RevokedAt.Equal(first) {
			t.Errorf("RevokedAt が変わった: %v", s.RevokedAt)
		}
	})

	t.Run("期限切れはログアウトできない", func(t *testing.T) {
		s := newSession()
		if err := s.Revoke(s.ExpiresAt); !errors.Is(err, domain.ErrSessionNotActive) {
			t.Errorf("err = %v, want ErrSessionNotActive", err)
		}
		if s.RevokedAt != nil {
			t.Errorf("RevokedAt が入った: %v", s.RevokedAt)
		}
	})
}
