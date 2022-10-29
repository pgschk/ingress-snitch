############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder
# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache 'git=~2'

# Install dependencies
ENV GO111MODULE=on
WORKDIR $GOPATH/src/packages/goginapp/
COPY . .

# Fetch dependencies.
# Using go get.
RUN go get -d -v

# Build the binary.
#RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/main .
RUN go build -o /app/ingress-snitch

############################
# STEP 2 build a small image
############################
FROM alpine:3

WORKDIR /

# Copy our static executable.
COPY --from=builder /app/ingress-snitch /app/ingress-snitch
COPY templates /app/templates

ENV PORT 8080
EXPOSE 8080
ENV GIN_MODE release

WORKDIR /app

# Run the Go Gin binary.
ENTRYPOINT ["/app/ingress-snitch"]