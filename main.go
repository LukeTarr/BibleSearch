package main

import (
	"BibleSearch/controllers"
	"BibleSearch/services"

	ginzerolog "github.com/dn365/gin-zerolog"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func main() {

	// Setup configs, services, and chroma client
	services.ReadDotEnv()
	configuration := services.NewDefaultConfig()

	chromaService := services.NewDefaultChromaService(configuration)
	_, err := chromaService.CreateCollection("bible")
	if err != nil {
		log.Fatal().Err(err).Msg("Error getting collection")
	}

	vectorizationService := services.NewDefaultVectorizationService(configuration, chromaService)

	// Gin router + middleware + static assets
	r := gin.New()
	r.Use(ginzerolog.Logger("gin"))
	r.Use(gin.Recovery())
	r.Static("./assets", "./assets")
	root := r.Group("/")

	// API routes
	controllers.RegisterAPIRoutes(root, vectorizationService, chromaService)

	// Pages routes
	controllers.RegisterPages(root, chromaService)

	err = r.Run()
	if err != nil {
		log.Fatal().Err(err).Msg("Error running server")
	}

}
