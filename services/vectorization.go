package services

import (
	"BibleSearch/model"
	"crypto/subtle"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
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
		log.Info().Msg("Resetting Client")
		err := v.ChromaService.ResetClient()
		if err != nil {
			log.Error().Err(err).Msg("Error resetting client")
			return
		}
	}

	log.Info().Msg("Creating Collection")
	err := v.ChromaService.CreateCollection(CollectionName)
	if err != nil {
		log.Error().Err(err).Msg("Error creating collection")
		return
	}

	bookSlice, err := GetBookSlice()
	if err != nil {
		log.Error().Err(err).Msg("Error getting book slice")
		return
	}

	log.Info().Msg("Adding Books to Collection")
	err = v.ChromaService.AddBooksToCollection(bookSlice)
	if err != nil {
		log.Error().Err(err).Msg("Error adding books to collection")
		return
	}

	countDocs, err := v.ChromaService.Count()
	if err != nil {
		log.Error().Err(err).Msg("Error counting documents")
		return
	}

	log.Info().Int("docsCounter", countDocs).Dur("elapsed", time.Since(start)).Msg("Counted documents")
}

// HandleVectorizationRequest starts vectorization in the background, if the password matches and it isn't already running
func (v *VectorizationService) HandleVectorizationRequest(ctx *gin.Context) {
	var vectorizeDTO model.VectorizeDTO
	err := ctx.ShouldBindJSON(&vectorizeDTO)
	if err != nil {
		log.Error().Err(err).Msg("Vectorize hit with invalid body")
		ctx.JSON(http.StatusBadRequest, model.ErrorDTO{
			Error: "string password required",
		})
		return
	}

	// An unset password would otherwise let an empty one through
	expected := v.ConfigService.VectorizationPassword
	if expected == "" || subtle.ConstantTimeCompare([]byte(vectorizeDTO.Password), []byte(expected)) != 1 {
		log.Error().Msg("Vectorize hit with invalid password")
		ctx.JSON(http.StatusUnauthorized, model.ErrorDTO{
			Error: "invalid password",
		})
		return
	}

	if !v.running.CompareAndSwap(false, true) {
		ctx.JSON(http.StatusConflict, model.ErrorDTO{
			Error: "vectorization already running",
		})
		return
	}

	go func() {
		defer v.running.Store(false)
		v.Vectorize(false)
	}()
	ctx.JSON(http.StatusOK, model.StatusDTO{
		Status:  "success",
		Message: "vectorization started",
	})
}
