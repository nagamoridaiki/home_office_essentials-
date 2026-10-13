package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrSessionNotFound  = errors.New("session not found")
	ErrSessionNotActive = errors.New("session not active")
)

// SessionLifetime はログインの有効期限。使っても延ばさない
const SessionLifetime = 30 * 24 * time.Hour

// SessionTokenHash は Cookie のトークンのハッシュ。作り方は外側が決め、ドメインは比べるだけ。
// 生のトークンはドメインに持ち込まない
type SessionTokenHash string

// CustomerSession は顧客のログイン状態。行は消さず、ログアウトは RevokedAt で表す
type CustomerSession struct {
	ID         int64
	CustomerID int64
	TokenHash  SessionTokenHash
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	CreatedAt  time.Time
}

// NewCustomerSession は ID が未採番のセッションを作る。ID はリポジトリが採番する
func NewCustomerSession(customerID int64, tokenHash SessionTokenHash, now time.Time) CustomerSession {
	return CustomerSession{
		CustomerID: customerID,
		TokenHash:  tokenHash,
		ExpiresAt:  now.Add(SessionLifetime),
		CreatedAt:  now,
	}
}

// IsActive はログアウトしておらず、期限前なら true
func (s CustomerSession) IsActive(now time.Time) bool {
	return s.RevokedAt == nil && now.Before(s.ExpiresAt)
}

// Revoke はログアウトする。有効でないセッションは ErrSessionNotActive。
// ログアウト済みを有効に戻す操作は作らない
func (s *CustomerSession) Revoke(now time.Time) error {
	if !s.IsActive(now) {
		return ErrSessionNotActive
	}
	s.RevokedAt = &now
	return nil
}

type CustomerSessionRepository interface {
	Create(ctx context.Context, s CustomerSession) (CustomerSession, error)
	// なければ ErrSessionNotFound
	FindByTokenHash(ctx context.Context, h SessionTokenHash) (CustomerSession, error)
	// ログアウトしていない行だけを更新する
	Revoke(ctx context.Context, id int64, at time.Time) error
}
