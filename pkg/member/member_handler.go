package member

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/ilovepitsa/beerLovers/pkg/sessions"
	randstring "github.com/ilovepitsa/beerLovers/pkg/uitls/randString"
	"golang.org/x/crypto/argon2"
)

var (
	errNoRec      = errors.New("no user record found")
	errBadPass    = errors.New("bad password")
	errUserExists = errors.New("user exists")
)

type userViewData struct {
	FIO         string
	Entry_Date  time.Time
	Address     sql.NullString
	PhoneNumber sql.NullString
	Email       string
	Balance     string
}

type MemberHandler struct {
	DB    *sql.DB
	Tmpls *template.Template
	SM    sessions.SessionManager
}

func NewMemberHandler(db *sql.DB, tmpls *template.Template, sm sessions.SessionManager) *MemberHandler {
	return &MemberHandler{
		DB:    db,
		Tmpls: tmpls,
		SM:    sm,
	}
}
func (mh *MemberHandler) checkPasswordByLogin(login, pass string) (*Member, error) {
	row := mh.DB.QueryRow("Select id, email, password from member where email = $1", login)
	return mh.passwordIsValid(pass, row)
}

func (mh *MemberHandler) hashPass(plainPassword, salt string) []byte {
	hashedPass := argon2.IDKey([]byte(plainPassword), []byte(salt), 1, 64*1024, 4, 32)
	res := make([]byte, len(salt))
	copy(res, salt[:len(salt)])
	return append(res, hashedPass...)
}

func (mh *MemberHandler) passwordIsValid(pass string, row *sql.Row) (*Member, error) {
	var (
		dbPass []byte
		user   = &Member{}
	)
	err := row.Scan(&user.Id, &user.Email, &dbPass)
	log.Println(err)
	if err == sql.ErrNoRows {
		return nil, errNoRec
	} else if err != nil {
		return nil, err
	}

	salt := string(dbPass[0:8])
	if !bytes.Equal(mh.hashPass(pass, salt), dbPass) {
		return nil, errBadPass
	}
	return user, nil
}

func (mh *MemberHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {

		tmpl := mh.Tmpls.Lookup("login.html")
		tmpl.Execute(w, nil)
		return
	}

	login := r.FormValue("login")
	pass := r.FormValue("password")

	user, err := mh.checkPasswordByLogin(login, pass)

	switch err {
	case nil:
		// all is ok
	case errNoRec:
		http.Error(w, "No user", http.StatusBadRequest)
	case errBadPass:
		http.Error(w, "Bad pass", http.StatusBadRequest)
	default:
		http.Error(w, "Db err", http.StatusInternalServerError)
	}
	if err != nil {
		return
	}

	mh.SM.Create(w, user)
	http.Redirect(w, r, "/events/", http.StatusFound)

	fmt.Fprintln(w, "login!")
}

func (mh *MemberHandler) Registry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		tmpl := mh.Tmpls.Lookup("reg.html")
		tmpl.Execute(w, nil)
		return
	}

	login := r.FormValue("login")
	pass := r.FormValue("password")
	fio := r.FormValue("fio")

	user, err := mh.createMember(login, pass, fio)
	switch err {
	case nil:
		// all is ok
	case errUserExists:
		http.Error(w, "Looks like user exists", http.StatusBadRequest)
	default:
		log.Println("db err", err)
		http.Error(w, "Db err", http.StatusInternalServerError)
	}
	if err != nil {
		return
	}

	mh.SM.Create(w, user)
	http.Redirect(w, r, "/", http.StatusFound)

}

func (mh *MemberHandler) createMember(login, passIn, fio string) (*Member, error) {
	salt := randstring.RandStringRunes(8)
	pass := mh.hashPass(passIn, salt)
	fmt.Printf("%x\n", []byte(pass))
	member := &Member{
		Id:         0,
		FIO:        fio,
		Entry_Date: time.Now(),
		Email:      login,
	}

	err := mh.DB.QueryRow("Select id, fio from member where email = $1;", member.Email).Scan(&member.Id, &member.FIO)

	if err != nil && err != sql.ErrNoRows {

		return nil, fmt.Errorf("db err  : %v", err)
	}

	if err != sql.ErrNoRows {
		return member, errUserExists
	}

	trans, err := mh.DB.Begin()
	if err != nil {
		trans.Rollback()
		return nil, fmt.Errorf("db err  : %v", err)
	}

	var walletId int
	err = trans.QueryRow(`insert into wallet (balance) values(0) RETURNING id;`).Scan(&walletId)
	if err != nil {
		trans.Rollback()

		return nil, fmt.Errorf("cant create wallet : %v", err)
	}

	err = trans.QueryRow(`insert into member (id, fio, entry_date,
	  email, password, wallet_id, level) 
	 values (DEFAULT, $1, $2, $3, $4, $5, 'user') RETURNING id;`,
		member.FIO,
		member.Entry_Date,
		member.Email,
		pass,
		walletId,
	).Scan(&member.Id)

	if err != nil {
		trans.Rollback()
		return nil, fmt.Errorf("cant create user  : %v", err)
	}
	trans.Commit()
	return member, nil
}

func (mh *MemberHandler) Logout(w http.ResponseWriter, r *http.Request) {
	mh.SM.DestroyCurrent(w, r)
	http.Redirect(w, r, "/user/login", http.StatusFound)
}

func (mh *MemberHandler) getUserInfo(uid uint32) (*userViewData, error) {
	trans, err := mh.DB.Begin()
	if err != nil {
		trans.Rollback()
		return nil, err
	}
	res := trans.QueryRow("select m.fio, m.entry_date, m.address, m.phone_number, "+
		"m.email, m.wallet_id, w.balance from member as m, wallet as w where m.id = $1 and m.wallet_id = w.id", uid)

	m := &Member{}
	err = res.Scan(&m.FIO, &m.Entry_Date, &m.Address, &m.PhoneNumber, &m.Email, &m.Wallet_id, &m.Balance)
	if err != nil {
		trans.Rollback()
		return nil, err
	}
	trans.Commit()
	vd := &userViewData{
		FIO:         m.FIO,
		Entry_Date:  m.Entry_Date,
		Address:     m.Address,
		PhoneNumber: m.PhoneNumber,
		Email:       m.Email,
		Balance:     fmt.Sprintf("%.2f", m.Balance),
	}
	return vd, nil

}

func (mh *MemberHandler) Profile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "wrong method", http.StatusMethodNotAllowed)
		return
	}
	sess, _ := sessions.SessionFromContext(r.Context())
	log.Println(sess.UserID)
	userInfo, err := mh.getUserInfo(sess.UserID)
	if err != nil {
		http.Error(w, "cant get user info", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	data := map[string]interface{}{
		"User": userInfo,
	}
	tmpls := mh.Tmpls.Lookup("profile.html")

	err = tmpls.Execute(w, data)
	log.Println(err)
}
