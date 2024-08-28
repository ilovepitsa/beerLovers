package beer

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/ilovepitsa/beerLovers/pkg/sessions"
)

type Beer struct {
}

type BeerHandler struct {
	DB    *sql.DB
	Tmpls *template.Template
	SM    sessions.SessionManager
}

func NewBeerHandler(DB *sql.DB, Tmpls *template.Template, SM sessions.SessionManager) *BeerHandler {
	return &BeerHandler{
		DB:    DB,
		Tmpls: Tmpls,
		SM:    SM,
	}
}

func (bh *BeerHandler) formatTableList(beer []Beer) string {
	return ""
}

func (bh *BeerHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `wrong method`, http.StatusMethodNotAllowed)
		return
	}
	tmpl := bh.Tmpls.Lookup("beer.html")
	sess, err := sessions.SessionFromContext(r.Context())
	if err != nil {
		log.Println("Event handler cant get session: ", err)
		http.Error(w, "cant get session", http.StatusInternalServerError)
	}
	beers, err := bh.getBeer()
	if err != nil {
		log.Println(err)
		http.Error(w, fmt.Sprintf("Get beer error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	input := map[string]interface{}{
		"IsAdmin": sess.IsAdmin,
		"Rows":    bh.formatTableList(beers),
	}

	err = tmpl.Execute(w, input)
	if err != nil {
		log.Println(err)
	}

}

func (bh *BeerHandler) AddBeer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "wornd method", http.StatusMethodNotAllowed)
		return
	}

}

func (bh *BeerHandler) getBeer() ([]Beer, error) {

	return nil, nil
}
