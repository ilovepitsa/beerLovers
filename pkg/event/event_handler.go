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
	"unicode/utf8"

	"github.com/ilovepitsa/beerLovers/pkg/beer"
	"github.com/ilovepitsa/beerLovers/pkg/sessions"
	httputils "github.com/ilovepitsa/beerLovers/pkg/uitls/httpUtils"
	"github.com/volatiletech/null"
)

// var (
// 	errEventExists = errors.New("event exists")
// )

type Review struct {
	MemberName string `json:"Reviewer"`
	PhotoPath  string `json:"PhotoPath"`
	Text       string `json:"Review"`
}

type Event struct {
	Id          int
	Name        string
	Date        time.Time
	Location    string
	Description string
}

type eventViewData struct {
	Event       Event
	IsExpired   bool
	IsTakePart  bool
	Responsible string
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
		if index%2 == 0 {
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
								select e.id, e.name, e.date, e.location, e.description,
								    case 
								        when pie.member_id is NULL then false
										when pie.member_id = $1 then true
								        else false 
								    end as IsTakePart,
								    case 
								        when e.date > CURRENT_DATE - INTEGER '1' then false
								        else true
								    end as IsExpired,
									m.fio
								from member as m, events as e left join part_in_event as pie on e.id = pie.event_id where e.responsible = m.id order by e.date;`, userId)
	} else {
		result, err = trans.Query(`
								select e.id, e.name, e.date, e.location, e.description,
								    case 
								        when pie.member_id is NULL then false
								        when pie.member_id = $1 then true
								        else false
								    end as IsTakePart,
								    case 
								        when e.date > CURRENT_DATE - INTEGER '1' then false
								        else true
								    end as IsExpired,
									m.fio
								from member as m, events as e left join part_in_event as pie on e.id = pie.event_id where e.date > $2 and e.responsible = m.id  order by e.date;`, userId, previosDay)
	}

	if err != nil {
		return nil, err
	}

	for result.Next() {
		e := eventViewData{}
		err = result.Scan(&e.Event.Id, &e.Event.Name, &e.Event.Date, &e.Event.Location, &e.Event.Description, &e.IsTakePart, &e.IsExpired, &e.Responsible)
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

	event, err := eh.createEvent(name, date, location, description, sess.UserID)

	switch err {
	case nil:

	default:
		log.Println("Create event error: ", err)
		http.Error(w, "Error event", http.StatusInternalServerError)
	}

	log.Println("Details: ", event)
	http.Redirect(w, r, "/events/", http.StatusFound)
}

func (eh *EventHandler) createEvent(name, date, location, description string, responsible uint32) (*Event, error) {

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
	log.Println(`insert into events (name, date, location, description, responsible) 
	values ($1, $2, $3, $4, $5) RETURNING id;`, event.Name, event.Date, event.Location, event.Description, responsible)
	err = trans.QueryRow(`insert into events (name, date, location, description, responsible) 
	values ($1, $2, $3, $4, $5) RETURNING id;`, event.Name, event.Date, event.Location, event.Description, responsible).Scan(&event.Id)
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
		_, err = trans.Exec(`insert into part_in_event (member_id, event_id) values ($1, $2) ON CONFLICT (member_id, event_id) DO NOTHING`, userID, id)
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

func (eh *EventHandler) addReview(r *http.Request) error {
	eid, err := strconv.ParseUint(r.FormValue("eid"), 10, 32)
	if err != nil {
		return err
	}
	sess, _ := sessions.SessionFromContext(r.Context())

	text := r.FormValue("review")

	if utf8.RuneCountInString(text) < 10 {
		return fmt.Errorf("review must contains at least 10 symbols")
	}

	photo, _, err := r.FormFile("messagePhoto")
	if err != nil {
		return err
	}
	defer photo.Close()
	photoPath, err := beer.SaveFile(photo, false)
	if err != nil {
		return err
	}

	trans, err := eh.DB.Begin()
	if err != nil {
		trans.Rollback()
		return err
	}
	id := 0
	photourl := null.StringFrom(photoPath)
	err = trans.QueryRow("insert into review (event_id, member_id, text, photo_url) values ($1, $2, $3) RETURNING id;", eid, sess.UserID, text, photourl).Scan(&id)
	if err != nil {
		trans.Rollback()
		return err
	}

	trans.Commit()
	return nil
}

func (eh *EventHandler) getReviewList(r *http.Request) ([]Review, error) {
	eid, err := strconv.ParseUint(r.FormValue("eid"), 10, 32)
	if err != nil {
		return nil, err
	}

	trans, err := eh.DB.Begin()
	if err != nil {
		trans.Rollback()
		return nil, err
	}
	res, err := trans.Query("select m.fio, r.text, r.photo_url from review as r, member as m where r.event_id = $1 and r.member_id = m.id order by r.id;", eid)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == sql.ErrNoRows {
		return nil, nil
	}

	var reviews []Review
	for res.Next() {
		review := Review{}
		var photoUrl sql.NullString
		res.Scan(&review.MemberName, &review.Text, &photoUrl)
		if photoUrl.Valid {
			review.PhotoPath = photoUrl.String
		}
		reviews = append(reviews, review)
	}

	return reviews, nil

}

func (eh *EventHandler) Review(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := eh.addReview(r)
		if err != nil {
			httputils.RespJSONError(w, http.StatusInternalServerError, err, "cant add review")
		}
		return
	}

	if r.Method == http.MethodGet {
		reviews, err := eh.getReviewList(r)
		if err != nil {
			httputils.RespJSONError(w, http.StatusInternalServerError, err, "cant get review")
			return
		}
		httputils.RespJSON(w, map[string]interface{}{
			"reviews": reviews,
		})
		return
	}
}
