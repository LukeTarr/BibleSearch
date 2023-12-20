package controllers

import (
	"BibleSearch/services"
	"github.com/gin-gonic/gin"
)

func RegisterAPIRoutes(supergroup *gin.RouterGroup, vectorizationService *services.VectorizationService, chromaService *services.ChromaService) {
	api := supergroup.Group("/api/v1")

	api.POST("/vectorize", vectorizationService.HandleVectorizationRequest)
	api.POST("/query", chromaService.HandleQueryRequest)
}
