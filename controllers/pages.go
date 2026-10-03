package controllers

import (
	"BibleSearch/docs"
	"BibleSearch/model"
	"BibleSearch/services"
	"BibleSearch/templates"
	"net/http"
	"strings"
)

func RegisterPages(mux *http.ServeMux, renderer *templates.Renderer, chromaService *services.ChromaService) {

	// Swagger UI + hand-written OpenAPI spec. The mux redirects /swagger to /swagger/.
	mux.Handle("GET /swagger/", http.StripPrefix("/swagger", http.FileServerFS(docs.FS)))

	// {$} matches only "/" itself, rather than every path
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		renderer.Render(w, http.StatusOK, "home", model.SearchResultsView{})
	})

	mux.HandleFunc("GET /about", func(w http.ResponseWriter, r *http.Request) {
		renderer.Render(w, http.StatusOK, "about", nil)
	})

	mux.HandleFunc("POST /search", chromaService.HandleHTMXQuery(renderer))
}

// RegisterAssets serves ./assets, without directory listings
func RegisterAssets(mux *http.ServeMux) {
	files := http.StripPrefix("/assets", http.FileServer(http.Dir("assets")))
	mux.HandleFunc("GET /assets/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		files.ServeHTTP(w, r)
	})
}
