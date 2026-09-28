package templates

import (
	"BibleSearch/model"
	"slices"
	"strings"
	"testing"
)

// Rendering every template catches field typos, which html/template only reports at execution time
func TestTemplatesRender(t *testing.T) {
	tmpl := Parse()
	results := model.SearchResultsView{
		Query: "Jesus cried",
		Results: []model.ChromaQueryResultsDTO{{
			Metadata: model.Metadata{
				Book:          "John",
				Chapter:       "11",
				Verse:         "35",
				ReferenceLink: "https://www.bible.com/bible/1/JHN.11",
			},
			Distance: 0.5,
			Text:     "Jesus wept.",
			Id:       "26580",
		}},
	}

	tests := []struct {
		name string
		data any
		want []string
	}{
		{"home", model.SearchResultsView{}, []string{"<html", "</html>", `id="searchResultArea"`, `hx-post="/search"`}},
		{"about", nil, []string{"<html", "</html>", "text-embedding-3-small"}},
		{"results", results, []string{"Jesus wept.", "John 11:35", `href="https://www.bible.com/bible/1/JHN.11"`, "Jesus cried"}},
		{"results", model.SearchResultsView{Error: "Something went wrong"}, []string{`id="searchResultArea"`, "Something went wrong"}},
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

func TestVerse(t *testing.T) {
	tests := []struct {
		in    string
		text  string
		notes []Note
	}{
		{"the tree {was} good for food, & <pleasant>", "the tree <em>was</em> good for food, &amp; &lt;pleasant&gt;", nil},
		{"And Adam called his wife's name Eve{Eve: Heb. Chavah: that is Living}; because she was", "And Adam called his wife&#39;s name Eve; because she was", []Note{{"Eve", "Heb. Chavah: that is Living"}}},
		{"that {it was} good. {the light from...: Heb. between the light}", "that <em>it was</em> good.", []Note{{"the light from...", "Heb. between the light"}}},
	}
	for _, tt := range tests {
		got := verse(tt.in)
		if string(got.Text) != tt.text {
			t.Errorf("text: got %q, want %q", got.Text, tt.text)
		}
		if !slices.Equal(got.Notes, tt.notes) {
			t.Errorf("notes: got %q, want %q", got.Notes, tt.notes)
		}
	}
}
