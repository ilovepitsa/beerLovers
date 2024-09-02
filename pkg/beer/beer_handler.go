package beer

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ilovepitsa/beerLovers/pkg/sessions"
)

type Beer struct {
	Name     string
	Producer string
	BeerType string
	Url      string
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

	// sess, _ := sessions.SessionFromContext(r.Context())
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

	md5Sum, err := saveFile(uploadData)
	if err != nil {
		http.Error(w, fmt.Sprintf("cant save file %v", err), http.StatusInternalServerError)
		return
	}
	b := Beer{
		Name:     name,
		Producer: producer,
		BeerType: beer_type,
		Url:      md5Sum,
	}
	err = bh.saveBeer(b)
	log.Println(err)
	http.Redirect(w, r, "/beer/", http.StatusFound)
}

func (bh *BeerHandler) saveBeer(b Beer) error {
	trans, err := bh.DB.Begin()
	if err != nil {
		trans.Rollback()
		return err
	}
	beer_id := -1
	err = trans.QueryRow("select id from beer_type where type_name = $1", b.BeerType).Scan(&beer_id)
	if err != nil {
		trans.Rollback()
		log.Println("Cant get beer_type id for ", b.BeerType)
		return err
	}
	res := 0
	err = trans.QueryRow("insert into beer (name, producer, beer_type, photo_url) values ($1, $2, $3, $4) returning1;", b.Name, b.Producer, beer_id, b.Url).Scan(&res)
	if err != nil {
		trans.Rollback()
		return err
	}

	if res != 1 {
		trans.Rollback()
		log.Println("1 != 1")
		return err
	}

	trans.Commit()
	return nil
}

func saveFile(in io.Reader) (string, error) {
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)
	b := make([]byte, 32)

	if _, err := r.Read(b); err != nil {
		return "", err
	}
	tmpFile := "./images/" + string(b) + ".jpg"
	newFile, err := os.Create(tmpFile)
	if err != nil {
		return "", err
	}

	hasher := md5.New()
	_, err = io.Copy(newFile, io.TeeReader(in, hasher))
	if err != nil {
		return "", err
	}
	newFile.Sync()
	newFile.Close()

	md5Sum := hex.EncodeToString(hasher.Sum(nil))

	realFile := "./images/" + md5Sum + ".jpg"
	err = os.Rename(tmpFile, realFile)
	if err != nil {
		return "", nil
	}

	return md5Sum, nil
}

func (bh *BeerHandler) getBeer() ([]Beer, error) {

	return nil, nil
}
