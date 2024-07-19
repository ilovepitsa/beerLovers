package middleware

import (
	"context"
	"net/http"

	"github.com/ilovepitsa/beerLovers/pkg/sessions"
)

// type SessionManager interface {
// 	Check(*http.Request) (*sessions.Session, error)
// 	Create(http.ResponseWriter, sessions.MemberInterface) error
// 	DestroyCurrent(http.ResponseWriter, *http.Request) error
// 	DestroyAll(http.ResponseWriter, sessions.MemberInterface) error
// 	CheckAdmin(session *sessions.Session) bool
// }

var (
	noAuthUrls = map[string]struct{}{
		"/user/login": struct{}{},
		"/user/reg":   struct{}{},
	}
)

func AuthMiddleware(sm sessions.SessionManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := noAuthUrls[r.URL.Path]; ok {
			next.ServeHTTP(w, r)
			return
		}
		sess, err := sm.Check(r)
		if err != nil {
			if err == sessions.ErrNoAuth {
				http.Redirect(w, r, "/user/login", http.StatusFound)
				return
			}
			http.Error(w, "No auth", http.StatusUnauthorized)
			return
		}
		sess.IsAdmin = sm.CheckAdmin(sess)
		ctx := context.WithValue(r.Context(), sessions.SessionKey, sess)
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}
