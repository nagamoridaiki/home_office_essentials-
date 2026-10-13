package controller

import (
	"net/http"

	"backend/domain"
	"backend/usecase"
)

// customerResponse は自分の情報。未登録の項目は null、顧客の id は返さない
type customerResponse struct {
	Email          string  `json:"email"`
	FamilyName     *string `json:"family_name"`
	GivenName      *string `json:"given_name"`
	FamilyNameKana *string `json:"family_name_kana"`
	GivenNameKana  *string `json:"given_name_kana"`
	PhoneNumber    *string `json:"phone_number"`
}

// toCustomerResponse は POST /customers・POST /sessions でも使う
func toCustomerResponse(c domain.Customer) customerResponse {
	return customerResponse{
		Email:          c.Email.String(),
		FamilyName:     c.FamilyName,
		GivenName:      c.GivenName,
		FamilyNameKana: c.FamilyNameKana,
		GivenNameKana:  c.GivenNameKana,
		PhoneNumber:    c.PhoneNumber,
	}
}

type MeController struct {
	customers *usecase.CustomerUsecase
}

func NewMeController(customers *usecase.CustomerUsecase) *MeController {
	return &MeController{customers: customers}
}

// Register は /me 配下を登録する。requireCustomer で包むので、未ログインならハンドラまで来ない
func (c *MeController) Register(mux *http.ServeMux, requireCustomer func(http.Handler) http.Handler) {
	mux.Handle("GET /me", requireCustomer(http.HandlerFunc(c.get)))
}

func (c *MeController) get(w http.ResponseWriter, r *http.Request) {
	// 顧客 ID はコンテキストからだけ読む。リクエストの値で他人を指定させない
	id, ok := CustomerIDFromContext(r.Context())
	if !ok {
		writeUnauthenticated(w)
		return
	}
	customer, err := c.customers.GetMe(r.Context(), id)
	if err != nil {
		writeInternalError(w)
		return
	}
	writeJSON(w, http.StatusOK, toCustomerResponse(customer))
}
