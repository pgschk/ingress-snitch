package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

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
	Name     string `json:"name"`
	Protocol string
	Port     string
}

var TraefikRouterList = []TraefikRouter{}
var TraefikEntryPointList = []TraefikEntryPoint{}
var traefikApiUrl string

func getTraefikRouteByName(name string) (*TraefikRouter, error) {
	for _, a := range TraefikRouterList {
		if a.Name == name {
			return &a, nil
		}
	}
	return nil, errors.New("Traefik Router not found")
}
func getTraefikEntryPointByName(name string) (*TraefikEntryPoint, error) {
	for _, a := range TraefikEntryPointList {
		if a.Name == name {
			return &a, nil
		}
	}
	return nil, errors.New("Traefik EntryPoint not found")
}

func getAllTraefikRouters() ([]TraefikRouter, error) {

	response, err := http.Get(traefikApiUrl + "/http/routers")

	if err != nil {
		fmt.Print(err.Error())
		return nil, err
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	var responseObject []TraefikRouter
	var routers []TraefikRouter
	json.Unmarshal(responseData, &responseObject)
	responseLength := len(responseObject)
	fmt.Printf("Received %d routers from Traefik API\n", responseLength)
	for i := 0; i < responseLength; i++ {
		routers = append(routers, parseTraefikRouterUrls(responseObject[i]))
	}
	return routers, nil
}

func getAllTraefikEntryPoints() ([]TraefikEntryPoint, error) {

	response, err := http.Get(traefikApiUrl + "/entrypoints")

	if err != nil {
		fmt.Print(err.Error())
		return nil, err
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	var responseObject []TraefikEntryPoint
	var EntryPoints []TraefikEntryPoint
	json.Unmarshal(responseData, &responseObject)
	responseLength := len(responseObject)
	fmt.Printf("Received %d EntryPoints from Traefik API\n", responseLength)
	for i := 0; i < responseLength; i++ {
		EntryPoints = append(EntryPoints, parseTraefikEntryPoint(responseObject[i]))
	}
	return EntryPoints, nil
}

func parseTraefikRouterUrls(router TraefikRouter) TraefikRouter {
	var ruleHostnames []string
	var rulePaths []string
	rule := router.Rule

	var entryPointPort string
	entryPointProto := "http://"
	for _, routerEntryPointName := range router.EntryPoints {
		entryPoint, err := getTraefikEntryPointByName(routerEntryPointName)
		if err != nil {
			log.Printf("Router %s uses unknown Entrypoint: %s\n", router.Name, routerEntryPointName)
			continue
		}
		if entryPoint.Protocol == "tcp" {
			entryPointPort = entryPoint.Port
			if entryPoint.HTTP.TLS != nil {
				fmt.Printf("%s is a HTTPS endpoint!\n", entryPoint.Name)
				entryPointProto = "https://"
			}
		}
	}
	if entryPointPort == "" {
		log.Printf("Did not find valid EntryPoint port for %s", router.Name)
	}
	// hostRegexp := regexp.MustCompile(`Host\(\140([^\140]*)\140(?:\s*\174{2}\s*\140*([^\140]*)\140)*`)
	hostRegexp := regexp.MustCompile(`Host\(\s*\140([^\051]*)`)
	hostMatches := hostRegexp.FindAllStringSubmatch(rule, -1)
	pathRegexp := regexp.MustCompile(`Path(?:Prefix)?\(\140([^\140]*)\140\)`)
	pathMatches := pathRegexp.FindAllStringSubmatch(rule, -1)
	for _, v := range hostMatches {
		for i, b := range v {
			if i == 0 || b == "" {
				continue
			}
			b = strings.ReplaceAll(b, " ", "")
			b = strings.ReplaceAll(b, "`", "")
			splitRegexp := regexp.MustCompile(`\174{2}|,`)
			hostnames := splitRegexp.Split(b, -1)
			for _, hostname := range hostnames {
				ruleHostnames = append(ruleHostnames, hostname)
			}
		}
	}
	for _, v := range pathMatches {
		rulePaths = append(rulePaths, v[1])
	}
	// if no paths found in rule, we add "/" as a path
	if len(rulePaths) == 0 {
		rulePaths = append(rulePaths, "/")
	}
	for _, v := range ruleHostnames {
		router.URLs = append(router.URLs, entryPointProto+v+":"+entryPointPort+rulePaths[0])
	}
	if len(router.URLs) == 0 {
		router.URLs = append(router.URLs, "Unknown")
	}
	fmt.Println(router.URLs)
	return router
}

func parseTraefikEntryPoint(entryPoint TraefikEntryPoint) TraefikEntryPoint {
	portRegexp := regexp.MustCompile(`:(\d+)\/(tcp|udp)`)
	portMatches := portRegexp.FindStringSubmatch(entryPoint.Address)
	if portMatches[2] == "tcp" || portMatches[2] == "udp" {
		entryPoint.Port = portMatches[1]
		entryPoint.Protocol = portMatches[2]
	}

	return entryPoint
}

func populizeTraefik(url string) error {
	err, err2 := errors.New(""), errors.New("")
	traefikApiUrl = url
	TraefikEntryPointList, err2 = getAllTraefikEntryPoints()
	TraefikRouterList, err = getAllTraefikRouters()
	if err != nil || err2 != nil {
		return err
	}
	return nil
}
