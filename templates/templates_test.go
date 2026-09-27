package templates

import (
	"BibleSearch/model"
	"strings"
	"testing"
)

// Rendering every template catches field typos, which html/template only reports at execution time
func TestTemplatesRender(t *testing.T) {
	tmpl := Parse()
	results := []model.ChromaQueryResultsDTO{{
		Metadata: model.Metadata{
			Book:          "John",
			Chapter:       "11",
			Verse:         "35",
			ReferenceLink: "https://www.bible.com/bible/1/JHN.11",
		},
		Distance: 0.5,
		Text:     "Jesus wept.",
		Id:       "26580",
	}}

	tests := []struct {
		name string
		data any
		want []string
	}{
		{"home", nil, []string{"<html", "</html>", `id="searchResultArea"`, `hx-post="/search"`}},
		{"about", nil, []string{"<html", "</html>", "text-embedding-3-small"}},
		{"results", results, []string{"Jesus wept.", "John 11:35", `href="https://www.bible.com/bible/1/JHN.11"`}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder
			if err := tmpl.ExecuteTemplate(&sb, tt.name, tt.data); err != nil {
				t.Fatal(err)
			}
			for _, want := range tt.want {
				if !strings.Contains(sb.String(), want) {
					t.Errorf("output missing %q", want)
				}
			}
		})
	}
}
