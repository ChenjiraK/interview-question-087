package httpapi

import (
	"net/http"

	"back-go/internal/handler"

	"github.com/gin-gonic/gin"
)

// NewRouter configures the API routes and logging/recovery middleware.
func NewRouter(persons handler.PersonService) *gin.Engine {
	router := gin.Default()
	// Only trust the direct connection until a reverse proxy is configured.
	if err := router.SetTrustedProxies(nil); err != nil {
		panic(err)
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	ping := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	}
	router.GET("/ping", ping)
	api := router.Group("/api/v1")
	api.GET("/ping", ping)

	personHandler := handler.NewPersonHandler(persons)
	router.POST("/person", personHandler.Create)
	router.GET("/persons", personHandler.List)
	router.GET("/person/:id", personHandler.GetByID)

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
	})

	return router
}
