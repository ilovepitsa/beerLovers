package beer

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/ilovepitsa/beerLovers/pkg/sessions"
	httputils "github.com/ilovepitsa/beerLovers/pkg/uitls/httpUtils"
	randstring "github.com/ilovepitsa/beerLovers/pkg/uitls/randString"
)

type Beer struct {
	Id         uint32
	Name       string
	Producer   string
	BeerType   string
	Url        string
	IsFavorite bool
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

func (bh *BeerHandler) formatTableList(beer []Beer) template.HTML {
	strB := strings.Builder{}

	for i, b := range beer {
		if i%4 == 0 {
			if i != 0 {
				strB.WriteString("</div><br>")
			}
			strB.WriteString("<div class='row'>")
		}
		strB.WriteString("<div class='col-sm-auto' style='max-width: max-content;'>")
		tmpl := bh.Tmpls.Lookup("beer.card.html")
		err := tmpl.Execute(&strB, b)
		if err != nil {
			log.Println("Error while executing eventCard: ", err)
		}
		strB.WriteString("</div>")
	}
	return template.HTML(strB.String())
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
	beers, err := bh.getBeer(sess.UserID)
	if err != nil {
		log.Println(err)
		http.Error(w, fmt.Sprintf("Get beer error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	log.Println(beers)
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
		sess, _ := sessions.SessionFromContext(r.Context())
		data := map[string]interface{}{
			"Options": bh.getOptions(),
			"IsAdmin": sess.IsAdmin,
		}
		tmpl.Execute(w, data)
		return
	}

	// sess, _ := sessions.SessionFromContext(r.Context())
	r.ParseMultipartForm(5 * 1024 * 1025)

	uploadData, _, err := r.FormFile("logo")
	if err != nil {
		http.Error(w, fmt.Sprintf("cant parse file %s", err.Error()), http.StatusInternalServerError)
		return
	}
	defer uploadData.Close()
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "cant parse form", http.StatusInternalServerError)
	}

	name := r.PostFormValue("beer_name")
	producer := r.PostFormValue("producer")
	beer_type := r.PostFormValue("beer_types")
	// log.Println("Beer name: ", name)

	md5Sum, err := SaveFile(uploadData, true)
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
	err = trans.QueryRow("insert into beer (name, producer, beer_type, photo_url) values ($1, $2, $3, $4) returning 1;", b.Name, b.Producer, beer_id, b.Url).Scan(&res)
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

func SaveFile(in io.Reader, isResize bool) (string, error) {
	tmpName := randstring.RandStringRunes(32)

	tmpFile := "./images/" + tmpName + ".jpg"
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

	if !isResize {
		return realFile, nil
	}

	srcImage, err := imaging.Open(realFile)
	if err != nil {
		return "", fmt.Errorf("failed to open image: %v", err)
	}
	resizedPath := "./images/" + md5Sum + "_res" + ".jpg"
	dstImageFill := imaging.Fill(srcImage, 150, 200, imaging.Center, imaging.Lanczos)
	err = imaging.Save(dstImageFill, resizedPath)
	if err != nil {
		return "", fmt.Errorf("failed to save image: %v", err)
	}
	return md5Sum + "_res", nil
}

func (bh *BeerHandler) getBeer(uid uint32) ([]Beer, error) {
	trans, err := bh.DB.Begin()
	if err != nil {
		trans.Rollback()
		return nil, err
	}

	res, err := trans.Query(`
	select b.id, b.name, b.producer, bt.type_name, b.photo_url,
                case
                        when fb.member_id = $1 then true
                        else false
                end as IsFavorite  
        from beer as b 
        left join favorite_beer as fb on b.id = fb.beer_id and fb.member_id = $1
        inner join beer_type as bt on b.beer_type = bt.id;
		`, uid)
	if err != nil {
		trans.Rollback()
		return nil, err
	}

	beers := []Beer{}
	for res.Next() {
		beer := Beer{}
		err = res.Scan(&beer.Id, &beer.Name, &beer.Producer, &beer.BeerType, &beer.Url, &beer.IsFavorite)
		if err != nil {
			log.Println(err)
		}
		beers = append(beers, beer)
	}

	return beers, nil
}

func (bh *BeerHandler) MakeFavorite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.RespJSONError(w, http.StatusMethodNotAllowed, nil, "bad method")
		return
	}
	sess, _ := sessions.SessionFromContext(r.Context())
	id, err := strconv.ParseUint(r.FormValue("id"), 10, 32)
	if err != nil {
		httputils.RespJSONError(w, http.StatusInternalServerError, nil, "internal")
		return
	}
	vote, err := strconv.Atoi(r.FormValue("vote"))
	if err != nil {
		httputils.RespJSONError(w, http.StatusInternalServerError, nil, "internal")
		return
	}
	err = bh.makeFavorite(uint32(id), vote, sess.UserID)
	if err != nil {
		httputils.RespJSONError(w, http.StatusInternalServerError, nil, "internal")
		return
	}
}

func (bh *BeerHandler) makeFavorite(bid uint32, vote int, uid uint32) error {
	trans, err := bh.DB.Begin()
	if err != nil {
		trans.Rollback()
		return err
	}
	log.Println(fmt.Printf("%d %d %d", bid, vote, uid))
	if vote >= 0 {
		_, err = trans.Exec(`insert into favorite_beer (member_id, beer_id) values ($1, $2) ON CONFLICT (member_id, beer_id) DO NOTHING`, uid, bid)
	} else {
		_, err = trans.Exec(`delete from favorite_beer where beer_id = $1 and member_id = $2`, bid, uid)
	}

	if err != nil {
		trans.Rollback()
		return err
	}
	trans.Commit()
	return nil
}
