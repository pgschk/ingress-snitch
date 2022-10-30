# ingress-snitch

## Overview
A small app designed to generate clickable links to differen Kubernetes ingresses as an easy starting point when using applications in your cluster.

## Running
Use container image `quay.io/pgschk/ingress-snitch` and set ENV variable `TRAEFIK_API_URL` to the URL where your Traefik's API is reachable (without authentication for now)
