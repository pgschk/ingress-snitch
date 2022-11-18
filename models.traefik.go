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
	URLs        []string
	HTMLName    string
	ServicePort uint
	ProtoPrefix string
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
func parseTraefikRouterUrls(traefikRouter TraefikRouter) TraefikRouter {
	rule := traefikRouter.Rule

	// Set the EntryPoints ServicePort and ProtoPrefix
	setEntryPointSpecsForRouter(&traefikRouter)

	if traefikRouter.ServicePort == 0 {
		log.Printf("Did not find valid EntryPoint port for %s. Possibly not exposed by Service\n", traefikRouter.Name)
	}

	ruleHostnames := getHostnameMatches(rule)
	rulePaths := getPathMatches(rule)

	for _, hostname := range ruleHostnames {
		portStr := getPortString(&traefikRouter)
		// for each hostname we build an URL using
		// - the entrypoints protocol (http:// or https://)
		// - the hostname
		// - the ServicePort
		// - the first matching path
		traefikRouter.URLs = append(traefikRouter.URLs, traefikRouter.ProtoPrefix+hostname+portStr+rulePaths[0])
	}

	// replace characters in the routers Name that are invalid in HTML ids
	traefikRouter.HTMLName = sanitizeHTMLName(traefikRouter.Name)
	fmt.Println(traefikRouter.HTMLName)

	// if there was no URL generated for this router it is most likely a very simple router
	// that we do not have enough information to work with
	if len(traefikRouter.URLs) == 0 {
		traefikRouter.URLs = append(traefikRouter.URLs, "Unknown")
	}

	fmt.Println(traefikRouter.URLs) // TODO: remove debug output
	return traefikRouter
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
	if entryPoint.ServicePort != 0 {
		fmt.Printf("Traefik Service is serving EntryPoint %s on port %d\n", entryPoint.Name, entryPoint.ServicePort)
	}
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

// sanitizeHTMLName replaces characters that are invalid in HTML ids
func sanitizeHTMLName(name string) string {
	HTMLName := strings.ReplaceAll(name, "@", "_")
	return HTMLName
}

// sanitizeTraefikRule normalizes a matched Traefik Rule
// to make it easier to parse
func sanitizeTraefikRule(rule string) string {
	// replace all spaces to normalize
	rule = strings.ReplaceAll(rule, " ", "")
	// replace all backticks to normalize
	rule = strings.ReplaceAll(rule, "`", "")
	// replace all linebreaks to normalize
	rule = strings.ReplaceAll(rule, "\n", "")
	return rule
}

// setEntryPointSpecsForRouter returns the port and protocol for the EntryPoint
func setEntryPointSpecsForRouter(traefikRouter *TraefikRouter) {
	traefikRouter.ProtoPrefix = "http://"
	for _, routerEntryPointName := range traefikRouter.EntryPoints {

		// retrieve the EntryPoint associated with this router
		entryPoint, err := getTraefikEntryPointByName(routerEntryPointName)
		if err != nil {
			log.Printf("Router %s uses unknown Entrypoint: %s\n", traefikRouter.Name, routerEntryPointName)
			continue
		}

		// ignore routers that don't use tcp for now
		// HTTP3 might be a problem here
		if entryPoint.Protocol == "tcp" {
			// set the port to use to this entrypoints port
			// this is often wrong, as a Kubernetes service
			// will do port translation
			traefikRouter.ServicePort = entryPoint.ServicePort

			// if there is TLS configuration associated with
			// this entrypoint, set the entrypoint url prefix to https://
			if entryPoint.HTTP.TLS != nil {
				fmt.Printf("%s is a HTTPS endpoint!\n", entryPoint.Name) // TODO: remove debug output
				traefikRouter.ProtoPrefix = "https://"
			}
		}
	}
}

// getPortString compiles a string to be used as port in an URL based on protocol and port
func getPortString(traefikRouter *TraefikRouter) string {
	portStr := ":" + strconv.FormatUint(uint64(traefikRouter.ServicePort), 10)
	if (traefikRouter.ProtoPrefix == "https://" && traefikRouter.ServicePort == 443) ||
		(traefikRouter.ProtoPrefix == "http://" && traefikRouter.ServicePort == 80) {
		portStr = ""
	}
	return portStr
}

// getHostnameMatches parses the Rule to extract hostnames
func getHostnameMatches(rule string) []string {
	var ruleHostnames []string
	// setup regexp to match Host() and Path() rules from Traefik routers
	hostRegexp := regexp.MustCompile(`Host\(\s*\140([^\051]*)`)

	// find matches in current router rules
	hostMatches := hostRegexp.FindAllStringSubmatch(rule, -1)

	// for all host matches, do some cleanup
	for _, hostRules := range hostMatches {
		for i, match := range hostRules {
			if i == 0 || match == "" {
				// skip the first and empty matches
				continue
			}

			// normalize and sanitize rule
			match = sanitizeTraefikRule(match)

			// split match but either comma or two pipes, seperating
			// all hostname used in the rule
			splitRegexp := regexp.MustCompile(`\174{2}|,`)
			hostnames := splitRegexp.Split(match, -1)

			// append all to ruleHostnames matching this group of matches
			ruleHostnames = append(ruleHostnames, hostnames...)
		}
	}
	return ruleHostnames
}

// getPathMatches extracts the paths from rules
func getPathMatches(rule string) []string {
	var rulePaths []string
	// setup regexp to match Host() and Path() rules from Traefik routers
	pathRegexp := regexp.MustCompile(`Path(?:Prefix)?\(\140([^\140]*)\140\)`)

	// find matches in current router rules
	pathMatches := pathRegexp.FindAllStringSubmatch(rule, -1)

	// add all path matches to the list rulePaths
	for _, pathRules := range pathMatches {
		rulePaths = append(rulePaths, pathRules[1])
	}

	// if no paths found in rule, we add "/" as a path
	if len(rulePaths) == 0 {
		rulePaths = append(rulePaths, "/")
	}

	return rulePaths
}
