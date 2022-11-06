package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// struct TraefikRouter represents a Traefik router
// as returned by Traefik API
type TraefikRouter struct {
	EntryPoints []string `json:"EntryPoints"`
	Service     string   `json:"service"`
	Rule        string   `json:"rule"`
	Priority    int64    `json:"priority,omitempty"`
	Status      string   `json:"status"`
	Using       []string `json:"using"`
	Name        string   `json:"name"`
	Provider    string   `json:"provider"`
	TLS         struct {
		Options string `json:"options"`
	} `json:"tls,omitempty"`
	URLs []string
}

// struct TraefikEntryPoint represents a Traefik entrypoint
// as returned by Traefik API
type TraefikEntryPoint struct {
	Address   string `json:"address"`
	Transport struct {
		LifeCycle struct {
			GraceTimeOut string `json:"graceTimeOut"`
		} `json:"lifeCycle"`
		RespondingTimeouts struct {
			IdleTimeout string `json:"idleTimeout"`
		} `json:"respondingTimeouts"`
	} `json:"transport"`
	ForwardedHeaders struct {
	} `json:"forwardedHeaders"`
	HTTP struct {
		TLS *struct {
		} `json:"tls"`
	} `json:"http,omitempty"`
	HTTP2 struct {
		MaxConcurrentStreams int `json:"maxConcurrentStreams"`
	} `json:"http2"`
	UDP struct {
		Timeout string `json:"timeout"`
	} `json:"udp"`
	Name        string `json:"name"`
	Protocol    string
	Port        string
	ServicePort uint
}

// TraefikRouterList holds a list if all TraefikRouter retrieved from Traefik API
var TraefikRouterList = []TraefikRouter{}

// TraefikEntryPointList holds a list if all TraefikEntryPoint retrieved from Traefik API
var TraefikEntryPointList = []TraefikEntryPoint{}
var traefikApiUrl string

// getTraefikRouteByName looks up cached TraefikRouter and
// returns the first one matching by name
func getTraefikRouteByName(name string) (*TraefikRouter, error) {
	for _, a := range TraefikRouterList {
		if a.Name == name {
			return &a, nil
		}
	}
	return nil, errors.New("traefik router not found")
}

// getTraefikEntryPointByName looks up cached TraefikEntryPoint and
// returns the first one matching by name
func getTraefikEntryPointByName(name string) (*TraefikEntryPoint, error) {
	for _, a := range TraefikEntryPointList {
		if a.Name == name {
			return &a, nil
		}
	}
	return nil, errors.New("traefik entrypoint not found")
}

// getAllTraefikRouters connects to a Traefik API and
// retrieves all routers
func getAllTraefikRouters() ([]TraefikRouter, error) {

	apiEndpoint := traefikApiUrl + "/http/routers"
	fmt.Printf("Trying to retrieve Traefik Routers from API at %s\n", apiEndpoint)
	response, err := http.Get(apiEndpoint)

	if err != nil {
		fmt.Print(err.Error())
		return nil, err
	}

	// read the responsedata from Traefik API
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	var responseObject []TraefikRouter
	var routers []TraefikRouter

	// Unmarshal the JSON response into TraefikRouter objects
	json.Unmarshal(responseData, &responseObject)
	responseLength := len(responseObject)
	fmt.Printf("Received %d routers from Traefik API\n", responseLength)

	// parse any TraefikRouter retrieved from the API and
	// append it to the routers object to be returned
	for i := 0; i < responseLength; i++ {
		routers = append(routers, parseTraefikRouterUrls(responseObject[i]))
	}
	return routers, nil
}

// getAllTraefikEntryPoints connects to a Traefik API and
// retrieves all entrypoints
func getAllTraefikEntryPoints() ([]TraefikEntryPoint, error) {

	apiEndpoint := traefikApiUrl + "/entrypoints"
	fmt.Printf("Trying to retrieve Traefik Routers from API at %s\n", apiEndpoint)
	response, err := http.Get(apiEndpoint)

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	// read the responsedata from Traefik API
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	var responseObject []TraefikEntryPoint
	var EntryPoints []TraefikEntryPoint

	// Unmarshal the JSON response into TraefikEntryPoint objects
	json.Unmarshal(responseData, &responseObject)
	responseLength := len(responseObject)
	fmt.Printf("Received %d EntryPoints from Traefik API\n", responseLength)

	// parse any TraefikEntryPoint retrieved from the API and
	// append it to the EntryPoints object to be returned
	for i := 0; i < responseLength; i++ {
		EntryPoints = append(EntryPoints, parseTraefikEntryPoint(responseObject[i]))
	}
	return EntryPoints, nil
}

