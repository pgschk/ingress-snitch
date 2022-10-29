// handlers.article.go

package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func showIndexPage(c *gin.Context) {

	// Call the render function with the name of the template to render
	render(c, gin.H{
		"title":   "Ingress Snitch",
		"payload": TraefikRouterList}, "index.html")
}

func getTraefikRouter(c *gin.Context) {
	// Check if the router exists
	routerName := c.Param("router_name")
	if router, err := getTraefikRouteByName(routerName); err == nil {
		// Call the HTML method of the Context to render a template
		c.HTML(
			// Set the HTTP status to 200 (OK)
			http.StatusOK,
			// Use the index.html template
			"traefikRouter.html",
			// Pass the data that the page uses
			gin.H{
				"title":   router.Name,
				"payload": router,
			},
		)

	} else {
		// If the router is not found, abort with an error
		c.AbortWithError(http.StatusNotFound, err)
	}
}
