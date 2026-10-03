package services

import (
	"BibleSearch/model"
	"crypto/subtle"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"
)

type VectorizationService struct {
	ChromaService *ChromaService
	ConfigService *ConfigService
	running       atomic.Bool
}

func NewDefaultVectorizationService(configService *ConfigService, chromaService *ChromaService) *VectorizationService {
	return &VectorizationService{
		ChromaService: chromaService,
		ConfigService: configService,
	}
}

func (v *VectorizationService) Vectorize(reset bool) {
	start := time.Now()

	if reset {
		slog.Info("Resetting Client")
		err := v.ChromaService.ResetClient()
		if err != nil {
			slog.Error("Error resetting client", "err", err)
			return
		}
	}

	slog.Info("Creating Collection")
	err := v.ChromaService.CreateCollection(CollectionName)
	if err != nil {
		slog.Error("Error creating collection", "err", err)
		return
	}

	bookSlice, err := GetBookSlice()
	if err != nil {
		slog.Error("Error getting book slice", "err", err)
		return
	}

	slog.Info("Adding Books to Collection")
	err = v.ChromaService.AddBooksToCollection(bookSlice)
	if err != nil {
		slog.Error("Error adding books to collection", "err", err)
		return
	}

	countDocs, err := v.ChromaService.Count()
	if err != nil {
		slog.Error("Error counting documents", "err", err)
		return
	}

	slog.Info("Counted documents", "docsCounter", countDocs, "elapsed", time.Since(start))
}

// HandleVectorizationRequest starts vectorization in the background, if the password matches and it isn't already running
func (v *VectorizationService) HandleVectorizationRequest(w http.ResponseWriter, r *http.Request) {
	var vectorizeDTO model.VectorizeDTO
	err := decodeJSON(w, r, &vectorizeDTO)
	if err != nil {
		slog.Error("Vectorize hit with invalid body", "err", err)
		writeJSON(w, http.StatusBadRequest, model.ErrorDTO{
			Error: "string password required",
		})
		return
	}

	// An unset password would otherwise let an empty one through
	expected := v.ConfigService.VectorizationPassword
	if expected == "" || subtle.ConstantTimeCompare([]byte(vectorizeDTO.Password), []byte(expected)) != 1 {
		slog.Warn("Vectorize hit with invalid password")
		writeJSON(w, http.StatusUnauthorized, model.ErrorDTO{
			Error: "invalid password",
		})
		return
	}

	if !v.running.CompareAndSwap(false, true) {
		writeJSON(w, http.StatusConflict, model.ErrorDTO{
			Error: "vectorization already running",
		})
		return
	}

	go func() {
		defer v.running.Store(false)
		v.Vectorize(false)
	}()
	writeJSON(w, http.StatusOK, model.StatusDTO{
		Status:  "success",
		Message: "vectorization started",
	})
}
