package index

import (
	"net/http"

	"github.com/ilovepitsa/beerLovers/pkg/sessions"
)

func Index(w http.ResponseWriter, r *http.Request) {
	_, err := sessions.SessionFromContext(r.Context())
	if err != nil {
		http.Redirect(w, r, "/user/login", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/events/", http.StatusFound)
}
