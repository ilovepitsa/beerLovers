package event

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ilovepitsa/beerLovers/pkg/sessions"
)

// var (
// 	errEventExists = errors.New("event exists")
// )

type Event struct {
	Id          int
	Name        string
	Date        time.Time
	Location    string
	Description string
}

type eventViewData struct {
	Event      Event
	IsExpired  bool
	IsTakePart bool
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

func (eh *EventHandler) formatTableList(events []Event) template.HTML {
	var rowsHTML strings.Builder
	// rowsHTML.WriteString("<div class='row'>")
	log.Printf("Found %v events\n", len(events))
	for index, elem := range events {
		if index%4 == 0 {
			if index != 0 {
				rowsHTML.WriteString("</div><br>")
			}
			rowsHTML.WriteString("<div class='row '>")
		}
		rowsHTML.WriteString("<div class='col-sm-auto' style='max-width: max-content;'>")
		tmpl := eh.Tmpls.Lookup("eventsCard.html")

		err := tmpl.Execute(&rowsHTML, elem)
		if err != nil {
			log.Println("Error while executing eventCard: ", err)
		}
		rowsHTML.WriteString("</div>")
	}

	return template.HTML(rowsHTML.String())
}

func (eh *EventHandler) List(w http.ResponseWriter, r *http.Request) {
	showOld := false
	if r.Method == http.MethodPost {
		r.ParseForm()
		flag := r.Form["show_old"]
		showOld = len(flag) > 0
	}
	tmpl := eh.Tmpls.Lookup("events.html")
	sess, err := sessions.SessionFromContext(r.Context())

	if err != nil {
		log.Println("Event handler cant get session: ", err)
	}
	events, err := eh.getAllEvents(showOld)
	if err != nil {
		log.Println(err)
		http.Error(w, "Event get err: ", http.StatusInternalServerError)
		return
	}
	input := map[string]interface{}{
		"IsAdmin": sess.IsAdmin,
		"Rows":    eh.formatTableList(events),
	}

	err = tmpl.Execute(w, input)
	if err != nil {
		log.Println(err)
	}
}

func (eh *EventHandler) getAllEvents(showOld bool) ([]Event, error) {
	events := []Event{}

	trans, err := eh.DB.Begin()

	if err != nil {
		return nil, err
	}
	currentTime := time.Now()
	previosDay := currentTime.AddDate(0, 0, -1)
	var result *sql.Rows
	if showOld {
		result, err = trans.Query("select id, name, date, location, description from events order by date;")
	} else {
		result, err = trans.Query("select id, name, date, location, description from events where date > $1 order by date;", previosDay)
	}

	if err != nil {
		return nil, err
	}

	for result.Next() {
		e := Event{}
		err = result.Scan(&e.Id, &e.Name, &e.Date, &e.Location, &e.Description)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
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
	location := r.FormValue("location")
	description := r.FormValue("description")

	event, err := eh.createEvent(name, date, location, description)

	switch err {
	case nil:

	default:
		log.Println("Create event error: ", err)
		http.Error(w, "Error event", http.StatusInternalServerError)
	}

	log.Println("Details: ", event)
	http.Redirect(w, r, "/events/", http.StatusFound)
}

func (eh *EventHandler) createEvent(name, date, location, description string) (*Event, error) {

	t, err := time.Parse("2006-01-02", date)

	if err != nil {
		return nil, err
	}

	event := &Event{
		Id:          0,
		Name:        name,
		Date:        t,
		Location:    location,
		Description: description,
	}

	trans, err := eh.DB.Begin()
	if err != nil {
		trans.Rollback()
		return nil, err
	}
	log.Println(`insert into events (name, date, location, description) 
	values ($1, $2, $3, $4) RETURNING id;`, event.Name, event.Date, event.Location, event.Description)
	err = trans.QueryRow(`insert into events (name, date, location, description) 
	values ($1, $2, $3, $4) RETURNING id;`, event.Name, event.Date, event.Location, event.Description).Scan(&event.Id)
	if err != nil {
		trans.Rollback()
		return nil, err
	}

	trans.Commit()
	return event, nil
}

func (eh *EventHandler) TakePart(w http.ResponseWriter, r *http.Request) {
	// log.Println(r.URL.Path, " taken part!")
	sess, err := sessions.SessionFromContext(r.Context())
	if err != nil {
		http.Error(w, `{"err": "no auth"}`, http.StatusUnauthorized)
		log.Println("Take part err: ", err)
		return
	}

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Error(w, `{"err": "bad id"}`, http.StatusBadRequest)
		log.Println("Take part err: ", err)
		return
	}
	vote, err := strconv.Atoi(r.FormValue("vote"))
	if err != nil {
		http.Error(w, `{"err": "bad vote"}`, http.StatusBadRequest)
		log.Println("Take part err: ", err)
	}
	err = eh.updateParticipation(id, vote, sess.UserID)

	if err != nil {
		http.Error(w, `{"err": "cant process part in event"}`, http.StatusInternalServerError)
		log.Println("update part in event error:", err)
	}

}

func (eh *EventHandler) updateParticipation(id, vote int, userID uint32) error {
	trans, err := eh.DB.Begin()
	if err != nil {
		trans.Rollback()
		return err
	}

	if vote >= 0 {
		_, err = trans.Exec(`insert ignore into part_in_event (member_id, event_id) values ($1, $2)`, userID, id)
	} else {
		_, err = trans.Exec(`delete from part_in_event where event_id = $1 and member_id = $2`, id, userID)
	}

	if err != nil {
		trans.Rollback()
		return err
	}
	trans.Commit()
	return nil
}
