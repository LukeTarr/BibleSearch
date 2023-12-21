package controllers

import (
	"BibleSearch/docs"
	"BibleSearch/services"
	"BibleSearch/templates"
	"context"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterPages(supergroup *gin.RouterGroup, chromaService *services.ChromaService) {

	// Swagger docs
	docs.SwaggerInfo.BasePath = "/api/v1"
	supergroup.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	supergroup.GET("/", func(c *gin.Context) {
		comp := templates.Home()
		c.Writer.Header().Set("Content-Type", "text/html")
		comp.Render(context.Background(), c.Writer)
	})

	supergroup.GET("/about", func(c *gin.Context) {
		comp := templates.About()
		c.Writer.Header().Set("Content-Type", "text/html")
		comp.Render(context.Background(), c.Writer)
	})

	supergroup.POST("/search", chromaService.HandleHTMXQuery)
}
