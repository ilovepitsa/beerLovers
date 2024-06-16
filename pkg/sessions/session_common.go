package sessions

import (
	"net/http"
)

type Session struct {
	UserId uint32
	ID     string
}

type SessionManager interface {
	Check(*http.Request) (*Session, error)
	Create(http.ResponseWriter, *Member) error
	DestroyCurrent(http.ResponseWriter, *http.Request) error
	DestroyAll(http.ResponseWriter, *Member) error
}
