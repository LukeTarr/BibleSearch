package main

import (
	"BibleSearch/controllers"
	"BibleSearch/docs"
	"BibleSearch/services"
	"BibleSearch/templates"
	"context"

	ginzerolog "github.com/dn365/gin-zerolog"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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

	// Swagger docs
	docs.SwaggerInfo.BasePath = "/api/v1"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	root := r.Group("/")

	// Web pages
	r.GET("/", func(c *gin.Context) {
		comp := templates.Home()
		c.Writer.Header().Set("Content-Type", "text/html")
		comp.Render(context.Background(), c.Writer)
	})

	// API routes
	controllers.RegisterAPIRoutes(root, vectorizationService, chromaService)

	err = r.Run()
	if err != nil {
		log.Fatal().Err(err).Msg("Error running server")
	}

}
