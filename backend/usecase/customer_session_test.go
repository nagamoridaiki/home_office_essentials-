package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"backend/domain"
	"backend/usecase"
)

// fakeSessionRepo は CustomerSessionRepository のフェイク。token_hash で引けるよう map に持つ
type fakeSessionRepo struct {
	sessions map[domain.SessionTokenHash]domain.CustomerSession
	nextID   int64
}

func newFakeSessionRepo() *fakeSessionRepo {
	return &fakeSessionRepo{sessions: map[domain.SessionTokenHash]domain.CustomerSession{}}
}

func (r *fakeSessionRepo) Create(_ context.Context, s domain.CustomerSession) (domain.CustomerSession, error) {
	r.nextID++
	s.ID = r.nextID
	r.sessions[s.TokenHash] = s
	return s, nil
}

func (r *fakeSessionRepo) FindByTokenHash(_ context.Context, h domain.SessionTokenHash) (domain.CustomerSession, error) {
	s, ok := r.sessions[h]
	if !ok {
		return domain.CustomerSession{}, domain.ErrSessionNotFound
	}
	return s, nil
}

func (r *fakeSessionRepo) Revoke(_ context.Context, id int64, at time.Time) error {
	for h, s := range r.sessions {
		if s.ID == id && s.RevokedAt == nil {
			s.RevokedAt = &at
			r.sessions[h] = s
			return nil
		}
	}
	return domain.ErrSessionNotActive
}

// fakeTokenGenerator は決まったトークンを返す。ハッシュは "hash:" を前に付けるだけ
type fakeTokenGenerator struct{ raw string }

func (g fakeTokenGenerator) Generate() (string, domain.SessionTokenHash, error) {
	return g.raw, g.Hash(g.raw), nil
}

func (g fakeTokenGenerator) Hash(raw string) domain.SessionTokenHash {
	return domain.SessionTokenHash("hash:" + raw)
}

var now = time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)

// clock は テストの中で時刻を進められる時刻関数
type clock struct{ t time.Time }

func (c *clock) now() time.Time { return c.t }

func newSessionUsecase() (*usecase.CustomerSessionUsecase, *fakeSessionRepo, *clock) {
	repo := newFakeSessionRepo()
	c := &clock{t: now}
	u := usecase.NewCustomerSessionUsecase(repo, fakeTokenGenerator{raw: "raw-token"}, c.now)
	return u, repo, c
}

func TestCustomerSessionUsecase_Issue(t *testing.T) {
	u, repo, _ := newSessionUsecase()

	raw, s, err := u.Issue(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if raw != "raw-token" {
		t.Errorf("raw = %q, want raw-token", raw)
	}
	if s.CustomerID != 7 {
		t.Errorf("CustomerID = %d, want 7", s.CustomerID)
	}

	t.Run("リポジトリにはハッシュだけが渡り、生のトークンは保存されない", func(t *testing.T) {
		for h := range repo.sessions {
			if string(h) == raw {
				t.Error("生のトークンが保存された")
			}
		}
		if _, ok := repo.sessions["hash:raw-token"]; !ok {
			t.Error("ハッシュで保存されていない")
		}
	})
}

func TestCustomerSessionUsecase_Authenticate(t *testing.T) {
	ctx := context.Background()

	t.Run("有効なトークンなら顧客 ID を返す", func(t *testing.T) {
		u, _, _ := newSessionUsecase()
		raw, _, _ := u.Issue(ctx, 7)
		id, err := u.Authenticate(ctx, raw)
		if err != nil {
			t.Fatal(err)
		}
		if id != 7 {
			t.Errorf("id = %d, want 7", id)
		}
	})

	tests := []struct {
		name  string
		setup func(u *usecase.CustomerSessionUsecase, repo *fakeSessionRepo, c *clock) string
	}{
		{"存在しないトークン", func(*usecase.CustomerSessionUsecase, *fakeSessionRepo, *clock) string {
			return "unknown"
		}},
		{"ログアウト済み", func(u *usecase.CustomerSessionUsecase, repo *fakeSessionRepo, _ *clock) string {
			raw, s, _ := u.Issue(ctx, 7)
			_ = repo.Revoke(ctx, s.ID, now)
			return raw
		}},
		{"期限切れ（30 日たった）", func(u *usecase.CustomerSessionUsecase, _ *fakeSessionRepo, c *clock) string {
			raw, _, _ := u.Issue(ctx, 7)
			c.t = now.Add(30 * 24 * time.Hour)
			return raw
		}},
		{"1 文字変えたトークン", func(u *usecase.CustomerSessionUsecase, _ *fakeSessionRepo, _ *clock) string {
			raw, _, _ := u.Issue(ctx, 7)
			return raw[:len(raw)-1] + "X"
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name+"は ErrUnauthenticated", func(t *testing.T) {
			u, repo, c := newSessionUsecase()
			raw := tt.setup(u, repo, c)
			if _, err := u.Authenticate(ctx, raw); !errors.Is(err, usecase.ErrUnauthenticated) {
				t.Errorf("err = %v, want ErrUnauthenticated", err)
			}
		})
	}
}
