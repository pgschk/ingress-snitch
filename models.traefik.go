package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

type TraefikRouter struct {
	EntryPoints []string `json:"entryPoints"`
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

var TraefikRouterList = []TraefikRouter{}
var traefikApiUrl string

func getTraefikRouteByName(name string) (*TraefikRouter, error) {
	for _, a := range TraefikRouterList {
		if a.Name == name {
			return &a, nil
		}
	}
	return nil, errors.New("Traefik Router not found")
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

func parseTraefikRouterUrls(router TraefikRouter) TraefikRouter {
	var ruleHostnames []string
	var rulePaths []string
	rule := router.Rule
	hostRegexp := regexp.MustCompile(`Host\(\140([^\140]*)\140\)`)
	hostMatches := hostRegexp.FindAllStringSubmatch(rule, -1)
	pathRegexp := regexp.MustCompile(`Path(?:Prefix)?\(\140([^\140]*)\140\)`)
	pathMatches := pathRegexp.FindAllStringSubmatch(rule, -1)
	for _, v := range hostMatches {
		ruleHostnames = append(ruleHostnames, v[1])
	}
	for _, v := range pathMatches {
		rulePaths = append(rulePaths, v[1])
	}
	for _, v := range ruleHostnames {
		router.URLs = append(router.URLs, "http://"+v+rulePaths[0])
	}
	if len(router.URLs) == 0 {
		router.URLs = append(router.URLs, "Unknown")
	}
	fmt.Println(router.URLs)
	return router
}

func populizeTraefikRouters(url string) error {
	err := errors.New("")
	traefikApiUrl = url
	TraefikRouterList, err = getAllTraefikRouters()
	if err != nil {
		return err
	}
	return nil
}
