package usecase

import (
	"context"

	"backend/domain"
)

// CustomerNotifier は顧客へのお知らせ。実装は外側（backend/notification）に置く
type CustomerNotifier interface {
	NotifyRegistered(ctx context.Context, c domain.Customer) error
}

type CustomerUsecase struct {
	repo domain.CustomerRepository
}

func NewCustomerUsecase(repo domain.CustomerRepository) *CustomerUsecase {
	return &CustomerUsecase{repo: repo}
}

// GetMe はログイン中の顧客自身の情報を返す
func (u *CustomerUsecase) GetMe(ctx context.Context, customerID int64) (domain.Customer, error) {
	return u.repo.FindByID(ctx, customerID)
}
