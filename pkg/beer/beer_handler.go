package beer

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

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

func (bh *BeerHandler) getOptions() template.HTML {
	strBuild := strings.Builder{}

	types, _ := bh.getBeerType()

	for _, t := range types {
		str := fmt.Sprintf("<option value=\"%s\">%s</option> \n ", t, t)
		log.Print(str)
		strBuild.WriteString(str)
	}

	return template.HTML(strBuild.String())
}

func (bh *BeerHandler) getBeerType() ([]string, error) {
	trans, err := bh.DB.Begin()
	if err != nil {
		trans.Rollback()
		return nil, err
	}

	row, err := trans.Query("select type_name from beer_type")
	if err != nil {
		trans.Rollback()
		return nil, err
	}
	res := []string{}

	for row.Next() {
		var str string
		err = row.Scan(&str)
		if err != nil {
			log.Println(err)
		}
		res = append(res, str)
	}

	return res, nil
}

func (bh *BeerHandler) AddBeer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		tmpl := bh.Tmpls.Lookup("beer.create.html")
		data := map[string]interface{}{
			"Options": bh.getOptions(),
		}
		tmpl.Execute(w, data)
		return
	}

	sess, _ := sessions.SessionFromContext(r.Context())
	r.ParseMultipartForm(5 * 1024 * 1025)

	uploadData, _, err := r.FormFile("logo")
	if err != nil {
		http.Error(w, fmt.Sprintf("cant parse file %v", err), http.StatusInternalServerError)
		return
	}
	defer uploadData.Close()
	name := r.FormValue("beer_name")
	producer := r.FormValue("producer")
	beer_type := r.FormValue("beer_types")

	md5Sum, err := 

}

func (bh *BeerHandler) getBeer() ([]Beer, error) {

	return nil, nil
}
