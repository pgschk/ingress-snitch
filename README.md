# ingress-snitch

## Overview
A small app designed to generate clickable links to differen Kubernetes ingresses as an easy starting point when using applications in your cluster.

## Running
Currently only usable in a Kubernetes cluster. Use container image `quay.io/pgschk/ingress-snitch`. Please check out `k8s/pod.yaml` example.
You will need to give the pod a ServiceAccount with elevated permissions (read all Services and IngressRoutes). You can find an example in `k8s/serviceaccount.yaml`.

The image accepts the following envs:

| Name               | Default | Description  |
|---|---|---|
| `TRAEFIK_API_URL`  | http://traefik.traefik:9000/api  | The URL where your Traefik's API is reachable (without authentication for now)  |
| `TRAEFIK_NAMESPACE`  | ""  | The namespace in which your Traefik Service is deployed. If left out it will check all Services in all namespaces   |
| `TRAEFIK_SERVICE_NAME`  | traefik   | The name of your Traefik Service  |

## Testing in Minikube
If you have [minikube](https://minikube.sigs.k8s.io/docs/start/) installed you can use the script `./minikube-dev.sh`, which should set up a test/dev environment for you with Traefik, a sample ingress/service and ingress-snitch. You need to have `kubectl` and `helm` installed.

To make sure this works you can delete your minikube (and all its contents!) with `minikube delete` before. 