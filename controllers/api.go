package controllers

import (
	"BibleSearch/services"
	"net/http"
)

func RegisterAPIRoutes(mux *http.ServeMux, vectorizationService *services.VectorizationService, chromaService *services.ChromaService) {
	mux.HandleFunc("POST /api/v1/vectorize", vectorizationService.HandleVectorizationRequest)
	mux.HandleFunc("POST /api/v1/query", chromaService.HandleQueryRequest)
}
