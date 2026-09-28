package templates

import (
	"embed"
	"html/template"
	"regexp"
	"strings"
)

// FS holds the templates compiled into the binary, used outside dev mode
//
//go:embed *.html
var FS embed.FS

// Funcs must be registered before parsing, in both dev (gin's LoadHTMLGlob) and production
var Funcs = template.FuncMap{
	"verse": verse,
}

func Parse() *template.Template {
	return template.Must(template.New("").Funcs(Funcs).ParseFS(FS, "*.html"))
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
