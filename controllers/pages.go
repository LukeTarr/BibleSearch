package controllers

import (
	"BibleSearch/docs"
	"BibleSearch/model"
	"BibleSearch/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterPages(supergroup *gin.RouterGroup, chromaService *services.ChromaService) {

	// Swagger UI + hand-written OpenAPI spec
	supergroup.StaticFS("/swagger", http.FS(docs.FS))

	supergroup.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home", model.SearchResultsView{})
	})

	supergroup.GET("/about", func(c *gin.Context) {
		c.HTML(http.StatusOK, "about", nil)
	})

	supergroup.POST("/search", chromaService.HandleHTMXQuery)
}
