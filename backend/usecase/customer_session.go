package usecase

import (
	"context"
	"errors"
	"time"

	"backend/domain"
)

// ErrUnauthenticated はログイン状態を確かめられないとき。理由（ない・ログアウト済み・期限切れ）は区別しない
var ErrUnauthenticated = errors.New("unauthenticated")

// SessionTokenGenerator はトークンの作成とハッシュ化。実装は外側（backend/auth）に置く
type SessionTokenGenerator interface {
	Generate() (raw string, hash domain.SessionTokenHash, err error)
	Hash(raw string) domain.SessionTokenHash
}

type CustomerSessionUsecase struct {
	repo   domain.CustomerSessionRepository
	tokens SessionTokenGenerator
	// テストで時刻を固定するため、現在時刻を外から受け取る
	now func() time.Time
}

func NewCustomerSessionUsecase(repo domain.CustomerSessionRepository, tokens SessionTokenGenerator, now func() time.Time) *CustomerSessionUsecase {
	return &CustomerSessionUsecase{repo: repo, tokens: tokens, now: now}
}

// Issue はセッションを作り、Cookie に入れる生のトークンを返す。DB にはハッシュだけを保存する
func (u *CustomerSessionUsecase) Issue(ctx context.Context, customerID int64) (string, domain.CustomerSession, error) {
	raw, hash, err := u.tokens.Generate()
	if err != nil {
		return "", domain.CustomerSession{}, err
	}
	s, err := u.repo.Create(ctx, domain.NewCustomerSession(customerID, hash, u.now()))
	if err != nil {
		return "", domain.CustomerSession{}, err
	}
	return raw, s, nil
}

// Authenticate は生のトークンから顧客 ID を引く。有効なセッションがなければ ErrUnauthenticated
func (u *CustomerSessionUsecase) Authenticate(ctx context.Context, raw string) (int64, error) {
	s, err := u.repo.FindByTokenHash(ctx, u.tokens.Hash(raw))
	if errors.Is(err, domain.ErrSessionNotFound) {
		return 0, ErrUnauthenticated
	}
	if err != nil {
		return 0, err
	}
	if !s.IsActive(u.now()) {
		return 0, ErrUnauthenticated
	}
	return s.CustomerID, nil
}
