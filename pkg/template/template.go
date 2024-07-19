package template

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/shurcooL/httpfs/html/vfstemplate"
)

type TemplateManager struct {
}

func NewTemplates(assets http.FileSystem) *template.Template {
	tmpl := template.New("").Funcs(template.FuncMap{
		"readableDate": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
	})
	tmpl, err := vfstemplate.ParseGlob(assets, tmpl, "/templates/*.html")
	if err != nil {
		log.Fatal(err)
	}
	return tmpl
}
