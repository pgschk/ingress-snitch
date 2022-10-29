// routes.go

package main

func initializeRoutes() {

	// Handle the index route
	router.GET("/", showIndexPage)

	// Handle GET requests at /traefik/router/view/some_router_name
	router.GET("/traefik/router/view/:router_name", getTraefikRouter)

}
