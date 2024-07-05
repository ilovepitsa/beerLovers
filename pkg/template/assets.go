//go:build dev
// +build dev

package template

import (
	"net/http"

	"github.com/shurcooL/httpfs/union"
)

var Assets http.FileSystem = union.New(map[string]http.FileSystem{
	"/templates": http.Dir("./templates/"),
	"/static":    http.Dir("./static/"),
})
