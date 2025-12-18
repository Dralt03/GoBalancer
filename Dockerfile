# BUILD
FROM golang:1.24-alpine AS builder
RUN apk add --no-cache 

WORKDIR /app

# Cache Dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy Source
COPY . .

#Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o gobalancer main.go

#RUNTIME
FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /app/gobalancer /app/gobalancer

# Run as non root for K8s
USER nonroot:nonroot

# Default PORT
EXPOSE 8080

ENTRYPOINT ["/app/gobalancer"]