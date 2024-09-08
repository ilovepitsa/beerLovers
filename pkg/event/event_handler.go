package event

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ilovepitsa/beerLovers/pkg/sessions"
	httputils "github.com/ilovepitsa/beerLovers/pkg/uitls/httpUtils"
)

// var (
// 	errEventExists = errors.New("event exists")
// )

type Event struct {
	Id          int
	Name        string
	Date        time.Time
	Location    string
	Cost        float32
	Description string
}

type eventViewData struct {
	Event          Event
	IsExpired      bool
	IsTakePart     bool
	EventCostPrint string
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

func (eh *EventHandler) formatTableList(uid uint32, events []eventViewData) template.HTML {
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
		data := map[string]interface{}{
			"UserId":  uid,
			"Element": elem,
		}
		err := tmpl.Execute(&rowsHTML, data)
		if err != nil {
			log.Println("Error while executing eventCard: ", err)
		}
		rowsHTML.WriteString("</div>")
	}

	return template.HTML(rowsHTML.String())
}

// func (eh *EventHandler) getUserPartition(userID uint32, events []Event) {

// }

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
		http.Error(w, "cant get session", http.StatusInternalServerError)
	}
	events, err := eh.getAllEvents(showOld, sess.UserID)
	// eventsList, err :=
	if err != nil {
		log.Println(err)
		http.Error(w, "Event get err: ", http.StatusInternalServerError)
		return
	}
	log.Println("Current user id: ", sess.UserID)
	input := map[string]interface{}{
		"IsAdmin": sess.IsAdmin,
		"Rows":    eh.formatTableList(sess.UserID, events),
	}

	err = tmpl.Execute(w, input)
	if err != nil {
		log.Println(err)
	}
}

func (eh *EventHandler) getAllEvents(showOld bool, userId uint32) ([]eventViewData, error) {
	events := []eventViewData{}

	trans, err := eh.DB.Begin()

	if err != nil {
		return nil, err
	}
	currentTime := time.Now()
	previosDay := currentTime.AddDate(0, 0, -1)
	var result *sql.Rows
	if showOld {
		result, err = trans.Query(`
								select e.id, e.name, e.date, e.location, e.description, e.cost,
								    case 
								        when pie.member_id is NULL then false
										when pie.member_id = $1 then true
								        else false 
								    end as IsTakePart,
								    case 
								        when e.date > CURRENT_DATE - INTEGER '1' then false
								        else true
								    end as IsExpired
								from events as e left join part_in_event as pie on e.id = pie.event_id order by e.date;`, userId)
	} else {
		result, err = trans.Query(`
								select e.id, e.name, e.date, e.location, e.description, e.cost,
								    case 
								        when pie.member_id is NULL then false
								        when pie.member_id = $1 then true
								        else false
								    end as IsTakePart,
								    case 
								        when e.date > CURRENT_DATE - INTEGER '1' then false
								        else true
								    end as IsExpired
								from events as e left join part_in_event as pie on e.id = pie.event_id where e.date > $2 order by e.date;`, userId, previosDay)
	}

	if err != nil {
		return nil, err
	}

	for result.Next() {
		e := eventViewData{}
		err = result.Scan(&e.Event.Id, &e.Event.Name, &e.Event.Date, &e.Event.Location, &e.Event.Description, &e.Event.Cost, &e.IsTakePart, &e.IsExpired)
		e.EventCostPrint = fmt.Sprintf("%.2f", e.Event.Cost)
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
	cost := r.FormValue("cost")

	event, err := eh.createEvent(name, date, location, cost, description)

	switch err {
	case nil:

	default:
		log.Println("Create event error: ", err)
		http.Error(w, "Error event", http.StatusInternalServerError)
	}

	log.Println("Details: ", event)
	http.Redirect(w, r, "/events/", http.StatusFound)
}

func (eh *EventHandler) createEvent(name, date, location, cost, description string) (*Event, error) {

	t, err := time.Parse("2006-01-02", date)

	if err != nil {
		return nil, err
	}

	costF, err := strconv.ParseFloat(cost, 32)
	if err != nil {
		return nil, err
	}
	event := &Event{
		Id:          0,
		Name:        name,
		Date:        t,
		Location:    location,
		Cost:        float32(costF),
		Description: description,
	}

	trans, err := eh.DB.Begin()
	if err != nil {
		trans.Rollback()
		return nil, err
	}
	log.Println(`insert into events (name, date, location, description, cost) 
	values ($1, $2, $3, $4, $5) RETURNING id;`, event.Name, event.Date, event.Location, event.Description, event.Cost)
	err = trans.QueryRow(`insert into events (name, date, location, description, cost) 
	values ($1, $2, $3, $4, $5) RETURNING id;`, event.Name, event.Date, event.Location, event.Description, event.Cost).Scan(&event.Id)
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
	cost, err := strconv.ParseFloat(r.FormValue("cost"), 32)
	if err != nil {
		httputils.RespJSONError(w, http.StatusBadRequest, err, "bad cost")
		return
	}

	err = eh.updateParticipation(id, vote, float32(cost), sess.UserID)

	if err != nil {
		http.Error(w, `{"err": "cant process part in event"}`, http.StatusInternalServerError)
		log.Println("update part in event error:", err)
	}

}

func (eh *EventHandler) updateParticipation(id, vote int, cost float32, userID uint32) error {
	trans, err := eh.DB.Begin()
	if err != nil {
		trans.Rollback()
		return err
	}

	if vote >= 0 {
		_, err = trans.Exec(`insert into part_in_event (member_id, event_id) values ($1, $2) ON CONFLICT (member_id, event_id) DO NOTHING`, userID, id)
	} else {
		_, err = trans.Exec(`delete from part_in_event where event_id = $1 and member_id = $2`, id, userID)
	}

	if err != nil {
		trans.Rollback()
		return err
	}
	cost = cost * float32(vote)
	_, err = trans.Exec(`update wallet set balance = balance - $1 where id = (select wallet_id from member where id = $2)`, cost, userID)
	if err != nil {
		trans.Rollback()
		return err
	}

	trans.Commit()
	return nil
}
