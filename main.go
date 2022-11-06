package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

var router *gin.Engine
var TraefikApiUrl string = "http://traefik.traefik:9000/api"
var TraefikNamespace string = ""
var TraefikServiceName string = "traefik"

func init() {
	// get various envs
	traefikApiUrlFromEnv, apiOk := os.LookupEnv("TRAEFIK_API_URL")
	if apiOk {
		TraefikApiUrl = traefikApiUrlFromEnv
	}

	traefikServiceNameFromEnv, nameOk := os.LookupEnv("TRAEFIK_SERVICE_NAME")
	if nameOk {
		TraefikServiceName = traefikServiceNameFromEnv
	}

	traefikNamespaceFromEnv, nsOk := os.LookupEnv("TRAEFIK_NAMESPACE")
	if nsOk {
		TraefikNamespace = traefikNamespaceFromEnv
	}
	fmt.Printf("Traefik Namespace: \"%s\"\n", TraefikNamespace)
}

func main() {

	GetTraefikService()

	//populize Traefik Routers
	err := populizeTraefik(TraefikApiUrl)
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
