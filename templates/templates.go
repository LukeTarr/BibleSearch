package templates

import (
	"embed"
	"html/template"
)

// FS holds the templates compiled into the binary, used outside dev mode
//
//go:embed *.html
var FS embed.FS

func Parse() *template.Template {
	return template.Must(template.ParseFS(FS, "*.html"))
}
