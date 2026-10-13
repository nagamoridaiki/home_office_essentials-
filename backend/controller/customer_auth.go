package controller

import (
	"errors"
	"net/http"

	"backend/domain"
	"backend/usecase"
)

// fieldErrorMessages は検証エラーの項目と理由ごとの文言（contracts の details 一覧）
var fieldErrorMessages = map[usecase.FieldError]string{
	{Field: "email", Code: usecase.ReasonRequired}:     "メールアドレスを入力してください",
	{Field: "email", Code: usecase.ReasonInvalidEmail}: "メールアドレスの形式が正しくありません",
	{Field: "password", Code: usecase.ReasonRequired}:  "パスワードを入力してください",
	{Field: "password", Code: usecase.ReasonTooShort}:  "パスワードは 8 文字以上で入力してください",
	{Field: "password", Code: usecase.ReasonTooLong}:   "パスワードが長すぎます",
}

// CustomerAuthController はアカウント作成・ログイン・ログアウト
type CustomerAuthController struct {
	passwordAuth *usecase.PasswordAuthUsecase
	sessions     *usecase.CustomerSessionUsecase
	cookie       SessionCookie
}

func NewCustomerAuthController(passwordAuth *usecase.PasswordAuthUsecase, sessions *usecase.CustomerSessionUsecase, cookie SessionCookie) *CustomerAuthController {
	return &CustomerAuthController{passwordAuth: passwordAuth, sessions: sessions, cookie: cookie}
}

func (c *CustomerAuthController) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /customers", c.signUp)
}

type credentialsRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// signUp はアカウントを作り、そのままログインした状態にする
func (c *CustomerAuthController) signUp(w http.ResponseWriter, r *http.Request) {
	var req credentialsRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	customer, err := c.passwordAuth.SignUp(r.Context(), req.Email, req.Password)
	var ve *usecase.ValidationError
	switch {
	case errors.As(err, &ve):
		writeValidationError(w, ve)
		return
	case errors.Is(err, domain.ErrEmailAlreadyUsed):
		writeError(w, http.StatusConflict, codeEmailAlreadyUsed, "このメールアドレスはすでに使われています", nil)
		return
	case err != nil:
		writeInternalError(w)
		return
	}

	raw, _, err := c.sessions.Issue(r.Context(), customer.ID)
	if err != nil {
		writeInternalError(w)
		return
	}
	c.cookie.Set(w, raw)
	writeJSON(w, http.StatusCreated, toCustomerResponse(customer))
}

func writeValidationError(w http.ResponseWriter, ve *usecase.ValidationError) {
	details := make([]errorDetail, 0, len(ve.Fields))
	for _, f := range ve.Fields {
		details = append(details, errorDetail{Field: f.Field, Code: f.Code, Message: fieldErrorMessages[f]})
	}
	writeError(w, http.StatusUnprocessableEntity, codeValidationFailed, "入力内容を確認してください", details)
}
