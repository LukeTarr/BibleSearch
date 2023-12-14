package services

import (
	"BibleSearch/model"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type VectorizationService struct {
	ChromaService *ChromaService
	ConfigService *ConfigService
}

func NewDefaultVectorizationService(configService *ConfigService, chromaService *ChromaService) *VectorizationService {
	return &VectorizationService{
		ChromaService: chromaService,
		ConfigService: configService,
	}
}

func (v *VectorizationService) Vectorize(reset bool) {
	if reset {
		log.Info().Msg("Resetting Client")
		err := v.ChromaService.ResetClient()
		if err != nil {
			log.Error().Err(err).Msg("Error resetting client")
			return
		}
	}

	log.Info().Msg("Creating Collection")
	_, err := v.ChromaService.CreateCollection("bible")
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

	countDocs, err := v.ChromaService.Collection.Count()
	if err != nil {
		log.Error().Err(err).Msg("Error `querying documents")
		return
	}

	log.Info().Int32("docsCounter", countDocs).Msg("Counted documents")
}

func (v *VectorizationService) HandleVectorizationRequest(ctx *gin.Context) {
	go v.Vectorize(false)
	ctx.JSON(200, model.StatusDTO{
		Status:  "success",
		Message: "vectorization started",
	})
}
