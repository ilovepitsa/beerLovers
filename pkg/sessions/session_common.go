package sessions

import (
	"context"
	"errors"
	"net/http"
)

var (
	ErrNoAuth = errors.New("no session found")
)

type ctxKey int

const SessionKey ctxKey = 1

func SessionFromContext(ctx context.Context) (*Session, error) {
	sess, ok := ctx.Value(SessionKey).(*Session)
	if !ok {
		return nil, ErrNoAuth
	}
	return sess, nil
}

type Session struct {
	UserID uint32
	ID     string
}

type MemberInterface interface {
	GetID() int
	IsAdmin() bool
}

type SessionManager interface {
	Check(*http.Request) (*Session, error)
	Create(http.ResponseWriter, MemberInterface) error
	DestroyCurrent(http.ResponseWriter, *http.Request) error
	DestroyAll(http.ResponseWriter, MemberInterface) error
	CheckAdmin(r *http.Request) bool
}
