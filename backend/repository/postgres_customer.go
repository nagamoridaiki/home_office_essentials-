package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend/domain"
	"backend/repository/internal/db"
)

// PostgresCustomerRepository は PostgreSQL の customers を読み書きする
type PostgresCustomerRepository struct {
	queries *db.Queries
}

func NewPostgresCustomerRepository(pool *pgxpool.Pool) *PostgresCustomerRepository {
	return &PostgresCustomerRepository{queries: db.New(pool)}
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
