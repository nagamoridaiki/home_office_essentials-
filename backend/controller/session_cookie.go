package controller

import (
	"net/http"

	"backend/domain"
)

const sessionCookieName = "customer_session"

// SessionCookie は Cookie customer_session の発行と削除
type SessionCookie struct {
	// secure が false のときだけ Secure を外す（http の localhost で動かすため）
	secure bool
}

func NewSessionCookie(secure bool) SessionCookie {
	return SessionCookie{secure: secure}
}

// Set は生のトークンを Cookie に入れる。有効期限はセッションと同じ 30 日
func (c SessionCookie) Set(w http.ResponseWriter, raw string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    raw,
		Path:     "/",
		MaxAge:   int(domain.SessionLifetime.Seconds()),
		HttpOnly: true,
		Secure:   c.secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// Clear は同じ名前・Path で Max-Age=0 を返し、ブラウザから消させる
func (c SessionCookie) Clear(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Go では負の値で「Max-Age=0」を出力する
		HttpOnly: true,
		Secure:   c.secure,
		SameSite: http.SameSiteLaxMode,
	})
}
