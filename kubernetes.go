package main

import (
	"context"
	"errors"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var TraefikService v1.Service

// GetTraefikService connects to the local K8s API and retrieves the
// Kubernetes Service LoadBalancer used by Traefik
func GetTraefikService() v1.Service {
	// create config from local ServiceAccount
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	// get services in all the namespaces by omitting namespace
	fmt.Printf("Querying for services in \"%s\" namespace\n", TraefikNamespace)
	services, err := clientset.CoreV1().Services(TraefikNamespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Found %d services\n", len(services.Items))

	for _, service := range services.Items {
		if service.Name == TraefikServiceName {
			TraefikService = service
		}
	}
	if TraefikService.Name == "" {
		err := errors.New("Traefik service not found in namespace " + TraefikNamespace + "\n")
		panic(err)
	}
	return TraefikService
}

// GetTraefikPortByName will return the numerical port associated with
// the Traefik EntryPoint provided as `name`
func GetTraefikPortByName(name string) (port uint, err error) {
	var traefikPort uint
	for _, port := range TraefikService.Spec.Ports {
		fmt.Println(port) // TODO: remove debug output
		if port.TargetPort.StrVal == name {
			traefikPort = uint(port.Port)
		}
	}
	if traefikPort == 0 {
		err := errors.New("could not find traefik port with name " + name)
		return traefikPort, err
	}
	return traefikPort, nil
}
