package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

var router *gin.Engine

func main() {
	//populize Traefik Routers
	traefikApiUrl := "http://traefik.traefik:9000/api"
	traefikApiUrlFromEnv, ok := os.LookupEnv("TRAEFIK_API_URL")
	if ok {
		traefikApiUrl = traefikApiUrlFromEnv
	}
	err := populizeTraefik(traefikApiUrl)
	if err != nil {
		log.Fatalf("Fatal Error: (while loading from Traefik API) %v\n", err)
	}

	router = gin.Default()

	// load templates
	router.LoadHTMLGlob("templates/*")

	// Initialize the routes
	initializeRoutes()

	// Start serving the application
	router.Run()
}

// render based in 'Accept' HTTP header
// if no header is provided HTML will be rendered using a template
func render(c *gin.Context, data gin.H, templateName string) {

	switch c.Request.Header.Get("Accept") {
	case "application/json":
		// Respond with JSON
		c.JSON(http.StatusOK, data["payload"])
	case "application/xml":
		// Respond with XML
		c.XML(http.StatusOK, data["payload"])
	default:
		// Respond with HTML
		c.HTML(http.StatusOK, templateName, data)
	}
}
