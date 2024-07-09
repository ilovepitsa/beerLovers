package event

import (
	"database/sql"
	"html/template"
	"log"
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
		sess, err := sessions.SessionFromContext(r.Context())

		if err != nil {
			log.Println("Event handler cant get session: ", err)
		}

		tmpl.Execute(w, sess.IsAdmin)
		return
	}
}

func (eh *EventHandler) Create(w http.ResponseWriter, r *http.Request) {

	sess, err := sessions.SessionFromContext(r.Context())
	if err != nil {
		return
	}
	if !sess.IsAdmin {
		http.Error(w, "not enought member level", http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodPost {
		tmpl := eh.Tmpls.Lookup("events.create.html")
		tmpl.Execute(w, nil)
		return
	}

	name := r.FormValue("name")
	date := r.FormValue("date")
	description := r.FormValue("description")

	log.Println("Details: ", name, " ", date, " ", description)

}
