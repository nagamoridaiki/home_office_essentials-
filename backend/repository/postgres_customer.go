package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend/domain"
	"backend/repository/internal/db"
)

// PostgresCustomerRepository は PostgreSQL の customers を読み書きする
type PostgresCustomerRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewPostgresCustomerRepository(pool *pgxpool.Pool) *PostgresCustomerRepository {
	return &PostgresCustomerRepository{pool: pool, queries: db.New(pool)}
}

func (r *PostgresCustomerRepository) FindByID(ctx context.Context, id int64) (domain.Customer, error) {
	row, err := r.queries.GetCustomerByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Customer{}, domain.ErrCustomerNotFound
	}
	if err != nil {
		return domain.Customer{}, err
	}
	return toDomainCustomer(row)
}

// uniqueViolation は PostgreSQL の UNIQUE 制約違反のエラーコード
const uniqueViolation = "23505"

// CreateCustomerWithPassword は顧客と認証情報を 1 つのトランザクションで作る。
// 重複の確認は事前の SELECT ではなく UNIQUE 制約に任せる（同時に届いても 1 件に絞れる）
func (r *PostgresCustomerRepository) CreateCustomerWithPassword(ctx context.Context, email domain.Email, passwordHash string) (domain.Customer, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Customer{}, err
	}
	// Commit の後に呼んでも何もしない。途中で失敗したときだけ取り消される
	defer tx.Rollback(ctx)

	q := r.queries.WithTx(tx)
	row, err := q.CreateCustomer(ctx, email.String())
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
		return domain.Customer{}, domain.ErrEmailAlreadyUsed
	}
	if err != nil {
		return domain.Customer{}, err
	}

	err = q.CreateCustomerCredential(ctx, db.CreateCustomerCredentialParams{
		CustomerID:   row.ID,
		PasswordHash: pgtype.Text{String: passwordHash, Valid: true},
	})
	if err != nil {
		return domain.Customer{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Customer{}, err
	}
	return toDomainCustomer(row)
}

// toDomainCustomer は生成コードの型をドメインの顧客に変換する
func toDomainCustomer(row db.Customer) (domain.Customer, error) {
	// Email のフィールドは外から入れられないので、NewEmail を通して作る
	email, err := domain.NewEmail(row.Email)
	if err != nil {
		return domain.Customer{}, err
	}
	return domain.Customer{
		ID:             row.ID,
		Email:          email,
		FamilyName:     textToPtr(row.FamilyName),
		GivenName:      textToPtr(row.GivenName),
		FamilyNameKana: textToPtr(row.FamilyNameKana),
		GivenNameKana:  textToPtr(row.GivenNameKana),
		PhoneNumber:    textToPtr(row.PhoneNumber),
	}, nil
}

// textToPtr は NULL を nil に変える
func textToPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}
