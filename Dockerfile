FROM golang:1.18-alpine 
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./
COPY templates ./
RUN go build -o /ingress-snitch
EXPOSE 8080
CMD [ "/ingress-snitch" ]
