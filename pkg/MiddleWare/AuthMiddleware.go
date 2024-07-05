package middleware

import (
	"context"
	"net/http"

	"github.com/ilovepitsa/beerLovers/pkg/sessions"
)

type SessionManager interface {
	Check(*http.Request) (*sessions.Session, error)
	Create(http.ResponseWriter, sessions.MemberInterface) error
	DestroyCurrent(http.ResponseWriter, *http.Request) error
	DestroyAll(http.ResponseWriter, sessions.MemberInterface) error
}

var (
	noAuthUrls = map[string]struct{}{
		"/user/login_oauth": struct{}{},
		"/user/login":       struct{}{},
		"/user/reg":         struct{}{},
		"/":                 struct{}{},
	}
)

func AuthMiddleware(sm SessionManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := noAuthUrls[r.URL.Path]; ok {
			next.ServeHTTP(w, r)
			return
		}
		sess, err := sm.Check(r)
		if err != nil {
			http.Error(w, "No auth", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), sessions.SessionKey, sess)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
