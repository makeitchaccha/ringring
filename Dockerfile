# syntax=docker/dockerfile:1.4
# Use for caching Go modules and build cache.
# --- Stage 1: Build the Go application ---
FROM golang:1.23-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod,rw \
    --mount=type=cache,target=/root/.cache/go-build,rw \
    go mod download

COPY . .

ENV CGO_ENABLED=0

RUN go build -ldflags="-s -w" -o build/ringring cmd/ringring/main.go
RUN go build -ldflags="-s -w" -o build/deploy cmd/deploy/main.go

# --- Stage 2: Create a small runtime image with OS-managed timezone data ---
FROM alpine:latest AS runner

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata && rm -rf /var/cache/apk/*

COPY --from=builder /build/build/ringring ./ringring
COPY --from=builder /build/build/deploy ./deploy

COPY --from=builder /build/locales ./locales

# DOCUMENTATION PURPOSE: you can set the timezone when launching the container
# ENV TZ=UTC

# Define the command to run your executable when the container starts.
CMD [ "/app/ringring" ]