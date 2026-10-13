package notification

import (
	"context"

	"backend/domain"
)

// NoopCustomerNotifier は何も送らない通知。確認メールの送信は口だけ用意している
type NoopCustomerNotifier struct{}

func (NoopCustomerNotifier) NotifyRegistered(context.Context, domain.Customer) error {
	return nil
}
