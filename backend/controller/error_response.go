package controller

import (
	"encoding/json"
	"mime"
	"net/http"
)

// エラーの code（docs/design/04 と contracts の一覧）
const (
	codeInvalidRequest       = "invalid_request"
	codeUnsupportedMediaType = "unsupported_media_type"
	codeValidationFailed     = "validation_failed"
	codeEmailAlreadyUsed     = "email_already_used"
	codeInvalidCredentials   = "invalid_credentials"
	codeUnauthenticated      = "unauthenticated"
	codeInternalError        = "internal_error"
)

// errorDetail は検証エラーの項目ごとの理由
type errorDetail struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type errorBody struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []errorDetail `json:"details"`
}

// writeError は {"error":{"code","message","details"}} を返す。details が nil なら空配列にする
func writeError(w http.ResponseWriter, status int, code, message string, details []errorDetail) {
	if details == nil {
		details = []errorDetail{}
	}
	writeJSON(w, status, map[string]errorBody{
		"error": {Code: code, Message: message, Details: details},
	})
}

func writeUnauthenticated(w http.ResponseWriter) {
	writeError(w, http.StatusUnauthorized, codeUnauthenticated, "ログインしてください", nil)
}

func writeInternalError(w http.ResponseWriter) {
	writeError(w, http.StatusInternalServerError, codeInternalError, "時間をおいてもう一度お試しください", nil)
}

// maxBodyBytes は JSON 本文の上限（1 MiB）
const maxBodyBytes = 1 << 20

// decodeJSON は JSON 本文を v に読む。失敗したらエラーを書き出して false を返す。
// Content-Type が application/json でなければ 415、JSON として読めなければ 400
func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		writeError(w, http.StatusUnsupportedMediaType, codeUnsupportedMediaType, "Content-Type は application/json にしてください", nil)
		return false
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes)).Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, codeInvalidRequest, "リクエストの形式が正しくありません", nil)
		return false
	}
	return true
}
