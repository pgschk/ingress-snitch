#!/usr/bin/env bash
#
# Script to start Minikube and install Traefik and a Test Ingress Deployment
#
############################################################################
set -e
set -x

KC='kubectl --context=minikube'
HC='helm --kube-context=minikube'

minikube start

$HC repo add traefik https://traefik.github.io/charts --force-update
$HC repo update

# Set docker env to minikubes docker
eval $(minikube docker-env)

# Build snitch
docker build -t ingress-snitch:latest .

# Install Traefik
$HC upgrade --install --create-namespace --namespace traefik traefik traefik/traefik -f ./k8s/traefik/traefik-values.yaml

# Apply a sample workload and IngressRoute
$KC apply -f k8s/traefik/whoami.yaml

# Add service-account to read services
$KC apply -f k8s/serviceaccount.yaml

# Add a snitch pod
$KC apply -f k8s/pod.minikube.yaml

# Add IngressRoute, Svc and Middleware for snitch
$KC apply -f k8s/ingressroute.yaml

# Give pods time to start
sleep 20

/usr/bin/env bash -c 'sleep 5; open "http://localhost:8080/snitch"'

# Forward Traefik web to localhost:8080
$KC port-forward -n traefik services/traefik 8080:8080