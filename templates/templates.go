package templates

import (
	"bytes"
	"embed"
	"html/template"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
)

// FS holds the templates compiled into the binary, used outside dev mode
//
//go:embed *.html
var FS embed.FS

// Funcs must be registered before parsing, in both dev and production
var Funcs = template.FuncMap{
	"verse": verse,
}

func Parse() *template.Template {
	return template.Must(template.New("").Funcs(Funcs).ParseFS(FS, "*.html"))
}

// Renderer executes the named templates. In dev it re-parses them from disk on every render, so edits show on
// refresh; otherwise it uses the ones embedded in the binary.
type Renderer struct {
	dev  bool
	tmpl *template.Template
}

func NewRenderer(dev bool) *Renderer {
	if dev {
		return &Renderer{dev: true}
	}
	return &Renderer{tmpl: Parse()}
}

// Render writes the template as an HTML response. It renders into a buffer first, so a template error becomes
// a clean 500 rather than half a page.
func (r *Renderer) Render(w http.ResponseWriter, status int, name string, data any) {
	tmpl := r.tmpl
	if r.dev {
		var err error
		tmpl, err = template.New("").Funcs(Funcs).ParseGlob("templates/*.html")
		if err != nil {
			slog.Error("Error parsing templates", "err", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		slog.Error("Error rendering template", "template", name, "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	buf.WriteTo(w)
}

// Verse is a verse split for display: the text with supplied words italicized, and its marginal notes
type Verse struct {
	Text  template.HTML
	Notes []Note
}

// Note is a translators' marginal note, like {firmament: Heb. expansion}
type Note struct {
	Term string
	Note string
}

var (
	// {term: note} is a marginal note. Any surrounding whitespace is dropped with it.
	marginalNote = regexp.MustCompile(`\s*\{([^{}:]*): ([^{}]*)\}`)
	// Any other {braces} mark words the translators supplied, which printed Bibles set in italics
	suppliedWords = regexp.MustCompile(`\{([^{}]*)\}`)
)

// verse pulls the marginal notes out of a verse, then escapes it and italicizes the supplied words
func verse(text string) Verse {
	var v Verse
	for _, m := range marginalNote.FindAllStringSubmatch(text, -1) {
		v.Notes = append(v.Notes, Note{Term: m[1], Note: m[2]})
	}
	text = marginalNote.ReplaceAllString(text, "")

	escaped := template.HTMLEscapeString(strings.TrimSpace(text))
	v.Text = template.HTML(suppliedWords.ReplaceAllString(escaped, "<em>$1</em>"))
	return v
}
