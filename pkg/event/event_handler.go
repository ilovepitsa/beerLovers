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

func (eh *EventHandler) formatTableList(sess sessions.Session, events []eventViewData) template.HTML {
	var rowsHTML strings.Builder

	for index, elem := range events {
		if index%4 == 0 {
			if index != 0 {
				rowsHTML.WriteString("</div><br>")
			}
			rowsHTML.WriteString("<div class='row '>")
		}
		rowsHTML.WriteString("<div class='col-sm-auto' style='max-width: max-content;'>")
		tmpl := eh.Tmpls.Lookup("eventsCard.html")
		log.Println("Event with id = ", elem.Event.Id)
		data := map[string]interface{}{
			"UserId":  sess.UserID,
			"Element": elem,
			"IsAdmin": sess.IsAdmin,
		}
		err := tmpl.Execute(&rowsHTML, data)
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
		"Rows":    eh.formatTableList(*sess, events),
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
		data := map[string]interface{}{
			"IsAdmin": sess.IsAdmin,
		}
		tmpl := eh.Tmpls.Lookup("events.create.html")
		tmpl.Execute(w, data)
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

func (eh *EventHandler) getUsersParticipants(eventId uint32) ([]string, error) {
	ans := []string{}
	trans, err := eh.DB.Begin()
	if err != nil {
		return nil, err
	}
	res, err := trans.Query(`select member.fio from member left join part_in_event on member.id = part_in_event.member_id where part_in_event.event_id = $1;`, eventId)
	if err != nil {
		return nil, err
	}

	for res.Next() {
		var fio string
		res.Scan(&fio)
		ans = append(ans, fio)
	}

	return ans, nil
}

func (eh *EventHandler) userList(usersName []string) template.HTML {
	var rowsHTML strings.Builder
	rowsHTML.WriteString(`<ul class="list-group">`)
	for _, name := range usersName {
		rowsHTML.WriteString(fmt.Sprintf(`	<li class="list-group-item">%s</li>%s`, name, "\n"))
	}
	rowsHTML.WriteString(`</ul>`)
	return template.HTML(rowsHTML.String())
}

func (eh *EventHandler) Participants(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.RespJSONError(w, http.StatusMethodNotAllowed, nil, "bad method")
		return
	}

	eventId, err := strconv.ParseUint(r.FormValue("eid"), 10, 32)
	if err != nil {
		httputils.RespJSONError(w, http.StatusInternalServerError, err, "bad eid")
		return
	}
	userNames, err := eh.getUsersParticipants(uint32(eventId))
	if err != nil {
		httputils.RespJSONError(w, http.StatusInternalServerError, err, "cant get users names")
		return
	}

	sess, _ := sessions.SessionFromContext(r.Context())
	data := map[string]interface{}{
		"ListUsers": eh.userList(userNames),
		"IsAdmin":   sess.IsAdmin,
	}

	tmpl := eh.Tmpls.Lookup("participants.html")

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println(err)
	}
}

func (eh *EventHandler) deleteEvent(uid uint32) error {
	trans, err := eh.DB.Begin()
	if err != nil {
		trans.Rollback()
		return err
	}
	_, err = trans.Exec(`delete from events where id = $1`, uid)
	if err != nil {
		trans.Rollback()
		return err
	}
	trans.Commit()
	return nil
}

func (eh *EventHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		httputils.RespJSONError(w, http.StatusMethodNotAllowed, nil, "bad method")
		return
	}
	sess, _ := sessions.SessionFromContext(r.Context())
	if !sess.IsAdmin {
		httputils.RespJSONError(w, http.StatusMethodNotAllowed, nil, "internal")
		return
	}
	uid, err := strconv.ParseUint(r.FormValue("uid"), 10, 32)
	if err != nil {
		httputils.RespJSONError(w, http.StatusMethodNotAllowed, nil, "bad uid")
		return
	}
	err = eh.deleteEvent(uint32(uid))
	if err != nil {
		httputils.RespJSONError(w, http.StatusMethodNotAllowed, nil, "bad uid")
		return
	}
}