// parseTraefikRouterUrls parses the rules in the Traefik router to
// interpret the URL they represent
func parseTraefikRouterUrls(router TraefikRouter) TraefikRouter {
	var ruleHostnames []string
	var rulePaths []string
	rule := router.Rule

	var entryPointPort uint
	entryPointProto := "http://"

	for _, routerEntryPointName := range router.EntryPoints {

		// retrieve the EntryPoint associated with this router
		entryPoint, err := getTraefikEntryPointByName(routerEntryPointName)
		if err != nil {
			log.Printf("Router %s uses unknown Entrypoint: %s\n", router.Name, routerEntryPointName)
			continue
		}

		// ignore routers that don't use tcp for now
		// HTTP3 might be a problem here
		if entryPoint.Protocol == "tcp" {
			// set the port to use to this entrypoints port
			// this is often wrong, as a Kubernetes service
			// will do port translation
			entryPointPort = entryPoint.ServicePort

			// if there is TLS configuration associated with
			// this entrypoint, set the entrypoint url prefix to https://
			if entryPoint.HTTP.TLS != nil {
				fmt.Printf("%s is a HTTPS endpoint!\n", entryPoint.Name) // TODO: remove debug output
				entryPointProto = "https://"
			}
		}
	}

	if entryPointPort == 0 {
		log.Printf("Did not find valid EntryPoint port for %s", router.Name)
	}

	// setup regexp to match Host() and Path() rules from Traefik routers
	hostRegexp := regexp.MustCompile(`Host\(\s*\140([^\051]*)`)
	pathRegexp := regexp.MustCompile(`Path(?:Prefix)?\(\140([^\140]*)\140\)`)

	// find matches in current router rules
	hostMatches := hostRegexp.FindAllStringSubmatch(rule, -1)
	pathMatches := pathRegexp.FindAllStringSubmatch(rule, -1)

	// for all host matches, do some cleanup
	for _, hostRules := range hostMatches {
		for i, match := range hostRules {
			if i == 0 || match == "" {
				// the first and any matches
				continue
			}

			// replace all spaces to normalize
			match = strings.ReplaceAll(match, " ", "")
			// replace all backticks to normalize
			match = strings.ReplaceAll(match, "`", "")

			// split match but either comma or two pipes, seperating
			// all hostname used in the rule
			splitRegexp := regexp.MustCompile(`\174{2}|,`)
			hostnames := splitRegexp.Split(match, -1)

			// append all to ruleHostnames matching this group of matches
			ruleHostnames = append(ruleHostnames, hostnames...)
		}
	}

	// add all path matches to the list rulePaths
	for _, pathRules := range pathMatches {
		rulePaths = append(rulePaths, pathRules[1])
	}

	// if no paths found in rule, we add "/" as a path
	if len(rulePaths) == 0 {
		rulePaths = append(rulePaths, "/")
	}

	for _, hostname := range ruleHostnames {
		// for each hostname we build an URL using
		// - the entrypoints protocol (http:// or https://)
		// - the hostname
		// - the entryPointPort
		// - the first matching path
		router.URLs = append(router.URLs, entryPointProto+hostname+":"+strconv.FormatUint(uint64(entryPointPort), 10)+rulePaths[0])
	}

	// if there was no URL generated for this router it is most likely a very simple router
	// that we do not have enough information to work with
	if len(router.URLs) == 0 {
		router.URLs = append(router.URLs, "Unknown")
	}

	fmt.Println(router.URLs) // TODO: remove debug output
	return router
}

// parseTraefikEntryPoint parses the a TraefikEntryPoint
// to extract its associated port and protocol if they are tcp or udp
func parseTraefikEntryPoint(entryPoint TraefikEntryPoint) TraefikEntryPoint {
	portRegexp := regexp.MustCompile(`:(\d+)\/(tcp|udp)`)
	portMatches := portRegexp.FindStringSubmatch(entryPoint.Address)
	if portMatches[2] == "tcp" || portMatches[2] == "udp" {
		var err error
		entryPoint.ServicePort, err = GetTraefikPortByName(entryPoint.Name)
		if err != nil {
			log.Println(err)
		}
		entryPoint.Protocol = portMatches[2]
	}
	fmt.Printf("Server %s on port %d", entryPoint.Name, entryPoint.ServicePort)
	return entryPoint
}

// populizeTraefik initially makes sure that the TraefikEntryPointList
// and TraefikRouterList are populated at startup (called from main)
func populizeTraefik(url string) error {
	var err, err2 error
	traefikApiUrl = url
	TraefikEntryPointList, err2 = getAllTraefikEntryPoints()
	TraefikRouterList, err = getAllTraefikRouters()
	if err != nil || err2 != nil {
		return err
	}
	return nil
}
