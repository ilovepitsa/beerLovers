package event

import (
	"database/sql"
	"html/template"
	"net/http"
	"time"

	"github.com/ilovepitsa/beerLovers/pkg/sessions"
)

type Event struct {
	Id          int
	Name        string
	Date        time.Time
	Location    string
	Description string
}

type EventHandler struct {
	DB    *sql.DB
	Tmpls *template.Template
	SM    sessions.SessionManager
}

func NewEventHander(DB *sql.DB, Tmpls *template.Template, SM sessions.SessionManager) *EventHandler {
	return &EventHandler{
		DB:    DB,
		Tmpls: Tmpls,
		SM:    SM,
	}
}
func (eh *EventHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		tmpl := eh.Tmpls.Lookup("events.html")
		tmpl.Execute(w, nil)
	}
}
