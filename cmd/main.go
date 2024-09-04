package main

import (
	"database/sql"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	middleware "github.com/ilovepitsa/beerLovers/pkg/MiddleWare"
	"github.com/ilovepitsa/beerLovers/pkg/beer"
	"github.com/ilovepitsa/beerLovers/pkg/event"
	"github.com/ilovepitsa/beerLovers/pkg/index"
	"github.com/ilovepitsa/beerLovers/pkg/member"
	"github.com/ilovepitsa/beerLovers/pkg/sessions"
	"github.com/ilovepitsa/beerLovers/pkg/template"
	_ "github.com/lib/pq"
)

func main() {
	connStr := "host=localhost port=5433 user=nikita password=12345 dbname=beer_lovers_party sslmode=disable"
	// connectDB := "user=nikita port=5432 password=12345 dbname=beer_lovers_party sslmode=disable host=localhost"

	db, err := sql.Open("postgres", connStr)
	err = db.Ping()
	if err != nil {
		log.Fatalf(" cant connect to db, err: %v\n", err)
	}

	tmpls := template.NewTemplates(template.Assets)
	sm := sessions.NewSessionsDB(db)

	mh := member.NewMemberHandler(db, tmpls, sm)
	bh := beer.NewBeerHandler(db, tmpls, sm)
	eh := event.NewEventHander(db, tmpls, sm)

	router := mux.NewRouter()
	router.HandleFunc("/", index.Index)
	router.HandleFunc("/user/login", mh.Login)
	router.HandleFunc("/user/reg", mh.Registry)
	router.HandleFunc("/user/logout", mh.Logout)
	router.HandleFunc("/events/", eh.List)
	router.HandleFunc("/events/take_part", eh.TakePart)
	router.HandleFunc("/events/create", eh.Create)
	router.HandleFunc("/user/", mh.Profile)
	router.HandleFunc("/beer/", bh.List)
	router.HandleFunc("/beer/create", bh.AddBeer)

	http.Handle("/", middleware.AuthMiddleware(sm, router))
	imageHandler := http.StripPrefix(
		"/images/",
		http.FileServer(http.Dir("./images/")),
	)

	http.Handle("/images/", imageHandler)
	http.Handle("/static/", http.FileServer(template.Assets))

	f, _ := template.Assets.Open("/static/favicon.ico")
	defer f.Close()
	favicon, _ := io.ReadAll(f)
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Write(favicon)
	})

	log.Println("Server starts: ", config.Address)
	http.ListenAndServe(config.Address, nil)
}
