package main

import (
	"BibleSearch/controllers"
	"BibleSearch/services"
	"BibleSearch/templates"
	"cmp"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {

	// Setup configs, logging, services, and chroma client
	services.ReadDotEnv()
	configuration := services.NewDefaultConfig()

	// Readable logs locally, JSON in production
	var handler slog.Handler = slog.NewJSONHandler(os.Stdout, nil)
	if configuration.Dev {
		handler = slog.NewTextHandler(os.Stdout, nil)
		slog.Info("Running in dev mode")
	}
	slog.SetDefault(slog.New(handler))

	chromaService := services.NewDefaultChromaService(configuration)
	err := chromaService.CreateCollection(services.CollectionName)
	if err != nil {
		slog.Error("Error getting collection", "err", err)
		os.Exit(1)
	}

	vectorizationService := services.NewDefaultVectorizationService(configuration, chromaService)

	// Dev reads templates from disk and re-parses them on every request, so edits show on refresh
	renderer := templates.NewRenderer(configuration.Dev)

	mux := http.NewServeMux()
	controllers.RegisterAssets(mux)

	// API routes
	controllers.RegisterAPIRoutes(mux, vectorizationService, chromaService)

	// Pages routes
	controllers.RegisterPages(mux, renderer, chromaService)

	server := &http.Server{
		Addr:              ":" + cmp.Or(os.Getenv("PORT"), "8080"),
		Handler:           controllers.Recover(controllers.LogRequests(mux)),
		ReadHeaderTimeout: 10 * time.Second,
	}

	slog.Info("Listening", "addr", server.Addr)
	err = server.ListenAndServe()
	if err != nil {
		slog.Error("Error running server", "err", err)
		os.Exit(1)
	}

}
