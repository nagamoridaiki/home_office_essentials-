package usecase_test

import (
	"context"
	"errors"
	"testing"

	"backend/domain"
	"backend/usecase"
)

type fakeCustomerRepo struct {
	customers map[int64]domain.Customer
}

func (r *fakeCustomerRepo) FindByID(_ context.Context, id int64) (domain.Customer, error) {
	c, ok := r.customers[id]
	if !ok {
		return domain.Customer{}, domain.ErrCustomerNotFound
	}
	return c, nil
}

func TestCustomerUsecase_GetMe(t *testing.T) {
	ctx := context.Background()
	email, err := domain.NewEmail("taro@example.com")
	if err != nil {
		t.Fatal(err)
	}
	repo := &fakeCustomerRepo{customers: map[int64]domain.Customer{
		1: {ID: 1, Email: email},
	}}
	u := usecase.NewCustomerUsecase(repo)

	t.Run("自分の顧客を返し、未登録の項目は nil のまま", func(t *testing.T) {
		c, err := u.GetMe(ctx, 1)
		if err != nil {
			t.Fatal(err)
		}
		if c.ID != 1 || c.Email.String() != "taro@example.com" {
			t.Errorf("got %+v", c)
		}
		for name, v := range map[string]*string{
			"FamilyName": c.FamilyName, "GivenName": c.GivenName,
			"FamilyNameKana": c.FamilyNameKana, "GivenNameKana": c.GivenNameKana,
			"PhoneNumber": c.PhoneNumber,
		} {
			if v != nil {
				t.Errorf("%s が nil でない: %q", name, *v)
			}
		}
	})

	t.Run("いない顧客 ID は ErrCustomerNotFound", func(t *testing.T) {
		if _, err := u.GetMe(ctx, 99); !errors.Is(err, domain.ErrCustomerNotFound) {
			t.Errorf("err = %v, want ErrCustomerNotFound", err)
		}
	})
}
