package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend/domain"
	"backend/repository/internal/db"
)

// PostgresCustomerSessionRepository は PostgreSQL の customer_sessions を読み書きする
type PostgresCustomerSessionRepository struct {
	queries *db.Queries
}

func NewPostgresCustomerSessionRepository(pool *pgxpool.Pool) *PostgresCustomerSessionRepository {
	return &PostgresCustomerSessionRepository{queries: db.New(pool)}
}

func (r *PostgresCustomerSessionRepository) Create(ctx context.Context, s domain.CustomerSession) (domain.CustomerSession, error) {
	row, err := r.queries.CreateCustomerSession(ctx, db.CreateCustomerSessionParams{
		CustomerID: s.CustomerID,
		TokenHash:  string(s.TokenHash),
		ExpiresAt:  pgtype.Timestamptz{Time: s.ExpiresAt, Valid: true},
	})
	if err != nil {
		return domain.CustomerSession{}, err
	}
	return toDomainCustomerSession(row), nil
}

func (r *PostgresCustomerSessionRepository) FindByTokenHash(ctx context.Context, h domain.SessionTokenHash) (domain.CustomerSession, error) {
	row, err := r.queries.GetCustomerSessionByTokenHash(ctx, string(h))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.CustomerSession{}, domain.ErrSessionNotFound
	}
	if err != nil {
		return domain.CustomerSession{}, err
	}
	return toDomainCustomerSession(row), nil
}

// Revoke は revoked_at が空の行だけ埋める。行は消さない。0 行ならすでに有効でない
func (r *PostgresCustomerSessionRepository) Revoke(ctx context.Context, id int64, at time.Time) error {
	n, err := r.queries.RevokeCustomerSession(ctx, db.RevokeCustomerSessionParams{
		ID:        id,
		RevokedAt: pgtype.Timestamptz{Time: at, Valid: true},
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrSessionNotActive
	}
	return nil
}

// toDomainCustomerSession は生成コードの型をドメインのセッションに変換する
func toDomainCustomerSession(row db.CustomerSession) domain.CustomerSession {
	var revokedAt *time.Time
	if row.RevokedAt.Valid {
		revokedAt = &row.RevokedAt.Time
	}
	return domain.CustomerSession{
		ID:         row.ID,
		CustomerID: row.CustomerID,
		TokenHash:  domain.SessionTokenHash(row.TokenHash),
		ExpiresAt:  row.ExpiresAt.Time,
		RevokedAt:  revokedAt,
		CreatedAt:  row.CreatedAt.Time,
	}
}
