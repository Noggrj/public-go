FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install git for fetching dependencies
RUN apk add --no-cache git

COPY go.mod ./
# We don't have go.sum yet, so we don't copy it.
# If we had it: COPY go.sum ./

COPY . .

# Tidy dependencies (will update go.mod and create go.sum in the layer)
RUN go mod tidy

# Build
RUN go build -o main cmd/api/main.go

FROM golang:1.23-alpine AS development

WORKDIR /app

# Install git and Air for live reloading (using version compatible with Go 1.23)
RUN apk add --no-cache git curl \
    && curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.49.0 \
    && go install github.com/cosmtrek/air@v1.49.0

# Install golangci-lint
COPY --from=golangci/golangci-lint:v1.64.5 /usr/bin/golangci-lint /usr/bin/golangci-lint

# Copy source code
COPY . .

# Install dependencies
RUN go mod download

EXPOSE 8080

# Default command for development (Air)
CMD ["air"]

FROM alpine:latest AS production
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations

HEALTHCHECK --interval=30s --timeout=3s \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

EXPOSE 8080
CMD ["./main"]
