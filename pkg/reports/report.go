package reports

import (
	"database/sql"
	"encoding/csv"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ilovepitsa/beerLovers/pkg/beer"
	"github.com/ilovepitsa/beerLovers/pkg/member"
	"github.com/ilovepitsa/beerLovers/pkg/sessions"
	httputils "github.com/ilovepitsa/beerLovers/pkg/uitls/httpUtils"
	randstring "github.com/ilovepitsa/beerLovers/pkg/uitls/randString"
)

type UserFavoriteBeer struct {
	User  member.Member
	Beers []beer.Beer
}

type ReportHandler struct {
	DB *sql.DB
	SM sessions.SessionManager
}

func NewReportHandler(DB *sql.DB, SM sessions.SessionManager) *ReportHandler {
	return &ReportHandler{
		DB: DB,
		SM: SM,
	}
}

func (rh *ReportHandler) getInfo() ([]UserFavoriteBeer, error) {
	trans, err := rh.DB.Begin()
	if err != nil {
		trans.Rollback()
		return nil, err
	}

	rows, err := trans.Query(`
	select 
		m.fio,
		string_agg( concat_ws('|',b.name,b.producer),',') 
	from 
		member as m 
	join favorite_beer as fb on m.id = fb.member_id 
	join beer as b on b.id = fb.beer_id group by m.fio;`)

	if err != nil {
		trans.Rollback()
		return nil, err
	}

	res := []UserFavoriteBeer{}
	for rows.Next() {
		ufb := UserFavoriteBeer{}
		usersBeer := ""
		rows.Scan(&ufb.User.FIO, &usersBeer)
		nameAdnProd := strings.Split(usersBeer, ",")
		for _, v := range nameAdnProd {
			b := beer.Beer{}
			nap := strings.Split(v, "|")
			b.Name, b.Producer = nap[0], nap[1]

			ufb.Beers = append(ufb.Beers, b)
		}

		res = append(res, ufb)
	}
	return res, nil
}

func (rh *ReportHandler) createFile(info []UserFavoriteBeer) (string, string, error) {
	filename := randstring.RandStringRunes(7)
	fullFile := filename + ".csv"
	filepath := "./reports/" + fullFile
	file, err := os.Create(filepath)
	if err != nil {
		return "", "", nil
	}
	csvFile := csv.NewWriter(file)
	csvFile.Write([]string{"Имя участника", "Название пива", "Производитель"})
	for _, v := range info {
		csvFile.Write([]string{v.User.FIO})
		for _, b := range v.Beers {
			csvFile.Write([]string{"", b.Name, b.Producer})
		}
		csvFile.Write([]string{})
	}
	csvFile.Flush()
	file.Sync()
	file.Close()

	return filepath, fullFile, nil

}

func (rh *ReportHandler) BeerFavoriteReport(w http.ResponseWriter, r *http.Request) {
	sess, _ := sessions.SessionFromContext(r.Context())
	if !sess.IsAdmin {
		httputils.RespJSONError(w, http.StatusInternalServerError, nil, "internal")
		return
	}
	res, err := rh.getInfo()
	if err != nil {
		httputils.RespJSONError(w, http.StatusInternalServerError, nil, "internal")
		return
	}
	filepath, filename, err := rh.createFile(res)
	if err != nil {
		httputils.RespJSONError(w, http.StatusInternalServerError, nil, "internal")
		return
	}
	file, err := os.Open(filepath)
	if err != nil {
		httputils.RespJSONError(w, http.StatusInternalServerError, nil, "internal")
		return
	}
	defer file.Close()

	http.ServeContent(w, r, filename, time.Time{}, file)

}
