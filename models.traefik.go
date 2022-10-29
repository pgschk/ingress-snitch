package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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
}

var TraefikRouterList = []TraefikRouter{}
var traefikApiUrl string

// type TraefikRouterResponse2 struct {
// 	Router []TraefikRouter `json:"array"`
// }

func getTraefikRouteByName(name string) (*TraefikRouter, error) {
	for _, a := range TraefikRouterList {
		if a.Name == name {
			return &a, nil
		}
	}
	return nil, errors.New("Traefik Router not found")
}

// func getAllPods() {
// 	fmt.Printf("Test")
// 	podName := "pasty-spa-557d94bd4-d6v6b"
// 	config, err := rest.InClusterConfig()
// 	if err != nil {
// 		panic(err.Error())
// 	}
// 	// creates the clientset
// 	clientset, err := kubernetes.NewForConfig(config)
// 	if err != nil {
// 		panic(err.Error())
// 	}
// 	// get pods in all the namespaces by omitting namespace
// 	// Or specify namespace to get pods in particular namespace
// 	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
// 	if err != nil {
// 		panic(err.Error())
// 	}
// 	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

// 	// Examples for error handling:
// 	// - Use helper functions e.g. errors.IsNotFound()
// 	// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
// 	_, err = clientset.CoreV1().Pods("pasty-staging").Get(context.TODO(), podName, metav1.GetOptions{})
// 	if k8serrors.IsNotFound(err) {
// 		fmt.Printf("Pod %s not found in pasty-staging namespace\n", podName)
// 	} else if statusError, isStatus := err.(*k8serrors.StatusError); isStatus {
// 		fmt.Printf("Error getting pod %v\n", statusError.ErrStatus.Message)
// 	} else if err != nil {
// 		panic(err.Error())
// 	} else {
// 		fmt.Printf("Found %s pod in pasty-staging namespace\n", podName)
// 	}
// }

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
	json.Unmarshal(responseData, &responseObject)
	responseLength := len(responseObject)
	fmt.Printf("Received %d routers from Traefik API\n", responseLength)
	return responseObject, nil
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
